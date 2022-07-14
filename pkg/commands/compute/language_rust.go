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
	fsterr "github.com/fastly/cli/pkg/errors"
	fstexec "github.com/fastly/cli/pkg/exec"
	"github.com/fastly/cli/pkg/filesystem"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	toml "github.com/pelletier/go-toml"
)

// RustSourceDirectory represents the source code directory.
const RustSourceDirectory = "src"

// RustManifestName represents the language file for configuring dependencies.
const RustManifestName = "Cargo.toml"

// CargoPackage models the package configuration properties of a Rust Cargo
// package which we are interested in and is embedded within CargoManifest and
// CargoLock.
type CargoPackage struct {
	Name    string `toml:"name" json:"name"`
	Version string `toml:"version" json:"version"`
}

// CargoManifest models the package configuration properties of a Rust Cargo
// manifest which we are interested in and are read from the Cargo.toml manifest
// file within the $PWD of the package.
type CargoManifest struct {
	Package CargoPackage `toml:"package"`
}

// Read the contents of the Cargo.toml manifest from filename.
func (m *CargoManifest) Read(path string) error {
	// gosec flagged this:
	// G304 (CWE-22): Potential file inclusion via variable.
	// Disabling as we need to load the Cargo.toml from the user's file system.
	// This file is decoded into a predefined struct, any unrecognised fields are dropped.
	/* #nosec */
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	err = toml.Unmarshal(data, m)
	return err
}

// SetPackageName into Cargo.toml manifest.
func (m *CargoManifest) SetPackageName(name, path string) error {
	if err := m.Read("Cargo.toml"); err != nil {
		return fmt.Errorf("error reading Cargo.toml manifest: %w", err)
	}
	// gosec flagged this:
	// G304 (CWE-22): Potential file inclusion via variable.
	// Disabling as we need to load the Cargo.toml from the user's file system.
	// This file is read, and the content is written back with the package
	// name updated.
	/* #nosec */
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("error reading Cargo.toml file: %w", err)
	}
	data = bytes.ReplaceAll(data, []byte(m.Package.Name), []byte(name))
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("error updating Cargo manifest file: %w", err)
	}
	return nil
}

// CargoMetadataPackage models the package structure returned when executing
// the command `cargo metadata`.
type CargoMetadataPackage struct {
	Name         string                 `toml:"name" json:"name"`
	Version      string                 `toml:"version" json:"version"`
	Dependencies []CargoMetadataPackage `toml:"dependencies" json:"dependencies"`
}

// CargoMetadata models information about the workspace members and resolved
// dependencies of the current package via `cargo metadata` command output.
type CargoMetadata struct {
	Package         []CargoMetadataPackage `json:"packages"`
	TargetDirectory string                 `json:"target_directory"`
}

// Read the contents of the Cargo.lock file from filename.
func (m *CargoMetadata) Read(errlog fsterr.LogInterface) error {
	cmd := exec.Command("cargo", "metadata", "--quiet", "--format-version", "1")
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		if len(stdoutStderr) > 0 {
			err = fmt.Errorf("%s", strings.TrimSpace(string(stdoutStderr)))
		}
		errlog.Add(err)
		return err
	}
	r := bytes.NewReader(stdoutStderr)
	if err := json.NewDecoder(r).Decode(&m); err != nil {
		errlog.Add(err)
		return err
	}
	return nil
}

// Rust implements a Toolchain for the Rust language.
type Rust struct {
	Shell

	build     string
	client    api.HTTPClient
	config    config.Rust
	errlog    fsterr.LogInterface
	pkgName   string
	postBuild string
	timeout   int
}

// NewRust constructs a new Rust toolchain.
func NewRust(pkgName string, scripts manifest.Scripts, errlog fsterr.LogInterface, client api.HTTPClient, timeout int, cfg config.Rust) *Rust {
	return &Rust{
		Shell:     Shell{},
		build:     scripts.Build,
		client:    client,
		config:    cfg,
		errlog:    errlog,
		pkgName:   pkgName,
		postBuild: scripts.PostBuild,
		timeout:   timeout,
	}
}

