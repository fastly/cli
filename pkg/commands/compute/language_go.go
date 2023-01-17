package compute

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"

	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
)

// GoCompilation is a language specific compilation target that converts the
// language code into a Wasm binary.
const GoCompilation = "tinygo"

// GoCompilationURL is the official TinyGo website URL.
const GoCompilationURL = "https://tinygo.org"

// GoCompilationTargetCommand is the shell command for returning the tinygo
// version.
const GoCompilationTargetCommand = "tinygo version"

// GoConstraints is the set of supported toolchain and compilation versions.
//
// NOTE: Two keys are supported: "toolchain" and "compilation", with the latter
// being optional as not all language compilation steps are separate tools from
// the toolchain itself.
var GoConstraints = make(map[string]string)

// GoDefaultBuildCommand is a build command compiled into the CLI binary so it
// can be used as a fallback for customer's who have an existing C@E project and
// are simply upgrading their CLI version and might not be familiar with the
// changes in the 4.0.0 release with regards to how build logic has moved to the
// fastly.toml manifest.
const GoDefaultBuildCommand = "tinygo build -target=wasi -gc=conservative -o bin/main.wasm ./"

// GoInstaller is the command used to install the dependencies defined within
// the Go language manifest.
const GoInstaller = "go mod download"

// GoManifest is the manifest file for defining project configuration.
const GoManifest = "go.mod"

// GoManifestCommand is the toolchain command to validate the manifest exists,
// and also enables parsing of the project's dependencies.
const GoManifestCommand = "go mod edit -json"

// GoManifestRemediation is a error remediation message for a missing manifest.
const GoManifestRemediation = "go mod init"

// GoSDK is the required Compute@Edge SDK.
// https://pkg.go.dev/github.com/fastly/compute-sdk-go
const GoSDK = "github.com/fastly/compute-sdk-go"

// GoSourceDirectory represents the source code directory.                                               │                                                           │
const GoSourceDirectory = "."

// GoToolchain is the executable responsible for managing dependencies.
const GoToolchain = "go"

// GoToolchainURL is the official Go website URL.
const GoToolchainURL = "https://go.dev/"

// GoToolchainVersionCommand is the shell command for returning the go version.
const GoToolchainVersionCommand = "go version"

// NewGo constructs a new Go toolchain.
func NewGo(
	fastlyManifest *manifest.File,
	errlog fsterr.LogInterface,
	timeout int,
	cfg config.Go,
	out io.Writer,
	ch chan string,
) *Go {
	GoConstraints["toolchain"] = cfg.ToolchainConstraint
	GoConstraints["compilation"] = cfg.TinyGoConstraint

	return &Go{
		Shell:     Shell{},
		errlog:    errlog,
		postBuild: fastlyManifest.Scripts.PostBuild,
		timeout:   timeout,
		validator: ToolchainValidator{
			Compilation:              GoCompilation,
			CompilationTargetCommand: GoCompilationTargetCommand,
			CompilationTargetPattern: regexp.MustCompile(`tinygo version (?P<version>\d[^\s]+)`),
			CompilationURL:           GoCompilationURL,
			Constraints:              GoConstraints,
			DefaultBuildCommand:      GoDefaultBuildCommand,
			ErrLog:                   errlog,
			FastlyManifestFile:       fastlyManifest,
			Installer:                GoInstaller,
			Manifest:                 GoManifest,
			ManifestCommand:          GoManifestCommand,
			ManifestRemediation:      GoManifestRemediation,
			Output:                   out,
			PatchedManifestNotifier:  ch,
			SDK:                      GoSDK,
			SDKCustomValidator:       validateGoSDK,
			Toolchain:                GoToolchain,
			ToolchainLanguage:        "Go",
			ToolchainVersionCommand:  GoToolchainVersionCommand,
			ToolchainVersionPattern:  regexp.MustCompile(`go version go(?P<version>\d[^\s]+)`),
			ToolchainURL:             GoToolchainURL,
		},
	}
}

// Go implements a Toolchain for the TinyGo language.
//
// NOTE: Two separate tools are required to support golang development.
//
// 1. Go: for defining required packages in a go.mod project module.
// 2. TinyGo: used to compile the go project.
type Go struct {
	Shell

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

// Initialize handles any non-build related set-up.
func (g Go) Initialize(_ io.Writer) error {
	return nil
}

// Verify ensures the user's environment has all the required resources/tools.
func (g *Go) Verify(_ io.Writer) error {
	return g.validator.Validate()
}

// Build compiles the user's source code into a Wasm binary.
func (g *Go) Build(out io.Writer, progress text.Progress, verbose bool, callback func() error) error {
	// NOTE: We deliberately reference the validator pointer to the fastly.toml
	// This is because the manifest.File might be updated when migrating a
	// pre-existing project to use the CLI v4.0.0 (as prior to this version the
	// manifest would not require [script.build] to be defined).
	// As of v4.0.0 if no value is set, then we provide a default.
	return build(buildOpts{
		buildScript: g.validator.FastlyManifestFile.Scripts.Build,
		buildFn:     g.Shell.Build,
		errlog:      g.errlog,
		postBuild:   g.postBuild,
		timeout:     g.timeout,
	}, out, progress, verbose, nil, callback)
}

// GoDependency represents the project's SDK and version.
type GoDependency struct {
	Path    string
	Version string
}

// GoMod represents the project's go.mod manifest.
type GoMod struct {
	Require []GoDependency
}

// validateGoSDK uses the Go toolchain to identify if the required SDK
// dependency is installed.
func validateGoSDK(sdk string, manifestCommandOutput []byte, _ chan string) error {
	var gm GoMod

	err := json.Unmarshal(manifestCommandOutput, &gm)
	if err != nil {
		return fmt.Errorf("failed to unmarshal manifest metadata: %w", err)
	}

	remediation := fmt.Sprintf("Ensure your %s is valid and contains the '%s' dependency.", GoManifest, sdk)

	if len(gm.Require) < 1 {
		return fsterr.RemediationError{
			Inner:       errors.New("no dependencies declared"),
			Remediation: remediation,
		}
	}

	for _, gd := range gm.Require {
		if gd.Path == sdk {
			return nil
		}
	}

	return fsterr.RemediationError{
		Inner:       errors.New("required dependency missing"),
		Remediation: remediation,
	}
}
