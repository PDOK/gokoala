package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type foo interface {
	SayHello() string
}

type bar struct {
	name string
}

func (b bar) SayHello() string {
	return "Hello " + b.name
}

func TestIsDate(t *testing.T) {
	tests := []struct {
		name string
		t    time.Time
		want bool
	}{
		{
			name: "time with zero hour, minute, second",
			t:    time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
			want: true,
		},
		{
			name: "time with non-zero hour",
			t:    time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC),
			want: false,
		},
		{
			name: "time with non-zero minute",
			t:    time.Date(2026, 1, 15, 0, 15, 0, 0, time.UTC),
			want: false,
		},
		{
			name: "time with non-zero second",
			t:    time.Date(2026, 1, 15, 0, 0, 45, 0, time.UTC),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsDate(tt.t)
			assert.Equal(t, tt.want, got)
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

func TestToInterfaceSlice(t *testing.T) {
	tests := []struct {
		name string
		in   []any
		want []any
	}{
		{
			name: "integers to interfaces",
			in:   []any{1, 2, 3},
			want: []any{1, 2, 3},
		},
		{
			name: "strings to interfaces",
			in:   []any{"a", "b", "c"},
			want: []any{"a", "b", "c"},
		},
		{
			name: "empty input slice",
			in:   []any{},
			want: []any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToInterfaceSlice[any, any](tt.in)
			assert.Equal(t, tt.want, got)
		})
	}

	t.Run("structs to interfaces", func(t *testing.T) {
		in := []bar{{name: "A"}, {name: "B"}}
		want := []foo{bar{name: "A"}, bar{name: "B"}}

		got := ToInterfaceSlice[bar, foo](in)

		assert.Equal(t, want, got)
		assert.Equal(t, "Hello A", got[0].SayHello())
	})
}
