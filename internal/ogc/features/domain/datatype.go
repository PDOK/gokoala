package domain

import "github.com/twpayne/go-geom"

// DataType is the type of the data in a collection.
type DataType string

const (
	Features   DataType = "features"   // Geospatial data, https://docs.ogc.org/is/12-128r19/12-128r19.html#features
	Attributes DataType = "attributes" // Non-geospatial data. Same as features but without geometry, https://docs.ogc.org/is/12-128r19/12-128r19.html#attributes
)

// ItemType indicator about the type of the items in a collection (the default value is 'feature').
// See https://docs.ogc.org/DRAFTS/20-024.html#collection-item-type-section
func (ct DataType) ItemType() string {
	switch ct {
	case Attributes:
		return "attribute"
	case Features:
		return "feature"
	default:
		return "feature"
	}
}

// IsSpatialRequestAllowed returns true if the collection supports spatial requests such as bbox or other spatial filters.
func (ct DataType) IsSpatialRequestAllowed(bbox *geom.Bounds) bool {
	return !(ct == Attributes && bbox != nil)
}
