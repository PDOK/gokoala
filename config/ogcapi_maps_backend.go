package config

import (
	"encoding/json"
	"errors"
	"fmt"

	mapserverv3 "github.com/pdok/mapserver-operator/api/v3"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
)

const (
	MapsBackendExternalWMS = "externalWms"
	MapsBackendManagedWMS  = "managedWms"
)

// +kubebuilder:object:generate=true
// +kubebuilder:validation:XValidation:rule="(self.type != 'externalWms' || has(self.externalWms)) && (self.type != 'managedWms' || has(self.managedWms))",message="configure the block matching the backend type"
type MapsBackend struct {
	// Type of backend used to render maps.
	// +kubebuilder:validation:Enum=externalWms;managedWms
	Type string `yaml:"type" json:"type"`

	// Configuration for an external WMS backend.
	ExternalWMS *ExternalWMSBackend `yaml:"externalWms,omitempty" json:"externalWms,omitempty"`

	// Configuration for a managed WMS provisioned by ogcapi-operator.
	ManagedWMS *ManagedWMSBackend `yaml:"managedWms,omitempty" json:"managedWms,omitempty"`
}

// +kubebuilder:object:generate=true
type ExternalWMSBackend struct {
	// Base URL of the external WMS service.
	URL *URL `yaml:"url" json:"url"`
}

// +kubebuilder:object:generate=true
type ManagedWMSBackend struct {
	// Internal URL of the managed WMS service.
	URL *URL `json:"url"`

	// CRS of the source data
	// +optional
	DataEPSG string `json:"dataEPSG"` //nolint:tagliatelle

	// Optional WMS title override. Defaults to the OGC API title.
	Title string `json:"title,omitempty"`

	// Root layer of the generated WMS.
	// The WMS CR validates the completed layer after the operator adds inherited metadata.
	// +kubebuilder:validation:Type=object
	// +kubebuilder:validation:Schemaless
	// +kubebuilder:pruning:PreserveUnknownFields
	// +optional
	Layer mapserverv3.Layer `json:"layer"`

	// ConfigMap and optional blob references for map styling.
	// +optional
	StylingAssets mapserverv3.StylingAssets `json:"stylingAssets"`

	// Optional override for the WMS probe.
	HealthCheck *mapserverv3.HealthCheckWMS `json:"healthCheck,omitempty"`

	// Optional resource requests and limits for MapServer.
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
}

// UnmarshalYAML uses the WMS API's JSON field names, including embedded data fields.
func (m *ManagedWMSBackend) UnmarshalYAML(node *yaml.Node) error {
	var input map[string]any
	if err := node.Decode(&input); err != nil {
		return err
	}
	if layer, ok := input["layer"].(map[string]any); ok {
		defaultLayerVisibility(layer)
	}
	data, err := json.Marshal(input)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, m)
}

func defaultLayerVisibility(layer map[string]any) {
	if _, ok := layer["visible"]; !ok {
		layer["visible"] = true
	}
	if layers, ok := layer["layers"].([]any); ok {
		for _, child := range layers {
			if layer, ok := child.(map[string]any); ok {
				defaultLayerVisibility(layer)
			}
		}
	}
}

func (m ManagedWMSBackend) MarshalYAML() (any, error) {
	data, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func validateMapBackend(backend *MapsBackend) error {
	if backend == nil {
		return errors.New("ogcApi.maps.backend is required")
	}
	// Both backend configurations may be present to facilitate migration; validate the selected one.
	switch backend.Type {
	case MapsBackendExternalWMS:
		if backend.ExternalWMS == nil || backend.ExternalWMS.URL == nil || backend.ExternalWMS.URL.URL == nil {
			return errors.New("ogcApi.maps.backend.externalWms.url is required for externalWms")
		}
	case MapsBackendManagedWMS:
		if backend.ManagedWMS == nil || backend.ManagedWMS.URL == nil || backend.ManagedWMS.URL.URL == nil {
			return errors.New("ogcApi.maps.backend.managedWms.url is required for managedWms")
		}
	default:
		return fmt.Errorf("unsupported ogcApi.maps.backend.type %q", backend.Type)
	}
	return nil
}

// WMSURL returns the endpoint for the selected WMS backend, or nil if it cannot be selected.
func (backend *MapsBackend) WMSURL() *URL {
	if backend == nil {
		return nil
	}
	switch backend.Type {
	case MapsBackendExternalWMS:
		if backend.ExternalWMS != nil {
			return backend.ExternalWMS.URL
		}
	case MapsBackendManagedWMS:
		if backend.ManagedWMS != nil {
			return backend.ManagedWMS.URL
		}
	}
	return nil
}
