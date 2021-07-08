package errors_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fastly/cli/pkg/errors"
)

func TestLogAdd(t *testing.T) {
	errors.Log.Add(fmt.Errorf("foo"))
	errors.Log.Add(fmt.Errorf("bar"))
	errors.Log.Add(fmt.Errorf("baz"))

	if len(*errors.Log) != 3 {
		t.Fatalf("want length %d, got: %d", 3, len(*errors.Log))
	}
}

func TestLogPersist(t *testing.T) {
	// The test will cause the testdata/errors.log file to be overwritten, and so
	// we need to reset that file at the end of each test run.
	segs := filepath.Join("testdata", "errors.log")
	path, err := filepath.Abs(segs)
	if err != nil {
		t.Fatal(err)
	}

	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	defer func(path string, b []byte) {
		err := os.WriteFile(path, b, 0644)
		if err != nil {
			t.Fatal(err)
		}
	}(path, b)

	errors.Now = func() (t time.Time) { return }

	errors.Log.Add(fmt.Errorf("foo"))
	errors.Log.Add(fmt.Errorf("bar"))
	errors.Log.Add(fmt.Errorf("baz"))

	err = errors.Log.Persist(path, []string{"command", "one", "--example"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = errors.Log.Persist(path, []string{"command", "two", "--example"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	have, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	segs = filepath.Join("testdata", "errors-expected.log")
	path, err = filepath.Abs(segs)
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if string(have) != string(want) {
		t.Fatalf("wanted content:\n%s\n, got:\n%s\n", string(want), string(have))
	}
}
