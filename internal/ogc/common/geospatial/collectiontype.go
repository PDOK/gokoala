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

// ItemType indicator about the type of the items in a collection. The default value is 'feature'.
// Other OGC-approved item types are e.g. 'record' and 'movingfeature'.
//
// See https://docs.ogc.org/DRAFTS/20-024.html#collection-item-type-section
func (ct CollectionType) ItemType() string {
	return "feature"
}

// AvailableFormats returns the output formats available for the current page.
func (ct CollectionType) AvailableFormats() []engine.OutputFormat {
	switch ct {
	case Attributes:
		return engine.OutputFormatDefault
	case Features:
		return []engine.OutputFormat{
			{Key: engine.FormatJSON, Name: "GeoJSON"},
			{Key: engine.FormatJSONFG, Name: "JSON-FG"},
		}
	default:
		return engine.OutputFormatDefault
	}
}

// IsSpatialRequestAllowed returns true if the collection supports spatial requests such as bbox or other spatial filters.
func (ct CollectionType) IsSpatialRequestAllowed(bbox *geom.Bounds) bool {
	return ct != Attributes || bbox == nil
}

// CollectionTypes one or more CollectionType.
type CollectionTypes struct {
	Types     map[string]CollectionType
	GeomTypes map[string]string
}

func NewCollectionTypes(types map[string]CollectionType, geomTypes map[string]string) CollectionTypes {
	return CollectionTypes{types, geomTypes}
}

func (cts CollectionTypes) GetCollectionType(collection string) CollectionType {
	return cts.Types[collection]
}

func (cts CollectionTypes) GetGeometryType(collection string) string {
	return cts.GeomTypes[collection]
}

func (cts CollectionTypes) HasAttributes() bool {
	for _, ct := range cts.Types {
		if ct == Attributes {
			return true
		}
	}

	return false
}
