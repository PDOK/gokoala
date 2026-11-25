package config

import (
	"bytes"
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
)

type JSONValue []byte

func NewJSONValue(value any) (JSONValue, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return JSONValue(encoded), nil
}

func (value JSONValue) Any() (any, error) {
	var decoded any
	if len(value) == 0 {
		return nil, nil
	}
	decoder := json.NewDecoder(bytes.NewReader(value))
	decoder.UseNumber()
	if err := decoder.Decode(&decoded); err != nil {
		return nil, err
	}
	return decoded, nil
}

func (value JSONValue) StringValue() string {
	decoded, err := value.Any()
	if err != nil {
		return string(value)
	}
	return fmt.Sprint(decoded)
}

func (value JSONValue) MarshalJSON() ([]byte, error) {
	if len(value) == 0 {
		return []byte("null"), nil
	}
	if !json.Valid(value) {
		return nil, fmt.Errorf("invalid JSON value")
	}
	return value, nil
}

func (value *JSONValue) UnmarshalJSON(data []byte) error {
	if !json.Valid(data) {
		return fmt.Errorf("invalid JSON value")
	}
	*value = append((*value)[:0], data...)
	return nil
}

func (value *JSONValue) UnmarshalYAML(node *yaml.Node) error {
	var decoded any
	if err := node.Decode(&decoded); err != nil {
		return err
	}
	encoded, err := json.Marshal(decoded)
	if err != nil {
		return err
	}
	*value = JSONValue(encoded)
	return nil
}

func (value JSONValue) MarshalYAML() (any, error) {
	return value.Any()
}
