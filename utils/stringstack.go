package utils

import "github.com/palantir/stacktrace"

type StringStack struct {
	stack []string
}

var ErrEmptyStack = stacktrace.NewError("empty stack")

func NewStringStack() *StringStack {
	return &StringStack{[]string{}}
}

func (s *StringStack) Push(value string) {
	s.stack = append(s.stack, value)
}

func (s *StringStack) Pop() (string, error) {
	if len(s.stack) == 0 {
		return "", ErrEmptyStack
	}
	value := s.stack[len(s.stack)-1]
	s.stack = s.stack[0 : len(s.stack)-1]
	return value, nil
}

func (s *StringStack) Head() (string, error) {
	if len(s.stack) == 0 {
		return "", ErrEmptyStack
	}
	return s.stack[len(s.stack)-1], nil
}
