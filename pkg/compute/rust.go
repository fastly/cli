package compute

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	fstexec "github.com/fastly/cli/pkg/exec"
	"github.com/fastly/cli/pkg/filesystem"
	"github.com/fastly/cli/pkg/text"
	toml "github.com/pelletier/go-toml"
)

// CargoPackage models the package configuration properties of a Rust Cargo
// package which we are interested in and is embedded within CargoManifest and
// CargoLock.
type CargoPackage struct {
	Name         string         `toml:"name" json:"name"`
	Version      string         `toml:"version" json:"version"`
	Dependencies []CargoPackage `toml:"-" json:"dependencies"`
}

// CargoManifest models the package configuration properties of a Rust Cargo
// manifest which we are interested in and are read from the Cargo.toml manifest
// file within the $PWD of the package.
type CargoManifest struct {
	Package CargoPackage
}

// Read the contents of the Cargo.toml manifest from filename.
func (m *CargoManifest) Read(fpath string) error {
	// gosec flagged this:
	// G304 (CWE-22): Potential file inclusion via variable.
	// Disabling as we need to load the Cargo.toml from the user's file system.
	// This file is decoded into a predefined struct, any unrecognised fields are dropped.
	/* #nosec */
	bs, err := os.ReadFile(fpath)
	if err != nil {
		return err
	}
	err = toml.Unmarshal(bs, m)
	return err
}

// CargoMetadata models information about the workspace members and resolved
// dependencies of the current package via `cargo metadata` command output.
type CargoMetadata struct {
	Package         []CargoPackage `json:"packages"`
	TargetDirectory string         `json:"target_directory"`
}

// Read the contents of the Cargo.lock file from filename.
func (m *CargoMetadata) Read() error {
	cmd := exec.Command("cargo", "metadata", "--quiet", "--format-version", "1")
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		if len(stdoutStderr) > 0 {
			return fmt.Errorf("%s", strings.TrimSpace(string(stdoutStderr)))
		}
		return err
	}
	r := bytes.NewReader(stdoutStderr)
	if err := json.NewDecoder(r).Decode(&m); err != nil {
		return err
	}
	return nil
}

// Rust implements a Toolchain for the Rust language.
type Rust struct {
	client    api.HTTPClient
	config    *config.Data
	toolchain *semver.Version
}

// NewRust constructs a new Rust.
func NewRust(client api.HTTPClient, config *config.Data) *Rust {
	return &Rust{
		client: client,
		config: config,
	}
}

// SourceDirectory implements the Toolchain interface and returns the source
// directory for Rust packages.
func (r Rust) SourceDirectory() string { return "src" }

// IncludeFiles implements the Toolchain interface and returns a list of
// additional files to include in the package archive for Rust packages.
func (r Rust) IncludeFiles() []string { return []string{"Cargo.toml"} }

