package transform

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/go-spatial/geom"
)

type RawRecord struct {
	FeatureID   int64
	FieldValues []any
	Bbox        geom.Extent
}

func (r RawRecord) Transform() (SearchIndexRecord, error) {
	fid := strconv.FormatInt(r.FeatureID, 10)

	values, err := toStringSlice(r.FieldValues)
	if err != nil {
		return SearchIndexRecord{}, err
	}
	return SearchIndexRecord{
		FeatureID:   fid,
		DisplayName: strings.Join(values, ","),
		Suggest:     strings.Join(values, ","),
		Bbox:        r.Bbox,
	}, nil
}

type SearchIndexRecord struct {
	FeatureID         string
	CollectionID      string
	CollectionVersion string
	DisplayName       string
	Suggest           string
	GeometryType      string
	Bbox              geom.Extent
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
