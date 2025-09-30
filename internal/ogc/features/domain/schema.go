package domain

import (
	"errors"
	"fmt"
	"log"
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

// Schema derived from the data source schema.
// Describes the schema of a single collection (table in the data source).
type Schema struct {
	Fields []Field
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

	schema := &Schema{publicFields}
	if externalFidColumn != "" && !schema.HasExternalFid() {
		return nil, fmt.Errorf("external feature ID column '%s' configured but not found in schema", externalFidColumn)
	}

	return schema, nil
}

// IsDate convenience function to check if the given field is a Date.
func (s Schema) IsDate(field string) bool {
	f := s.findField(field)

	return f.ToTypeFormat().Format == formatDateOnly
}

// HasExternalFid convenience function to check if this schema defines an external feature ID.
func (s Schema) HasExternalFid() bool {
	for _, field := range s.Fields {
		if field.IsExternalFid {
			return true
		}
	}

	return false
}

func (s Schema) findField(name string) Field {
	for _, f := range s.Fields {
		if f.Name == name {
			return f
		}
	}

	return Field{}
}

func (s Schema) findFeatureRelation(name string) *FeatureRelation {
	for _, field := range s.Fields {
		if field.FeatureRelation != nil && field.FeatureRelation.Name == name {
			return field.FeatureRelation
		}
	}

	return nil
}

// Field a field/column/property in the schema. Contains at least a name and data type.
type Field struct {
	FeatureRelation *FeatureRelation
	Name            string // required
	Type            string // required, can be data source specific
	Description     string // optional

	IsRequired             bool
	IsPrimaryGeometry      bool
	IsPrimaryIntervalStart bool
	IsPrimaryIntervalEnd   bool
	IsFid                  bool
	IsExternalFid          bool
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
	// sometimes data sources mention the length of fields within parenthesis, this is irrelevant.
	// also, SQLite accepts for example TEXT(5) but ignores the length: https://sqlite.org/datatype3.html#affinity_name_examples
	normalizedType = prefixBeforeParenthesis(normalizedType)

	switch normalizedType {
	case "boolean", "bool":
		return TypeFormat{Type: "boolean"}
	case "text", "char", "character", "charactervarying", "varchar", "nvarchar", "clob":
		return TypeFormat{Type: "string"}
	case "int", "integer", "tinyint", "smallint", "mediumint", "bigint", "int2", "int4", "int8":
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
	case "datetime", "timestamp", "timestampwithtimezone", "timestampwithouttimezone":
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
		// handle geometry types with additional Z and/or M dimensions e.g., LineStringZ, or PointZM.
		// OAF part 5 only supports simple 2D types, so advertise the 2D variant.
		for _, geomType := range geometryTypes {
			if strings.HasPrefix(normalizedType, geomType) {
				return TypeFormat{Type: normalizedType, Format: "geometry-" + geomType}
			}
		}
		log.Printf("Warning: unknown data type '%s' for field '%s', falling back to string", f.Type, f.Name)

		return TypeFormat{Type: "string"}
	}
}

func prefixBeforeParenthesis(s string) string {
	idx := strings.Index(s, "(")
	if idx != -1 {
		return s[:idx]
	}

	return s
}
