package config

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// +kubebuilder:object:generate=true
type OgcAPIRecords struct {
	// Database that stores DCAT-AP-NL catalogs, datasets, and distributions.
	Datasource *Datasource `yaml:"datasource" json:"datasource" validate:"required"`

	// Catalog collections containing records served by this API.
	Collections RecordsCollections `yaml:"collections,omitempty" json:"collections,omitempty" validate:"omitempty,dive"`
}

// +kubebuilder:object:generate=true
type RecordsCollections []RecordsCollection

func (cs RecordsCollections) ContainsID(id string) bool {
	for _, collection := range cs {
		if collection.ID == id {
			return true
		}
	}
	return false
}

// +kubebuilder:object:generate=false
type RecordsCollection struct {
	// Unique ID of the records collection.
	// +kubebuilder:validation:Pattern=`^[a-z0-9"]([a-z0-9_-]*[a-z0-9"]+|)$`
	ID string `yaml:"id" json:"id" validate:"required,lowercase_id"`

	// Part 1 catalog properties, such as title, description, extent, and links.
	Metadata *GeoSpatialCollectionMetadata `yaml:"metadata,omitempty" json:"metadata,omitempty"`

	// Links pertaining to this records collection.
	Links *CollectionLinks `yaml:"links,omitempty" json:"links,omitempty"`

	// Conformance classes that apply to this catalog.
	ConformsTo []string `yaml:"conformsTo,omitempty" json:"conformsTo,omitempty" validate:"omitempty,dive,url"`

	// DCAT-AP-NL properties added to the OGC Record Collection representation.
	CatalogProperties map[string]JSONValue `yaml:"catalogProperties,omitempty" json:"catalogProperties,omitempty"`

	// Record properties that may be used by the sortby parameter.
	Sortables []string `yaml:"sortables,omitempty" json:"sortables,omitempty" validate:"omitempty,dive,required"`

	// Default record ordering when no sortby parameter is supplied.
	DefaultSortOrder []RecordsSortOrder `yaml:"defaultSortOrder,omitempty" json:"defaultSortOrder,omitempty" validate:"omitempty,dive"`

	// CQL queryables and operators supported for this Records catalog.
	Filters FeatureFilters `yaml:"filters,omitempty" json:"filters,omitempty"`

	// Records describing the resources in this catalog.
	Records Records `yaml:"records,omitempty" json:"records,omitempty" validate:"omitempty,dive"`
}

func (in *RecordsCollection) DeepCopyInto(out *RecordsCollection) {
	encoded, err := json.Marshal(in)
	if err == nil && json.Unmarshal(encoded, out) == nil {
		return
	}
	*out = *in
}

func (in *RecordsCollection) DeepCopy() *RecordsCollection {
	if in == nil {
		return nil
	}
	out := new(RecordsCollection)
	in.DeepCopyInto(out)
	return out
}

func (rc RecordsCollection) GetID() string {
	return rc.ID
}

func (rc RecordsCollection) GetMetadata() *GeoSpatialCollectionMetadata {
	return rc.Metadata
}

func (rc RecordsCollection) GetLinks() *CollectionLinks {
	return rc.Links
}

func (rc RecordsCollection) Merge(other GeoSpatialCollection) GeoSpatialCollection {
	otherRecords, ok := other.(RecordsCollection)
	if !ok {
		return rc
	}
	rc.Metadata = mergeMetadata(rc, otherRecords)
	rc.Links = mergeLinks(rc, otherRecords)
	rc.Records = append(rc.Records, otherRecords.Records...)
	return rc
}

// +kubebuilder:object:generate=true
type RecordsSortOrder struct {
	Field     string `yaml:"field" json:"field" validate:"required"`
	Direction string `yaml:"direction" json:"direction" validate:"required,oneof=asc desc"`
}

// +kubebuilder:object:generate=true
type Records []*Record

// +kubebuilder:object:generate=true
type Record struct {
	// Unique identifier of this record.
	ID string `yaml:"id" json:"id" validate:"required"`

	// STAC version implemented by this record.
	STACVersion string `yaml:"stacVersion" json:"stacVersion" validate:"required"`

	// STAC extension schema URIs used by this record.
	STACExtensions []string `yaml:"stacExtensions,omitempty" json:"stacExtensions,omitempty" validate:"omitempty,dive,url"`

	// Spatial footprint in GeoJSON, or nil when the resource has no geometry.
	Geometry JSONValue `yaml:"geometry" json:"geometry"`

	// GeoJSON bounding box, required when geometry is non-null.
	Bbox []float64 `yaml:"bbox,omitempty" json:"bbox,omitempty"`

	// Temporal extent as defined by OGC API - Records Part 1.
	Time *RecordTime `yaml:"time,omitempty" json:"time,omitempty"`

	// Record and STAC metadata, including the STAC datetime property.
	Properties map[string]JSONValue `yaml:"properties" json:"properties" validate:"required"`

	// Part 1 record links and STAC item links.
	Links []RecordsLink `yaml:"links,omitempty" json:"links,omitempty"`

	// STAC data assets keyed by asset identifier.
	Assets map[string]RecordsAsset `yaml:"assets" json:"assets" validate:"required,dive"`

	// STAC Collection identifier, when the record belongs to a STAC Collection.
	Collection string `yaml:"collection,omitempty" json:"collection,omitempty"`

	// OGC API - Records conformance classes used by this record.
	ConformsTo []string `yaml:"conformsTo,omitempty" json:"conformsTo,omitempty"`
}

