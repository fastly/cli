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
	toml "github.com/pelletier/go-toml"

	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/filesystem"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
)

// RustDefaultBuildCommand is a build command compiled into the CLI binary so it
// can be used as a fallback for customer's who have an existing Compute project and
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
	globals *global.Data,
	flags Flags,
	in io.Reader,
	manifestFilename string,
	out io.Writer,
	spinner text.Spinner,
) *Rust {
	return &Rust{
		Shell: Shell{},

		autoYes:          globals.Flags.AutoYes,
		build:            fastlyManifest.Scripts.Build,
		config:           globals.Config.Language.Rust,
		env:              fastlyManifest.Scripts.EnvVars,
		errlog:           globals.ErrLog,
		input:            in,
		manifestFilename: manifestFilename,
		nonInteractive:   globals.Flags.NonInteractive,
		output:           out,
		postBuild:        fastlyManifest.Scripts.PostBuild,
		spinner:          spinner,
		timeout:          flags.Timeout,
		verbose:          globals.Verbose(),
	}
}

// Rust implements a Toolchain for the Rust language.
type Rust struct {
	Shell

	// autoYes is the --auto-yes flag.
	autoYes bool
	// build is a shell command defined in fastly.toml using [scripts.build].
	build string
	// config is the Rust specific application configuration.
	config config.Rust
	// defaultBuild indicates if the default build script was used.
	defaultBuild bool
	// env is environment variables to be set.
	env []string
	// errlog is an abstraction for recording errors to disk.
	errlog fsterr.LogInterface
	// input is the user's terminal stdin stream
	input io.Reader
	// manifestFilename is the name of the manifest file.
	manifestFilename string
	// nonInteractive is the --non-interactive flag.
	nonInteractive bool
	// output is the users terminal stdout stream
	output io.Writer
	// packageName is the resolved package name from the project Cargo.toml
	packageName string
	// postBuild is a custom script executed after the build but before the Wasm
	// binary is added to the .tar.gz archive.
	postBuild string
	// projectRoot is the root directory where the Cargo.toml is located.
	projectRoot string
	// spinner is a terminal progress status indicator.
	spinner text.Spinner
	// timeout is the build execution threshold.
	timeout int
	// verbose indicates if the user set --verbose
	verbose bool
}

// DefaultBuildScript indicates if a custom build script was used.
func (r *Rust) DefaultBuildScript() bool {
	return r.defaultBuild
}

// CargoLockFilePackage represents a package within a Rust lockfile.
type CargoLockFilePackage struct {
	Name    string `toml:"name"`
	Version string `toml:"version"`
}

// CargoLockFile represents a Rust lockfile.
type CargoLockFile struct {
	Packages []CargoLockFilePackage `toml:"package"`
}

// Dependencies returns all dependencies used by the project.
func (r *Rust) Dependencies() map[string]string {
	deps := make(map[string]string)

	var clf CargoLockFile
	if data, err := os.ReadFile("Cargo.lock"); err == nil {
		if err := toml.Unmarshal(data, &clf); err == nil {
			for _, v := range clf.Packages {
				deps[v.Name] = v.Version
			}
		}
	}

	return deps
}

// Build compiles the user's source code into a Wasm binary.
func (r *Rust) Build() error {
	if r.build == "" {
		r.build = fmt.Sprintf(RustDefaultBuildCommand, RustDefaultPackageName)
		r.defaultBuild = true
	}

	err := r.modifyCargoPackageName(r.defaultBuild)
	if err != nil {
		return err
	}

	if r.defaultBuild && r.verbose {
		text.Info(r.output, "No [scripts.build] found in %s. The following default build command for Rust will be used: `%s`\n\n", r.manifestFilename, r.build)
	}

	r.toolchainConstraint()

	bt := BuildToolchain{
		autoYes:                   r.autoYes,
		buildFn:                   r.Shell.Build,
		buildScript:               r.build,
		env:                       r.env,
		errlog:                    r.errlog,
		in:                        r.input,
		internalPostBuildCallback: r.ProcessLocation,
		manifestFilename:          r.manifestFilename,
		nonInteractive:            r.nonInteractive,
		out:                       r.output,
		postBuild:                 r.postBuild,
		spinner:                   r.spinner,
		timeout:                   r.timeout,
		verbose:                   r.verbose,
	}

	return bt.Build()
}

