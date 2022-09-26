package compute

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/filesystem"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	toml "github.com/pelletier/go-toml"
)

// RustCompilation is a language specific compilation target that converts the
// language code into a Wasm binary.
const RustCompilation = "wasm32-wasi"

// RustCompilationURL is the specification URL for the wasm32-wasi target.
const RustCompilationURL = "https://doc.rust-lang.org/stable/nightly-rustc/rustc_target/spec/wasm32_wasi/index.html"

// RustCompilationCommandRemediation is the command to execute to fix the
// missing compilation target.
const RustCompilationCommandRemediation = "rustup target add %s --toolchain $ACTIVE_TOOLCHAIN"

// RustCompilationTargetCommand is the shell command for returning the list of
// installed compilation targets.
const RustCompilationTargetCommand = "rustup target list --installed --toolchain %s"

// RustConstraints is the set of supported toolchain and compilation versions.
//
// NOTE: Two keys are supported: "toolchain" and "compilation", with the latter
// being optional as not all language compilation steps are separate tools from
// the toolchain itself.
var RustConstraints = make(map[string]string)

// RustDefaultBuildCommand is a build command compiled into the CLI binary so it
// can be used as a fallback for customer's who have an existing C@E project and
// are simply upgrading their CLI version and might not be familiar with the
// changes in the 4.0.0 release with regards to how build logic has moved to the
// fastly.toml manifest.
const RustDefaultBuildCommand = "cargo build --bin compute-starter-kit-rust --release --target wasm32-wasi --color always"

// RustManifest is the manifest file for defining project configuration.
const RustManifest = "Cargo.toml"

// RustManifestRemediation is a error remediation message for a missing manifest.
const RustManifestRemediation = "cargo new $NAME --bin"

// RustSDK is the required Compute@Edge SDK.
// https://crates.io/crates/fastly
const RustSDK = "fastly"

// RustSourceDirectory represents the source code directory.                                               │                                                           │
const RustSourceDirectory = "src"

// RustToolchain is the executable responsible for managing dependencies.
const RustToolchain = "cargo"

// RustToolchainCommandRemediation is the command to execute to fix the
// toolchain.
const RustToolchainCommandRemediation = "Run `rustup update stable`, or ensure your `rust-toolchain` file specifies a version matching the constraint (e.g. `channel = \"stable\"`)."

// RustToolchainURL is the official Rust website URL.
const RustToolchainURL = "https://doc.rust-lang.org/stable/cargo/"

// RustToolchainVersionCommand is the shell command for returning the Rust
// version.
const RustToolchainVersionCommand = "cargo version"

// NewRust constructs a new Rust toolchain.
func NewRust(
	fastlyManifest *manifest.File,
	errlog fsterr.LogInterface,
	timeout int,
	cfg config.Rust,
	out io.Writer,
	ch chan string,
) *Rust {
	RustConstraints["toolchain"] = cfg.ToolchainConstraint

	return &Rust{
		Shell:     Shell{},
		config:    cfg,
		errlog:    errlog,
		postBuild: fastlyManifest.Scripts.PostBuild,
		timeout:   timeout,
		validator: ToolchainValidator{
			Compilation:                   RustCompilation,
			CompilationIntegrated:         true,
			CompilationCommandRemediation: fmt.Sprintf(RustCompilationCommandRemediation, cfg.WasmWasiTarget),
			CompilationTargetCommand:      fmt.Sprintf(RustCompilationTargetCommand, rustupToolchain()),
			CompilationTargetPattern:      regexp.MustCompile(fmt.Sprintf(`(?P<version>)%s`, RustCompilation)),
			CompilationURL:                RustCompilationURL,
			Constraints:                   RustConstraints,
			DefaultBuildCommand:           RustDefaultBuildCommand,
			ErrLog:                        errlog,
			FastlyManifestFile:            fastlyManifest,
			Manifest:                      RustManifest,
			ManifestRemediation:           RustManifestRemediation,
			Output:                        out,
			PatchedManifestNotifier:       ch,
			SDK:                           RustSDK,
			SDKCustomValidator:            validateRustSDK,
			Toolchain:                     RustToolchain,
			ToolchainCommandRemediation:   RustToolchainCommandRemediation,
			ToolchainLanguage:             "Rust",
			ToolchainVersionCommand:       RustToolchainVersionCommand,
			ToolchainVersionPattern:       regexp.MustCompile(`cargo (?P<version>\d[^\s]+)`),
			ToolchainURL:                  RustToolchainURL,
		},
	}
}

