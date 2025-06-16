package domain

import (
	"errors"
	"log"
	"regexp"
	"slices"
	"strings"
)

const (
	formatDateOnly = "date"
	formatTimeOnly = "time"
	formatDateTime = "date-time"
)

const (
	MinxField = "minx"
	MinyField = "miny"
	MaxxField = "maxx"
	MaxyField = "maxy"
)

var fieldsToSkip = []string{
	MinxField,
	MinyField,
	MaxxField,
	MaxyField,
}

const (
	geometryType           = "geometry"
	geometryCollectionType = "geometrycollection"
	pointType              = "point"
	linestringType         = "linestring"
	polygonType            = "polygon"
	multipointType         = "multipoint"
	multilinestringType    = "multilinestring"
	multipolygonType       = "multipolygon"
)

var geometryTypes = []string{
	geometryType,
	geometryCollectionType,
	pointType,
	linestringType,
	polygonType,
	multipointType,
	multilinestringType,
	multipolygonType,
}

func NewSchema(fields []Field, fidColumn, externalFidColumn string) (*Schema, error) {
	publicFields := make([]Field, 0, len(fields))
	nrOfGeomsFound := 0

	for _, field := range fields {
		if field.Name == "" {
			return nil, errors.New("empty field name found, field name is required")
		}
		if field.Type == "" {
			return nil, errors.New("empty field type found, field type is required")
		}
		// Don't include internal/non-public fields in schema
		if slices.Contains(fieldsToSkip, strings.ToLower(field.Name)) {
			continue
		}
		// Don't allow multiple geometries. OAF Part 5 does support multiple geometries, but GeoPackage and GeoJSON don't
		if slices.Contains(geometryTypes, strings.ToLower(field.Type)) {
			nrOfGeomsFound++
			if nrOfGeomsFound > 1 {
				return nil, errors.New("more than one geometry field found! Currently only a single geometry " +
					"per collection is supported (also a restriction of GeoJSON and GeoPackage)")
			}
		}

		field.IsFid = field.Name == fidColumn
		field.IsExternalFid = field.Name == externalFidColumn

		publicFields = append(publicFields, field)
	}
	return &Schema{publicFields}, nil
}

// Schema derived from the data source schema.
// Describes the schema of a single collection (table in the data source).
type Schema struct {
	Fields []Field
}

// HasExternalFid convenience function to check if this schema defines an external feature ID
func (s Schema) HasExternalFid() bool {
	for _, field := range s.Fields {
		if field.IsExternalFid {
			return true
		}
	}
	return false
}

// FindField convenience function to get a Field by name
func (s Schema) FindField(name string) Field {
	for _, f := range s.Fields {
		if f.Name == name {
			return f
		}
	}
	return Field{}
}

// IsDate convenience function to check if field is a Date
func (s Schema) IsDate(field string) bool {
	f := s.FindField(field)
	return f.ToTypeFormat().Format == formatDateOnly
}

// Field a field/column/property in the schema. Contains at least a name and data type.
type Field struct {
	Name        string // required
	Type        string // required, can be data source specific
	Description string // optional

	IsRequired             bool
	IsPrimaryGeometry      bool
	IsPrimaryIntervalStart bool
	IsPrimaryIntervalEnd   bool
	IsFid                  bool
	IsExternalFid          bool

	FeatureRelation *FeatureRelation
}

// FeatureRelation a relation/reference from one feature to another in a different
// collection, according to OAF Part 5: https://docs.ogc.org/DRAFTS/23-058r1.html#rc_feature-references.
type FeatureRelation struct {
	Name         string
	CollectionID string
}

func NewFeatureRelation(name, externalFidColumn string, collectionNames []string) *FeatureRelation {
	if !isFeatureRelation(name, externalFidColumn) {
		return nil
	}
	regex, _ := regexp.Compile(regexRemoveSeparators + externalFidColumn + regexRemoveSeparators)
	referencePropertyName := regex.ReplaceAllString(name, "")
	return &FeatureRelation{
		Name:         referencePropertyName,
		CollectionID: findReferencedCollection(collectionNames, referencePropertyName),
	}
}

