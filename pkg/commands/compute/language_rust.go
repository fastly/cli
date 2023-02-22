package compute

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/filesystem"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	toml "github.com/pelletier/go-toml"
	"github.com/theckman/yacspin"
)

// RustDefaultBuildCommand is a build command compiled into the CLI binary so it
// can be used as a fallback for customer's who have an existing C@E project and
// are simply upgrading their CLI version and might not be familiar with the
// changes in the 4.0.0 release with regards to how build logic has moved to the
// fastly.toml manifest.
//
// NOTE: In the 5.x CLI releases we persisted the default to the fastly.toml
// We no longer do that. In 6.x we use the default and just inform the user.
// This makes the experience less confusing as users didn't expect file changes.
const RustDefaultBuildCommand = "cargo build --bin %s --release --target wasm32-wasi --color always"

// RustManifest is the manifest file for defining project configuration.
const RustManifest = "Cargo.toml"

// RustDefaultPackageName is the expected binary create/package name to be built.
const RustDefaultPackageName = "fastly-compute-project"

// RustSourceDirectory represents the source code directory.                                               │                                                           │
const RustSourceDirectory = "src"

// NewRust constructs a new Rust toolchain.
func NewRust(
	fastlyManifest *manifest.File,
	errlog fsterr.LogInterface,
	timeout int,
	cfg config.Rust,
	out io.Writer,
	verbose bool,
) *Rust {
	return &Rust{
		Shell:     Shell{},
		build:     fastlyManifest.Scripts.Build,
		config:    cfg,
		errlog:    errlog,
		output:    out,
		postBuild: fastlyManifest.Scripts.PostBuild,
		timeout:   timeout,
		verbose:   verbose,
	}
}

// Rust implements a Toolchain for the Rust language.
type Rust struct {
	Shell

	// build is a shell command defined in fastly.toml using [scripts.build].
	build string
	// config is the Rust specific application configuration.
	config config.Rust
	// errlog is an abstraction for recording errors to disk.
	errlog fsterr.LogInterface
	// output is the users terminal stdout stream
	output io.Writer
	// postBuild is a custom script executed after the build but before the Wasm
	// binary is added to the .tar.gz archive.
	postBuild string
	// projectRoot is the root directory where the Cargo.toml is located.
	projectRoot string
	// timeout is the build execution threshold.
	timeout int
	// verbose indicates if the user set --verbose
	verbose bool
}

// Build compiles the user's source code into a Wasm binary.
func (r *Rust) Build(out io.Writer, spinner *yacspin.Spinner, verbose bool, callback func() error) error {
	var noBuildScript bool
	if r.build == "" {
		r.build = fmt.Sprintf(RustDefaultBuildCommand, RustDefaultPackageName)
		noBuildScript = true
	}

	err := r.modifyCargoPackageName()
	if err != nil {
		return err
	}

	if noBuildScript && r.verbose {
		text.Info(out, "No [scripts.build] found in fastly.toml. The following default build command for Rust will be used: `%s`\n", r.build)
	}

	r.toolchainConstraint()

	bt := BuildToolchain{
		buildFn:                   r.Shell.Build,
		buildScript:               r.build,
		errlog:                    r.errlog,
		internalPostBuildCallback: r.ProcessLocation,
		postBuild:                 r.postBuild,
		timeout:                   r.timeout,
		out:                       out,
		postBuildCallback:         callback,
		spinner:                   spinner,
		verbose:                   verbose,
	}

	return bt.Build()
}