// Verify implements the Toolchain interface and verifies whether the Rust
// language toolchain is correctly configured on the host.
//
// NOTE:
// Steps for validation are:
//
// 1. Validate `rustc` installed.
// 2. Validate `rustc --version` meets the constraint.
// 3. Validate `wasm32-wasi` target is added to the relevant toolchain.
// 4. Validate `cargo` is installed.
// 5. Validate `fastly-sys` crate version.
// 6. Validate `fastly` crate version (optional upgrade suggestion).
func (r *Rust) Verify(out io.Writer) (err error) {
	fmt.Fprintf(out, "Checking if `rustc` is installed...\n")

	err = validateCompilerExists(r.errlog)
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "Checking the `rustc` version...\n")

	err = validateCompilerVersion(r.config.ToolchainConstraint, r.errlog)
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "Checking the `wasm32-wasi` target is installed...\n")

	err = validateWasmTarget(r.config.WasmWasiTarget, r.errlog)
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "Checking if `cargo` is installed...\n")

	err = validateCargoExists(r.errlog)
	if err != nil {
		return err
	}

	// Validate the fastly and fastly-sys crates...

	latestFastlyCrate, err := GetLatestCrateVersion(r.client, "fastly", r.errlog)
	if err != nil {
		return fmt.Errorf("error fetching latest `fastly` crate version: %w", err)
	}

	var metadata CargoMetadata
	if err := metadata.Read(r.errlog); err != nil {
		return fmt.Errorf("error reading cargo metadata: %w", err)
	}

	err = validateFastlySysCrate(metadata, r.config.FastlySysConstraint, latestFastlyCrate.String(), r.errlog)
	if err != nil {
		return err
	}

	err = validateFastlyCrate(metadata, latestFastlyCrate, out, r.errlog)
	if err != nil {
		return err
	}

	return nil
}

// validateCompilerExists checks if `rustc` is installed.
func validateCompilerExists(errlog fsterr.LogInterface) error {
	_, err := exec.LookPath("rustc")
	if err != nil {
		errlog.Add(err)
		return fsterr.RemediationError{
			Inner:       err,
			Remediation: "Ensure the `rustc` compiler is installed:\n\n\thttps://www.rust-lang.org/tools/install",
		}
	}
	return nil
}

// validateCompilerVersion checks the `rustc` version meets our constraint.
func validateCompilerVersion(constraint string, errlog fsterr.LogInterface) error {
	version, err := rustcVersion(errlog)
	if err != nil {
		return err
	}

	rustcVersion, err := semver.NewVersion(version)
	if err != nil {
		errlog.Add(err)
		return fmt.Errorf("error parsing `%s` output %q into a semver: %w", "rustc --version", version, err)
	}

	rustcConstraint, err := semver.NewConstraint(constraint)
	if err != nil {
		errlog.Add(err)
		return fmt.Errorf("error parsing rustup constraint: %w", err)
	}

	if !rustcConstraint.Check(rustcVersion) {
		err := fsterr.RemediationError{
			Inner:       fmt.Errorf("rustc constraint '%s' not met: %s", constraint, version),
			Remediation: "Run `rustup update stable`, or ensure your `rust-toolchain` file specifies a version matching the constraint (e.g. `channel = \"stable\"`).",
		}
		errlog.Add(err)
		return err
	}

	return nil
}

