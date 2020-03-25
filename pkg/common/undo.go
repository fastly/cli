package common

import (
	"fmt"
	"io"
)

// UndoFn is a function with no arguments which returns an error or nil.
type UndoFn func() error

// UndoStack models a simple undo stack which consumers can use to store undo
// stateful functions, such as a function to teardown API state if something
// goes wrong during procedural commands, for example deleting a Fastly service
// after it's been created.
type UndoStack struct {
	states []UndoFn
}

// NewUndoStack constructs a new UndoStack.
func NewUndoStack() *UndoStack {
	s := make([]UndoFn, 0, 1)
	stack := &UndoStack{
		states: s,
	}
	return stack
}

// Pop method pops last added UndoFn element oof the stack and returns it.
// If stack is empty Pop() returns nil.
func (s *UndoStack) Pop() UndoFn {
	n := len(s.states)
	if n == 0 {
		return nil
	}
	v := s.states[n-1]
	s.states = s.states[:n-1]
	return v
}

// Push method pushes an Undoer element onto the UndoStack.
func (s *UndoStack) Push(elem UndoFn) {
	s.states = append(s.states, elem)
}

// Len method returns the number of elements in the UndoStack.
func (s *UndoStack) Len() int {
	return len(s.states)
}

// RunIfError unwinds the stack if a non-nil error is passed, by serially
// calling each UndoFn function state in FIFO order. If any UndoFn returns an
// error, it gets logged to the provided writer. Should be deferrerd, such as:
//
//     undoStack := common.NewUndoStack()
//     defer func() { undoStack.RunIfError(w, err) }()
//
func (s *UndoStack) RunIfError(w io.Writer, err error) {
	if err == nil {
		return
	}
	for i := len(s.states) - 1; i >= 0; i-- {
		if err := s.states[i](); err != nil {
			fmt.Fprintln(w, err)
		}
	}
}
