package compute

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
)

// JsCompilation is a language specific compilation target that converts the
// language code into a Wasm binary.
const JsCompilation = "js-compute-runtime"

// JsCompilationCommandRemediation is the command to execute to fix the missing
// compilation target.
const JsCompilationCommandRemediation = "npm install --save-dev %s"

// JsCompilationURL is the official Fastly C@E JS runtime package URL.
const JsCompilationURL = "https://www.npmjs.com/package/@fastly/js-compute"

// JsDefaultBuildCommand is a build command compiled into the CLI binary so it
// can be used as a fallback for customer's who have an existing C@E project and
// are simply upgrading their CLI version and might not be familiar with the
// changes in the 4.0.0 release with regards to how build logic has moved to the
// fastly.toml manifest.
const JsDefaultBuildCommand = "npm exec js-compute-runtime ./src/index.js ./bin/main.wasm"

// JsDefaultBuildCommandForWebpack is a build command compiled into the CLI
// binary so it can be used as a fallback for customer's who have an existing
// C@E project using the 'default' JS Starter Kit, and are simply upgrading
// their CLI version and might not be familiar with the changes in the 4.0.0
// release with regards to how build logic has moved to the fastly.toml manifest.
//
// NOTE: For this variation of the build script to be added to the user's
// fastly.toml will require a successful check for the npm task:
// `prebuild: webpack` in the user's package.json manifest.
const JsDefaultBuildCommandForWebpack = "npm exec webpack && npm exec js-compute-runtime ./bin/index.js ./bin/main.wasm"

// JsInstaller is the command used to install the dependencies defined within
// the Js language manifest.
const JsInstaller = "npm install"

// JsManifest is the manifest file for defining project configuration.
const JsManifest = "package.json"

// JsManifestRemediation is a error remediation message for a missing manifest.
const JsManifestRemediation = "npm init"

// JsSDK is the required Compute@Edge SDK.
// https://www.npmjs.com/package/@fastly/js-compute
const JsSDK = "@fastly/js-compute"

// JsSourceDirectory represents the source code directory.                                               │                                                           │
const JsSourceDirectory = "src"

// JsToolchain is the executable responsible for managing dependencies.
const JsToolchain = "npm"

// JsToolchainURL is the official JS website URL.
const JsToolchainURL = "https://nodejs.org/"

// NewJavaScript constructs a new JavaScript toolchain.
func NewJavaScript(
	fastlyManifest *manifest.File,
	errlog fsterr.LogInterface,
	timeout int,
	out io.Writer,
	ch chan string,
) *JavaScript {
	return &JavaScript{
		Shell:     Shell{},
		errlog:    errlog,
		postBuild: fastlyManifest.Scripts.PostBuild,
		timeout:   timeout,
		validator: ToolchainValidator{
			Compilation:             JsCompilation,
			CompilationIntegrated:   true,
			CompilationSkipVersion:  true,
			CompilationURL:          JsCompilationURL,
			DefaultBuildCommand:     JsDefaultBuildCommand,
			ErrLog:                  errlog,
			FastlyManifestFile:      fastlyManifest,
			Installer:               JsInstaller,
			Manifest:                JsManifest,
			ManifestRemediation:     JsManifestRemediation,
			Output:                  out,
			PatchedManifestNotifier: ch,
			SDK:                     JsSDK,
			SDKCustomValidator:      validateJsSDK,
			Toolchain:               JsToolchain,
			ToolchainLanguage:       "JavaScript",
			ToolchainSkipVersion:    true,
			ToolchainURL:            JsToolchainURL,
		},
	}
}

// JavaScript implements a Toolchain for the JavaScript language.
type JavaScript struct {
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
func (j JavaScript) Initialize(_ io.Writer) error {
	return nil
}

// Verify ensures the user's environment has all the required resources/tools.
func (j JavaScript) Verify(_ io.Writer) error {
	return j.validator.Validate()
}

// Build compiles the user's source code into a Wasm binary.
func (j JavaScript) Build(out io.Writer, progress text.Progress, verbose bool, callback func() error) error {
	// NOTE: We deliberately reference the validator pointer to the fastly.toml
	// This is because the manifest.File might be updated when migrating a
	// pre-existing project to use the CLI v4.0.0 (as prior to this version the
	// manifest would not require [script.build] to be defined).
	// As of v4.0.0 if no value is set, then we provide a default.
	return build(buildOpts{
		buildScript: j.validator.FastlyManifestFile.Scripts.Build,
		buildFn:     j.Shell.Build,
		errlog:      j.errlog,
		postBuild:   j.postBuild,
		timeout:     j.timeout,
	}, out, progress, verbose, nil, callback)
}

// JsPackage represents a package.json manifest.
type JsPackage struct {
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
	Scripts         map[string]string `json:"scripts"`
}

// validateJsSDK marshals the JS manifest into JSON to check if the dependency
// has been defined in the package.json manifest.
//
// NOTE: This function also causes a side-effect of modifying the default build
// script based on the user's project context (e.g does it require webpack).
func validateJsSDK(name string, bs []byte, notifier chan string) error {
	e := fmt.Errorf(SDKErrMessageFormat, name, JsManifest)

	var p JsPackage

	err := json.Unmarshal(bs, &p)
	if err != nil {
		return fsterr.RemediationError{
			Inner:       fmt.Errorf("failed to unmarshal package.json: %w", err),
			Remediation: fmt.Sprintf("Ensure your package.json is valid and contains the '%s' dependency.", name),
		}
	}

	var needsWebpack bool
	for k, v := range p.Scripts {
		if k == "prebuild" && strings.Contains(v, "webpack") {
			needsWebpack = true
			break
		}
	}

	go func() {
		if needsWebpack {
			notifier <- JsDefaultBuildCommandForWebpack
		} else {
			notifier <- JsDefaultBuildCommand
		}
	}()

	for k := range p.Dependencies {
		if k == name {
			return nil
		}
	}
	for k := range p.DevDependencies {
		if k == name {
			return nil
		}
	}

	return e
}
