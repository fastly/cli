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
	"time"

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
	timeout   int
}

// NewRust constructs a new Rust.
func NewRust(client api.HTTPClient, config *config.Data, timeout int) *Rust {
	return &Rust{
		client:  client,
		config:  config,
		timeout: timeout,
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
//
// NOTE:
// Steps for validation are:
//
// 1. Lookup `rustc` in `$PATH`
// 2. Execute `rustc --version`
// 3. Validate `wasm32-wasi` target
// 4. Lookup `cargo` in `$PATH`
// 5. Lookup `Cargo.toml`
// 6. Validate `fastly-sys` crate version
// 7. Validate `fastly` crate version (optional upgrade suggestion)
//
// We used to use the `rustup` command to help validate some of the toolchain
// requirements but we no longer presume that to be how `rustc` is exposed to
// the user's running system. This means we have to use a low-level mechanism
// for identifying the wasm32-wasi target it available.
func (r *Rust) Verify(out io.Writer) (err error) {
	fmt.Fprintf(out, "Checking if `rustc` is installed...\n")

	err = validateCompilerExists()
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "Checking the `rustc` version...\n")

	err = validateCompilerVersion(r.config.File.Language.Rust.ToolchainConstraint)
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "Checking the `wasm32-wasi` target is installed...\n")

	err = validateWasmTarget(r.config.File.Language.Rust.WasmWasiTarget)
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "Checking if `cargo` is installed...\n")

	err = validateCargoExists()
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "Checking for a `Cargo.toml` file...\n")

	err = validateCargoToml()
	if err != nil {
		return err
	}

	latestFastlyCrate, err := GetLatestCrateVersion(r.client, "fastly")
	if err != nil {
		return fmt.Errorf("error fetching latest `fastly` crate version: %w", err)
	}

	var metadata CargoMetadata
	if err := metadata.Read(); err != nil {
		return fmt.Errorf("error reading cargo metadata: %w", err)
	}

	err = validateFastlySysCrate(metadata, r.config.File.Language.Rust.FastlySysConstraint, latestFastlyCrate.String())
	if err != nil {
		return err
	}

	err = validateFastlyCrate(metadata, latestFastlyCrate, out)
	if err != nil {
		return err
	}

	return nil
}

// validateCompilerExists checks if `rustc` is installed.
func validateCompilerExists() error {
	_, err := exec.LookPath("rustc")
	if err != nil {
		return errors.RemediationError{
			Inner:       err,
			Remediation: "Ensure the `rustc` compiler is installed:\n\n\thttps://www.rust-lang.org/tools/install",
		}
	}
	return nil
}

// validateCompilerVersion checks the `rustc` version meets our constraint.
func validateCompilerVersion(constraint string) error {
	cmd := []string{"rustc", "--version"}
	c := exec.Command(cmd[0], cmd[1:]...) // #nosec G204
	stdoutStderr, err := c.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error executing `%s`: %w", strings.Join(cmd, " "), err)
	}

	reader := bufio.NewReader(bytes.NewReader(stdoutStderr))
	line, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("error reading `%s` output: %w", strings.Join(cmd, " "), err)
	}

	// Example outputs:
	// rustc 1.54.0 (a178d0322 2021-07-26)
	// rustc 1.56.0-nightly (2d2bc94c8 2021-08-15)
	parts := strings.Split(line, " ")
	if len(parts) < 2 {
		return fmt.Errorf("error reading `%s` output", strings.Join(cmd, " "))
	}
	version := strings.Split(parts[1], "-")[0]
	rustcVersion, err := semver.NewVersion(version)
	if err != nil {
		return fmt.Errorf("error parsing `%s` output into a semver: %w", strings.Join(cmd, " "), err)
	}

	rustcConstraint, err := semver.NewConstraint(constraint)
	if err != nil {
		return fmt.Errorf("error parsing rustup constraint: %w", err)
	}

	if !rustcConstraint.Check(rustcVersion) {
		return errors.RemediationError{
			Inner:       fmt.Errorf("rustc constraint not met: %s", constraint),
			Remediation: "Ensure the `rustc` compiler version meets the constraint by installing an appropriate version of `rustc`.",
		}
	}

	return nil
}

// validateWasmTarget checks the `wasm32-wasi` target is installed.
func validateWasmTarget(target string) error {
	cmd := []string{"rustc", "--print", "sysroot"}
	c := exec.Command(cmd[0], cmd[1:]...) // #nosec G204
	stdoutStderr, err := c.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error executing `%s`: %w", strings.Join(cmd, " "), err)
	}

	sysroot := strings.TrimSpace(string(stdoutStderr))
	path := filepath.Join(sysroot, "lib", "rustlib", "wasm32-wasi")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return errors.RemediationError{
			Inner:       fmt.Errorf("the Rust target directory `%s` doesn't exist", path),
			Remediation: fmt.Sprintf("Ensure the following target is installed:\n\n\t%s", target),
		}
	}

	return nil
}

