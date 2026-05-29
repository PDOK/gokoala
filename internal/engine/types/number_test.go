package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsNumeric(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected bool
	}{
		// Valid numbers
		{"int", 123, true},
		{"int8", int8(123), true},
		{"int16", int16(123), true},
		{"int32", int32(123), true},
		{"int64", int64(123), true},
		{"uint", uint(123), true},
		{"uint8", uint8(123), true},
		{"uint16", uint16(123), true},
		{"uint32", uint32(123), true},
		{"uint64", uint64(123), true},
		{"float32", float32(123.45), true},
		{"float64", 123.45, true},
		{"complex64", complex(1.2, 3.4), true},
		{"complex128", complex(1.2, 3.4), true},
		{"zero int", 0, true},
		{"zero float", 0.0, true},
		{"zero complex", complex(0, 0), true},
		// Invalid numbers
		{"string", "123", false},
		{"bool", true, false},
		{"nil", nil, false},
		{"struct", struct{}{}, false},
		{"slice", []int{1, 2, 3}, false},
		{"map", map[string]int{"a": 1}, false},
		{"empty string", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsNumeric(tt.input))
		})
	}
}

func TestIsFloat(t *testing.T) {
	tests := []struct {
		name string
		f    float64
		want bool
	}{
		{
			name: "integer value",
			f:    42.0,
			want: false,
		},
		{
			name: "float with decimal value",
			f:    42.5,
			want: true,
		},
		{
			name: "negative integer value",
			f:    -100.0,
			want: false,
		},
		{
			name: "negative float with decimal value",
			f:    -100.123,
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsFloat(tt.f)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestToInt64(t *testing.T) {
	tests := []struct {
		name    string
		v       any
		want    int64
		wantErr bool
	}{
		{
			name:    "integer input",
			v:       42,
			want:    42,
			wantErr: false,
		},
		{
			name:    "int32 input",
			v:       int32(123),
			want:    123,
			wantErr: false,
		},
		{
			name:    "int64 input",
			v:       int64(987654321),
			want:    987654321,
			wantErr: false,
		},
		{
			name:    "string input",
			v:       "123",
			want:    0,
			wantErr: true,
		},
		{
			name:    "float input",
			v:       123.45,
			want:    0,
			wantErr: true,
		},
		{
			name:    "nil input",
			v:       nil,
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToInt64(tt.v)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
