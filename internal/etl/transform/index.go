package transform

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/PDOK/gomagpie/config"
	"github.com/go-spatial/geom"
)

type RawRecord struct {
	FeatureID    int64
	FieldValues  []any
	Bbox         *geom.Extent
	GeometryType string
}

// Transform the 'T' of ETL
func (r RawRecord) Transform(collection config.GeoSpatialCollection) (SearchIndexRecord, error) {
	fid := strconv.FormatInt(r.FeatureID, 10)

	values, err := toStringSlice(r.FieldValues)
	if err != nil {
		return SearchIndexRecord{}, err
	}
	if r.Bbox.Area() <= 0 {
		r.Bbox = nil
	}
	return SearchIndexRecord{
		FeatureID:         fid,
		CollectionID:      collection.ID,
		CollectionVersion: collection.Search.Version,
		DisplayName:       strings.Join(values, ","),
		Suggest:           strings.Join(values, ","),
		GeometryType:      r.GeometryType,
		Bbox:              r.Bbox,
	}, nil
}

type SearchIndexRecord struct {
	FeatureID         string
	CollectionID      string
	CollectionVersion int
	DisplayName       string
	Suggest           string
	GeometryType      string
	Bbox              *geom.Extent
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
