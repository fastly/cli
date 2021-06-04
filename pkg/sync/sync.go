package sync

import (
	"io"
	"sync"
)

// Writer protects any io.Writer with a mutex.
type Writer struct {
	mtx sync.Mutex
	w   io.Writer
}

// NewWriter wraps an io.Writer with a mutex.
func NewWriter(w io.Writer) *Writer {
	return &Writer{
		w: w,
	}
}

// Write implements io.Writer with mutex protection.
func (w *Writer) Write(p []byte) (int, error) {
	w.mtx.Lock()
	defer w.mtx.Unlock()
	return w.w.Write(p)
}
