package testutil

import (
	"bytes"
	"sync"
)

// MockStdout provides a concurrent-safe wrapper around bytes.Buffer to
// replicate the behavior of os.Stdout (or os.Stderr) without
// introducing potential data races from the buffer resizing itself
// during Write calls (see bytes.(*Buffer).tryGrowByReslice()). In
// particular, yacspin uses multiple goroutines internally that can
// cause these races when using a plain bytes.Buffer.
type MockStdout struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (m *MockStdout) Write(p []byte) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.buf.Write(p)
}

func (m *MockStdout) String() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.buf.String()
}
