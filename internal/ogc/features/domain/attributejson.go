package domain

// -----
// "Attribute JSON" is NOT a formal standard. It's defined by PDOK as a way to return non-spatial data from certain collections
// alongside collections that do contain spatial data. Attribute JSON is modeled after GeoJSON but does NOT contain a geometry.
// -----

// AttributeCollection is a FeatureCollection with only attributes and NO geometries.
type AttributeCollection struct {
	Features []*Attribute `json:"features"`
	FeatureCollection
}

// Attribute is a Feature with only attributes and NO geometry.
type Attribute struct {
	Type       featureType       `json:"type"`
	Properties FeatureProperties `json:"properties"`
	// We expect ids to be auto-incrementing integers (which is the default in geopackages)
	// since we use it for cursor-based pagination.
	ID    string `json:"id"`
	Links []Link `json:"links,omitempty"`
}

// Keys of the Attribute properties.
func (f *Attribute) Keys() []string {
	return f.Properties.Keys()
}
