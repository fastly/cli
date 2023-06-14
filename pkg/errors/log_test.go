package errors_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/testutil"
)

func TestLogAdd(t *testing.T) {
	le := new(errors.LogEntries)
	le.Add(fmt.Errorf("foo"))
	le.Add(fmt.Errorf("bar"))
	le.Add(fmt.Errorf("baz"))

	m := make(map[string]any)
	m["beep"] = "boop"
	m["this"] = "that"
	m["nums"] = 123
	le.AddWithContext(fmt.Errorf("qux"), m)

	want := 4
	got := len(le.Entries)
	if got != want {
		t.Fatalf("want length %d, got: %d", want, got)
	}
}

func TestLogPersist(t *testing.T) {
	var path string

	// Create temp environment to run test code within.
	{
		wd, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}

		rootdir := testutil.NewEnv(testutil.EnvOpts{
			T: t,
			Write: []testutil.FileIO{
				{Src: string(""), Dst: "errors.log"},
			},
			Copy: []testutil.FileIO{
				{
					Src: filepath.Join("testdata", "errors-expected.log"),
					Dst: filepath.Join("errors-expected.log"),
				},
			},
		})
		path = filepath.Join(rootdir, "errors.log")
		defer os.RemoveAll(rootdir)

		if err := os.Chdir(rootdir); err != nil {
			t.Fatal(err)
		}
		defer os.Chdir(wd)
	}

	errors.Now = func() (t time.Time) { return }

	le := new(errors.LogEntries)
	le.Add(fmt.Errorf("foo"))
	le.Add(fmt.Errorf("bar"))
	le.Add(fmt.Errorf("baz"))

	m := make(map[string]any)
	m["beep"] = "boop"
	m["this"] = "that"
	m["nums"] = 123
	le.AddWithContext(fmt.Errorf("qux"), m)

	err := le.Persist(path, []string{"command", "one", "--example"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = le.Persist(path, []string{"command", "two", "--example"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	have, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	wantPath, err := filepath.Abs("errors-expected.log")
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile(wantPath)
	if err != nil {
		t.Fatal(err)
	}

	r := strings.NewReplacer("\n", "", "\r", "")
	wanttrim := r.Replace(string(want))
	havetrim := r.Replace(string(have))

	testutil.AssertEqual(t, wanttrim, havetrim)
}

// TestLogPersistLogRotation validates that if an audit log file exceeds the
// specified threshold, then the file will be deleted and recreated.
//
// The way this is achieved is by creating an errors.log file that has a
// specific size, and then overriding the package level variable that
// determines the threshold so that it matches the size of the file we created.
// This means we can be sure our logic will trigger the file to be replaced
// with a new empty file, to which we'll then write our log content into.
func TestLogPersistLogRotation(t *testing.T) {
	var (
		fi   os.FileInfo
		path string
	)

	// Create temp environment to run test code within.
	{
		wd, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}

		// We want to start off with an existing audit log file that we expect to
		// be rotated because it exceeded our defined threshold.
		seedPath, err := filepath.Abs(filepath.Join("testdata", "errors-expected.log"))
		if err != nil {
			t.Fatal(err)
		}
		seed, err := os.ReadFile(seedPath)
		if err != nil {
			t.Fatal(err)
		}
		f, err := os.Open(seedPath)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		fi, err = f.Stat()
		if err != nil {
			t.Fatal(err)
		}

		rootdir := testutil.NewEnv(testutil.EnvOpts{
			T: t,
			Write: []testutil.FileIO{
				{Src: string(seed), Dst: "errors.log"},
			},
			Copy: []testutil.FileIO{
				{
					Src: filepath.Join("testdata", "errors-expected-rotation.log"),
					Dst: filepath.Join("errors-expected-rotation.log"),
				},
			},
		})
		path = filepath.Join(rootdir, "errors.log")
		defer os.RemoveAll(rootdir)

		if err := os.Chdir(rootdir); err != nil {
			t.Fatal(err)
		}
		defer os.Chdir(wd)
	}

	errors.Now = func() (t time.Time) { return }
	errors.FileRotationSize = fi.Size()

	le := new(errors.LogEntries)
	le.Add(fmt.Errorf("foo"))
	le.Add(fmt.Errorf("bar"))
	le.Add(fmt.Errorf("baz"))

	m := make(map[string]any)
	m["beep"] = "boop"
	m["this"] = "that"
	m["nums"] = 123
	le.AddWithContext(fmt.Errorf("qux"), m)

	err := le.Persist(path, []string{"command", "one", "--example"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	have, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	wantPath, err := filepath.Abs("errors-expected-rotation.log")
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile(wantPath)
	if err != nil {
		t.Fatal(err)
	}

	r := strings.NewReplacer("\n", "", "\r", "")
	wanttrim := r.Replace(string(want))
	havetrim := r.Replace(string(have))

	testutil.AssertEqual(t, wanttrim, havetrim)
}
