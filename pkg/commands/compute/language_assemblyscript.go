package compute

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	fstexec "github.com/fastly/cli/pkg/exec"
	"github.com/fastly/cli/pkg/filesystem"
)

// AssemblyScript implements a Toolchain for the AssemblyScript language.
//
// NOTE: We embed the JavaScript type as the behaviours across both languages
// are fundamentally the same with some minor differences. This means we don't
// need to duplicate the Verify() or Toolchain() implementations, while the
// Build() methods can be kept unique between the two languages. Additionally
// the JavaScript Verify() method has an extra validation step not needed by
// AssemblyScript and that step is skipped for AssemblyScript as it doesn't set
// the `validateScriptBuild` field.
type AssemblyScript struct {
	JavaScript
}

// NewAssemblyScript constructs a new AssemblyScript.
func NewAssemblyScript(timeout int, toolchain string) *AssemblyScript {
	return &AssemblyScript{
		JavaScript{
			packageDependency: "assemblyscript",
			packageExecutable: "asc",
			timeout:           timeout,
			toolchain:         toolchain,
		},
	}
}

// Build implements the Toolchain interface and attempts to compile the package
// AssemblyScript source to a Wasm binary.
func (a AssemblyScript) Build(out io.Writer, verbose bool) error {
	// Check if bin directory exists and create if not.
	pwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting current working directory: %w", err)
	}
	binDir := filepath.Join(pwd, "bin")
	if err := filesystem.MakeDirectoryIfNotExists(binDir); err != nil {
		return fmt.Errorf("making bin directory: %w", err)
	}

	toolchaindir, err := getJsToolchainBinPath(a.toolchain)
	if err != nil {
		return fmt.Errorf("getting npm path: %w", err)
	}

	args := []string{
		"assembly/index.ts",
		"--binaryFile",
		filepath.Join(binDir, "main.wasm"),
		"--optimize",
		"--noAssert",
	}
	if verbose {
		args = append(args, "--verbose")
	}

	cmd := fstexec.Streaming{
		Command: filepath.Join(toolchaindir, "asc"),
		Args:    args,
		Env:     []string{},
		Output:  out,
	}
	if a.timeout > 0 {
		cmd.Timeout = time.Duration(a.timeout) * time.Second
	}
	if err := cmd.Exec(); err != nil {
		return err
	}

	return nil
}
