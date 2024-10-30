package transform

import "github.com/go-spatial/geom"

type RawRecord struct {
	FeatureID        int64
	FieldsWithValues map[string]any
	Bbox             geom.Extent
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
