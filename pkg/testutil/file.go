package testutil

import (
	"os"
	"testing"
)

// MakeTempFile creates a tempfile with the given contents and returns its path
func MakeTempFile(t *testing.T, contents string) string {
	t.Helper()

	tmpfile, err := os.CreateTemp("", "fastly-*")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tmpfile.Write([]byte(contents)); err != nil {
		t.Fatal(err)
	}
	return tmpfile.Name()
}
