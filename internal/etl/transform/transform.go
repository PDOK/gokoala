package transform

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"text/template"

	"github.com/PDOK/gomagpie/config"
	"github.com/go-spatial/geom"
	pggeom "github.com/twpayne/go-geom" // this lib has a large overlap with github.com/go-spatial/geom but we need it to integrate with postgres
)

type RawRecord struct {
	FeatureID    int64
	FieldValues  []any
	Bbox         *geom.Extent
	GeometryType string
}

type SearchIndexRecord struct {
	FeatureID         string
	CollectionID      string
	CollectionVersion int
	DisplayName       string
	Suggest           string
	GeometryType      string
	Bbox              *pggeom.Polygon
}

type Transformer struct{}

func (t Transformer) Transform(records []RawRecord, collection config.GeoSpatialCollection) ([]SearchIndexRecord, error) {
	result := make([]SearchIndexRecord, 0, len(records))
	for _, r := range records {
		fieldValuesByName, err := slicesToMap(collection.Search.Fields, r.FieldValues)
		if err != nil {
			return nil, err
		}
		displayName, err := t.renderTemplate(collection.Search.DisplayName, fieldValuesByName)
		if err != nil {
			return nil, err
		}
		suggestions := make([]string, 0, len(collection.Search.Suggest))
		for _, suggestTemplate := range collection.Search.Suggest {
			suggestion, err := t.renderTemplate(suggestTemplate, fieldValuesByName)
			if err != nil {
				return nil, err
			}
			suggestions = append(suggestions, suggestion)
		}
		bbox, err := r.transformBbox()
		if err != nil {
			return nil, err
		}
		for _, suggestion := range suggestions {
			resultRecord := SearchIndexRecord{
				FeatureID:         strconv.FormatInt(r.FeatureID, 10),
				CollectionID:      collection.ID,
				CollectionVersion: collection.Search.Version,
				DisplayName:       displayName,
				Suggest:           suggestion,
				GeometryType:      r.GeometryType,
				Bbox:              bbox,
			}
			result = append(result, resultRecord)
		}
	}
	return result, nil
}

func (t Transformer) renderTemplate(templateFromConfig string, fieldValuesByName map[string]any) (string, error) {
	parsedTemplate, err := template.New("").Parse(templateFromConfig)
	if err != nil {
		return "", err
	}
	var b bytes.Buffer
	if err = parsedTemplate.Execute(&b, fieldValuesByName); err != nil {
		return "", err
	}
	return b.String(), err
}

func (r RawRecord) transformBbox() (*pggeom.Polygon, error) {
	if strings.EqualFold(r.GeometryType, "POINT") {
		r.Bbox = r.Bbox.ExpandBy(0.1) // create valid bbox in case original geom is a point by expanding it a bit
	}
	if r.Bbox.Area() <= 0 {
		return nil, errors.New("bbox area must be greater than zero")
	}
	// convert bbox to polygon type supported by Postgres db driver
	polygon, err := pggeom.NewPolygon(pggeom.XY).SetCoords([][]pggeom.Coord{{
		{r.Bbox.MinX(), r.Bbox.MinY()},
		{r.Bbox.MaxX(), r.Bbox.MinY()},
		{r.Bbox.MaxX(), r.Bbox.MaxY()},
		{r.Bbox.MinX(), r.Bbox.MaxY()},
		{r.Bbox.MinX(), r.Bbox.MinY()},
	}})
	if err != nil {
		return nil, err
	}
	polygon = polygon.SetSRID(4326)
	if polygon.Area() <= 0 {
		return nil, errors.New("polygon area must be greater than zero")
	}
	return polygon, nil
}

func toStringSlice[T any](slice []T) ([]string, error) {
	result := make([]string, 0, len(slice))
	for _, v := range slice {
		str, ok := any(v).(string)
		if !ok {
			return nil, fmt.Errorf("non-string element found: %v", v)
		}
		result = append(result, str)
	}
	return result, nil
}

func slicesToMap(keys []string, values []any) (map[string]any, error) {
	if len(keys) != len(values) {
		return nil, fmt.Errorf("slices must be of the same length, got %d keys and %d values", len(keys), len(values))
	}
	result := make(map[string]any, len(keys))
	for i := range keys {
		result[keys[i]] = values[i]
	}
	return result, nil
}