// isFeatureRelation "Algorithm" to determine feature reference:
//
//	When externalFidColumn (e.g. 'external_fid') is part of the column name (e.g. 'foobar_external_fid')
//	we treat the field as a relation/reference to another feature.
func isFeatureRelation(columnName string, externalFidColumn string) bool {
	if externalFidColumn == "" || columnName == externalFidColumn {
		return false
	}
	return strings.Contains(columnName, externalFidColumn)
}

func findReferencedCollection(collectionNames []string, name string) string {
	// prefer exact matches first
	for _, collName := range collectionNames {
		if name == collName {
			return collName
		}
	}
	// then prefer fuzzy match (to support infix)
	for _, collName := range collectionNames {
		if strings.HasPrefix(name, collName) {
			return collName
		}
	}
	log.Printf("Warning: could not find collection for feature reference '%s'", name)
	return ""
}

// TypeFormat type and optional format according to JSON schema (https://json-schema.org/).
type TypeFormat struct {
	Type   string
	Format string
}

// ToTypeFormat converts the Field's data type (from SQLite or Postgres) to a valid JSON data type
// and optional format as specified in OAF Part 5.
func (f Field) ToTypeFormat() TypeFormat {
	// lowercase, no spaces
	normalizedType := strings.ReplaceAll(strings.ToLower(f.Type), " ", "")

	switch normalizedType {
	case "boolean", "bool":
		return TypeFormat{Type: "boolean"}
	case "text", "char", "character", "charactervarying", "varchar", "nvarchar", "clob":
		return TypeFormat{Type: "string"}
	case "int", "integer", "tinyint", "smallint", "mediumint", "bigint", "int2", "int8":
		return TypeFormat{Type: "integer"}
	case "real", "float", "double", "doubleprecision", "numeric", "decimal":
		return TypeFormat{Type: "number", Format: "double"}
	case "uuid":
		// From OAF Part 5: Properties that represent a UUID SHOULD be represented as a string with format "uuid".
		return TypeFormat{Type: "string", Format: "uuid"}
	case "date":
		// From OAF Part 5: Each temporal property SHALL be a "string" literal with the appropriate format
		// (e.g., "date-time" or "date" for instances, depending on the temporal granularity).
		return TypeFormat{Type: "string", Format: formatDateOnly}
	case "time":
		// From OAF Part 5: Each temporal property SHALL be a "string" literal with the appropriate format
		// (e.g., "date-time" or "date" for instances, depending on the temporal granularity).
		return TypeFormat{Type: "string", Format: formatTimeOnly}
	case "datetime", "timestamp":
		// From OAF Part 5: Each temporal property SHALL be a "string" literal with the appropriate format
		// (e.g., "date-time" or "date" for instances, depending on the temporal granularity).
		return TypeFormat{Type: "string", Format: formatDateTime}
	case geometryType, geometryCollectionType:
		// From OAF Part 5: the following special value is supported: "geometry-any" as the wildcard for any geometry type
		return TypeFormat{Type: normalizedType, Format: "geometry-any"}
	case pointType, linestringType, polygonType, multipointType, multilinestringType, multipolygonType:
		// From OAF Part 5: Each spatial property SHALL include a "format" member with a string value "geometry",
		// followed by a hyphen, followed by the name of the geometry type in lower case
		return TypeFormat{Type: normalizedType, Format: "geometry-" + normalizedType}
	default:
		if strings.Contains(normalizedType, "text(") || strings.Contains(normalizedType, "varchar(") {
			// Sometimes datasources mention the length of fields, this is irrelevant.
			// Also, SQLite accepts for example TEXT(5) but ignores the length: https://sqlite.org/datatype3.html#affinity_name_examples
			return TypeFormat{Type: "string"}
		}
		log.Printf("Warning: unknown data type '%s' for field '%s', falling back to string", f.Type, f.Name)
		return TypeFormat{Type: "string"}
	}
}
