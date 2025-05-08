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
	case "uuid":
		return TypeFormat{Type: "string", Format: "uuid"}
	case "date":
		return TypeFormat{Type: "string", Format: "date"}
	case "time":
		return TypeFormat{Type: "string", Format: "time"}
	case "datetime", "timestamp":
		return TypeFormat{Type: "string", Format: "date-time"}
	case "real", "float", "double", "double precision", "numeric", "decimal":
		return TypeFormat{Type: "number"}
	case "int", "integer", "tinyint", "smallint", "mediumint", "bigint", "int2", "int8":
		return TypeFormat{Type: "integer"}
	case "geometry", "geometrycollection":
		return TypeFormat{Type: lowerCaseType, Format: "geometry-any"}
	case "point", "linestring", "polygon", "multipoint", "multilinestring", "multipolygon":
		return TypeFormat{Type: lowerCaseType, Format: "geometry-" + lowerCaseType}
	default:
		return TypeFormat{Type: lowerCaseType}
	}
}