// modifyCargoPackageName validates whether the --bin flag matches the
// Cargo.toml package name. If it doesn't match, update the default build script
// to match.
func (r *Rust) modifyCargoPackageName() error {
	s := "cargo locate-project --quiet"
	args := strings.Split(s, " ")

	var stdout, stderr bytes.Buffer

	// gosec flagged this:
	// G204 (CWE-78): Subprocess launched with variable
	// Disabling as we control this command.
	// #nosec
	// nosemgrep
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		if stderr.Len() > 0 {
			err = fmt.Errorf("%w: %s", err, stderr.String())
		}
		return fmt.Errorf("failed to execute command '%s': %w", s, err)
	}

	if r.verbose {
		text.Break(r.output)
		text.Output(r.output, "Command output for '%s': %s", s, stdout.String())
	}

	var cp *CargoLocateProject
	err = json.Unmarshal(stdout.Bytes(), &cp)
	if err != nil {
		return fmt.Errorf("failed to unmarshal manifest project root metadata: %w", err)
	}

	r.projectRoot = cp.Root

	var m CargoManifest
	if err := m.Read(cp.Root); err != nil {
		return fmt.Errorf("error reading %s manifest: %w", RustManifest, err)
	}

	if m.Package.Name != RustDefaultPackageName {
		r.build = fmt.Sprintf(RustDefaultBuildCommand, m.Package.Name)
	}

	return nil
}

// toolchainConstraint warns the user if the required constraint is not met.
//
// NOTE: We don't stop the build as their toolchain may compile successfully.
// The warning is to help a user know something isn't quite right and gives them
// the opportunity to do something about it if they choose.
func (r *Rust) toolchainConstraint() {
	versionCommand := "cargo version --quiet"
	args := strings.Split(versionCommand, " ")

	// gosec flagged this:
	// G204 (CWE-78): Subprocess launched with function call as argument or cmd arguments
	// Disabling as we trust the source of the variable.
	// #nosec
	// nosemgrep
	cmd := exec.Command(args[0], args[1:]...)
	stdoutStderr, err := cmd.CombinedOutput()
	output := string(stdoutStderr)
	if err != nil {
		return
	}

	versionPattern := regexp.MustCompile(`cargo (?P<version>\d[^\s]+)`)
	match := versionPattern.FindStringSubmatch(output)
	if len(match) < 2 { // We expect a pattern with one capture group.
		return
	}
	version := match[1]

	v, err := semver.NewVersion(version)
	if err != nil {
		return
	}

	c, err := semver.NewConstraint(r.config.ToolchainConstraint)
	if err != nil {
		return
	}

	if r.verbose {
		text.Info(r.output, "The Fastly CLI requires a Rust version '%s'. ", r.config.ToolchainConstraint)
	}

	if !c.Check(v) {
		text.Warning(r.output, "The Rust version '%s' didn't meet the constraint '%s'", version, r.config.ToolchainConstraint)
		text.Break(r.output)
	}
}

// ProcessLocation ensures the generated Rust Wasm binary is moved to the
// required location for packaging.
func (r *Rust) ProcessLocation() error {
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
	var m CargoManifest
	if err := m.Read(r.projectRoot); err != nil {
		return fmt.Errorf("error reading %s manifest: %w", RustManifest, err)
	}

	src := filepath.Join(metadata.TargetDirectory, r.config.WasmWasiTarget, "release", fmt.Sprintf("%s.wasm", m.Package.Name))
	dst := filepath.Join(dir, "bin", "main.wasm")

	err = filesystem.CopyFile(src, dst)
	if err != nil {
		r.errlog.Add(err)
		return fmt.Errorf("failed to copy wasm binary: %w", err)
	}
	return nil
}

// CargoLocateProject represents the metadata for where to find the project's
// Cargo.toml manifest file.
type CargoLocateProject struct {
	Root string `json:"root"`
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

// CargoPackage models the package configuration properties of a Rust Cargo
// package which we are interested in and is embedded within CargoManifest and
// CargoLock.
type CargoPackage struct {
	Name    string `toml:"name" json:"name"`
	Version string `toml:"version" json:"version"`
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

// CargoMetadataPackage models the package structure returned when executing
// the command `cargo metadata`.
type CargoMetadataPackage struct {
	Name         string                 `toml:"name" json:"name"`
	Version      string                 `toml:"version" json:"version"`
	Dependencies []CargoMetadataPackage `toml:"dependencies" json:"dependencies"`
}
