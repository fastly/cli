package compute

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/BurntSushi/toml"
	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

const (
	// RustToolchainVersion is the `rustup` toolchain string for the compiler
	// that we support
	RustToolchainVersion = "1.43.0"
	// WasmWasiTarget is the Rust compilation target for Wasi capable Wasm.
	WasmWasiTarget = "wasm32-wasi"
)

// CargoPackage models the package confuiguration properties of a Rust Cargo
// package which we are interested in and is embedded within CargoManifest.
type CargoPackage struct {
	Name string `toml:"name"`
}

// CargoManifest models the package confuiguration properties of a Rust Cargo
// manifest which we are interested in and are read from the Cargo.toml manifest
// file within the $PWD of the package.
type CargoManifest struct {
	Package CargoPackage
}

// Read the contents of the Cargo.toml manifest from filename.
func (m *CargoManifest) Read(filename string) error {
	_, err := toml.DecodeFile(filename, m)
	return err
}

// Rust is an implments Toolchain for the Rust lanaguage.
type Rust struct{}

// Verify implments the Toolchain interface and verifies whether the Rust
// language toolchain is correctly configured on the host.
func (r Rust) Verify(out io.Writer) error {
	// 1) Check `rustup` is on $PATH
	//
	// Rustup is Rust's toolchain installer and manager, it is needed to assert
	// that the correct WASI WASM compiler target is installed correctly. We
	// only check whether the binary exists on the users $PATH and error with
	// installation help text.

	fmt.Fprintf(out, "Checking if rustup is installed...\n")

	p, err := exec.LookPath("rustup")
	if err != nil {
		return errors.RemediationError{
			Inner:       fmt.Errorf("`rustup` not found in $PATH"),
			Remediation: fmt.Sprintf("To fix this error, run the following command:\n\n\t$ %s", text.Bold("curl https://sh.rustup.rs -sSf | sh")),
		}
	}

	fmt.Fprintf(out, "Found rustup at %s\n", p)

	// 2) Check that the `1.43.0` toolchain is installed
	//
	// We use rustup to assert that the toolchain is installed by streaming the output of
	// `rustup toolchain list` and looking for a toolchain whose prefix matches our desired
	// version.
	fmt.Fprintf(out, "Checking if Rust %s is installed...\n", RustToolchainVersion)

	cmd := exec.Command("rustup", "toolchain", "list")
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error executing rustup: %w", err)
	}

	scanner := bufio.NewScanner(strings.NewReader(string(stdoutStderr)))
	scanner.Split(bufio.ScanLines)
	var found bool
	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), RustToolchainVersion) {
			found = true
			break
		}
	}

	if !found {
		return errors.RemediationError{
			Inner:       fmt.Errorf("rust toolchain %s not found", RustToolchainVersion),
			Remediation: fmt.Sprintf("To fix this error, run the following command:\n\n\t$ %s\n", text.Bold("rustup toolchain install "+RustToolchainVersion)),
		}
	}

	// 3) Check `wasm32-wasi` target exists
	//
	// We use rustup to assert that the target is installed for our toolchain by streaming the
	// output of `rustup target list` and looking for the the `wasm32-wasi` value. If not found,
	// we error with help text suggesting how to install.

	fmt.Fprintf(out, "Checking if %s target is installed...\n", WasmWasiTarget)

	cmd = exec.Command("rustup", "target", "list", "--installed", "--toolchain", RustToolchainVersion)
	stdoutStderr, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error executing rustup: %w", err)
	}

	scanner = bufio.NewScanner(strings.NewReader(string(stdoutStderr)))
	scanner.Split(bufio.ScanWords)
	found = false
	for scanner.Scan() {
		if scanner.Text() == WasmWasiTarget {
			found = true
			break
		}
	}

	if !found {
		return errors.RemediationError{
			Inner:       fmt.Errorf("rust target %s not found", WasmWasiTarget),
			Remediation: fmt.Sprintf("To fix this error, run the following command:\n\n\t$ %s\n", text.Bold(fmt.Sprintf("rustup target add %s --toolchain %s", WasmWasiTarget, RustToolchainVersion))),
		}
	}

	fmt.Fprintf(out, "Found wasm32-wasi target\n")

	// 4) Check Cargo.toml file exists in $PWD
	//
	// A valid Cargo.toml file is needed for the `cargo build` compilation
	// process. Therefore, we assert whether one exists in the current $PWD.

	fpath, err := filepath.Abs("Cargo.toml")
	if err != nil {
		return fmt.Errorf("error getting Cargo.toml path: %w", err)
	}

	if !common.FileExists(fpath) {
		return fmt.Errorf("%s not found", fpath)
	}

	fmt.Fprintf(out, "Found Cargo.toml at %s\n", fpath)

	return nil
}

// Build implments the Toolchain interface and attempts to compile the package
// Rust source to a Wasm binary.
func (r Rust) Build(out io.Writer) error {
	// Get binary name from Cargo.toml.
	var m CargoManifest
	if err := m.Read("Cargo.toml"); err != nil {
		return fmt.Errorf("error reading Cargo.toml manifest: %w", err)
	}
	binName := m.Package.Name

	// Specify the toolchain using the `cargo +<version>` syntax.
	toolchain := fmt.Sprintf("+%s", RustToolchainVersion)

	// Call cargo build with Wasm Wasi target and release flags.
	// gosec flagged this:
	// G204 (CWE-78): Subprocess launched with variable
	// Disabling as the variables come from trusted sources.
	/* #nosec */
	cmd := exec.Command("cargo", toolchain, "build", "--bin", binName, "--release", "--target", WasmWasiTarget)

	// Add debuginfo RUSTFLAGS to command environment to ensure DWARF debug
	// infomation (such as, source mappings) are compiled into the binary.
	cmd.Env = append(os.Environ(),
		`RUSTFLAGS=-C debuginfo=2`,
	)

	// Pipe the child process stdout and stderr to our own writer.
	var stdoutBuf, stderrBuf bytes.Buffer
	stdoutIn, _ := cmd.StdoutPipe()
	stderrIn, _ := cmd.StderrPipe()
	stdout := io.MultiWriter(out, &stdoutBuf)
	stderr := io.MultiWriter(out, &stderrBuf)

	// Start the command.
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start compilation process: %w", err)
	}

	var errStdout, errStderr error
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		_, errStdout = io.Copy(stdout, stdoutIn)
		wg.Done()
	}()

	_, errStderr = io.Copy(stderr, stderrIn)
	wg.Wait()

	// Wait for the command to exit.
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("error during compilation process: %w", err)
	}
	if errStdout != nil {
		return fmt.Errorf("error during compilation process: %w", errStdout)
	}
	if errStderr != nil {
		return fmt.Errorf("error during compilation process: %w", errStderr)
	}

	// Get working directory.
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current working directory: %w", err)
	}
	src := filepath.Join(dir, "target", WasmWasiTarget, "release", "deps", fmt.Sprintf("%s.wasm", binName))
	dst := filepath.Join(dir, "bin", "main.wasm")

	// Check if bin directory exists and create if not.
	binDir := filepath.Join(dir, "bin")
	fi, err := os.Stat(binDir)
	switch {
	case err == nil && fi.IsDir():
		// no problem
	case err == nil && !fi.IsDir():
		return fmt.Errorf("error creating bin directory: target already exists as a regular file")
	case os.IsNotExist(err):
		if err := os.MkdirAll(binDir, 0750); err != nil {
			return err
		}
	case err != nil:
		return err
	}

	err = common.CopyFile(src, dst)
	if err != nil {
		return fmt.Errorf("error copying wasm binary: %w", err)
	}

	return nil
}