// validateCargoExists checks `cargo` is installed.
func validateCargoExists() error {
	_, err := exec.LookPath("cargo")
	if err != nil {
		return errors.RemediationError{
			Inner:       err,
			Remediation: "Ensure the `cargo` package manager is installed:\n\n\thttps://doc.rust-lang.org/cargo/getting-started/installation.html",
		}
	}
	return nil
}

// validateCargoToml checks the `Cargo.toml` file exists.
func validateCargoToml() error {
	file := "Cargo.toml"
	path, err := filepath.Abs(file)
	if err != nil {
		return fmt.Errorf("error parsing the %s file path: %w", file, err)
	}

	if !filesystem.FileExists(path) {
		return errors.RemediationError{
			Inner:       fmt.Errorf("a Cargo.toml file is missing from the current directory"),
			Remediation: "Ensure you have a valid `Cargo.toml` file defined.",
		}
	}

	return nil
}

// validateFastlySysCrate checks the `fastly-sys` crate version meets our constraint.
//
// The following logic is an requirement that we don't want a customer to
// have to think about and so we don't indicate to the user that we're
// validating the fastly-sys crate specifically (i.e. we make the messaging
// generic towards the fastly crate).
func validateFastlySysCrate(metadata CargoMetadata, constraint string, latestFastlyCrateVersion string) error {
	fastlySysConstraint, err := semver.NewConstraint(constraint)
	if err != nil {
		return fmt.Errorf("error parsing crate constraint: %w", err)
	}

	fastlySysVersion, err := GetCrateVersionFromMetadata(metadata, "fastly-sys")
	if err != nil {
		return newCargoUpdateRemediationErr(err, latestFastlyCrateVersion)
	}

	if ok := fastlySysConstraint.Check(fastlySysVersion); !ok {
		return newCargoUpdateRemediationErr(fmt.Errorf("fastly crate not up-to-date"), latestFastlyCrateVersion)
	}

	return nil
}

// validateFastlyCrate checks the `fastly` crate version meets our constraint.
//
// The folllowing logic is an optional upgrade suggestion and so we don't
// display any up front message to say we're checking the fastly crate.
func validateFastlyCrate(metadata CargoMetadata, v *semver.Version, out io.Writer) error {
	fastlyVersion, err := GetCrateVersionFromMetadata(metadata, "fastly")
	if err != nil {
		return newCargoUpdateRemediationErr(err, v.String())
	}

	// If the fastly crate version is a prerelease, exit early. We assume that the
	// user knows what they are doing and avoids any confusing messaging to
	// "upgrade" to an older version.
	if fastlyVersion.Prerelease() != "" {
		return nil
	}

	// If the fastly crate version is lower than the latest, suggest user should
	// update, but don't error.
	if fastlyVersion.LessThan(v) {
		text.Break(out)
		text.Info(out, fmt.Sprintf(
			"an optional upgrade for the fastly crate is available, edit %s with:\n\n\t %s\n\nAnd then run the following command:\n\n\t$ %s\n",
			text.Bold("Cargo.toml"),
			text.Bold(fmt.Sprintf(`fastly = "^%s"`, v.String())),
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

	args := []string{
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
	cmd := fstexec.Streaming{
		Command: "cargo",
		Args:    args,
		Env:     os.Environ(),
		Output:  out,
	}
	if r.timeout > 0 {
		cmd.Timeout = time.Duration(r.timeout) * time.Second
	}
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
	cmd := exec.Command("rustc", "--version")
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error executing `rustc --version`: %w", err)
	}

	remediation := fmt.Sprintf("To fix this error, run the following command:\n\n\t$ %s\n", text.Bold("rustup update stable"))

	// output should look like:
	// rustc 1.54.0 (a178d0322 2021-07-26)
	version := strings.Split(string(stdoutStderr), " ")
	if len(version) < 2 {
		return errors.RemediationError{
			Inner:       fmt.Errorf("rust toolchain %s not found: %s", r.config.File.Language.Rust.ToolchainConstraint, string(stdoutStderr)),
			Remediation: remediation,
		}
	}

	r.toolchain, err = semver.NewVersion(version[1])
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

// newCargoUpdateRemediationErr constructs a new RemediationError which wraps a
// cargo error and suggests to update the fastly crate to a specified version as
// its remediation message.
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
