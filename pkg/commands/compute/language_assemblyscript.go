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

// AsCompilationURL is the official assemblyscript package URL.
const AsCompilationURL = "https://www.npmjs.com/package/assemblyscript"

// AsSDK is the required Compute@Edge SDK.
// https://www.npmjs.com/package/@fastly/as-compute
const AsSDK = "@fastly/as-compute"

// AsSourceDirectory represents the source code directory.
const AsSourceDirectory = "assembly"

// AsDefaultBuildCommand is a build command compiled into the CLI binary so it
// can be used as a fallback for customer's who have an existing C@E project and
// are simply upgrading their CLI version and might not be familiar with the
// changes in the 4.0.0 release with regards to how build logic has moved to the
// fastly.toml manifest.
var AsDefaultBuildCommand = "%s exec -- asc assembly/index.ts --target release"

// NewAssemblyScript constructs a new AssemblyScript toolchain.
func NewAssemblyScript(
	fastlyManifest *manifest.File,
	errlog fsterr.LogInterface,
	timeout int,
	out io.Writer,
	ch chan string,
) *AssemblyScript {
	packageManager := "npm"
	asSkipErr := true                       // Refer to NOTE under ManifestCommandSkipError
	asManifestMetaData := JsManifestCommand // npm only needs JsManifestCommand (yarn needs a separate command)

	if fastlyManifest.PackageManager == "yarn" {
		JsCompilationCommandRemediation = "yarn add --dev %s"
		JsManifestCommand = "yarn workspaces list"
		asManifestMetaData = "yarn info --json"
		asSkipErr = false // we're unsetting as yarn doesn't need it, though npm does
		JsToolchain = fastlyManifest.PackageManager
		JsToolchainURL = "https://yarnpkg.com/"
		packageManager = fastlyManifest.PackageManager
	}

	// Dynamically insert the package manager name.
	AsDefaultBuildCommand = fmt.Sprintf(AsDefaultBuildCommand, packageManager)
	JsInstaller = fmt.Sprintf(JsInstaller, packageManager)

	a := &AssemblyScript{
		JavaScript: JavaScript{
			errlog:  errlog,
			timeout: timeout,
			validator: ToolchainValidator{
				Compilation:                   AsCompilation,
				CompilationIntegrated:         true,
				CompilationCommandRemediation: fmt.Sprintf(JsCompilationCommandRemediation, AsSDK),
				CompilationSkipVersion:        true,
				CompilationURL:                AsCompilationURL,
				DefaultBuildCommand:           AsDefaultBuildCommand,
				ErrLog:                        errlog,
				FastlyManifestFile:            fastlyManifest,
				Installer:                     JsInstaller,
				Manifest:                      JsManifest,
				ManifestExist:                 JsManifestCommand,
				ManifestExistSkipError:        asSkipErr,
				ManifestMetaData:              asManifestMetaData,
				ManifestRemediation:           JsManifestRemediation,
				Output:                        out,
				PatchedManifestNotifier:       ch,
				SDK:                           AsSDK,
				SDKCustomValidator:            validateJsSDK(packageManager, AsDefaultBuildCommand, ""),
				Toolchain:                     JsToolchain,
				ToolchainLanguage:             "AssemblyScript",
				ToolchainSkipVersion:          true,
				ToolchainURL:                  JsToolchainURL,
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
