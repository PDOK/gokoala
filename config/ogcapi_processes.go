package config

// +kubebuilder:object:generate=true
type OgcAPIProcesses struct {
	// Enable to advertise dismiss operations on the conformance page
	SupportsDismiss bool `yaml:"supportsDismiss" json:"supportsDismiss"`

	// Enable to advertise callback operations on the conformance page
	SupportsCallback bool `yaml:"supportsCallback" json:"supportsCallback"`

	// Reference to an external service implementing the process API. GoKoala acts only as a proxy for OGC API Processes.
	ProcessesServer URL `yaml:"processesServer" json:"processesServer" validate:"required"`
}
