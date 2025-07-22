package common

import (
	"context"
	"fmt"
	"time"

	"github.com/PDOK/gokoala/config"
	"github.com/PDOK/gokoala/internal/engine/types"
	"github.com/PDOK/gokoala/internal/ogc/features/domain"
	"github.com/twpayne/go-geom"
)

// MapRelation abstract function type to map feature relations
type MapRelation func(columnName string, columnValue any, externalFidColumn string) (newColumnName, newColumnNameWithoutProfile string, newColumnValue any)

// MapGeom abstract function type to map datasource-specific geometry
// (in GeoPackage, PostGIS, WKB, etc. format) to general-purpose geometry
type MapGeom func(columnValue any) (geom.T, error)

// MapRowsToFeatureIDs datasource agnostic mapper from SQL rows set feature IDs, including prev/next feature ID
func MapRowsToFeatureIDs(ctx context.Context, rows DatasourceRows) (featureIDs []int64, prevNextID *domain.PrevNextFID, err error) {
	firstRow := true
	for rows.Next() {
		var values []any
		if values, err = rows.SliceScan(); err != nil {
			return nil, nil, err
		}
		if len(values) != 3 {
			return nil, nil, fmt.Errorf("expected 3 columns containing the feature id, "+
				"the previous feature id and the next feature id. Got: %v", values)
		}
		featureID := values[0].(int64)
		featureIDs = append(featureIDs, featureID)
		if firstRow {
			prev := int64(0)
			if values[1] != nil {
				prev = values[1].(int64)
			}
			next := int64(0)
			if values[2] != nil {
				next = values[2].(int64)
			}
			prevNextID = &domain.PrevNextFID{Prev: prev, Next: next}
			firstRow = false
		}
	}
	if ctx.Err() != nil {
		err = ctx.Err()
	}
	return
}

type FormatOpts struct {
	MaxDecimals int
	ForceUTC    bool
}

// MapRowsToFeatures datasource agnostic mapper from SQL rows/result set to Features domain model
func MapRowsToFeatures(ctx context.Context, rows DatasourceRows,
	fidColumn string, externalFidColumn string, geomColumn string,
	propConfig *config.FeatureProperties, schema *domain.Schema,
	mapGeom MapGeom, mapRel MapRelation, formatOpts FormatOpts) ([]*domain.Feature, *domain.PrevNextFID, error) {

	result := make([]*domain.Feature, 0)
	columns, err := rows.Columns()
	if err != nil {
		return result, nil, err
	}

	propertiesOrder := propConfig != nil && propConfig.PropertiesInSpecificOrder
	firstRow := true
	var prevNextID *domain.PrevNextFID
	for rows.Next() {
		var values []any
		if values, err = rows.SliceScan(); err != nil {
			return result, nil, err
		}
		feature := &domain.Feature{Properties: domain.NewFeatureProperties(propertiesOrder)}
		np, err := mapColumnsToFeature(ctx, firstRow, feature, columns, values, fidColumn,
			externalFidColumn, geomColumn, schema, mapGeom, mapRel, formatOpts)
		if err != nil {
			return result, nil, err
		} else if firstRow {
			prevNextID = np
			firstRow = false
		}
		result = append(result, feature)
	}
	return result, prevNextID, ctx.Err()
}

//nolint:cyclop,funlen
func mapColumnsToFeature(ctx context.Context, firstRow bool, feature *domain.Feature, columns []string, values []any,
	fidColumn string, externalFidColumn string, geomColumn string, schema *domain.Schema, mapGeom MapGeom, mapRel MapRelation,
	formatOpts FormatOpts) (*domain.PrevNextFID, error) {

	prevNextID := domain.PrevNextFID{}
	for i, columnName := range columns {
		columnValue := values[i]

		switch columnName {
		case fidColumn:
			feature.ID = fmt.Sprint(columnValue)

		case geomColumn:
			if columnValue == nil {
				feature.Properties.Set(columnName, nil)
				continue
			}
			mappedGeom, err := mapGeom(columnValue)
			if err != nil {
				return nil, fmt.Errorf("failed to map/decode geometry from datasource, error: %w", err)
			}
			if err = feature.SetGeom(mappedGeom, formatOpts.MaxDecimals); err != nil {
				return nil, fmt.Errorf("failed to map/encode geometry to JSON, error: %w", err)
			}

		case domain.MinxField, domain.MinyField, domain.MaxxField, domain.MaxyField:
			// Skip these columns used for bounding box handling
			continue

		case domain.PrevFid:
			// Only the first row in the result set contains the previous feature id
			if firstRow && columnValue != nil {
				val, err := types.ToInt64(columnValue)
				if err != nil {
					return nil, err
				}
				prevNextID.Prev = val
			}

		case domain.NextFid:
			// Only the first row in the result set contains the next feature id
			if firstRow && columnValue != nil {
				val, err := types.ToInt64(columnValue)
				if err != nil {
					return nil, err
				}
				prevNextID.Next = val
			}

		default:
			if columnValue == nil {
				feature.Properties.Set(columnName, nil)
				continue
			}
			// Grab any non-nil, non-id, non-bounding box & non-geometry column as a feature property
			switch v := columnValue.(type) {
			case []byte:
				asBytes := make([]byte, len(v))
				copy(asBytes, v)
				feature.Properties.Set(columnName, string(asBytes))
			case int32, int64:
				feature.Properties.Set(columnName, v)
			case float64:
				// Check to determine whether the content of the columnValue is truly a floating point value.
				// (Because of non-strict tables in SQLite)
				if !types.IsFloat(v) {
					feature.Properties.Set(columnName, int64(v))
				} else {
					feature.Properties.Set(columnName, v)
				}
			case time.Time:
				timeVal := v
				if formatOpts.ForceUTC {
					timeVal = timeVal.UTC()
				}
				// Map as date (= without time) only when defined as such in the schema AND when no time component is present
				if types.IsDate(timeVal) && schema.IsDate(columnName) {
					feature.Properties.Set(columnName, types.NewDate(timeVal))
				} else {
					feature.Properties.Set(columnName, timeVal)
				}
			case string:
				feature.Properties.Set(columnName, v)
			case bool:
				feature.Properties.Set(columnName, v)
			default:
				return nil, fmt.Errorf("unexpected type: %v: %T", columns[i], v)
			}
		}
	}

	mapExternalFid(columns, values, externalFidColumn, feature, mapRel)
	return &prevNextID, ctx.Err()
}

// mapExternalFid run a second pass over columns to map external feature ID, including relations to other features
func mapExternalFid(columns []string, values []any, externalFidColumn string, feature *domain.Feature, mapRel MapRelation) {
	for i, columnName := range columns {
		columnValue := values[i]

		switch {
		case externalFidColumn == "":
			continue
		case columnName == externalFidColumn:
			// When externalFidColumn is configured, overwrite feature ID and drop externalFidColumn.
			// Note: This happens in a second pass over the feature, since we want to overwrite the
			// feature ID irrespective of the order of columns in the table
			feature.ID = fmt.Sprint(columnValue)
			feature.Properties.Delete(columnName)
		case domain.IsFeatureRelation(columnName, externalFidColumn):
			// When externalFidColumn is part of the column name (e.g. 'foobar_external_fid') we treat
			// it as a relation to another feature.
			newColumnName, newColumnNameWithoutProfile, newColumnValue := mapRel(columnName, columnValue, externalFidColumn)
			if newColumnName != "" {
				feature.Properties.SetRelation(newColumnName, newColumnValue, newColumnNameWithoutProfile)
				feature.Properties.Delete(columnName)
			}
		}
	}
}
