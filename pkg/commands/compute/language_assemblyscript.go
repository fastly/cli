package compute

import (
	"fmt"
	"io"

	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
)

// AsCompilation is a language specific compilation target that converts the
// language code into a Wasm binary.
const AsCompilation = "asc"

// AsCompilationCommandRemediation is the command to execute to fix the missing
// compilation target.
const AsCompilationCommandRemediation = "npm install --save-dev %s"

// AsCompilationURL is the official assemblyscript package URL.
const AsCompilationURL = "https://www.npmjs.com/package/assemblyscript"

// AsDefaultBuildCommand is a build command compiled into the CLI binary so it
// can be used as a fallback for customer's who have an existing C@E project and
// are simply upgrading their CLI version and might not be familiar with the
// changes in the 4.0.0 release with regards to how build logic has moved to the
// fastly.toml manifest.
const AsDefaultBuildCommand = "%s exec -- asc assembly/index.ts --target release"

// AsInstaller is the command used to install the dependencies defined within
// the Js language manifest.
const AsInstaller = "%s install"

// AsManifest is the manifest file for defining project configuration.
const AsManifest = "package.json"

// AsManifestCommand is the toolchain command to validate the manifest exists,
// and also enables parsing of the project's dependencies.
const AsManifestCommand = "npm list --json --depth 0"

// AsManifestRemediation is a error remediation message for a missing manifest.
const AsManifestRemediation = "%s init"

// AsSDK is the required Compute@Edge SDK.
// https://www.npmjs.com/package/@fastly/as-compute
const AsSDK = "@fastly/as-compute"

// AsSourceDirectory represents the source code directory.
const AsSourceDirectory = "assembly"

// AsToolchain is the executable responsible for managing dependencies.
const AsToolchain = "npm"

// AsToolchainURL is the official JS website URL.
const AsToolchainURL = "https://nodejs.org/"

// NewAssemblyScript constructs a new AssemblyScript toolchain.
func NewAssemblyScript(
	fastlyManifest *manifest.File,
	errlog fsterr.LogInterface,
	timeout int,
	out io.Writer,
	ch chan string,
) *AssemblyScript {
	installerPreHook := ""
	asCompilationCommandRemediation := AsCompilationCommandRemediation
	asManifestCommand := AsManifestCommand
	asManifestMetaData := AsManifestCommand // npm only needs AsManifestCommand (yarn needs a separate command)
	asSkipErr := true                       // Refer to NOTE under ManifestCommandSkipError
	asToolchain := AsToolchain
	asToolchainURL := AsToolchainURL
	packageManager := "npm"

	if fastlyManifest.PackageManager == "yarn" {
		asCompilationCommandRemediation = "yarn add --dev %s"
		asManifestCommand = "yarn workspaces list"
		asToolchain = fastlyManifest.PackageManager
		asToolchainURL = "https://yarnpkg.com/"
		asManifestMetaData = "yarn info --json"
		asSkipErr = false // we're unsetting as yarn doesn't need it, though npm does
		installerPreHook = "yarn config set nodeLinker node-modules"
		packageManager = fastlyManifest.PackageManager
	}

	// Dynamically insert the package manager name.
	asDefaultBuildCommand := fmt.Sprintf(AsDefaultBuildCommand, packageManager)
	asInstaller := fmt.Sprintf(AsInstaller, packageManager)
	asManifestRemediation := fmt.Sprintf(AsManifestRemediation, packageManager)

	a := &AssemblyScript{
		JavaScript: JavaScript{
			errlog:  errlog,
			timeout: timeout,
			validator: ToolchainValidator{
				Compilation:                   AsCompilation,
				CompilationCommandRemediation: fmt.Sprintf(asCompilationCommandRemediation, AsSDK),
				CompilationIntegrated:         true,
				CompilationSkipVersion:        true,
				CompilationURL:                AsCompilationURL,
				DefaultBuildCommand:           asDefaultBuildCommand,
				ErrLog:                        errlog,
				FastlyManifestFile:            fastlyManifest,
				Installer:                     asInstaller,
				Manifest:                      AsManifest,
				ManifestExist:                 asManifestCommand,
				ManifestExistSkipError:        asSkipErr,
				ManifestMetaData:              asManifestMetaData,
				ManifestRemediation:           asManifestRemediation,
				Output:                        out,
				PatchedManifestNotifier:       ch,
				SDK:                           AsSDK,
				SDKCustomValidator:            validateJsSDK(packageManager, asDefaultBuildCommand, ""),
				Toolchain:                     asToolchain,
				ToolchainLanguage:             "AssemblyScript",
				ToolchainSkipVersion:          true,
				ToolchainURL:                  asToolchainURL,
				InstallerPreHook:              installerPreHook,
			},
		},
		errlog:    errlog,
		postBuild: fastlyManifest.Scripts.PostBuild,
	}

	return a
}

// AssemblyScript implements a Toolchain for the AssemblyScript language.
//
// NOTE: We embed the JavaScript type as the behaviours across both languages
// are fundamentally the same with some minor differences. This means we don't
// need to duplicate the Verify() implementation, while the Build() method can
// be kept unique between the two languages. Additionally the JavaScript
// Verify() method has an extra validation step that is skipped for
// AssemblyScript as it doesn't set the `validateScriptBuild` field.
type AssemblyScript struct {
	JavaScript

	// errlog is an abstraction for recording errors to disk.
	errlog fsterr.LogInterface
	// postBuild is a custom script executed after the build but before the Wasm
	// binary is added to the .tar.gz archive.
	postBuild string
}

// Build compiles the user's source code into a Wasm binary.
func (a AssemblyScript) Build(out io.Writer, progress text.Progress, verbose bool, callback func() error) error {
	// NOTE: We deliberately reference the validator pointer to the fastly.toml
	// This is because the manifest.File might be updated when migrating a
	// pre-existing project to use the CLI v4.0.0 (as prior to this version the
	// manifest would not require [script.build] to be defined).
	// As of v4.0.0 if no value is set, then we provide a default.
	return build(buildOpts{
		buildScript: a.validator.FastlyManifestFile.Scripts.Build,
		buildFn:     a.Shell.Build,
		errlog:      a.errlog,
		postBuild:   a.postBuild,
		timeout:     a.timeout,
	}, out, progress, verbose, nil, callback)
}
