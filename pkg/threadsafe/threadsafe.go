package threadsafe

import (
	"bytes"
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