// rustcVersion returns the active rustc compiler version.
func rustcVersion(errlog fsterr.LogInterface) (string, error) {
	cmd := []string{"rustc", "--version"}
	c := exec.Command(cmd[0], cmd[1:]...) // #nosec G204
	stdoutStderr, err := c.CombinedOutput()
	if err != nil {
		errlog.Add(err)
		return "", fmt.Errorf("error executing `%s`: %w", strings.Join(cmd, " "), err)
	}

	// Note: when `rustc` is managed by `rustup`, and the toolchain that
	// `rustup` considers active is not installed, then the first use of `rustc`
	// (or any Rust distribution executable) will install it. This produces some
	// `rustup` output before the real `rustc --version` command fires.
	scanner := bufio.NewScanner(bytes.NewReader(stdoutStderr))
	line := ""
	for scanner.Scan() {
		line = scanner.Text()
	}
	err = scanner.Err()
	if line == "" || err != nil {
		err = fmt.Errorf("error reading `%s` output: %w", strings.Join(cmd, " "), err)
		errlog.Add(err)
		return "", err
	}
	line = strings.TrimSpace(line)

	// Example outputs:
	// rustc 1.54.0 (a178d0322 2021-07-26)
	// rustc 1.56.0-nightly (2d2bc94c8 2021-08-15)
	parts := strings.Split(line, " ")
	if len(parts) < 2 {
		err := fmt.Errorf("error reading `%s` output", strings.Join(cmd, " "))
		errlog.Add(err)
		return "", err
	}

	version := strings.Split(parts[1], "-")
	if len(version) > 1 {
		err := fsterr.RemediationError{
			Inner:       fmt.Errorf("non-stable releases are not supported: %s", parts[1]),
			Remediation: "Run `rustup update stable`, or ensure your `rust-toolchain` file specifies a version matching the constraint (e.g. `channel = \"stable\"`). Alternatively utilise the CLI's `--force` flag.",
		}
		errlog.Add(err)
		return "", err
	}

	return version[0], nil
}

// validateWasmTarget checks the `wasm32-wasi` target is installed.
//
// If the user has `rustup` installed then we use it to identify if the target
// is installed, otherwise we fallback to a low-level check of the target
// directory using `rustc --print sysroot`.
func validateWasmTarget(target string, errlog fsterr.LogInterface) error {
	_, err := exec.LookPath("rustup")
	if err != nil {
		errlog.Add(err)
		return rustcSysroot(target, errlog)
	}

	toolchain, err := rustupToolchain(errlog)
	if err != nil {
		return err
	}

	cmd := []string{"rustup", "target", "list", "--installed", "--toolchain", toolchain}
	c := exec.Command(cmd[0], cmd[1:]...) // #nosec G204
	stdoutStderr, err := c.CombinedOutput()
	if err != nil {
		errlog.Add(err)
		return fmt.Errorf("error executing `%s`: %w", strings.Join(cmd, " "), err)
	}

	scanner := bufio.NewScanner(strings.NewReader(string(stdoutStderr)))
	scanner.Split(bufio.ScanWords)
	found := false
	for scanner.Scan() {
		if scanner.Text() == target {
			found = true
			break
		}
	}

	if !found {
		err := fsterr.RemediationError{
			Inner:       fmt.Errorf("rust target %s not found", target),
			Remediation: fmt.Sprintf("Run the following command:\n\n\t$ %s\n", text.Bold(fmt.Sprintf("rustup target add %s --toolchain %s", target, toolchain))),
		}
		errlog.Add(err)
		return err
	}

	return nil
}

// rustupToolchain returns the active rustup toolchain.
func rustupToolchain(errlog fsterr.LogInterface) (string, error) {
	cmd := []string{"rustup", "show", "active-toolchain"}
	c := exec.Command(cmd[0], cmd[1:]...) // #nosec G204
	stdoutStderr, err := c.CombinedOutput()
	if err != nil {
		errlog.Add(err)
		return "", fmt.Errorf("error executing `%s`: %w", strings.Join(cmd, " "), err)
	}

	reader := bufio.NewReader(bytes.NewReader(stdoutStderr))
	line, err := reader.ReadString('\n')
	if err != nil {
		errlog.Add(err)
		return "", fmt.Errorf("error reading `%s` output: %w", strings.Join(cmd, " "), err)
	}

	// Example outputs:
	// stable-x86_64-apple-darwin (default)
	// 1.54.0-x86_64-apple-darwin (directory override for '/Users/integralist/Code/fastly/cli')
	parts := strings.Split(line, "-")
	if len(parts) < 2 {
		err := fmt.Errorf("error reading `%s` output", strings.Join(cmd, " "))
		errlog.Add(err)
		return "", err
	}

	return parts[0], nil
}

