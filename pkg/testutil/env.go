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
	Dirs  []string // expect path to have a trailing slash (will be added if missing)
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

	if err := os.MkdirAll(rootdir, 0o750); err != nil {
		opts.T.Fatal(err)
	}

	for _, d := range opts.Dirs {
		d = strings.TrimRight(d, "/") + "/filename-required.txt"
		createIntermediaryDirectories(d, rootdir, opts.T)
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
		createIntermediaryDirectories(f.Dst, rootdir, opts.T)

		if err := os.WriteFile(dst, []byte(src), 0o777); err != nil /* #nosec */ {
			opts.T.Fatal(err)
		}
	}

	if len(opts.Exec) > 0 {
		// gosec flagged this:
		// G204 (CWE-78): Subprocess launched with function call as argument or cmd arguments
		// Disabling as we trust the source of the variable.
		// #nosec
		// nosemgrep: go.lang.security.audit.dangerous-exec-command.dangerous-exec-command
		cmd := exec.Command(opts.Exec[0], opts.Exec[1:]...)
		cmd.Dir = rootdir
		if err := cmd.Run(); err != nil {
			opts.T.Fatal(err)
		}
	}

	return rootdir
}

// createIntermediaryDirectories strips the filename from the given path and
// appends it to the rootdir so that we can use MkdirAll to create the
// directory and all its intermediary directories.
//
// EXAMPLE: /foo/bar/baz.txt will create the foo and bar directories if they
// don't already exist.
//
// NOTE: If path is just a filename (e.g. config.toml), then this function
// won't necessarily trigger a test failure because we would end up appending
// an empty string to the rootdir and so the MkdirAll call still succeeds.
func createIntermediaryDirectories(path, rootdir string, t *testing.T) {
	intermediary := strings.Replace(path, filepath.Base(path), "", 1)
	intermediary = filepath.Join(rootdir, intermediary)
	if err := os.MkdirAll(intermediary, 0o750); err != nil {
		t.Fatal(err)
	}
}
