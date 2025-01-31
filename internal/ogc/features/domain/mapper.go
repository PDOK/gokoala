package domain

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/PDOK/gokoala/config"
	"github.com/jmoiron/sqlx"

	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/geojson"
)

// MapRelation abstract function type to map feature relations
type MapRelation func(columnName string, columnValue any, externalFidColumn string) (newColumnName string, newColumnValue any)

// MapGeom abstract function type to map geometry from bytes to Geometry
type MapGeom func([]byte) (geom.T, error)

// MapRowsToFeatureIDs datasource agnostic mapper from SQL rows set feature IDs, including prev/next feature ID
func MapRowsToFeatureIDs(ctx context.Context, rows *sqlx.Rows) (featureIDs []int64, prevNextID *PrevNextFID, err error) {
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
			prevNextID = &PrevNextFID{Prev: prev, Next: next}
			firstRow = false
		}
	}
	if ctx.Err() != nil {
		err = ctx.Err()
	}
	return
}

// MapRowsToFeatures datasource agnostic mapper from SQL rows/result set to Features domain model
func MapRowsToFeatures(ctx context.Context, rows *sqlx.Rows, fidColumn string, externalFidColumn string,
	geomColumn string, propConfig *config.FeatureProperties, mapGeom MapGeom, mapRel MapRelation) ([]*Feature, *PrevNextFID, error) {

	result := make([]*Feature, 0)
	columns, err := rows.Columns()
	if err != nil {
		return result, nil, err
	}

	propertiesOrder := propConfig != nil && propConfig.PropertiesInSpecificOrder
	firstRow := true
	var prevNextID *PrevNextFID
	for rows.Next() {
		var values []any
		if values, err = rows.SliceScan(); err != nil {
			return result, nil, err
		}
		feature := &Feature{Properties: NewFeatureProperties(propertiesOrder)}
		np, err := mapColumnsToFeature(ctx, firstRow, feature, columns, values, fidColumn, externalFidColumn,
			geomColumn, mapGeom, mapRel)
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
func mapColumnsToFeature(ctx context.Context, firstRow bool, feature *Feature, columns []string, values []any,
	fidColumn string, externalFidColumn string, geomColumn string, mapGeom MapGeom, mapRel MapRelation) (*PrevNextFID, error) {

	prevNextID := PrevNextFID{}
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
			rawGeom, ok := columnValue.([]byte)
			if !ok {
				return nil, fmt.Errorf("failed to read geometry from %s column in datasource", geomColumn)
			}
			mappedGeom, err := mapGeom(rawGeom)
			if err != nil {
				return nil, fmt.Errorf("failed to map/decode geometry from datasource, error: %w", err)
			}
			if mappedGeom != nil {
				feature.Geometry, err = geojson.Encode(mappedGeom)
				if err != nil {
					return nil, fmt.Errorf("failed to map/encode geometry to JSON, error: %w", err)
				}
			}

		case "minx", "miny", "maxx", "maxy", "min_zoom", "max_zoom":
			// Skip these columns used for bounding box and zoom filtering
			continue

		case PrevFid:
			// Only the first row in the result set contains the previous feature id
			if firstRow && columnValue != nil {
				prevNextID.Prev = columnValue.(int64)
			}

		case NextFid:
			// Only the first row in the result set contains the next feature id
			if firstRow && columnValue != nil {
				prevNextID.Next = columnValue.(int64)
			}

		default:
			if columnValue == nil {
				feature.Properties.Set(columnName, nil)
				continue
			}
			// Grab any non-nil, non-id, non-bounding box, & non-geometry column as a tag
			switch v := columnValue.(type) {
			case []uint8:
				asBytes := make([]byte, len(v))
				copy(asBytes, v)
				feature.Properties.Set(columnName, string(asBytes))
			case int64:
				feature.Properties.Set(columnName, v)
			case float64:
				feature.Properties.Set(columnName, v)
			case time.Time:
				feature.Properties.Set(columnName, v)
			case string:
				feature.Properties.Set(columnName, v)
			case bool:
				feature.Properties.Set(columnName, v)
			default:
				return nil, fmt.Errorf("unexpected type for sqlite column data: %v: %T", columns[i], v)
			}
		}
	}

	mapExternalFid(columns, values, externalFidColumn, feature, mapRel)
	return &prevNextID, ctx.Err()
}

// mapExternalFid run a second pass over columns to map external feature ID, including relations to other features
func mapExternalFid(columns []string, values []any, externalFidColumn string, feature *Feature, mapRel MapRelation) {
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
		case strings.Contains(columnName, externalFidColumn):
			// When externalFidColumn is part of the column name (e.g. 'foobar_external_fid') we treat
			// it as a relation to another feature.
			newColumnName, newColumnValue := mapRel(columnName, columnValue, externalFidColumn)
			if newColumnName != "" {
				columnNameWithoutExternalFID := strings.ReplaceAll(columnName, externalFidColumn, "")
				feature.Properties.SetRelation(newColumnName, newColumnValue, columnNameWithoutExternalFID)
				feature.Properties.Delete(columnName)
			}
		}
	}
}
