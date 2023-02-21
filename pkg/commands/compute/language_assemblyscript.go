package compute

import (
	"encoding/json"
	"io"
	"os"

	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/theckman/yacspin"
)

// AsDefaultBuildCommand is a build command compiled into the CLI binary so it
// can be used as a fallback for customer's who have an existing C@E project and
// are simply upgrading their CLI version and might not be familiar with the
// changes in the 4.0.0 release with regards to how build logic has moved to the
// fastly.toml manifest.
//
// NOTE: In the 5.x CLI releases we persisted the default to the fastly.toml
// We no longer do that. In 6.x we use the default and just inform the user.
// This makes the experience less confusing as users didn't expect file changes.
const AsDefaultBuildCommand = "npm exec -- asc assembly/index.ts --outFile bin/main.wasm --optimize --noAssert"

// AsDefaultBuildCommandForWebpack is a build command compiled into the CLI
// binary so it can be used as a fallback for customer's who have an existing
// C@E project using the 'default' JS Starter Kit, and are simply upgrading
// their CLI version and might not be familiar with the changes in the 4.0.0
// release with regards to how build logic has moved to the fastly.toml manifest.
//
// NOTE: For this variation of the build script to be added to the user's
// fastly.toml will require a successful check for the webpack dependency.
const AsDefaultBuildCommandForWebpack = "npm exec webpack && npm exec -- asc assembly/index.ts --outFile bin/main.wasm --optimize --noAssert"

// AsSourceDirectory represents the source code directory.                                               │                                                           │
const AsSourceDirectory = "assembly"

// NewAssemblyScript constructs a new AssemblyScript toolchain.
func NewAssemblyScript(
	fastlyManifest *manifest.File,
	errlog fsterr.LogInterface,
	timeout int,
	out io.Writer,
	verbose bool,
) *AssemblyScript {
	return &AssemblyScript{
		Shell:     Shell{},
		build:     fastlyManifest.Scripts.Build,
		errlog:    errlog,
		output:    out,
		postBuild: fastlyManifest.Scripts.PostBuild,
		timeout:   timeout,
		verbose:   verbose,
	}
}

// AssemblyScript implements a Toolchain for the AssemblyScript language.
type AssemblyScript struct {
	Shell

	// build is a shell command defined in fastly.toml using [scripts.build].
	build string
	// errlog is an abstraction for recording errors to disk.
	errlog fsterr.LogInterface
	// output is the users terminal stdout stream
	output io.Writer
	// postBuild is a custom script executed after the build but before the Wasm
	// binary is added to the .tar.gz archive.
	postBuild string
	// timeout is the build execution threshold.
	timeout int
	// verbose indicates if the user set --verbose
	verbose bool
}

// Build compiles the user's source code into a Wasm binary.
func (a *AssemblyScript) Build(out io.Writer, spinner *yacspin.Spinner, verbose bool, callback func() error) error {
	var noBuildScript bool
	if a.build == "" {
		a.build = AsDefaultBuildCommand
		noBuildScript = true
	}

	usesWebpack, err := a.checkForWebpack()
	if err != nil {
		return err
	}
	if usesWebpack {
		a.build = AsDefaultBuildCommandForWebpack
	}

	if noBuildScript && a.verbose {
		text.Info(out, "No [scripts.build] found in fastly.toml. The following default build command for AssemblyScript will be used: `%s`\n", a.build)
		text.Break(out)
	}

	bt := BuildToolchain{
		buildFn:           a.Shell.Build,
		buildScript:       a.build,
		errlog:            a.errlog,
		postBuild:         a.postBuild,
		timeout:           a.timeout,
		out:               out,
		postBuildCallback: callback,
		spinner:           spinner,
		verbose:           verbose,
	}

	return bt.Build()
}

func (a AssemblyScript) checkForWebpack() (bool, error) {
	wd, err := os.Getwd()
	if err != nil {
		return false, err
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return false, err
	}

	found, path, err := search("package.json", wd, home)
	if err != nil {
		return false, err
	}

	if found {
		// gosec flagged this:
		// G304 (CWE-22): Potential file inclusion via variable
		//
		// Disabling as the path is determined by our own logic.
		/* #nosec */
		data, err := os.ReadFile(path)
		if err != nil {
			return false, err
		}

		var pkg NPMPackage

		err = json.Unmarshal(data, &pkg)
		if err != nil {
			return false, err
		}

		for k := range pkg.DevDependencies {
			if k == "webpack" {
				return true, nil
			}
		}

		for k := range pkg.Dependencies {
			if k == "webpack" {
				return true, nil
			}
		}
	}

	return false, nil
}
