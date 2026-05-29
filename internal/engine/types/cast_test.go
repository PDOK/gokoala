package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