// RustToolchainManifest models a [toolchain] from a rust-toolchain.toml manifest.
type RustToolchainManifest struct {
	Toolchain RustToolchain `toml:"toolchain"`
}

// RustToolchain models the rust-toolchain targets.
type RustToolchain struct {
	Targets []string `toml:"targets"`
}

// modifyCargoPackageName validates whether the --bin flag matches the
// Cargo.toml package name. If it doesn't match, update the default build script
// to match.
func (r *Rust) modifyCargoPackageName(noBuildScript bool) error {
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

	hasCustomBuildScript := !noBuildScript

	switch {
	case m.Package.Name != "":
		// If using standard project structure.
		// Cargo.toml won't be a Workspace, so it will contain a package name.
		r.packageName = m.Package.Name
	case len(m.Workspace.Members) > 0 && noBuildScript:
		// If user has a Cargo Workspace AND no custom script.
		// We need to identify which Workspace package is their application.
		// Then extract the package name from its Cargo.toml manifest.
		// We do this by checking for a rust-toolchain.toml containing a wasm32-wasi target.
		//
		// NOTE: This logic will need to change in the future.
		// Specifically, when we support linking multiple Wasm binaries.
		for _, m := range m.Workspace.Members {
			var rtm RustToolchainManifest
			// G304 (CWE-22): Potential file inclusion via variable.
			// #nosec
			data, err := os.ReadFile(filepath.Join(m, "rust-toolchain.toml"))
			if err != nil {
				return err
			}
			toml.Unmarshal(data, &rtm)
			if len(rtm.Toolchain.Targets) > 0 && rtm.Toolchain.Targets[0] == "wasm32-wasi" {
				var cm CargoManifest
				cm.Read(filepath.Join(m, "Cargo.toml"))
				r.packageName = cm.Package.Name
			}
		}
	case len(m.Workspace.Members) > 0 && hasCustomBuildScript:
		// If user has a Cargo Workspace AND a custom script.
		// Trust their custom script aligns with the relevant Workspace package name.
		// i.e. we parse the package name specified in their custom script.
		parts := strings.Split(r.build, " ")
		for i, p := range parts {
			if p == "--bin" {
				r.packageName = parts[i+1]
				break
			}
		}
	}

	// Ensure the default build script matches the Cargo.toml package name.
	if noBuildScript && r.packageName != "" && r.packageName != RustDefaultPackageName {
		r.build = fmt.Sprintf(RustDefaultBuildCommand, r.packageName)
	}

	return nil
}

// toolchainConstraint warns the user if the required constraint is not met.
//
// NOTE: We don't stop the build as their toolchain may compile successfully.
// The warning is to help a user know something isn't quite right and gives them
// the opportunity to do something about it if they choose.
func (r *Rust) toolchainConstraint() {
	if r.verbose {
		text.Info(r.output, "The Fastly CLI requires a Rust version '%s'.\n\n", r.config.ToolchainConstraint)
	}

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

	if !c.Check(v) {
		text.Warning(r.output, "The Rust version '%s' didn't meet the constraint '%s'\n\n", version, r.config.ToolchainConstraint)
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

	src := filepath.Join(metadata.TargetDirectory, r.config.WasmWasiTarget, "release", fmt.Sprintf("%s.wasm", r.packageName))
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
	Package   CargoPackage   `toml:"package"`
	Workspace CargoWorkspace `toml:"workspace"`
}

// Read the contents of the Cargo.toml manifest from filename.
func (m *CargoManifest) Read(path string) error {
	// gosec flagged this:
	// G304 (CWE-22): Potential file inclusion via variable.
	// Disabling as we need to load the Cargo.toml from the user's file system.
	// This file is decoded into a predefined struct, any unrecognised fields are dropped.
	// #nosec
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return toml.Unmarshal(data, m)
}

// CargoWorkspace models the [workspace] config inside Cargo.toml
type CargoWorkspace struct {
	Members []string `toml:"members" json:"members"`
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
