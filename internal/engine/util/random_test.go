package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultRandomizer(t *testing.T) {
	tests := []struct {
		name        string
		input       int
		expectPanic bool
	}{
		{name: "positive input", input: 10, expectPanic: false},
		{name: "zero input", input: 0, expectPanic: true},
		{name: "negative input", input: -5, expectPanic: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			randomizer := DefaultRandomizer

			if tt.expectPanic {
				require.Panics(t, func() {
					randomizer.IntN(tt.input)
				})
				return
			}

			require.NotPanics(t, func() {
				_ = randomizer.IntN(tt.input)
			})

			// when
			result := randomizer.IntN(tt.input)

			// then
			assert.GreaterOrEqual(t, result, 0)
			assert.Less(t, result, tt.input)
		})
	}
}

func TestMockRandomizer(t *testing.T) {
	tests := []struct {
		name    string
		counter int
	}{
		{name: "initial state", counter: 0},
		{name: "incremented once", counter: 1},
		{name: "incremented twice", counter: 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			mock := MockRandomizer{counter: tt.counter}

			// when
			result := mock.IntN(100)

			// then
			assert.Equal(t, tt.counter+1, result)
		})
	}
}
