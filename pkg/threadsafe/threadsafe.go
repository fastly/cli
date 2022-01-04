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

func (b *Buffer) Read(p []byte) (n int, err error) {
	b.m.Lock()
	defer b.m.Unlock()
	return b.b.Read(p)
}

func (b *Buffer) Write(p []byte) (n int, err error) {
	b.m.Lock()
	defer b.m.Unlock()
	return b.b.Write(p)
}

func (b *Buffer) String() string {
	b.m.Lock()
	defer b.m.Unlock()
	return b.b.String()
}

func (b *Buffer) Len() int {
	b.m.Lock()
	defer b.m.Unlock()
	return b.b.Len()
}
