package testutil

import "testing"

// LogWriter is used to debug issues with our tests.
type LogWriter struct{ T *testing.T }

func (w LogWriter) Write(p []byte) (int, error) {
	// NOTE: text printed only if test fails or -test.v set
	w.T.Log(string(p))
	return len(p), nil
}
