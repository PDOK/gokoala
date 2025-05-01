package domain

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
