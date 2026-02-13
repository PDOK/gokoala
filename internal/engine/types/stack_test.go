package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStack_Push(t *testing.T) {
	tests := []struct {
		name     string
		values   []string
		expected []string
	}{
		{name: "single item", values: []string{"a"}, expected: []string{"a"}},
		{name: "multiple items", values: []string{"a", "b", "c"}, expected: []string{"a", "b", "c"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stack := NewStack()
			for _, v := range tt.values {
				stack.Push(v)
			}

			require.Equal(t, tt.expected, stack.stack)
		})
	}
}

func TestStack_Pop(t *testing.T) {
	tests := []struct {
		name         string
		initialStack []string
		popResult    string
		finalStack   []string
	}{
		{name: "pop from empty stack", initialStack: []string{}, popResult: "", finalStack: []string{}},
		{name: "pop last item", initialStack: []string{"a"}, popResult: "a", finalStack: []string{}},
		{name: "pop multiple items", initialStack: []string{"a", "b", "c"}, popResult: "c", finalStack: []string{"a", "b"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stack := &Stack{stack: tt.initialStack}
			result := stack.Pop()

			require.Equal(t, tt.popResult, result)
			require.Equal(t, tt.finalStack, stack.stack)
		})
	}
}

func TestStack_PopMany(t *testing.T) {
	tests := []struct {
		name         string
		initialStack []string
		count        int
		popResult    []string
		finalStack   []string
	}{
		{name: "pop many on empty stack", initialStack: []string{}, count: 3, popResult: []string{}, finalStack: []string{}},
		{name: "pop 0 items", initialStack: []string{"a", "b", "c"}, count: 0, popResult: nil, finalStack: []string{"a", "b", "c"}},
		{name: "pop less than stack size", initialStack: []string{"a", "b", "c"}, count: 2, popResult: []string{"b", "c"}, finalStack: []string{"a"}},
		{name: "pop all items", initialStack: []string{"a", "b", "c"}, count: 3, popResult: []string{"a", "b", "c"}, finalStack: []string{}},
		{name: "pop more than available", initialStack: []string{"a", "b"}, count: 5, popResult: []string{"a", "b"}, finalStack: []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stack := &Stack{stack: tt.initialStack}
			result := stack.PopMany(tt.count)

			if tt.popResult == nil {
				require.Nil(t, result)
			} else {
				require.Equal(t, tt.popResult, result)
			}
			require.Equal(t, tt.finalStack, stack.stack)
		})
	}
}

func TestStack_Peek(t *testing.T) {
	tests := []struct {
		name         string
		initialStack []string
		expected     string
		finalStack   []string
	}{
		{name: "peek on empty stack", initialStack: []string{}, expected: "", finalStack: []string{}},
		{name: "peek single item", initialStack: []string{"a"}, expected: "a", finalStack: []string{"a"}},
		{name: "peek multiple items", initialStack: []string{"a", "b", "c"}, expected: "c", finalStack: []string{"a", "b", "c"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stack := &Stack{stack: tt.initialStack}
			result := stack.Peek()

			require.Equal(t, tt.expected, result)
			require.Equal(t, tt.finalStack, stack.stack)
		})
	}
}
