package domain

import "github.com/twpayne/go-geom"

type CollectionType string

const (
	Features   CollectionType = "features"   // Geospatial data
	Attributes CollectionType = "attributes" // Non-geospatial data. Same as features but without geometry.
)

// ItemType indicator about the type of the items in a collection (the default value is 'feature').
// See https://docs.ogc.org/DRAFTS/20-024.html#collection-item-type-section
func (ct CollectionType) ItemType() string {
	switch ct {
	case Attributes:
		return "attribute"
	case Features:
		return "feature"
	default:
		return "feature"
	}
}

func (ct CollectionType) IsSpatialRequestAllowed(bbox *geom.Bounds) bool {
	return !(ct == Attributes && bbox != nil)
}
