package testutil

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// FileIO represents a source file and a destination.
type FileIO struct {
	Src string // path to a file inside ./testdata/ OR file content
	Dst string // path to a file relative to test environment's root directory
}

// EnvOpts represents configuration when creating a new environment.
type EnvOpts struct {
	T     *testing.T
	Copy  []FileIO // .Src expected to be file path
	Write []FileIO // .Src expected to be file content
	Exec  []string // e.g. []string{"npm", "install"}
}

// NewEnv creates a new test environment and returns the root directory.
func NewEnv(opts EnvOpts) (rootdir string) {
	rootdir, err := os.MkdirTemp("", "fastly-temp-*")
	if err != nil {
		opts.T.Fatal(err)
	}

	if err := os.MkdirAll(rootdir, 0750); err != nil {
		opts.T.Fatal(err)
	}

	for _, f := range opts.Copy {
		src := f.Src
		dst := filepath.Join(rootdir, f.Dst)
		CopyFile(opts.T, src, dst)
	}

	for _, f := range opts.Write {
		if f.Src == "" {
			continue
		}
		src := f.Src
		dst := filepath.Join(rootdir, f.Dst)

		// Ensure any intermediary directories exist before trying to write the
		// given file to disk.
		intermediary := strings.Replace(f.Dst, filepath.Base(f.Dst), "", 1)
		intermediary = filepath.Join(rootdir, intermediary)
		if err := os.MkdirAll(intermediary, 0750); err != nil {
			opts.T.Fatal(err)
		}

		if err := os.WriteFile(dst, []byte(src), 0777); err != nil {
			opts.T.Fatal(err)
		}
	}

	if len(opts.Exec) > 0 {
		// gosec flagged this:
		// G204 (CWE-78): Subprocess launched with function call as argument or cmd arguments
		// Disabling as we trust the source of the variable.
		/* #nosec */
		cmd := exec.Command(opts.Exec[0], opts.Exec[1:]...)
		cmd.Dir = rootdir
		if err := cmd.Run(); err != nil {
			opts.T.Fatal(err)
		}
	}

	return rootdir
}
