package config

import (
	"encoding/json"

	"gopkg.in/yaml.v3"
)

type JSONSchema struct {
	Content map[string]any `yaml:"-" json:"-"`
}

// MarshalJSON turn JSONSchema into JSON
// Value instead of pointer receiver because only that way it can be used for both.
func (js JSONSchema) MarshalJSON() ([]byte, error) {
	return json.Marshal(js.Content)
}

// UnmarshalJSON turn JSON into JSONSchema
func (js *JSONSchema) UnmarshalJSON(b []byte) error {
	c := map[string]any{}
	if err := json.Unmarshal(b, &c); err != nil {
		return err
	}
	js.Content = c
	return nil
}

// MarshalYAML turns JSONSchema into YAML.
// Value instead of pointer receiver because only that way it can be used for both.
func (js JSONSchema) MarshalYAML() (interface{}, error) {
	return yaml.Marshal(js.Content)
}

// UnmarshalYAML parses a string to JSONSchema
func (js *JSONSchema) UnmarshalYAML(unmarshal func(any) error) error {
	c := map[string]any{}
	if err := unmarshal(c); err != nil {
		return err
	}
	js.Content = c
	return nil
}

// DeepCopyInto copies the receiver, writes into out.
func (js *JSONSchema) DeepCopyInto(out *JSONSchema) {
	*out = *js
	if js.Content != nil {
		in, out := &js.Content, &out.Content
		*out = make(map[string]any, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy copy the receiver, create a new MediaType.
func (js *JSONSchema) DeepCopy() *JSONSchema {
	if js == nil {
		return nil
	}
	out := new(JSONSchema)
	js.DeepCopyInto(out)
	return out
}

// ParseRelations parses relations between features as defined in
// OAF Part 5: https://docs.ogc.org/DRAFTS/23-058r1.html#rc_feature-references
func (js *JSONSchema) ParseRelations() map[string]string {
	relations := make(map[string]string)
	if js.Content == nil {
		return relations
	}
	if props, ok := js.Content["properties"].(map[string]any); ok {
		for propName, prop := range props {
			propFields := prop.(map[string]any)
			collectionReference := propFields["x-ogc-collectionId"]
			if propFields["x-ogc-role"] == "reference" && collectionReference != nil {
				relations[propName] = collectionReference.(string)
			}
		}
	}
	return relations
}
