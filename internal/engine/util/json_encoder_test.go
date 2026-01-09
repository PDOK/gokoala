package util

import (
	"bytes"
	"os"
	"strconv"
	"testing"

	stdjson "encoding/json"

	perfjson "github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetJSONEncoder(t *testing.T) {
	type testCase struct {
		name         string
		envVarValue  string
		expectedType string
	}
	var testCases = []testCase{
		{
			name:         "perf optimization enabled",
			envVarValue:  "false",
			expectedType: "*gojson.Encoder",
		},
		{
			name:         "perf optimization disabled",
			envVarValue:  "true",
			expectedType: "*json.Encoder",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("DISABLE_JSON_PERF_OPTIMIZATION", tc.envVarValue)
			disableJSONPerfOptimization, _ = strconv.ParseBool(os.Getenv("DISABLE_JSON_PERF_OPTIMIZATION"))

			buffer := &bytes.Buffer{}
			encoder := GetJSONEncoder(buffer)

			// Assert the type of the returned encoder
			assert.Equal(t, tc.expectedType, assertTypeName(encoder))
		})
	}
}

func TestJSONEncoderFunctionality(t *testing.T) {
	type testCase struct {
		name      string
		input     any
		expected  string
		shouldErr bool
	}
	var testCases = []testCase{
		{
			name:     "simple object",
			input:    map[string]string{"key": "value"},
			expected: `{"key":"value"}`,
		},
		{
			name:     "nested object",
			input:    map[string]any{"outer": map[string]string{"inner": "value"}},
			expected: `{"outer":{"inner":"value"}}`,
		},
		{
			name:     "empty object",
			input:    map[string]string{},
			expected: `{}`,
		},
		{
			name:     "array",
			input:    []int{1, 2, 3},
			expected: `[1,2,3]`,
		},
		{
			name:      "unsupported type",
			input:     func() {},
			shouldErr: true,
		},
	}

	t.Setenv("DISABLE_JSON_PERF_OPTIMIZATION", "true")
	disableJSONPerfOptimization, _ = strconv.ParseBool(os.Getenv("DISABLE_JSON_PERF_OPTIMIZATION"))
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// given
			buffer := &bytes.Buffer{}
			encoder := GetJSONEncoder(buffer)

			// when
			err := encoder.Encode(tc.input)

			// then
			if tc.shouldErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.JSONEq(t, tc.expected, buffer.String())
			}
		})
	}
}

func assertTypeName(i any) string {
	switch i.(type) {
	case *stdjson.Encoder:
		return "*json.Encoder"
	case *perfjson.Encoder:
		return "*gojson.Encoder"
	default:
		return "unknown"
	}
}
