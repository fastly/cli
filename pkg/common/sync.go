package common

import (
	"io"
	"sync"
)

// SyncWriter protects any io.Writer with a mutex.
type SyncWriter struct {
	mtx sync.Mutex
	w   io.Writer
}

// NewSyncWriter wraps an io.Writer with a mutex.
func NewSyncWriter(w io.Writer) *SyncWriter {
	return &SyncWriter{
		w: w,
	}
}

// Write implements io.Writer with mutex protection.
func (w *SyncWriter) Write(p []byte) (int, error) {
	w.mtx.Lock()
	defer w.mtx.Unlock()
	return w.w.Write(p)
}