// Verify implments the Toolchain interface and verifies whether the Rust
// language toolchain is correctly configured on the host.
func (r *Rust) Verify(out io.Writer) error {
	// 1) Check `rustup` is on $PATH
	//
	// Rustup is Rust's toolchain installer and manager, it is needed to assert
	// that the correct WASI Wasm compiler target is installed correctly. We
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

	// 2) Check that the desired rustup version is installed
	fmt.Fprintf(out, "Checking if rustup %s is installed...\n", r.config.File.Language.Rust.RustupConstraint)

	cmd := exec.Command("rustup", "--version")
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error executing rustup: %w", err)
	}

	reader := bufio.NewReader(bytes.NewReader(stdoutStderr))
	line, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("error reading rustup output: %w", err)
	}
	parts := strings.Split(line, " ")
	// Either `rustup <version> (<date>)` or `rustup <version> (<sha> <date>)`
	if len(parts) != 3 && len(parts) != 4 {
		return fmt.Errorf("error reading rustup version")
	}
	rustupVersion, err := semver.NewVersion(parts[1])
	if err != nil {
		return fmt.Errorf("error parsing rustup version: %w", err)
	}

	rustupConstraint, err := semver.NewConstraint(r.config.File.Language.Rust.RustupConstraint)
	if err != nil {
		return fmt.Errorf("error parsing rustup constraint: %w", err)
	}
	if !rustupConstraint.Check(rustupVersion) {
		pre := "To fix this error, run the following command"
		cmd := text.Bold("rustup self update")
		alt := fmt.Sprintf("%s If you installed rustup using a package manager, you may need to follow your package manager's documentation to update the rustup package.", text.Bold("INFO:"))

		return errors.RemediationError{
			Inner:       fmt.Errorf("rustup constraint not met: %s", r.config.File.Language.Rust.RustupConstraint),
			Remediation: fmt.Sprintf("%s:\n\n\t$ %s\n\n%s\n", pre, cmd, alt),
		}
	}

	// 3) Check that the desired toolchain version is installed
	//
	// We use rustup to assert that the toolchain is installed by streaming the output of
	// `rustup toolchain list` and looking for a toolchain whose prefix matches our desired
	// version.
	fmt.Fprintf(out, "Checking if Rust %s is installed...\n", r.config.File.Language.Rust.ToolchainConstraint)

	rustConstraint, err := semver.NewConstraint(r.config.File.Language.Rust.ToolchainConstraint)
	if err != nil {
		return fmt.Errorf("error parsing rust toolchain constraint: %w", err)
	}

	// Side-effect: sets r.toolchain
	err = r.toolchainVersion(rustConstraint)
	if err != nil {
		return err
	}

	// 4) Check `wasm32-wasi` target exists
	//
	// We use rustup to assert that the target is installed for our toolchain by streaming the
	// output of `rustup target list` and looking for the the `wasm32-wasi` value. If not found,
	// we error with help text suggesting how to install.

	fmt.Fprintf(out, "Checking if %s target is installed...\n", r.config.File.Language.Rust.WasmWasiTarget)

	// gosec flagged this:
	// G204 (CWE-78): Subprocess launched with function call as argument or cmd arguments
	//
	// TODO: decide if this is safe or not. It should be as the only affected
	// user should be the person making the local configuration change.
	//
	/* #nosec */
	cmd = exec.Command("rustup", "target", "list", "--installed", "--toolchain", r.toolchain.String())
	stdoutStderr, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error executing rustup: %w", err)
	}

	scanner := bufio.NewScanner(strings.NewReader(string(stdoutStderr)))
	scanner.Split(bufio.ScanWords)
	found := false
	for scanner.Scan() {
		if scanner.Text() == r.config.File.Language.Rust.WasmWasiTarget {
			found = true
			break
		}
	}

	if !found {
		return errors.RemediationError{
			Inner:       fmt.Errorf("rust target %s not found", r.config.File.Language.Rust.WasmWasiTarget),
			Remediation: fmt.Sprintf("To fix this error, run the following command:\n\n\t$ %s\n", text.Bold(fmt.Sprintf("rustup target add %s --toolchain %s", r.config.File.Language.Rust.WasmWasiTarget, r.toolchain.String()))),
		}
	}

	fmt.Fprintf(out, "Found wasm32-wasi target\n")

	// 5) Check Cargo.toml file exists in $PWD
	//
	// A valid Cargo.toml file is needed for the `cargo build` compilation
	// process. Therefore, we assert whether one exists in the current $PWD.

	fpath, err := filepath.Abs("Cargo.toml")
	if err != nil {
		return fmt.Errorf("error getting Cargo.toml path: %w", err)
	}

	if !filesystem.FileExists(fpath) {
		return fmt.Errorf("%s not found", fpath)
	}

	fmt.Fprintf(out, "Found Cargo.toml at %s\n", fpath)

	// 6) Verify `fastly` and  `fastly-sys` crate version
	//
	// A valid and up-to-date version of the fastly-sys crate is required.
	if !filesystem.FileExists(fpath) {
		return fmt.Errorf("%s not found", fpath)
	}

	var metadata CargoMetadata
	if err := metadata.Read(); err != nil {
		return fmt.Errorf("error reading cargo metadata: %w", err)
	}

	// Fetch the latest crate versions from cargo.io API.
	latestFastly, err := GetLatestCrateVersion(r.client, "fastly")
	if err != nil {
		return fmt.Errorf("error fetching latest crate version: %w", err)
	}

	fastlySysConstraint, err := semver.NewConstraint(r.config.File.Language.Rust.FastlySysConstraint)
	if err != nil {
		return fmt.Errorf("error parsing latest crate version: %w", err)
	}

	fastlySysVersion, err := GetCrateVersionFromMetadata(metadata, "fastly-sys")
	// If fastly-sys crate not found, error with dual remediation steps.
	if err != nil {
		return newCargoUpdateRemediationErr(err, latestFastly.String())
	}

	fastlyVersion, err := GetCrateVersionFromMetadata(metadata, "fastly")
	// If fastly crate not found, error with dual remediation steps.
	if err != nil {
		return newCargoUpdateRemediationErr(err, latestFastly.String())
	}

	// If fastly crate version is a prerelease, exit early. We assume that the
	// user knows what they are doing and avoids any confusing messaging to
	// "upgrade" to an older version.
	if fastlyVersion.Prerelease() != "" {
		return nil
	}

	// If fastly-sys version doesn't meet our constraint, error with dual remediation steps.
	if ok := fastlySysConstraint.Check(fastlySysVersion); !ok {
		return newCargoUpdateRemediationErr(fmt.Errorf("fastly crate not up-to-date"), latestFastly.String())
	}

	// If fastly crate version is lower than the latest, suggest user should
	// update, but don't error.
	if fastlyVersion.LessThan(latestFastly) {
		text.Break(out)
		text.Info(out, fmt.Sprintf(
			"an optional upgrade for the fastly crate is available, edit %s with:\n\n\t %s\n\nAnd then run the following command:\n\n\t$ %s\n",
			text.Bold("Cargo.toml"),
			text.Bold(fmt.Sprintf(`fastly = "^%s"`, latestFastly)),
			text.Bold("cargo update -p fastly"),
		))
		text.Break(out)
	}

	return nil
}

