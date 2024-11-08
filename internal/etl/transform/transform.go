package transform

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

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

type Transformer struct{}

func (t Transformer) Transform(records []RawRecord, collection config.GeoSpatialCollection) ([]SearchIndexRecord, error) {
	result := make([]SearchIndexRecord, 0, len(records))
	for _, record := range records {
		tr, err := record.transform(collection)
		if err != nil {
			return nil, err
		}
		result = append(result, tr)
	}
	return result, nil
}

func (r RawRecord) transform(collection config.GeoSpatialCollection) (SearchIndexRecord, error) {
	fid := strconv.FormatInt(r.FeatureID, 10)

	values, err := toStringSlice(r.FieldValues)
	if err != nil {
		return SearchIndexRecord{}, err
	}
	bbox, err := r.transformBbox()
	if err != nil {
		return SearchIndexRecord{}, err
	}
	return SearchIndexRecord{
		FeatureID:         fid,
		CollectionID:      collection.ID,
		CollectionVersion: collection.Search.Version,
		DisplayName:       strings.Join(values, ","),
		Suggest:           strings.Join(values, ","),
		GeometryType:      r.GeometryType,
		Bbox:              bbox,
	}, nil
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

type SearchIndexRecord struct {
	FeatureID         string
	CollectionID      string
	CollectionVersion int
	DisplayName       string
	Suggest           string
	GeometryType      string
	Bbox              *pggeom.Polygon
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
