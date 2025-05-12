package domain

import (
	"errors"
	"log"
	"slices"
	"strings"
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
	publicFields := make(map[string]Field)
	nrOfGeomsFound := 0
	for _, field := range fields {
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

		publicFields[field.Name] = field
	}
	return &Schema{publicFields}, nil
}

// Schema derived from the data source (database) schema.
type Schema struct {
	Fields map[string]Field
}

// FieldsWithDataType flatten fields to name=>datatype
func (s Schema) FieldsWithDataType() map[string]string {
	result := make(map[string]string)
	for _, field := range s.Fields {
		result[field.Name] = field.Type
	}
	return result
}

// HasExternalFid convenience function
func (s Schema) HasExternalFid() bool {
	for _, field := range s.Fields {
		if field.IsExternalFid {
			return true
		}
	}
	return false
}

// Field a field/column/property in the schema. Contains at least a name and data type.
type Field struct {
	Name        string
	Type        string // can be data source specific
	Description string

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

// ToTypeFormat converts the Field's data type to a valid JSON data type and optional format as specified in OAF Part 5.
func (f Field) ToTypeFormat() TypeFormat {
	lowerCaseType := strings.ToLower(f.Type)

	switch lowerCaseType {
	case "boolean", "bool":
		return TypeFormat{Type: "boolean"}
	case "text", "char", "character", "character varying", "varchar", "nvarchar", "clob":
		return TypeFormat{Type: "string"}
	case "int", "integer", "tinyint", "smallint", "mediumint", "bigint", "int2", "int8":
		return TypeFormat{Type: "integer"}
	case "real", "float", "double", "double precision", "numeric", "decimal":
		return TypeFormat{Type: "number", Format: "double"}
	case "uuid":
		// From OAF Part 5: Properties that represent a UUID SHOULD be represented as a string with format "uuid".
		return TypeFormat{Type: "string", Format: "uuid"}
	case "date":
		// From OAF Part 5: Each temporal property SHALL be a "string" literal with the appropriate format
		// (e.g., "date-time" or "date" for instances, depending on the temporal granularity).
		return TypeFormat{Type: "string", Format: "date"}
	case "time":
		// From OAF Part 5: Each temporal property SHALL be a "string" literal with the appropriate format
		// (e.g., "date-time" or "date" for instances, depending on the temporal granularity).
		return TypeFormat{Type: "string", Format: "time"}
	case "datetime", "timestamp":
		// From OAF Part 5: Each temporal property SHALL be a "string" literal with the appropriate format
		// (e.g., "date-time" or "date" for instances, depending on the temporal granularity).
		return TypeFormat{Type: "string", Format: "date-time"}
	case geometryType, geometryCollectionType:
		// From OAF Part 5: the following special value is supported: "geometry-any" as the wildcard for any geometry type
		return TypeFormat{Type: lowerCaseType, Format: "geometry-any"}
	case pointType, linestringType, polygonType, multipointType, multilinestringType, multipolygonType:
		// From OAF Part 5: Each spatial property SHALL include a "format" member with a string value "geometry",
		// followed by a hyphen, followed by the name of the geometry type in lower case
		return TypeFormat{Type: lowerCaseType, Format: "geometry-" + lowerCaseType}
	default:
		log.Printf("Warning: unknown data type '%s' for field '%s', falling back to string", f.Type, f.Name)
		return TypeFormat{Type: "string"}
	}
}
