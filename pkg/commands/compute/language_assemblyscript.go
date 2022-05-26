package compute

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	fsterr "github.com/fastly/cli/pkg/errors"
	fstexec "github.com/fastly/cli/pkg/exec"
	"github.com/fastly/cli/pkg/filesystem"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
)

// ASSourceDirectory represents the source code directory.
const ASSourceDirectory = "assembly"

// AssemblyScript implements a Toolchain for the AssemblyScript language.
//
// NOTE: We embed the JavaScript type as the behaviours across both languages
// are fundamentally the same with some minor differences. This means we don't
// need to duplicate the Verify() implementation, while the Build() method can
// be kept unique between the two languages. Additionally the JavaScript
// Verify() method has an extra validation step that is skipped for
// AssemblyScript as it doesn't set the `validateScriptBuild` field.
type AssemblyScript struct {
	JavaScript

	build     string
	errlog    fsterr.LogInterface
	postBuild string
}

// NewAssemblyScript constructs a new AssemblyScript toolchain.
func NewAssemblyScript(pkgName string, scripts manifest.Scripts, errlog fsterr.LogInterface, timeout int) *AssemblyScript {
	return &AssemblyScript{
		JavaScript: JavaScript{
			build:             scripts.Build,
			errlog:            errlog,
			packageDependency: "assemblyscript",
			packageExecutable: "asc",
			pkgName:           pkgName,
			timeout:           timeout,
			toolchain:         JsToolchain,
		},
		build:     scripts.Build,
		errlog:    errlog,
		postBuild: scripts.PostBuild,
	}
}

// Build implements the Toolchain interface and attempts to compile the package
// AssemblyScript source to a Wasm binary.
func (a AssemblyScript) Build(out io.Writer, progress text.Progress, verbose bool, callback func() error) error {
	// Check if bin directory exists and create if not.
	pwd, err := os.Getwd()
	if err != nil {
		a.errlog.Add(err)
		return fmt.Errorf("getting current working directory: %w", err)
	}
	binDir := filepath.Join(pwd, "bin")
	if err := filesystem.MakeDirectoryIfNotExists(binDir); err != nil {
		a.errlog.Add(err)
		return fmt.Errorf("making bin directory: %w", err)
	}

	toolchaindir, err := getJsToolchainBinPath(a.toolchain)
	if err != nil {
		a.errlog.Add(err)
		return fmt.Errorf("getting npm path: %w", err)
	}

	cmd := filepath.Join(toolchaindir, "asc")
	args := []string{
		"assembly/index.ts",
		"--binaryFile",
		filepath.Join(binDir, "main.wasm"),
		"--optimize",
		"--noAssert",
	}

	if a.build != "" {
		cmd, args = a.Shell.Build(a.build)
	}

	err = a.execCommand(cmd, args, out, progress, verbose)
	if err != nil {
		return err
	}

	// NOTE: We set the progress indicator to Done() so that any output we now
	// print via the post_build callback doesn't get hidden by the progress status.
	// The progress is 'reset' inside the main build controller `build.go`.
	progress.Done()

	if a.postBuild != "" {
		if err = callback(); err == nil {
			cmd, args := a.Shell.Build(a.postBuild)
			err := a.execCommand(cmd, args, out, progress, verbose)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (a AssemblyScript) execCommand(cmd string, args []string, out, progress io.Writer, verbose bool) error {
	s := fstexec.Streaming{
		Command:  cmd,
		Args:     args,
		Env:      os.Environ(),
		Output:   out,
		Progress: progress,
		Verbose:  verbose,
	}
	if a.timeout > 0 {
		s.Timeout = time.Duration(a.timeout) * time.Second
	}
	if err := s.Exec(); err != nil {
		a.errlog.Add(err)
		return err
	}
	return nil
}
