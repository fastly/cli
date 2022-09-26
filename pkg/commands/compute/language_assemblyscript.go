package compute

import (
	"fmt"
	"io"
	"path/filepath"

	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
)

// AsCompilation is a language specific compilation target that converts the
// language code into a Wasm binary.
const AsCompilation = "asc"

// AsCompilationURL is the official assemblyscript package URL.
const AsCompilationURL = "https://www.npmjs.com/package/assemblyscript"

// AsDefaultBuildCommand is a build command compiled into the CLI binary so it
// can be used as a fallback for customer's who have an existing C@E project and
// are simply upgrading their CLI version and might not be familiar with the
// changes in the 4.0.0 release with regards to how build logic has moved to the
// fastly.toml manifest.
const AsDefaultBuildCommand = "$(npm bin)/asc assembly/index.ts --outFile bin/main.wasm --optimize --noAssert"

// AsSDK is the required Compute@Edge SDK.
// https://www.npmjs.com/package/@fastly/as-compute
const AsSDK = "@fastly/as-compute"

// AsSourceDirectory represents the source code directory.
const AsSourceDirectory = "assembly"

// NewAssemblyScript constructs a new AssemblyScript toolchain.
func NewAssemblyScript(
	fastlyManifest *manifest.File,
	errlog fsterr.LogInterface,
	timeout int,
	out io.Writer,
	ch chan string,
) *AssemblyScript {
	return &AssemblyScript{
		JavaScript: JavaScript{
			errlog:  errlog,
			timeout: timeout,
			validator: ToolchainValidator{
				Compilation: AsCompilation,
				CompilationDirectPath: func() (string, error) {
					p, err := getJsToolchainBinPath(JsToolchain)
					if err != nil {
						errlog.Add(err)
						remediation := "npm install --global npm@latest"
						return "", fsterr.RemediationError{
							Inner:       fmt.Errorf("could not determine %s bin path", JsToolchain),
							Remediation: fmt.Sprintf(fsterr.FormatTemplate, text.Bold(remediation)),
						}
					}

					return filepath.Join(p, AsCompilation), nil
				},
				CompilationCommandRemediation: fmt.Sprintf(JsCompilationCommandRemediation, AsSDK),
				CompilationSkipVersion:        true,
				CompilationURL:                AsCompilationURL,
				DefaultBuildCommand:           AsDefaultBuildCommand,
				ErrLog:                        errlog,
				FastlyManifestFile:            fastlyManifest,
				Installer:                     JsInstaller,
				Manifest:                      JsManifest,
				ManifestRemediation:           JsManifestRemediation,
				Output:                        out,
				PatchedManifestNotifier:       ch,
				SDK:                           AsSDK,
				SDKCustomValidator:            validateJsSDK,
				Toolchain:                     JsToolchain,
				ToolchainLanguage:             "AssemblyScript",
				ToolchainSkipVersion:          true,
				ToolchainURL:                  JsToolchainURL,
			},
		},
		errlog:    errlog,
		postBuild: fastlyManifest.Scripts.PostBuild,
	}
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
	// NOTE: We purposes reference the validator pointer to the fastly.toml file.
	// This is because the manifest.File might be updated when migrating a
	// pre-existing project to use the CLI v4.0.0 (as prior to this version the
	// manifest would not require [script.build] to be defined. As of v4.0.0 if no
	// value is set, then we provide a default.
	return build(language{
		buildScript: a.validator.FastlyManifestFile.Scripts.Build,
		buildFn:     a.Shell.Build,
		errlog:      a.errlog,
		postBuild:   a.postBuild,
		timeout:     a.timeout,
	}, out, progress, verbose, nil, callback)
}
