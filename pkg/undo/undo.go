package undo

import (
	"fmt"
	"io"
)

// Fn is a function with no arguments which returns an error or nil.
type Fn func() error

// Stack models a simple undo stack which consumers can use to store undo
// stateful functions, such as a function to teardown API state if something
// goes wrong during procedural commands, for example deleting a Fastly service
// after it's been created.
type Stack struct {
	states []Fn
}

// Stacker represents the API of a Stack.
type Stacker interface {
	Pop() Fn
	Push(elem Fn)
	Len() int
	RunIfError(w io.Writer, err error)
}

// NewStack constructs a new Stack.
func NewStack() *Stack {
	s := make([]Fn, 0, 1)
	stack := &Stack{
		states: s,
	}
	return stack
}

// Pop method pops last added Fn element off the stack and returns it.
// If stack is empty Pop() returns nil.
func (s *Stack) Pop() Fn {
	n := len(s.states)
	if n == 0 {
		return nil
	}
	v := s.states[n-1]
	s.states = s.states[:n-1]
	return v
}

// Push method pushes an element onto the Stack.
func (s *Stack) Push(elem Fn) {
	s.states = append(s.states, elem)
}

// Len method returns the number of elements in the Stack.
func (s *Stack) Len() int {
	return len(s.states)
}

// RunIfError unwinds the stack if a non-nil error is passed, by serially
// calling each Fn function state in FIFO order. If any Fn returns an
// error, it gets logged to the provided writer. Should be deferrerd, such as:
//
//	undoStack := undo.NewStack()
//	defer func() { undoStack.RunIfError(w, err) }()
func (s *Stack) RunIfError(w io.Writer, err error) {
	if err == nil {
		return
	}
	for i := len(s.states) - 1; i >= 0; i-- {
		if err := s.states[i](); err != nil {
			fmt.Fprintln(w, err)
		}
	}
}