// rustcSysroot validates if the wasm32-wasi target is installed by using the
// low-level rustc compiler `--print sysroot` flag.
//
// This is called only when the user doesn't have `rustup` installed.
func rustcSysroot(target string, errlog fsterr.LogInterface) error {
	cmd := []string{"rustc", "--print", "sysroot"}
	c := exec.Command(cmd[0], cmd[1:]...) // #nosec G204
	stdoutStderr, err := c.CombinedOutput()
	if err != nil {
		errlog.Add(err)
		return fmt.Errorf("error executing `%s`: %w", strings.Join(cmd, " "), err)
	}

	sysroot := strings.TrimSpace(string(stdoutStderr))
	path := filepath.Join(sysroot, "lib", "rustlib", "wasm32-wasi")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		errlog.Add(err)
		return fsterr.RemediationError{
			Inner:       fmt.Errorf("the Rust target directory `%s` doesn't exist", path),
			Remediation: fmt.Sprintf("Ensure the following target is installed:\n\n\t%s", target),
		}
	}

	return nil
}

// validateCargoExists checks `cargo` is installed.
func validateCargoExists(errlog fsterr.LogInterface) error {
	_, err := exec.LookPath("cargo")
	if err != nil {
		errlog.Add(err)
		return fsterr.RemediationError{
			Inner:       err,
			Remediation: "Ensure the `cargo` package manager is installed:\n\n\thttps://doc.rust-lang.org/cargo/getting-started/installation.html",
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
func validateFastlySysCrate(metadata CargoMetadata, constraint string, latestFastlyCrateVersion string, errlog fsterr.LogInterface) error {
	fastlySysConstraint, err := semver.NewConstraint(constraint)
	if err != nil {
		errlog.Add(err)
		return fmt.Errorf("error parsing crate constraint: %w", err)
	}

	fastlySysVersion, err := GetCrateVersionFromMetadata(metadata, "fastly-sys")
	if err != nil {
		errlog.Add(err)
		return newCargoUpdateRemediationErr(err, latestFastlyCrateVersion)
	}

	if ok := fastlySysConstraint.Check(fastlySysVersion); !ok {
		err := fmt.Errorf("fastly crate not up-to-date")
		errlog.Add(err)
		return newCargoUpdateRemediationErr(err, latestFastlyCrateVersion)
	}

	return nil
}

// validateFastlyCrate checks the `fastly` crate version meets our constraint.
//
// The folllowing logic is an optional upgrade suggestion and so we don't
// display any up front message to say we're checking the fastly crate.
func validateFastlyCrate(metadata CargoMetadata, v *semver.Version, out io.Writer, errlog fsterr.LogInterface) error {
	fastlyVersion, err := GetCrateVersionFromMetadata(metadata, "fastly")
	if err != nil {
		errlog.Add(err)
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
			text.Bold(RustManifestName),
			text.Bold(fmt.Sprintf(`fastly = "^%s"`, v.String())),
			text.Bold("cargo update -p fastly"),
		))
		text.Break(out)
	}

	return nil
}

// Initialize implements the Toolchain interface and initializes a newly cloned
// package. It is a noop for Rust as the Cargo toolchain handles these steps.
func (r Rust) Initialize(out io.Writer) error {
	var m CargoManifest
	if err := m.setPackageName(r.pkgName, RustManifestName); err != nil {
		r.errlog.Add(err)
		return fmt.Errorf("error updating %s manifest: %w", RustManifestName, err)
	}
	return nil
}

// Build implements the Toolchain interface and attempts to compile the package
// Rust source to a Wasm binary.
//
// NOTE: The callback function is called before executing any potential custom
// post_build script defined, allowing the controlling build logic to display a
// message to the user informing them a post_build is going to execute.
func (r *Rust) Build(out io.Writer, progress text.Progress, verbose bool, callback func() error) error {
	// Get binary name from Cargo.toml.
	var m CargoManifest
	if err := m.Read(RustManifestName); err != nil {
		r.errlog.Add(err)
		return fmt.Errorf("error reading %s manifest: %w", RustManifestName, err)
	}
	binName := m.Package.Name

	cmd := "cargo"
	args := []string{
		"build",
		"--bin",
		binName,
		"--release",
		"--target",
		r.config.WasmWasiTarget,
		"--color",
		"always",
	}
	if verbose {
		args = append(args, "--verbose")
	}

	if r.build != "" {
		cmd, args = r.Shell.Build(r.build)
	}

	// Execute the `cargo build` commands with the Wasm WASI target, release
	// flags and env vars.
	err := r.execCommand(cmd, args, out, progress, verbose)
	if err != nil {
		return err
	}

	// Get working directory.
	dir, err := os.Getwd()
	if err != nil {
		r.errlog.Add(err)
		return fmt.Errorf("getting current working directory: %w", err)
	}
	var metadata CargoMetadata
	if err := metadata.Read(r.errlog); err != nil {
		r.errlog.Add(err)
		return fmt.Errorf("error reading cargo metadata: %w", err)
	}
	src := filepath.Join(metadata.TargetDirectory, r.config.WasmWasiTarget, "release", fmt.Sprintf("%s.wasm", binName))
	dst := filepath.Join(dir, "bin", "main.wasm")

	// Check if bin directory exists and create if not.
	binDir := filepath.Join(dir, "bin")
	if err := filesystem.MakeDirectoryIfNotExists(binDir); err != nil {
		r.errlog.Add(err)
		return fmt.Errorf("creating bin directory: %w", err)
	}

	err = filesystem.CopyFile(src, dst)
	if err != nil {
		r.errlog.Add(err)
		return fmt.Errorf("copying wasm binary: %w", err)
	}

	// NOTE: We set the progress indicator to Done() so that any output we now
	// print via the post_build callback doesn't get hidden by the progress status.
	// The progress is 'reset' inside the main build controller `build.go`.
	progress.Done()

	if r.postBuild != "" {
		if err = callback(); err == nil {
			cmd, args := r.Shell.Build(r.postBuild)
			err := r.execCommand(cmd, args, out, progress, verbose)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// TODO: Consider generics to avoid re-implementing this same logic.
func (r Rust) execCommand(cmd string, args []string, out, progress io.Writer, verbose bool) error {
	s := fstexec.Streaming{
		Command:  cmd,
		Args:     args,
		Env:      os.Environ(),
		Output:   out,
		Progress: progress,
		Verbose:  verbose,
	}
	if r.timeout > 0 {
		s.Timeout = time.Duration(r.timeout) * time.Second
	}
	if err := s.Exec(); err != nil {
		r.errlog.Add(err)
		return err
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
func GetLatestCrateVersion(client api.HTTPClient, name string, errlog fsterr.LogInterface) (*semver.Version, error) {
	url := fmt.Sprintf("https://crates.io/api/v1/crates/%s/versions", name)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		errlog.Add(err)
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		errlog.Add(err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errlog.Add(err)
		return nil, fmt.Errorf("error fetching latest crate version: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		errlog.Add(err)
		return nil, err
	}

	crate := CargoCrateVersions{}
	err = json.Unmarshal(body, &crate)
	if err != nil {
		errlog.Add(err)
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
	var c CargoMetadataPackage
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
func newCargoUpdateRemediationErr(err error, version string) fsterr.RemediationError {
	return fsterr.RemediationError{
		Inner: err,
		Remediation: fmt.Sprintf(
			"To fix this error, edit %s with:\n\n\t %s\n\nAnd then run the following command:\n\n\t$ %s\n",
			text.Bold(RustManifestName),
			text.Bold(fmt.Sprintf(`fastly = "^%s"`, version)),
			text.Bold("cargo update -p fastly"),
		),
	}
}