// Initialize implements the Toolchain interface and initializes a newly cloned
// package. It is a noop for Rust as the Cargo toolchain handles these steps.
func (r Rust) Initialize(out io.Writer) error { return nil }

// Build implements the Toolchain interface and attempts to compile the package
// Rust source to a Wasm binary.
func (r *Rust) Build(out io.Writer, verbose bool) error {
	// Get binary name from Cargo.toml.
	var m CargoManifest
	if err := m.Read("Cargo.toml"); err != nil {
		return fmt.Errorf("error reading Cargo.toml manifest: %w", err)
	}
	binName := m.Package.Name

	if r.toolchain == nil {
		rustConstraint, err := semver.NewConstraint(r.config.File.Language.Rust.ToolchainConstraint)
		if err != nil {
			return fmt.Errorf("error parsing rust toolchain constraint: %w", err)
		}

		// Side-effect: sets r.toolchain
		err = r.toolchainVersion(rustConstraint)
		if err != nil {
			return err
		}
	}

	toolchain := fmt.Sprintf("+%s", r.toolchain.String())

	args := []string{
		toolchain,
		"build",
		"--bin",
		binName,
		"--release",
		"--target",
		r.config.File.Language.Rust.WasmWasiTarget,
		"--color",
		"always",
	}
	if verbose {
		args = append(args, "--verbose")
	}
	// Append debuginfo RUSTFLAGS to command environment to ensure DWARF debug
	// information (such as, source mappings) are compiled into the binary.
	rustflags := "-C debuginfo=2"
	if val, ok := os.LookupEnv("RUSTFLAGS"); ok {
		os.Setenv("RUSTFLAGS", fmt.Sprintf("%s %s", val, rustflags))
	} else {
		os.Setenv("RUSTFLAGS", rustflags)
	}

	// Execute the `cargo build` commands with the Wasm WASI target, release
	// flags and env vars.
	cmd := fstexec.NewStreaming("cargo", args, os.Environ(), verbose, out)
	if err := cmd.Exec(); err != nil {
		return err
	}

	// Get working directory.
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting current working directory: %w", err)
	}
	var metadata CargoMetadata
	if err := metadata.Read(); err != nil {
		return fmt.Errorf("error reading cargo metadata: %w", err)
	}
	src := filepath.Join(metadata.TargetDirectory, r.config.File.Language.Rust.WasmWasiTarget, "release", fmt.Sprintf("%s.wasm", binName))
	dst := filepath.Join(dir, "bin", "main.wasm")

	// Check if bin directory exists and create if not.
	binDir := filepath.Join(dir, "bin")
	if err := filesystem.MakeDirectoryIfNotExists(binDir); err != nil {
		return fmt.Errorf("creating bin directory: %w", err)
	}

	err = filesystem.CopyFile(src, dst)
	if err != nil {
		return fmt.Errorf("copying wasm binary: %w", err)
	}

	return nil
}

