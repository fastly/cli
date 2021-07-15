package errors_test

import (
	"fmt"
	"os"
	"path/filepath"
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

	want := 3
	got := len(*le)
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

	if string(have) != string(want) {
		t.Fatalf("wanted content:\n%s\n, got:\n%s\n", string(want), string(have))
	}
}
