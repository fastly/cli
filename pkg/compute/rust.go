package compute

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/BurntSushi/toml"
	"github.com/blang/semver"
	"github.com/fastly/cli/pkg/api"
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
// package which we are interested in and is embedded within CargoManifest and
// CargoLock.
type CargoPackage struct {
	Name    string `toml:"name"`
	Version string `toml:"version"`
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

// CargoLock models the package confuiguration properties of a Rust Cargo
// lock file which we are interested in and are read from the Cargo.lock file
// within the $PWD of the package.
type CargoLock struct {
	Package []CargoPackage
}

// Read the contents of the Cargo.lock file from filename.
func (m *CargoLock) Read(filename string) error {
	_, err := toml.DecodeFile(filename, m)
	return err
}

// Rust is an implments Toolchain for the Rust lanaguage.
type Rust struct {
	client api.HTTPClient
}

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

	// 5) Verify `fastly` crate version
	//
	// A valid and up-to-date version of the fastly crate is required.
	fpath, err = filepath.Abs("Cargo.lock")
	if err != nil {
		return fmt.Errorf("error getting Cargo.lock path: %w", err)
	}

	if !common.FileExists(fpath) {
		return fmt.Errorf("%s not found", fpath)
	}

	var lock CargoLock
	if err := lock.Read("Cargo.lock"); err != nil {
		return fmt.Errorf("error reading Cargo.lock file: %w", err)
	}

	var crate CargoPackage
	for _, p := range lock.Package {
		if p.Name == "fastly" {
			crate = p
		}
	}

	version, err := semver.Parse(crate.Version)
	if err != nil {
		return fmt.Errorf("error parsing Cargo.lock file: %w", err)
	}

	l, err := GetLatestCrateVersion(r.client, "fastly")
	if err != nil {
		return err
	}
	latest, err := semver.Parse(l)
	if err != nil {
		return err
	}

	if version.LT(latest) {
		return errors.RemediationError{
			Inner:       fmt.Errorf("fastly crate not up-to-date"),
			Remediation: fmt.Sprintf("To fix this error, run the following command:\n\n\t$ %s\n", text.Bold("cargo update -p fastly")),
		}
	}

	return nil
}

// Build implments the Toolchain interface and attempts to compile the package
// Rust source to a Wasm binary.
func (r Rust) Build(out io.Writer, verbose bool) error {
	// Get binary name from Cargo.toml.
	var m CargoManifest
	if err := m.Read("Cargo.toml"); err != nil {
		return fmt.Errorf("error reading Cargo.toml manifest: %w", err)
	}
	binName := m.Package.Name

	// Specify the toolchain using the `cargo +<version>` syntax.
	toolchain := fmt.Sprintf("+%s", RustToolchainVersion)

	args := []string{
		toolchain,
		"build",
		"--bin",
		binName,
		"--release",
		"--target",
		WasmWasiTarget,
		"--color",
		"always",
	}
	if verbose {
		args = append(args, "--verbose")
	}

	// Call cargo build with Wasm Wasi target and release flags.
	// gosec flagged this:
	// G204 (CWE-78): Subprocess launched with variable
	// Disabling as the variables come from trusted sources.
	/* #nosec */
	cmd := exec.Command("cargo", args...)

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

	if errStdout != nil {
		return fmt.Errorf("error streaming stdout output from child process: %w", errStdout)
	}
	if errStderr != nil {
		return fmt.Errorf("error streaming stderr output from child process: %w", errStderr)
	}

	// Wait for the command to exit.
	if err := cmd.Wait(); err != nil {
		// If we're not in verbose mode return the bufferred stderr output
		// from cargo as the error.
		if !verbose && stderrBuf.Len() > 0 {
			return fmt.Errorf("error during compilation process:\n%s", strings.TrimSpace(stderrBuf.String()))
		}
		return fmt.Errorf("error during compilation process")
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

// CargoCrateVersion models a Cargo crate version returned by the crates.io API.
type CargoCrateVersion struct {
	Version string `json:"num"`
}

// CargoCrateVersions models a Cargo crate version returned by the crates.io API.
type CargoCrateVersions struct {
	Versions []CargoCrateVersion `json:"versions"`
}

// GetLatestCrateVersion fetches all versions of a given Rust crate from the
// crates.io HTTP API and returns the latest valid semver version.
func GetLatestCrateVersion(client api.HTTPClient, name string) (string, error) {
	url := fmt.Sprintf("https://crates.io/api/v1/crates/%s/versions", name)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error fetching latest crate version %s", resp.Status)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	crate := CargoCrateVersions{}
	err = json.Unmarshal(body, &crate)
	if err != nil {
		return "", err
	}

	var versions []semver.Version
	for _, v := range crate.Versions {
		if version, err := semver.Parse(v.Version); err == nil {
			versions = append(versions, version)
		}
	}

	if len(versions) < 1 {
		return "", fmt.Errorf("no valid crate versions found")
	}

	semver.Sort(versions)

	latest := versions[len(versions)-1]

	return latest.String(), nil
}
