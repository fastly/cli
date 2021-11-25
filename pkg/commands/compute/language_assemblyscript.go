package compute

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	fsterr "github.com/fastly/cli/pkg/errors"
	fstexec "github.com/fastly/cli/pkg/exec"
	"github.com/fastly/cli/pkg/filesystem"
	"github.com/fastly/cli/pkg/text"
)

// AssemblyScript implements a Toolchain for the AssemblyScript language.
type AssemblyScript struct {
	Shell

	build   string
	errlog  fsterr.LogInterface
	timeout int
}

// NewAssemblyScript constructs a new AssemblyScript.
func NewAssemblyScript(timeout int, build string, errlog fsterr.LogInterface) *AssemblyScript {
	return &AssemblyScript{
		Shell:   Shell{},
		build:   build,
		errlog:  errlog,
		timeout: timeout,
	}
}

// Verify implements the Toolchain interface and verifies whether the
// AssemblyScript language toolchain is correctly configured on the host.
func (a AssemblyScript) Verify(out io.Writer) error {
	// 1) Check `npm` is on $PATH
	//
	// npm is Node/AssemblyScript's toolchain installer and manager, it is
	// needed to assert that the correct versions of the asc compiler and
	// @fastly/as-compute package are installed. We only check whether the
	// binary exists on the users $PATH and error with installation help text.
	fmt.Fprintf(out, "Checking if npm is installed...\n")

	p, err := exec.LookPath("npm")
	if err != nil {
		a.errlog.Add(err)
		return fsterr.RemediationError{
			Inner:       fmt.Errorf("`npm` not found in $PATH"),
			Remediation: fmt.Sprintf("To fix this error, install Node.js and npm by visiting:\n\n\t$ %s", text.Bold("https://nodejs.org/")),
		}
	}

	fmt.Fprintf(out, "Found npm at %s\n", p)

	// 2) Check package.json file exists in $PWD
	//
	// A valid npm package is needed for compilation and to assert whether the
	// required dependencies are installed locally. Therefore, we first assert
	// whether one exists in the current $PWD.
	pkg, err := filepath.Abs("package.json")
	if err != nil {
		a.errlog.Add(err)
		return fmt.Errorf("getting package.json path: %w", err)
	}

	if !filesystem.FileExists(pkg) {
		err = fsterr.RemediationError{
			Inner:       fmt.Errorf("package.json not found"),
			Remediation: fmt.Sprintf("To fix this error, run the following command:\n\n\t$ %s", text.Bold("npm init")),
		}
		a.errlog.Add(err)
		return err
	}

	fmt.Fprintf(out, "Found package.json at %s\n", pkg)

	// 3) Check if `asc` is installed.
	//
	// asc is the AssemblyScript compiler. We first check if it exists in the
	// package.json and then whether the binary exists in the npm bin directory.
	fmt.Fprintf(out, "Checking if AssemblyScript is installed...\n")
	if !checkPackageDependencyExists("assemblyscript") {
		err = fsterr.RemediationError{
			Inner:       fmt.Errorf("`assemblyscript` not found in package.json"),
			Remediation: fmt.Sprintf("To fix this error, run the following command:\n\n\t$ %s", text.Bold("npm install --save-dev assemblyscript")),
		}
		a.errlog.Add(err)
		return err
	}

	p, err = getNpmBinPath()
	if err != nil {
		a.errlog.Add(err)
		return fsterr.RemediationError{
			Inner:       fmt.Errorf("could not determine npm bin path"),
			Remediation: fmt.Sprintf("To fix this error, run the following command:\n\n\t$ %s", text.Bold("npm install --global npm@latest")),
		}
	}

	path, err := exec.LookPath(filepath.Join(p, "asc"))
	if err != nil {
		a.errlog.Add(err)
		return fmt.Errorf("getting asc path: %w", err)
	}
	if !filesystem.FileExists(path) {
		err = fsterr.RemediationError{
			Inner:       fmt.Errorf("`asc` binary not found in %s", p),
			Remediation: fmt.Sprintf("To fix this error, run the following command:\n\n\t$ %s", text.Bold("npm install --save-dev assemblyscript")),
		}
		a.errlog.Add(err)
		return err
	}

	fmt.Fprintf(out, "Found asc at %s\n", path)

	return nil
}

// Initialize implements the Toolchain interface and initializes a newly cloned
// package by installing required dependencies.
func (a AssemblyScript) Initialize(out io.Writer) error {
	// 1) Check `npm` is on $PATH
	//
	// npm is Node/AssemblyScript's toolchain package manager, it is needed to
	// install the package dependencies on initialization. We only check whether
	// the binary exists on the users $PATH and error with installation help text.
	fmt.Fprintf(out, "Checking if npm is installed...\n")

	p, err := exec.LookPath("npm")
	if err != nil {
		err = fsterr.RemediationError{
			Inner:       fmt.Errorf("`npm` not found in $PATH"),
			Remediation: fmt.Sprintf("To fix this error, install Node.js and npm by visiting:\n\n\t$ %s", text.Bold("https://nodejs.org/")),
		}
		a.errlog.Add(err)
		return err
	}

	fmt.Fprintf(out, "Found npm at %s\n", p)

	// 2) Check package.json file exists in $PWD
	//
	// A valid npm package manifest file is needed for the install command to
	// work. Therefore, we first assert whether one exists in the current $PWD.
	pkg, err := filepath.Abs("package.json")
	if err != nil {
		a.errlog.Add(err)
		return fmt.Errorf("getting package.json path: %w", err)
	}

	if !filesystem.FileExists(pkg) {
		err = fsterr.RemediationError{
			Inner:       fmt.Errorf("package.json not found"),
			Remediation: fmt.Sprintf("To fix this error, run the following command:\n\n\t$ %s", text.Bold("npm init")),
		}
		a.errlog.Add(err)
		return err
	}

	fmt.Fprintf(out, "Found package.json at %s\n", pkg)
	fmt.Fprintf(out, "Installing package dependencies...\n")

	cmd := fstexec.Streaming{
		Command: "npm",
		Args:    []string{"install"},
		Env:     []string{},
		Output:  out,
	}
	if err := cmd.Exec(); err != nil {
		a.errlog.Add(err)
	}
	return nil
}

// Build implements the Toolchain interface and attempts to compile the package
// AssemblyScript source to a Wasm binary.
func (a AssemblyScript) Build(out io.Writer, verbose bool) error {
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

	npmdir, err := getNpmBinPath()
	if err != nil {
		a.errlog.Add(err)
		return fmt.Errorf("getting npm path: %w", err)
	}

	cmd := filepath.Join(npmdir, "asc")
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

	if a.build != "" {
		cmd, args = a.Shell.Build(a.build)
	}

	s := fstexec.Streaming{
		Command: cmd,
		Args:    args,
		Env:     []string{},
		Output:  out,
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
