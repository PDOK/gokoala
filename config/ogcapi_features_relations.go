package config

// +kubebuilder:object:generate=true
type Relation struct {
	// Name of the related collection
	RelatedCollection string `yaml:"collection" json:"collection" validate:"required"`

	// Optional prefix (infix actually) with a functional name for the relation. To distinguish multiple relations to the same collection/table.
	// +optional
	Prefix string `yaml:"prefix,omitempty" json:"prefix,omitempty"`

	// Column mappings between collections
	Columns ColumnRelation `yaml:"columns" json:"columns" validate:"required"`

	// Junction defines a junction/mapping table between collections in case this is a many-to-many relationship
	Junction JunctionTable `yaml:"junction,omitempty" json:"junction,omitempty" validate:"required"`
}

func (r *Relation) Name() string {
	result := r.RelatedCollection
	if r.Prefix != "" {
		result += "_" + r.Prefix
	}
	return result
}

// +kubebuilder:object:generate=true
type JunctionTable struct {
	// Name of the junction table
	// +kubebuilder:validation:Pattern=`^[a-zA-Z0-9_]+$`
	Name string `yaml:"name" json:"name" validate:"required"`

	// Column mappings for the junction table
	Columns ColumnRelation `yaml:"columns" json:"columns" validate:"required"`
}

// +kubebuilder:object:generate=true
type ColumnRelation struct {
	// Column name in the current/source collection
	// +kubebuilder:validation:Pattern=`^[a-zA-Z0-9_]+$`
	Source string `yaml:"source" json:"source" validate:"required"`

	// Column name in the target collection
	// +kubebuilder:validation:Pattern=`^[a-zA-Z0-9_]+$`
	Target string `yaml:"target" json:"target" validate:"required"`
}

// FeatureRelationsByID returns a map of collection IDs to their corresponding Relation.
// Skips collections that do not have features defined.
func (csf CollectionsFeatures) FeatureRelationsByID() map[string][]Relation {
	result := make(map[string][]Relation)
	for _, collection := range csf {
		result[collection.ID] = collection.Relations
	}

	return result
}
