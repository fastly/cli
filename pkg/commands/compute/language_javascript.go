package compute

import (
	"bytes"
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

// JsCompilationURL is the official Fastly C@E JS runtime package URL.
const JsCompilationURL = "https://www.npmjs.com/package/@fastly/js-compute"

// JsManifest is the manifest file for defining project configuration.
const JsManifest = "package.json"

// JsSDK is the required Compute@Edge SDK.
// https://www.npmjs.com/package/@fastly/js-compute
const JsSDK = "@fastly/js-compute"

// JsSourceDirectory represents the source code directory.                                               │                                                           │
const JsSourceDirectory = "src"

// JsCompilationCommandRemediation is the command to execute to fix the missing
// compilation target.
var JsCompilationCommandRemediation = "npm install --save-dev %s"

// JsDefaultBuildCommand is a build command compiled into the CLI binary so it
// can be used as a fallback for customer's who have an existing C@E project and
// are simply upgrading their CLI version and might not be familiar with the
// changes in the 4.0.0 release with regards to how build logic has moved to the
// fastly.toml manifest.
var JsDefaultBuildCommand = "%s exec js-compute-runtime ./src/index.js ./bin/main.wasm"

// JsDefaultBuildCommandForWebpack is a build command compiled into the CLI
// binary so it can be used as a fallback for customer's who have an existing
// C@E project using the 'default' JS Starter Kit, and are simply upgrading
// their CLI version and might not be familiar with the changes in the 4.0.0
// release with regards to how build logic has moved to the fastly.toml manifest.
//
// NOTE: For this variation of the build script to be added to the user's
// fastly.toml will require a successful check for the npm task:
// `prebuild: webpack` in the user's package.json manifest.
var JsDefaultBuildCommandForWebpack = "%s exec webpack && %s exec js-compute-runtime ./bin/index.js ./bin/main.wasm"

// JsInstaller is the command used to install the dependencies defined within
// the Js language manifest.
var JsInstaller = "%s install"

// JsManifestCommand is the toolchain command to validate the manifest exists,
// and also enables parsing of the project's dependencies.
var JsManifestCommand = "npm list --json --depth 0"

// JsManifestRemediation is a error remediation message for a missing manifest.
var JsManifestRemediation = "%s init"

// JsToolchain is the executable responsible for managing dependencies.
var JsToolchain = "npm"

// JsToolchainURL is the official JS website URL.
var JsToolchainURL = "https://nodejs.org/"

// NewJavaScript constructs a new JavaScript toolchain.
func NewJavaScript(
	fastlyManifest *manifest.File,
	errlog fsterr.LogInterface,
	timeout int,
	out io.Writer,
	ch chan string,
) *JavaScript {
	packageManager := "npm"
	jsSkipErr := true                       // Refer to NOTE under ManifestCommandSkipError
	jsManifestMetaData := JsManifestCommand // npm only needs JsManifestCommand (yarn needs a separate command)
	installerPreHook := ""

	if fastlyManifest.PackageManager == "yarn" {
		JsCompilationCommandRemediation = "yarn add --dev %s"
		JsManifestCommand = "yarn workspaces list"
		jsManifestMetaData = "yarn info --json"
		jsSkipErr = false // we're unsetting as yarn doesn't need it, though npm does
		JsToolchain = fastlyManifest.PackageManager
		JsToolchainURL = "https://yarnpkg.com/"
		packageManager = fastlyManifest.PackageManager
		installerPreHook = "yarn config set nodeLinker node-modules"
	}

	// Dynamically insert the package manager name.
	JsDefaultBuildCommand = fmt.Sprintf(JsDefaultBuildCommand, packageManager)
	JsDefaultBuildCommandForWebpack = fmt.Sprintf(JsDefaultBuildCommandForWebpack, packageManager, packageManager)
	JsInstaller = fmt.Sprintf(JsInstaller, packageManager)
	JsManifestRemediation = fmt.Sprintf(JsManifestRemediation, packageManager)

	return &JavaScript{
		Shell:     Shell{},
		errlog:    errlog,
		postBuild: fastlyManifest.Scripts.PostBuild,
		timeout:   timeout,
		validator: ToolchainValidator{
			Compilation:                   JsCompilation,
			CompilationIntegrated:         true,
			CompilationCommandRemediation: fmt.Sprintf(JsCompilationCommandRemediation, JsSDK),
			CompilationSkipVersion:        true,
			CompilationURL:                JsCompilationURL,
			DefaultBuildCommand:           JsDefaultBuildCommand,
			ErrLog:                        errlog,
			FastlyManifestFile:            fastlyManifest,
			Installer:                     JsInstaller,
			InstallerPreHook:              installerPreHook,
			Manifest:                      JsManifest,
			ManifestExist:                 JsManifestCommand,
			ManifestExistSkipError:        jsSkipErr,
			ManifestMetaData:              jsManifestMetaData,
			ManifestRemediation:           JsManifestRemediation,
			Output:                        out,
			PatchedManifestNotifier:       ch,
			SDK:                           JsSDK,
			SDKCustomValidator:            validateJsSDK(packageManager, JsDefaultBuildCommand, JsDefaultBuildCommandForWebpack),
			Toolchain:                     JsToolchain,
			ToolchainLanguage:             "JavaScript",
			ToolchainSkipVersion:          true,
			ToolchainURL:                  JsToolchainURL,
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

// NPMDependency represents a JS dependency.
type NPMDependency struct {
	Version string `json:"version"`
}

// NPMPackage represents a package.json manifest and its dependencies.
type NPMPackage struct {
	Dependencies map[string]NPMDependency `json:"dependencies"`
}

func validateWithNPM(sdk string, manifestCommandOutput []byte, notifier chan string, buildCommand, buildCommandForWebpack string, e error) error {
	var p NPMPackage

	err := json.Unmarshal(manifestCommandOutput, &p)
	if err != nil {
		return fsterr.RemediationError{
			Inner:       fmt.Errorf("failed to unmarshal package.json: %w", err),
			Remediation: fmt.Sprintf("Ensure your package.json is valid and contains the '%s' dependency.", sdk),
		}
	}

	var needsWebpack bool
	for k := range p.Dependencies {
		if k == "webpack" {
			needsWebpack = true
			break
		}
	}

	go func() {
		if needsWebpack {
			notifier <- buildCommandForWebpack
		} else {
			notifier <- buildCommand
		}
	}()

	for k := range p.Dependencies {
		if k == sdk {
			return nil
		}
	}

	return e
}

// YarnPackage represents a Yarn data structure for dependency metadata.
type YarnPackage struct {
	Value string `json:"value"`
}

// NOTE: Yarn produces a stream of JSON.
func validateWithYarn(sdk string, manifestCommandOutput []byte, notifier chan string, buildCommand, buildCommandForWebpack string, e error) error {
	dec := json.NewDecoder(bytes.NewReader(manifestCommandOutput))

	var needsWebpack bool
	for {
		var yp YarnPackage
		if err := dec.Decode(&yp); err == io.EOF {
			break
		} else if err != nil {
			return fsterr.RemediationError{
				Inner:       fmt.Errorf("failed to unmarshal yarn metadata: %w", err),
				Remediation: fmt.Sprintf("Ensure your package.json is valid and contains the '%s' dependency.", sdk),
			}
		}
		if strings.HasPrefix(yp.Value, "webpack@") {
			needsWebpack = true
			break
		}
	}

	go func() {
		if needsWebpack {
			notifier <- buildCommandForWebpack
		} else {
			notifier <- buildCommand
		}
	}()

	dec = json.NewDecoder(bytes.NewReader(manifestCommandOutput))

	for {
		var yp YarnPackage
		if err := dec.Decode(&yp); err == io.EOF {
			break
		} else if err != nil {
			return fsterr.RemediationError{
				Inner:       fmt.Errorf("failed to unmarshal yarn metadata: %w", err),
				Remediation: fmt.Sprintf("Ensure your package.json is valid and contains the '%s' dependency.", sdk),
			}
		}
		if strings.HasPrefix(yp.Value, sdk+"@") {
			return nil
		}
	}

	return e
}

// validateJsSDK marshals the JS manifest into JSON to check if the dependency
// has been defined in the package.json manifest.
//
// NOTE: This function also causes a side-effect of modifying the default build
// script based on the user's project context (e.g does it require webpack).
func validateJsSDK(packageManager, buildCommand, buildCommandForWebpack string) func(string, []byte, chan string) error {
	return func(sdk string, manifestCommandOutput []byte, notifier chan string) error {
		e := fmt.Errorf(SDKErrMessageFormat, sdk, JsManifest)

		if packageManager == "yarn" {
			return validateWithYarn(sdk, manifestCommandOutput, notifier, buildCommand, buildCommandForWebpack, e)
		}
		return validateWithNPM(sdk, manifestCommandOutput, notifier, buildCommand, buildCommandForWebpack, e)
	}
}
