package threadsafe

import (
	"bytes"
	"io"
	"sync"
)

// Buffer is a thread-safe bytes.Buffer instance.
type Buffer struct {
	b bytes.Buffer
	m sync.Mutex
}

// Read reads the next len(p) bytes from the buffer.
func (b *Buffer) Read(p []byte) (n int, err error) {
	b.m.Lock()
	defer b.m.Unlock()
	return b.b.Read(p)
}

// Write appends the contents of p to the buffer.
func (b *Buffer) Write(p []byte) (n int, err error) {
	b.m.Lock()
	defer b.m.Unlock()
	return b.b.Write(p)
}

// String returns the contents of the unread portion of the buffer
// as a string.
func (b *Buffer) String() string {
	b.m.Lock()
	defer b.m.Unlock()
	return b.b.String()
}

// Len returns the number of bytes of the unread portion of the buffer.
func (b *Buffer) Len() int {
	b.m.Lock()
	defer b.m.Unlock()
	return b.b.Len()
}

// NewWriter returns an instance of a thread-safe Writer.
func NewWriter(w io.Writer) *Writer {
	return &Writer{
		w: w,
	}
}

// Writer is a thread-safe io.Writer.
type Writer struct {
	w io.Writer
	m sync.Mutex
}

// Write writes the contents of bs to the io.Writer.
func (w *Writer) Write(bs []byte) (n int, err error) {
	w.m.Lock()
	defer w.m.Unlock()
	return w.w.Write(bs)
}
