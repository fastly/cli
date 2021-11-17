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

var errFormat = "To fix this error, run the following command:\n\n\t$ %s"

// JavaScript implements a Toolchain for the JavaScript language.
type JavaScript struct {
	timeout   int
	toolchain string
}

// NewJavaScript constructs a new JavaScript.
func NewJavaScript(timeout int, toolchain string) *JavaScript {
	return &JavaScript{
		timeout:   timeout,
		toolchain: toolchain,
	}
}

// Initialize implements the Toolchain interface and initializes a newly cloned
// package by installing required dependencies.
func (a JavaScript) Initialize(out io.Writer) error {
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
// JavaScript language toolchain is correctly configured on the host.
func (a JavaScript) Verify(out io.Writer) error {
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

	// 3) Check if `js-compute-runtime` is installed.
	//
	// js-compute-runtime is the JavaScript compiler. We first check if the
	// required dependency exists in the package.json and then whether the
	// js-compute-runtime binary exists in the toolchain bin directory.
	fmt.Fprintf(out, "Checking if @fastly/js-compute is installed...\n")
	if !checkJsPackageDependencyExists(a.toolchain, "@fastly/js-compute") {
		var remediation string
		switch a.toolchain {
		case "npm":
			remediation = "npm install --save-dev @fastly/js-compute"
		case "yarn":
			remediation = "yarn install --save-dev @fastly/js-compute"
		}
		return errors.RemediationError{
			Inner:       fmt.Errorf("`@fastly/js-compute` not found in package.json"),
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

	path, err := exec.LookPath(filepath.Join(p, "js-compute-runtime"))
	if err != nil {
		return fmt.Errorf("getting js-compute-runtime path: %w", err)
	}
	if !filesystem.FileExists(path) {
		var remediation string
		switch a.toolchain {
		case "npm":
			remediation = "npm install --save-dev @fastly/js-compute"
		case "yarn":
			remediation = "yarn install --save-dev @fastly/js-compute"
		}
		return errors.RemediationError{
			Inner:       fmt.Errorf("`js-compute-runtime` binary not found in %s", p),
			Remediation: fmt.Sprintf(errFormat, text.Bold(remediation)),
		}
	}

	fmt.Fprintf(out, "Found js-compute-runtime at %s\n", path)

	var remediation string
	switch a.toolchain {
	case "npm":
		remediation = "npm run"
	case "yarn":
		remediation = "yarn run"
	}

	pkgErr := "package.json requires a `script` field with a `build` step defined that calls the `js-compute-runtime` binary"
	remediation = fmt.Sprintf("Check your package.json has a `script` field with a `build` step defined:\n\n\t$ %s", text.Bold(remediation))

	// gosec flagged this:
	// G204 (CWE-78): Subprocess launched with variable
	// Disabling as the variables come from trusted sources:
	// The CLI parser enforces supported values via EnumVar.
	/* #nosec */
	cmd := exec.Command(a.toolchain, "run")
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return errors.RemediationError{
			Inner:       fmt.Errorf("%s: %w", pkgErr, err),
			Remediation: remediation,
		}
	}

	if !strings.Contains(string(stdoutStderr), " build\n") {
		return errors.RemediationError{
			Inner:       fmt.Errorf("%s:\n\n%s", pkgErr, stdoutStderr),
			Remediation: remediation,
		}
	}

	return nil
}

// Build implements the Toolchain interface and attempts to compile the package
// JavaScript source to a Wasm binary.
func (a JavaScript) Build(out io.Writer, verbose bool) error {
	cmd := fstexec.Streaming{
		Command: a.toolchain,
		Args:    []string{"run", "build"},
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
