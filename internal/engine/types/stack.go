package types

// Stack of strings. See https://en.wikipedia.org/wiki/Stack_(abstract_data_type)
type Stack struct {
	stack []string
}

func NewStack() *Stack {
	return &Stack{make([]string, 0)}
}

func (s *Stack) Push(value string) {
	s.stack = append(s.stack, value)
}

func (s *Stack) Pop() string {
	length := len(s.stack)
	if length == 0 {
		return ""
	}
	value := s.stack[length-1]
	s.stack = s.stack[:length-1]
	return value
}

func (s *Stack) PopMany(count int) []string {
	if count <= 1 {
		return nil
	}
	if count > len(s.stack) {
		count = len(s.stack)
	}
	items := make([]string, count)

	// Pop in reverse to maintain the correct ordering
	for i := count - 1; i >= 0; i-- {
		items[i] = s.Pop()
	}
	return items
}

func (s *Stack) Peek() string {
	length := len(s.stack)
	if length == 0 {
		return ""
	}
	return s.stack[length-1]
}
