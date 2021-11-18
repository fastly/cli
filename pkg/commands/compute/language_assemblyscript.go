package compute

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/fastly/cli/pkg/errors"
	fstexec "github.com/fastly/cli/pkg/exec"
	"github.com/fastly/cli/pkg/filesystem"
	"github.com/fastly/cli/pkg/text"
)

// AssemblyScript implements a Toolchain for the AssemblyScript language.
type AssemblyScript struct {
	timeout   int
	toolchain string
}

// NewAssemblyScript constructs a new AssemblyScript.
func NewAssemblyScript(timeout int, toolchain string) *AssemblyScript {
	return &AssemblyScript{
		timeout:   timeout,
		toolchain: toolchain,
	}
}

// Initialize implements the Toolchain interface and initializes a newly cloned
// package by installing required dependencies.
func (a AssemblyScript) Initialize(out io.Writer) error {
	// 1) Check a.toolchain is on $PATH
	//
	// npm and yarn, two popular Node/JavaScript toolchain installers/managers,
	// is needed to install the package dependencies on initialization. We only
	// check whether the binary exists on the users $PATH and error with
	// installation help text.
	fmt.Fprintf(out, "Checking if %s is installed...\n", a.toolchain)

	p, err := exec.LookPath(a.toolchain)
	if err != nil {
		nodejsURL := "https://nodejs.org/"
		var remediation string
		switch a.toolchain {
		case "npm":
			remediation = fmt.Sprintf("To fix this error, install Node.js and %s by visiting:\n\n\t$ %s", a.toolchain, text.Bold(nodejsURL))
		case "yarn":
			remediation = fmt.Sprintf("To fix this error, install Node.js by visiting %s, and install %s by visiting:\n\n\t$ %s", text.Bold(nodejsURL), a.toolchain, text.Bold("https://yarnpkg.com/"))
		}

		return errors.RemediationError{
			Inner:       fmt.Errorf("`%s` not found in $PATH", a.toolchain),
			Remediation: remediation,
		}
	}

	fmt.Fprintf(out, "Found %s at %s\n", a.toolchain, p)

	// 2) Check package.json file exists in $PWD
	//
	// A valid package manifest file is needed for the install command to work.
	// Therefore, we first assert whether one exists in the current $PWD.
	fpath, err := filepath.Abs("package.json")
	if err != nil {
		return fmt.Errorf("getting package.json path: %w", err)
	}

	if !filesystem.FileExists(fpath) {
		var remediation string
		switch a.toolchain {
		case "npm":
			remediation = "npm init"
		case "yarn":
			remediation = "yarn init"
		}

		return errors.RemediationError{
			Inner:       fmt.Errorf("package.json not found"),
			Remediation: fmt.Sprintf(errFormat, text.Bold(remediation)),
		}
	}

	fmt.Fprintf(out, "Found package.json at %s\n", fpath)
	fmt.Fprintf(out, "Installing package dependencies...\n")

	cmd := fstexec.Streaming{
		Command: a.toolchain,
		Args:    []string{"install"},
		Env:     []string{},
		Output:  out,
	}
	return cmd.Exec()
}

// Verify implements the Toolchain interface and verifies whether the
// AssemblyScript language toolchain is correctly configured on the host.
func (a AssemblyScript) Verify(out io.Writer) error {
	// 1) Check a.toolchain is on $PATH
	//
	// npm and yarn, two popular Node/JavaScript toolchain installers/managers,
	// which is needed to assert that the correct versions of the
	// js-compute-runtime compiler and @fastly/js-compute package are installed.
	// We only check whether the binary exists on the users $PATH and error with
	// installation help text.
	fmt.Fprintf(out, "Checking if %s is installed...\n", a.toolchain)

	p, err := exec.LookPath(a.toolchain)
	if err != nil {
		nodejsURL := "https://nodejs.org/"
		var remediation string
		switch a.toolchain {
		case "npm":
			remediation = fmt.Sprintf("To fix this error, install Node.js and %s by visiting:\n\n\t$ %s", a.toolchain, text.Bold(nodejsURL))
		case "yarn":
			remediation = fmt.Sprintf("To fix this error, install Node.js by visiting %s and %s by visiting:\n\n\t$ %s", text.Bold(nodejsURL), a.toolchain, text.Bold("https://yarnpkg.com/"))
		}

		return errors.RemediationError{
			Inner:       fmt.Errorf("`%s` not found in $PATH", a.toolchain),
			Remediation: remediation,
		}
	}

	fmt.Fprintf(out, "Found %s at %s\n", a.toolchain, p)

	// 2) Check package.json file exists in $PWD
	//
	// A valid package is needed for compilation and to assert whether the
	// required dependencies are installed locally. Therefore, we first assert
	// whether one exists in the current $PWD.
	fpath, err := filepath.Abs("package.json")
	if err != nil {
		return fmt.Errorf("getting package.json path: %w", err)
	}

	if !filesystem.FileExists(fpath) {
		var remediation string
		switch a.toolchain {
		case "npm":
			remediation = "npm init"
		case "yarn":
			remediation = "yarn init"
		}
		return errors.RemediationError{
			Inner:       fmt.Errorf("package.json not found"),
			Remediation: fmt.Sprintf(errFormat, text.Bold(remediation)),
		}
	}

	fmt.Fprintf(out, "Found package.json at %s\n", fpath)

	// 3) Check if `asc` is installed.
	//
	// asc is the AssemblyScript compiler. We first check if it exists in the
	// package.json and then whether the binary exists in the toolchain bin directory.
	fmt.Fprintf(out, "Checking if AssemblyScript is installed...\n")
	if !checkJsPackageDependencyExists(a.toolchain, "assemblyscript") {
		var remediation string
		switch a.toolchain {
		case "npm":
			remediation = "npm install --save-dev assemblyscript"
		case "yarn":
			remediation = "yarn install --save-dev assemblyscript"
		}
		return errors.RemediationError{
			Inner:       fmt.Errorf("`assemblyscript` not found in package.json"),
			Remediation: fmt.Sprintf(errFormat, text.Bold(remediation)),
		}
	}

	p, err = getJsToolchainBinPath(a.toolchain)
	if err != nil {
		var remediation string
		switch a.toolchain {
		case "npm":
			remediation = "npm install --global npm@latest"
		case "yarn":
			remediation = "yarn install --global yarn@latest"
		}
		return errors.RemediationError{
			Inner:       fmt.Errorf("could not determine %s bin path", a.toolchain),
			Remediation: fmt.Sprintf(errFormat, text.Bold(remediation)),
		}
	}

	path, err := exec.LookPath(filepath.Join(p, "asc"))
	if err != nil {
		return fmt.Errorf("getting asc path: %w", err)
	}
	if !filesystem.FileExists(path) {
		var remediation string
		switch a.toolchain {
		case "npm":
			remediation = "npm install --save-dev assemblyscript"
		case "yarn":
			remediation = "yarn install --save-dev assemblyscript"
		}
		return errors.RemediationError{
			Inner:       fmt.Errorf("`asc` binary not found in %s", p),
			Remediation: fmt.Sprintf(errFormat, text.Bold(remediation)),
		}
	}

	fmt.Fprintf(out, "Found asc at %s\n", path)

	return nil
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

// Toolchain updates the toolchain used at runtime based on a user's
// prompt input.
func (a *AssemblyScript) Toolchain(toolchain string) {
	a.toolchain = toolchain
}
