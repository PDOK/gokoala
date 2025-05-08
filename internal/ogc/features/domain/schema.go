package domain

import "strings"

type Schema struct {
	Fields map[string]Field
}

func (s *Schema) FieldsWithDataType() map[string]string {
	result := make(map[string]string)
	for _, field := range s.Fields {
		result[field.Name] = field.Type
	}
	return result
}

type Field struct {
	Name            string
	Type            string
	Description     string
	Required        bool
	PrimaryGeometry bool
}

type TypeFormat struct {
	Type   string
	Format string
}

func (f Field) ToJSONSchemaTypeFormat() TypeFormat {
	lowerCaseType := strings.ToLower(f.Type)

	switch lowerCaseType {
	case "boolean", "bool":
		return TypeFormat{Type: "boolean"}
	case "text", "char", "character", "character varying", "varchar", "nvarchar", "clob":
		return TypeFormat{Type: "string"}
	case "int", "integer", "tinyint", "smallint", "mediumint", "bigint", "int2", "int8":
		return TypeFormat{Type: "integer"}
	case "real", "float":
		return TypeFormat{Type: "number", Format: "float"}
	case "double", "double precision", "numeric", "decimal":
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
	case "geometry", "geometrycollection":
		// From OAF Part 5: the following special value is supported: "geometry-any" as the wildcard for any geometry type
		return TypeFormat{Type: lowerCaseType, Format: "geometry-any"}
	case "point", "linestring", "polygon", "multipoint", "multilinestring", "multipolygon":
		// From OAF Part 5: Each spatial property SHALL include a "format" member with a string value "geometry",
		// followed by a hyphen, followed by the name of the geometry type in lower case
		return TypeFormat{Type: lowerCaseType, Format: "geometry-" + lowerCaseType}
	default:
		return TypeFormat{Type: lowerCaseType}
	}
}
