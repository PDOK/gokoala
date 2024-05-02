package domain

import (
	"fmt"
	"time"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/encoding/geojson"
	"github.com/jmoiron/sqlx"
)

// MapRowsToFeatureIDs datasource agnostic mapper from SQL rows set feature IDs, including prev/next feature ID
func MapRowsToFeatureIDs(rows *sqlx.Rows) (featureIDs []int64, prevNextID *PrevNextFID, err error) {
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
	return
}

// MapRowsToFeatures datasource agnostic mapper from SQL rows/result set to Features domain model
func MapRowsToFeatures(rows *sqlx.Rows, fidColumn string, geomColumn string,
	geomMapper func([]byte) (geom.Geometry, error)) ([]*Feature, *PrevNextFID, error) {

	result := make([]*Feature, 0)
	columns, err := rows.Columns()
	if err != nil {
		return result, nil, err
	}

	firstRow := true
	var prevNextID *PrevNextFID
	for rows.Next() {
		var values []any
		if values, err = rows.SliceScan(); err != nil {
			return result, nil, err
		}

		feature := &Feature{Feature: geojson.Feature{Properties: make(map[string]any)}}
		np, err := mapColumnsToFeature(firstRow, feature, columns, values, fidColumn, geomColumn, geomMapper)
		if err != nil {
			return result, nil, err
		} else if firstRow {
			prevNextID = np
			firstRow = false
		}
		result = append(result, feature)
	}
	return result, prevNextID, nil
}

//nolint:cyclop,funlen
func mapColumnsToFeature(firstRow bool, feature *Feature, columns []string, values []any,
	fidColumn string, geomColumn string, geomMapper func([]byte) (geom.Geometry, error)) (*PrevNextFID, error) {

	prevNextID := PrevNextFID{}
	for i, columnName := range columns {
		columnValue := values[i]

		switch columnName {
		case fidColumn:
			feature.ID = columnValue.(int64)

		case geomColumn:
			if columnValue == nil {
				feature.Properties[columnName] = nil
				continue
			}
			rawGeom, ok := columnValue.([]byte)
			if !ok {
				return nil, fmt.Errorf("failed to read geometry from %s column in datasource", geomColumn)
			}
			mappedGeom, err := geomMapper(rawGeom)
			if err != nil {
				return nil, fmt.Errorf("failed to map/decode geometry from datasource, error: %w", err)
			}
			feature.Geometry = geojson.Geometry{Geometry: mappedGeom}

		case "minx", "miny", "maxx", "maxy", "min_zoom", "max_zoom":
			// Skip these columns used for bounding box and zoom filtering
			continue

		case "prevfid":
			// Only the first row in the result set contains the previous feature id
			if firstRow && columnValue != nil {
				prevNextID.Prev = columnValue.(int64)
			}

		case "nextfid":
			// Only the first row in the result set contains the next feature id
			if firstRow && columnValue != nil {
				prevNextID.Next = columnValue.(int64)
			}

		default:
			if columnValue == nil {
				feature.Properties[columnName] = nil
				continue
			}
			// Grab any non-nil, non-id, non-bounding box, & non-geometry column as a tag
			switch v := columnValue.(type) {
			case []uint8:
				asBytes := make([]byte, len(v))
				copy(asBytes, v)
				feature.Properties[columnName] = string(asBytes)
			case int64:
				feature.Properties[columnName] = v
			case float64:
				feature.Properties[columnName] = v
			case time.Time:
				feature.Properties[columnName] = v
			case string:
				feature.Properties[columnName] = v
			case bool:
				feature.Properties[columnName] = v
			default:
				return nil, fmt.Errorf("unexpected type for sqlite column data: %v: %T", columns[i], v)
			}
		}
	}
	return &prevNextID, nil
}