// +kubebuilder:object:generate=true
type RecordTime struct {
	Date       string   `yaml:"date,omitempty" json:"date,omitempty"`
	Timestamp  string   `yaml:"timestamp,omitempty" json:"timestamp,omitempty"`
	Interval   []string `yaml:"interval,omitempty" json:"interval,omitempty"`
	Resolution string   `yaml:"resolution,omitempty" json:"resolution,omitempty"`
}

// +kubebuilder:object:generate=true
type RecordsLink struct {
	Rel      string   `yaml:"rel" json:"rel" validate:"required"`
	Href     string   `yaml:"href" json:"href" validate:"required,url"`
	Type     string   `yaml:"type,omitempty" json:"type,omitempty"`
	Title    string   `yaml:"title,omitempty" json:"title,omitempty"`
	Hreflang string   `yaml:"hreflang,omitempty" json:"hreflang,omitempty"`
	Profile  []string `yaml:"profile,omitempty" json:"profile,omitempty"`
}

// +kubebuilder:object:generate=true
type RecordsAsset struct {
	Href        string               `yaml:"href" json:"href" validate:"required,url"`
	Title       string               `yaml:"title,omitempty" json:"title,omitempty"`
	Description string               `yaml:"description,omitempty" json:"description,omitempty"`
	Type        string               `yaml:"type,omitempty" json:"type,omitempty"`
	Roles       []string             `yaml:"roles,omitempty" json:"roles,omitempty"`
	Fields      map[string]JSONValue `yaml:"fields,omitempty" json:"fields,omitempty"`
}

func validateRecords(records *OgcAPIRecords) error {
	if records == nil {
		return nil
	}
	seenCollections := make(map[string]bool)
	for _, collection := range records.Collections {
		if seenCollections[collection.ID] {
			return fmt.Errorf("duplicate Records collection ID %q", collection.ID)
		}
		seenCollections[collection.ID] = true
		seenRecords := make(map[string]bool)
		for _, record := range collection.Records {
			if seenRecords[record.ID] {
				return fmt.Errorf("duplicate record ID %q in collection %q", record.ID, collection.ID)
			}
			seenRecords[record.ID] = true
			datetime, exists := record.Properties["datetime"]
			if !exists {
				return fmt.Errorf("record %q in collection %q is missing STAC properties.datetime", record.ID, collection.ID)
			}
			datetimeValue, err := datetime.Any()
			if err != nil {
				return fmt.Errorf("record %q in collection %q has invalid properties.datetime", record.ID, collection.ID)
			}
			if datetimeValue == nil {
				for _, boundary := range []string{"start_datetime", "end_datetime"} {
					value, exists := record.Properties[boundary]
					if !exists || !validSTACTime(value) {
						return fmt.Errorf("record %q in collection %q requires valid properties.%s when datetime is null", record.ID, collection.ID, boundary)
					}
				}
			} else if !validSTACTime(datetimeValue) {
				return fmt.Errorf("record %q in collection %q has invalid properties.datetime", record.ID, collection.ID)
			}
			hasGeometry := record.Geometry != nil
			if record.Geometry != nil {
				geometry, err := record.Geometry.Any()
				hasGeometry = err != nil || geometry != nil
			}
			if !hasGeometry && len(record.Bbox) != 0 {
				return fmt.Errorf("record %q in collection %q cannot include bbox when geometry is null", record.ID, collection.ID)
			}
			if hasGeometry && len(record.Bbox) != 4 && len(record.Bbox) != 6 {
				return fmt.Errorf("record %q in collection %q requires a four- or six-value bbox when geometry is present", record.ID, collection.ID)
			}
			if err := validateExtensionDeclarations(record); err != nil {
				return fmt.Errorf("record %q in collection %q: %w", record.ID, collection.ID, err)
			}
		}
	}
	return nil
}

func validSTACTime(value any) bool {
	if encoded, ok := value.(JSONValue); ok {
		decoded, err := encoded.Any()
		if err != nil {
			return false
		}
		value = decoded
	}
	text, ok := value.(string)
	if !ok {
		if timestamp, ok := value.(time.Time); ok {
			text = timestamp.Format(time.RFC3339Nano)
		} else {
			return false
		}
	}
	parsed, err := time.Parse(time.RFC3339Nano, text)
	if err != nil {
		return false
	}
	_, offset := parsed.Zone()
	return offset == 0
}

func validateExtensionDeclarations(record *Record) error {
	used := make(map[string]bool)
	collect := func(fields map[string]JSONValue) {
		for name := range fields {
			prefix, _, hasPrefix := strings.Cut(name, ":")
			if hasPrefix {
				used[prefix] = true
			}
		}
	}
	collect(record.Properties)
	for _, asset := range record.Assets {
		collect(asset.Fields)
	}
	if len(used) == 0 {
		return nil
	}
	if len(record.STACExtensions) == 0 {
		return fmt.Errorf("extension fields require stacExtensions")
	}
	return nil
}

func ValidateRecords(records *OgcAPIRecords) error {
	return validateRecords(records)
}

func (rc RecordsCollection) SupportsCQL() bool {
	return rc.Filters.CQL.IsEnabled()
}

func (cs RecordsCollections) SupportsCQL() bool {
	for _, collection := range cs {
		if collection.SupportsCQL() {
			return true
		}
	}
	return false
}
