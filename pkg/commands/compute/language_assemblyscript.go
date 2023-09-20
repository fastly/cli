package compute

import (
	"encoding/json"
	"errors"
	"io"
	"os"

	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
)

// AsDefaultBuildCommand is a build command compiled into the CLI binary so it
// can be used as a fallback for customer's who have an existing Compute project and
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
// Compute project using the 'default' JS Starter Kit, and are simply upgrading
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
	globals *global.Data,
	flags Flags,
	in io.Reader,
	manifestFilename string,
	out io.Writer,
	spinner text.Spinner,
) *AssemblyScript {
	return &AssemblyScript{
		Shell: Shell{},

		build:            fastlyManifest.Scripts.Build,
		errlog:           globals.ErrLog,
		input:            in,
		manifestFilename: manifestFilename,
		output:           out,
		postBuild:        fastlyManifest.Scripts.PostBuild,
		spinner:          spinner,
		timeout:          flags.Timeout,
		verbose:          globals.Verbose(),
	}
}

// AssemblyScript implements a Toolchain for the AssemblyScript language.
type AssemblyScript struct {
	Shell

	// autoYes is the --auto-yes flag.
	autoYes bool
	// build is a shell command defined in fastly.toml using [scripts.build].
	build string
	// defaultBuild indicates if the default build script was used.
	defaultBuild bool
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
	// postBuild is a custom script executed after the build but before the Wasm
	// binary is added to the .tar.gz archive.
	postBuild string
	// spinner is a terminal progress status indicator.
	spinner text.Spinner
	// timeout is the build execution threshold.
	timeout int
	// verbose indicates if the user set --verbose
	verbose bool
}

// DefaultBuildScript indicates if a custom build script was used.
func (a *AssemblyScript) DefaultBuildScript() bool {
	return a.defaultBuild
}

// Dependencies returns all dependencies used by the project.
func (a *AssemblyScript) Dependencies() map[string]string {
	deps := make(map[string]string)

	lockfile := "npm-shrinkwrap.json"
	_, err := os.Stat(lockfile)
	if errors.Is(err, os.ErrNotExist) {
		lockfile = "package-lock.json"
	}

	var jlf JavaScriptLockFile
	if f, err := os.Open(lockfile); err == nil {
		if err := json.NewDecoder(f).Decode(&jlf); err == nil {
			for k, v := range jlf.Packages {
				if k != "" { // avoid "root" package
					deps[k] = v.Version
				}
			}
		}
	}

	return deps
}

// Build compiles the user's source code into a Wasm binary.
func (a *AssemblyScript) Build() error {
	if !a.verbose {
		text.Break(a.output)
	}
	text.Deprecated(a.output, "The Fastly AssemblyScript SDK is being deprecated in favor of the more up-to-date and feature-rich JavaScript SDK. You can learn more about the JavaScript SDK on our Developer Hub Page - https://developer.fastly.com/learning/compute/javascript/\n\n")

	if a.build == "" {
		a.build = AsDefaultBuildCommand
		a.defaultBuild = true
	}

	usesWebpack, err := a.checkForWebpack()
	if err != nil {
		return err
	}
	if usesWebpack {
		a.build = AsDefaultBuildCommandForWebpack
	}

	if a.defaultBuild && a.verbose {
		text.Info(a.output, "No [scripts.build] found in %s. The following default build command for AssemblyScript will be used: `%s`\n\n", a.manifestFilename, a.build)
	}

	bt := BuildToolchain{
		autoYes:          a.autoYes,
		buildFn:          a.Shell.Build,
		buildScript:      a.build,
		errlog:           a.errlog,
		in:               a.input,
		manifestFilename: a.manifestFilename,
		nonInteractive:   a.nonInteractive,
		out:              a.output,
		postBuild:        a.postBuild,
		spinner:          a.spinner,
		timeout:          a.timeout,
		verbose:          a.verbose,
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
