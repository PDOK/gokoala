package types

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewDate(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected Date
	}{
		{
			name:     "valid date",
			input:    time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC),
			expected: Date{time: time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC)},
		},
		{
			name:     "zero date",
			input:    time.Time{},
			expected: Date{time: time.Time{}},
		},
		{
			name:     "date with time component",
			input:    time.Date(2023, 12, 25, 15, 30, 45, 123456789, time.UTC),
			expected: Date{time: time.Date(2023, 12, 25, 15, 30, 45, 123456789, time.UTC)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewDate(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDate_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		date     Date
		expected string
		wantErr  bool
	}{
		{
			name:     "valid date",
			date:     NewDate(time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC)),
			expected: `"2023-12-25"`,
			wantErr:  false,
		},
		{
			name:     "zero date",
			date:     NewDate(time.Time{}),
			expected: "null",
			wantErr:  false,
		},
		{
			name:     "date with time component",
			date:     NewDate(time.Date(2023, 12, 25, 15, 30, 45, 0, time.UTC)),
			expected: `"2023-12-25"`,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.date.MarshalJSON()

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.JSONEq(t, tt.expected, string(result))
		})
	}
}

func TestDate_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Date
		wantErr  bool
	}{
		{
			name:     "valid date",
			input:    `"2023-12-25"`,
			expected: NewDate(time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC)),
			wantErr:  false,
		},
		{
			name:     "empty string",
			input:    `""`,
			expected: Date{},
			wantErr:  false,
		},
		{
			name:     "null value",
			input:    "null",
			expected: Date{},
			wantErr:  false,
		},
		{
			name:     "invalid date format",
			input:    `"2023/12/25"`,
			expected: Date{},
			wantErr:  true,
		},
		{
			name:     "invalid date value",
			input:    `"2023-13-45"`,
			expected: Date{},
			wantErr:  true,
		},
		{
			name:     "datetime string",
			input:    `"2023-12-25T15:30:45Z"`,
			expected: Date{},
			wantErr:  true,
		},
		{
			name:     "invalid JSON",
			input:    `"2023-12-25`,
			expected: Date{},
			wantErr:  true,
		},
		{
			name:     "number instead of string",
			input:    `12345`,
			expected: Date{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result Date
			err := result.UnmarshalJSON([]byte(tt.input))

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDate_String(t *testing.T) {
	tests := []struct {
		name     string
		date     Date
		expected string
	}{
		{
			name:     "valid date",
			date:     NewDate(time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC)),
			expected: "2023-12-25",
		},
		{
			name:     "zero date",
			date:     NewDate(time.Time{}),
			expected: "",
		},
		{
			name:     "date with time component",
			date:     NewDate(time.Date(2023, 12, 25, 15, 30, 45, 0, time.UTC)),
			expected: "2023-12-25",
		},
		{
			name:     "single digit month and day",
			date:     NewDate(time.Date(2023, 1, 5, 0, 0, 0, 0, time.UTC)),
			expected: "2023-01-05",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.date.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDate_JSONRoundTrip(t *testing.T) {
	tests := []struct {
		name string
		date Date
	}{
		{
			name: "valid date",
			date: NewDate(time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC)),
		},
		{
			name: "zero date",
			date: NewDate(time.Time{}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, err := json.Marshal(tt.date)
			assert.NoError(t, err)
			var result Date
			err = json.Unmarshal(jsonData, &result)
			assert.NoError(t, err)

			// For zero dates, both should be zero
			if tt.date.time.IsZero() {
				assert.True(t, result.time.IsZero())
			} else {
				// For non-zero dates, should match the date part only
				assert.Equal(t, tt.date.time.Format(time.DateOnly), result.time.Format(time.DateOnly))
			}
		})
	}
}

func TestDate_StructMarshaling(t *testing.T) {
	type TestStruct struct {
		Date         Date   `json:"date"`
		Name         string `json:"name"`
		OptionalDate *Date  `json:"optionalDate,omitempty"`
	}

	tests := []struct {
		name     string
		input    TestStruct
		expected string
	}{
		{
			name: "valid date",
			input: TestStruct{
				Date: NewDate(time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC)),
				Name: "test",
			},
			expected: `{"date":"2023-12-25","name":"test"}`,
		},
		{
			name: "zero date",
			input: TestStruct{
				Date: NewDate(time.Time{}),
				Name: "test",
			},
			expected: `{"date":null,"name":"test"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := json.Marshal(tt.input)
			assert.NoError(t, err)
			assert.JSONEq(t, tt.expected, string(result))
		})
	}
}