// Rust implements a Toolchain for the Rust language.
type Rust struct {
	Shell

	// config is the Rust specific application configuration.
	config config.Rust
	// errlog is an abstraction for recording errors to disk.
	errlog fsterr.LogInterface
	// postBuild is a custom script executed after the build but before the Wasm
	// binary is added to the .tar.gz archive.
	postBuild string
	// timeout is the build execution threshold.
	timeout int
	// validator is an abstraction to validate required resources are installed.
	validator ToolchainValidator
}

// Initialize implements the Toolchain interface and initializes a newly cloned
// package. It is a noop for Rust as the Cargo toolchain handles these steps.
func (r Rust) Initialize(_ io.Writer) error {
	return nil
}

// Verify ensures the user's environment has all the required resources/tools.
func (r *Rust) Verify(_ io.Writer) error {
	return r.validator.Validate()
}

// Build compiles the user's source code into a Wasm binary.
func (r *Rust) Build(out io.Writer, progress text.Progress, verbose bool, callback func() error) error {
	// NOTE: We purposes reference the validator pointer to the fastly.toml file.
	// This is because the manifest.File might be updated when migrating a
	// pre-existing project to use the CLI v4.0.0 (as prior to this version the
	// manifest would not require [script.build] to be defined. As of v4.0.0 if no
	// value is set, then we provide a default.
	return build(language{
		buildScript: r.validator.FastlyManifestFile.Scripts.Build,
		buildFn:     r.Shell.Build,
		errlog:      r.errlog,
		postBuild:   r.postBuild,
		timeout:     r.timeout,
	}, out, progress, verbose, r.ProcessLocation, callback)
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
	if err := m.Read("Cargo.toml"); err != nil {
		return fmt.Errorf("error reading Cargo.toml manifest: %w", err)
	}
	src := filepath.Join(metadata.TargetDirectory, r.config.WasmWasiTarget, "release", fmt.Sprintf("%s.wasm", m.Package.Name))
	dst := filepath.Join(dir, "bin", "main.wasm")

	err = filesystem.CopyFile(src, dst)
	if err != nil {
		r.errlog.Add(err)
		return fmt.Errorf("copying wasm binary: %w", err)
	}
	return nil
}

// rustupToolchain returns the active rustup toolchain and falls back to stable
// if there was an error.
func rustupToolchain() string {
	stable := "stable"
	cmd := []string{"rustup", "show", "active-toolchain"}
	c := exec.Command(cmd[0], cmd[1:]...) // #nosec G204
	stdoutStderr, err := c.CombinedOutput()
	if err != nil {
		return stable
	}

	reader := bufio.NewReader(bytes.NewReader(stdoutStderr))
	line, err := reader.ReadString('\n')
	if err != nil {
		return stable
	}

	// Example outputs:
	// stable-x86_64-apple-darwin (default)
	// 1.54.0-x86_64-apple-darwin (directory override for '/Users/integralist/Code/fastly/cli')
	parts := strings.Split(line, "-")
	if len(parts) < 2 {
		return "stable"
	}

	return parts[0]
}

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

// validateRustSDK marshals the Rust manifest into toml to check if the
// dependency has been defined in the Cargo.toml manifest.
func validateRustSDK(name string, bs []byte) error {
	e := fmt.Errorf(SDKErrMessageFormat, name, RustManifest)

	tree, err := toml.LoadBytes(bs)
	if err != nil {
		return e
	}

	if v, ok := tree.GetArray("dependencies").(*toml.Tree); ok {
		if dependency, ok := v.Get("fastly").(string); ok && dependency != "" {
			return nil
		}
	}

	return e
}
