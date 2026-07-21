package domain

const (
	PropertyFilterWildcard = "*" // used as wildcard in OAF part 1
	Wildcard               = "%" // used as wildcard in CQL and SQL
)

// Queryables one or more QueryableWithAllowedValues indexed by queryable name.
type Queryables map[string]QueryableWithAllowedValues

// Fields flatten queryables to a slice of fields.
func (q Queryables) Fields() []Field {
	result := make([]Field, 0, len(q))
	for _, v := range q {
		result = append(result, v.Field)
	}
	return result
}

// QueryableWithAllowedValues a field from the datasource that can be used as a "queryable", optionally enriched
// with allowed values. A "queryable" is a field that can be used in a filter (part 1 filter or part 3 CQL filter).
type QueryableWithAllowedValues struct {
	Field

	// static or dynamic values that are allowed to be used in this queryable
	AllowedValues []string
}