func (r *Rust) toolchainVersion(rustConstraint *semver.Constraints) error {
	cmd := exec.Command("rustup", "toolchain", "list")
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error executing rustup: %w", err)
	}

	remediation := fmt.Sprintf("To fix this error, run the following command with a version within the given range %s:\n\n\t$ %s\n", r.config.File.Language.Rust.ToolchainConstraint, text.Bold("rustup toolchain install <version>"))

	versions := strings.Split(strings.Trim(string(stdoutStderr), "\n"), "\n")
	if len(versions) < 1 {
		return errors.RemediationError{
			Inner:       fmt.Errorf("rust toolchain %s not found", r.config.File.Language.Rust.ToolchainConstraint),
			Remediation: remediation,
		}
	}
	version := strings.Split(versions[len(versions)-1], "-")[0]

	r.toolchain, err = semver.NewVersion(version)
	if err != nil {
		return fmt.Errorf("error parsing rust toolchain version: %w", err)
	}

	if ok := rustConstraint.Check(r.toolchain); !ok {
		return errors.RemediationError{
			Inner:       fmt.Errorf("rust toolchain %s is incompatible with the constraint %s", r.toolchain, r.config.File.Language.Rust.ToolchainConstraint),
			Remediation: remediation,
		}
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
func GetLatestCrateVersion(client api.HTTPClient, name string) (*semver.Version, error) {
	url := fmt.Sprintf("https://crates.io/api/v1/crates/%s/versions", name)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error fetching latest crate version: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	crate := CargoCrateVersions{}
	err = json.Unmarshal(body, &crate)
	if err != nil {
		return nil, err
	}

	var versions []*semver.Version
	for _, v := range crate.Versions {
		// Parse version string and only append if not a prerelease.
		if version, err := semver.NewVersion(v.Version); err == nil && version.Prerelease() == "" {
			versions = append(versions, version)
		}
	}

	if len(versions) < 1 {
		return nil, fmt.Errorf("no valid crate versions found")
	}

	sort.Sort(semver.Collection(versions))

	latest := versions[len(versions)-1]

	return latest, nil
}

// GetCrateVersionFromMetadata searches for a crate inside a CargoMetadata tree
// and returns the crates version as a semver.Version.
func GetCrateVersionFromMetadata(metadata CargoMetadata, crate string) (*semver.Version, error) {
	// Search for crate in metadata tree.
	var c CargoPackage
	for _, p := range metadata.Package {
		if p.Name == crate {
			c = p
			break
		}
		for _, pp := range p.Dependencies {
			if pp.Name == crate {
				c = pp
				break
			}
		}
	}

	if c.Name == "" {
		return nil, fmt.Errorf("%s crate not found", crate)
	}

	// Parse lockfile version to semver.Version.
	version, err := semver.NewVersion(c.Version)
	if err != nil {
		return nil, fmt.Errorf("error parsing cargo metadata: %w", err)
	}

	return version, nil
}

// newCargoUpdateRemediationErr constructs a new a new RemediationError which
// wraps a cargo error and suggests to update the fastly crate to a specified
// version as its remediation message.
func newCargoUpdateRemediationErr(err error, version string) errors.RemediationError {
	return errors.RemediationError{
		Inner: err,
		Remediation: fmt.Sprintf(
			"To fix this error, edit %s with:\n\n\t %s\n\nAnd then run the following command:\n\n\t$ %s\n",
			text.Bold("Cargo.toml"),
			text.Bold(fmt.Sprintf(`fastly = "^%s"`, version)),
			text.Bold("cargo update -p fastly"),
		),
	}
}
