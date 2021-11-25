package compute

import (
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/fastly/cli/pkg/errors"
	fstexec "github.com/fastly/cli/pkg/exec"
	"github.com/fastly/cli/pkg/filesystem"
	"github.com/fastly/cli/pkg/text"
)

// JavaScript implements a Toolchain for the JavaScript language.
type JavaScript struct {
	Shell

	build   string
	timeout int
}

// NewJavaScript constructs a new JavaScript.
func NewJavaScript(timeout int, build string) *JavaScript {
	return &JavaScript{
		Shell:   Shell{},
		build:   build,
		timeout: timeout,
	}
}

// Initialize implements the Toolchain interface and initializes a newly cloned
// package by installing required dependencies.
func (a JavaScript) Initialize(out io.Writer) error {
	// 1) Check `npm` is on $PATH
	//
	// npm is Node/JavaScript's toolchain package manager, it is needed to
	// install the package dependencies on initialization. We only check whether
	// the binary exists on the users $PATH and error with installation help text.
	fmt.Fprintf(out, "Checking if npm is installed...\n")

	p, err := exec.LookPath("npm")
	if err != nil {
		return errors.RemediationError{
			Inner:       fmt.Errorf("`npm` not found in $PATH"),
			Remediation: fmt.Sprintf("To fix this error, install Node.js and npm by visiting:\n\n\t$ %s", text.Bold("https://nodejs.org/")),
		}
	}

	fmt.Fprintf(out, "Found npm at %s\n", p)

	// 2) Check package.json file exists in $PWD
	//
	// A valid npm package manifest file is needed for the install command to
	// work. Therefore, we first assert whether one exists in the current $PWD.
	fpath, err := filepath.Abs("package.json")
	if err != nil {
		return fmt.Errorf("getting package.json path: %w", err)
	}

	if !filesystem.FileExists(fpath) {
		return errors.RemediationError{
			Inner:       fmt.Errorf("package.json not found"),
			Remediation: fmt.Sprintf("To fix this error, run the following command:\n\n\t$ %s", text.Bold("npm init")),
		}
	}

	fmt.Fprintf(out, "Found package.json at %s\n", fpath)
	fmt.Fprintf(out, "Installing package dependencies...\n")

	cmd := fstexec.Streaming{
		Command: "npm",
		Args:    []string{"install"},
		Env:     []string{},
		Output:  out,
	}
	return cmd.Exec()
}

// Verify implements the Toolchain interface and verifies whether the
// JavaScript language toolchain is correctly configured on the host.
func (a JavaScript) Verify(out io.Writer) error {
	// 1) Check `npm` is on $PATH
	//
	// npm is Node/JavaScript's toolchain installer and manager, it is
	// needed to assert that the correct versions of the js-compute-runtime
	// compiler and @fastly/js-compute package are installed. We only check
	// whether the binary exists on the users $PATH and error with installation
	// help text.
	fmt.Fprintf(out, "Checking if npm is installed...\n")

	p, err := exec.LookPath("npm")
	if err != nil {
		return errors.RemediationError{
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
	fpath, err := filepath.Abs("package.json")
	if err != nil {
		return fmt.Errorf("getting package.json path: %w", err)
	}

	if !filesystem.FileExists(fpath) {
		return errors.RemediationError{
			Inner:       fmt.Errorf("package.json not found"),
			Remediation: fmt.Sprintf("To fix this error, run the following command:\n\n\t$ %s", text.Bold("npm init")),
		}
	}

	fmt.Fprintf(out, "Found package.json at %s\n", fpath)

	// 3) Check if `js-compute-runtime` is installed.
	//
	// js-compute-runtime is the JavaScript compiler. We first check if the
	// required dependency exists in the package.json and then whether the
	// js-compute-runtime binary exists in the npm bin directory.
	fmt.Fprintf(out, "Checking if @fastly/js-compute is installed...\n")
	if !checkPackageDependencyExists("@fastly/js-compute") {
		return errors.RemediationError{
			Inner:       fmt.Errorf("`@fastly/js-compute` not found in package.json"),
			Remediation: fmt.Sprintf("To fix this error, run the following command:\n\n\t$ %s", text.Bold("npm install --save-dev @fastly/js-compute")),
		}
	}

	p, err = getNpmBinPath()
	if err != nil {
		return errors.RemediationError{
			Inner:       fmt.Errorf("could not determine npm bin path"),
			Remediation: fmt.Sprintf("To fix this error, run the following command:\n\n\t$ %s", text.Bold("npm install --global npm@latest")),
		}
	}

	path, err := exec.LookPath(filepath.Join(p, "js-compute-runtime"))
	if err != nil {
		return fmt.Errorf("getting js-compute-runtime path: %w", err)
	}
	if !filesystem.FileExists(path) {
		return errors.RemediationError{
			Inner:       fmt.Errorf("`js-compute-runtime` binary not found in %s", p),
			Remediation: fmt.Sprintf("To fix this error, run the following command:\n\n\t$ %s", text.Bold("npm install --save-dev @fastly/js-compute")),
		}
	}

	fmt.Fprintf(out, "Found js-compute-runtime at %s\n", path)

	pkgErr := "package.json requires a `script` field with a `build` step defined that calls the `js-compute-runtime` binary"
	remediation := fmt.Sprintf("Check your package.json has a `script` field with a `build` step defined:\n\n\t$ %s", text.Bold("npm run"))

	cmd := exec.Command("npm", "run")
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return errors.RemediationError{
			Inner:       fmt.Errorf("%s: %w", pkgErr, err),
			Remediation: remediation,
		}
	}

	if !strings.Contains(string(stdoutStderr), "  build\n") {
		return errors.RemediationError{
			Inner:       fmt.Errorf(pkgErr),
			Remediation: remediation,
		}
	}

	return nil
}

// Build implements the Toolchain interface and attempts to compile the package
// JavaScript source to a Wasm binary.
func (a JavaScript) Build(out io.Writer, verbose bool) error {
	cmd := "npm"
	args := []string{"run", "build"}

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
		return err
	}

	return nil
}
