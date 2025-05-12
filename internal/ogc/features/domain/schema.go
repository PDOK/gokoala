package domain

import (
	"errors"
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

		field.Fid = field.Name == fidColumn
		field.ExternalFid = field.Name == externalFidColumn

		publicFields[field.Name] = field
	}
	return &Schema{publicFields}, nil
}

// Schema derived from the data source (database) schema.
type Schema struct {
	Fields map[string]Field
}

func (s Schema) FieldsWithDataType() map[string]string {
	result := make(map[string]string)
	for _, field := range s.Fields {
		result[field.Name] = field.Type
	}
	return result
}

func (s Schema) HasExternalFid() bool {
	for _, field := range s.Fields {
		if field.ExternalFid {
			return true
		}
	}
	return false
}

type Field struct {
	Name        string
	Type        string // can be data source specific
	Description string

	Required             bool
	PrimaryGeometry      bool
	PrimaryIntervalStart bool
	PrimaryIntervalEnd   bool
	Fid                  bool
	ExternalFid          bool
}

// TypeFormat type and optional format according to JSON schema (https://json-schema.org/).
type TypeFormat struct {
	Type   string
	Format string
}

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
		return TypeFormat{Type: lowerCaseType}
	}
}
