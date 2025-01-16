package config

// +kubebuilder:object:generate=true
type OgcAPIStyles struct {
	// ID of the style to use a default
	Default string `yaml:"default" json:"default" validate:"required"`

	// Location on disk where the styles are hosted
	StylesDir string `yaml:"stylesDir" json:"stylesDir" validate:"required,dirpath|filepath"`

	// Styles exposed though this API
	SupportedStyles []Style `yaml:"supportedStyles" json:"supportedStyles" validate:"required,dive"`
}

// +kubebuilder:object:generate=true
type Style struct {
	// Unique ID of this style
	// +kubebuilder:validation:Pattern=`^[a-z0-9"]([a-z0-9_-]*[a-z0-9"]+|)$`
	ID string `yaml:"id" json:"id" validate:"required,lowercase_id"`

	// Human-friendly name of this style
	Title string `yaml:"title" json:"title" validate:"required"`

	// Explains what is visualized by this style
	// +optional
	Description *string `yaml:"description,omitempty" json:"description,omitempty"`

	// Keywords to make this style better discoverable
	// +optional
	Keywords []string `yaml:"keywords,omitempty" json:"keywords,omitempty"`

	// Moment in time when the style was last updated
	// +optional
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Format="date-time"
	LastUpdated *string `yaml:"lastUpdated,omitempty" json:"lastUpdated,omitempty" validate:"omitempty,datetime=2006-01-02T15:04:05Z"`

	// Optional version of this style
	// +optional
	Version *string `yaml:"version,omitempty" json:"version,omitempty"`

	// Reference to a PNG image to use a thumbnail on the style metadata page.
	// The full path is constructed by appending Resources + Thumbnail.
	// +optional
	Thumbnail *string `yaml:"thumbnail,omitempty" json:"thumbnail,omitempty"`

	// This style is offered in the following formats
	Formats []StyleFormat `yaml:"formats" json:"formats" validate:"required,dive"`
}

// +kubebuilder:object:generate=true
type StyleFormat struct {
	// Name of the format
	// +kubebuilder:default="mapbox"
	// +optional
	Format string `yaml:"format,omitempty" json:"format,omitempty" default:"mapbox" validate:"required,oneof=mapbox sld10"`
}
