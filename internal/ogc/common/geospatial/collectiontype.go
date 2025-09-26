package geospatial

import (
	"github.com/PDOK/gokoala/internal/engine"
	"github.com/twpayne/go-geom"
)

// CollectionType is the type of the data in a collection.
type CollectionType string

const (
	Features   CollectionType = "features"   // Geospatial data, https://docs.ogc.org/is/12-128r19/12-128r19.html#features
	Attributes CollectionType = "attributes" // Non-geospatial data. Same as features but without geometry, https://docs.ogc.org/is/12-128r19/12-128r19.html#attributes
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

func (ct CollectionType) AvailableOutputFormats() []engine.OutputFormat {
	switch ct {
	case Attributes:
		return engine.OutputFormatDefault
	case Features:
		return []engine.OutputFormat{
			{Key: engine.FormatJSON, Value: "GeoJSON"},
			{Key: engine.FormatJSONFG, Value: "JSON-FG"},
		}
	default:
		return engine.OutputFormatDefault
	}
}

// IsSpatialRequestAllowed returns true if the collection supports spatial requests such as bbox or other spatial filters.
func (ct CollectionType) IsSpatialRequestAllowed(bbox *geom.Bounds) bool {
	return !(ct == Attributes && bbox != nil)
}
