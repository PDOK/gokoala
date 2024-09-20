package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFeatureProperties(t *testing.T) {
	tests := []struct {
		name      string
		inOrder   bool
		data      map[string]any
		wantOrder bool
	}{
		{
			name:      "Unordered properties",
			inOrder:   false,
			data:      map[string]any{"key1": "value1"},
			wantOrder: false,
		},
		{
			name:      "Ordered properties",
			inOrder:   true,
			data:      map[string]any{"key1": "value1"},
			wantOrder: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewFeaturePropertiesWithData(tt.inOrder, tt.data)

			if tt.wantOrder {
				assert.Nil(t, p.unordered)
				val, _ := p.ordered.Get("key1")
				assert.Equal(t, "value1", val)
			} else {
				assert.NotNil(t, p.unordered)
				assert.Equal(t, "value1", p.unordered["key1"])
			}
		})
	}
}

func TestMarshalJSON(t *testing.T) {
	tests := []struct {
		name      string
		inOrder   bool
		data      map[string]any
		expected  string
		expectErr bool
	}{
		{
			name:     "Unordered JSON marshal",
			inOrder:  false,
			data:     map[string]any{"key1": "value1"},
			expected: `{"key1":"value1"}`,
		},
		{
			name:     "Ordered JSON marshal",
			inOrder:  true,
			data:     map[string]any{"key1": "value1"},
			expected: `{"key1":"value1"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewFeaturePropertiesWithData(tt.inOrder, tt.data)
			bytes, err := p.MarshalJSON()

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.JSONEq(t, tt.expected, string(bytes))
			}
		})
	}
}

func TestValue(t *testing.T) {
	tests := []struct {
		name     string
		inOrder  bool
		data     map[string]any
		key      string
		expected any
	}{
		{
			name:     "Unordered value retrieval",
			inOrder:  false,
			data:     map[string]any{"key1": "value1"},
			key:      "key1",
			expected: "value1",
		},
		{
			name:     "Ordered value retrieval",
			inOrder:  true,
			data:     map[string]any{"key1": "value1"},
			key:      "key1",
			expected: "value1",
		},
		{
			name:     "Missing key",
			inOrder:  false,
			data:     map[string]any{"key1": "value1"},
			key:      "key2",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewFeaturePropertiesWithData(tt.inOrder, tt.data)
			value := p.Value(tt.key)
			assert.Equal(t, tt.expected, value)
		})
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		name        string
		inOrder     bool
		data        map[string]any
		deleteKey   string
		shouldExist bool
	}{
		{
			name:        "Delete item from unordered",
			inOrder:     false,
			data:        map[string]any{"key1": "value1", "key2": "value2"},
			deleteKey:   "key1",
			shouldExist: false,
		},
		{
			name:        "Delete item from ordered",
			inOrder:     true,
			data:        map[string]any{"key1": "value1", "key2": "value2"},
			deleteKey:   "key1",
			shouldExist: false,
		},
		{
			name:        "Missing key",
			inOrder:     true,
			data:        map[string]any{"key1": "value1"},
			deleteKey:   "key2",
			shouldExist: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewFeaturePropertiesWithData(tt.inOrder, tt.data)
			p.Delete(tt.deleteKey)

			if tt.shouldExist {
				assert.NotNil(t, p.Value(tt.deleteKey))
			} else {
				assert.Nil(t, p.Value(tt.deleteKey))
			}
		})
	}
}

func TestSet(t *testing.T) {
	tests := []struct {
		name     string
		inOrder  bool
		data     map[string]any
		key      string
		value    any
		expected any
	}{
		{
			name:     "Set unordered",
			inOrder:  false,
			data:     map[string]any{},
			key:      "key1",
			value:    "value1",
			expected: "value1",
		},
		{
			name:     "Set ordered",
			inOrder:  true,
			data:     map[string]any{},
			key:      "key1",
			value:    "value1",
			expected: "value1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewFeaturePropertiesWithData(tt.inOrder, tt.data)
			p.Set(tt.key, tt.value)
			assert.Equal(t, tt.expected, p.Value(tt.key))
		})
	}
}

func TestSetRelation(t *testing.T) {
	tests := []struct {
		name           string
		inOrder        bool
		data           map[string]any
		key            string
		value          any
		existingPrefix string
		expectedOrder  []string
	}{
		{
			name:           "Unordered set relation",
			inOrder:        false,
			data:           map[string]any{"related_key": "related_value"},
			key:            "new_key",
			value:          "new_value",
			existingPrefix: "related",
			expectedOrder:  []string{"new_key", "related_key"}, // unordered, so it doesn't matter
		},
		{
			name:           "Ordered set relation, item before related key",
			inOrder:        true,
			data:           map[string]any{"related_key": "related_value"},
			key:            "new_key",
			value:          "new_value",
			existingPrefix: "related",
			expectedOrder:  []string{"new_key", "related_key"}, // 'new_key' should appear before 'related_key'
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewFeaturePropertiesWithData(tt.inOrder, tt.data)
			p.SetRelation(tt.key, tt.value, tt.existingPrefix)

			keys := p.Keys()
			assert.Equal(t, tt.expectedOrder, keys)
		})
	}
}

func TestKeys(t *testing.T) {
	tests := []struct {
		name     string
		inOrder  bool
		data     map[string]any
		expected []string
	}{
		{
			name:     "Unordered keys",
			inOrder:  false,
			data:     map[string]any{"key2": "value2", "key1": "value1"},
			expected: []string{"key1", "key2"}, // sorted alphabetically
		},
		{
			name:     "Ordered keys",
			inOrder:  true,
			data:     map[string]any{"key1": "value1", "key2": "value2"},
			expected: []string{"key1", "key2"}, // insertion order maintained
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewFeaturePropertiesWithData(tt.inOrder, tt.data)
			keys := p.Keys()
			assert.Equal(t, tt.expected, keys)
		})
	}
}
