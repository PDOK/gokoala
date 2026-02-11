package util

import "math/rand/v2"

// Randomizer is used to generate random numbers
type Randomizer interface {
	IntN(n int) int
}

// DefaultRandomizer is used for production and wraps the stdlib math/rand/v2 package
var DefaultRandomizer = defaultRandomizer{}

type defaultRandomizer struct{}

func (defaultRandomizer) IntN(n int) int { return rand.IntN(n) } //nolint:gosec

// MockRandomizer is used for testing and provides predictable results
type MockRandomizer struct {
	counter int
}

func (m *MockRandomizer) IntN(n int) int {
	if n <= 0 {
		return 0
	}
	m.counter++ // not so random, instead it's predictable for testing
	return m.counter % n
}
