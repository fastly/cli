package compute

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"

	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
)

// JsDefaultBuildCommand is a build command compiled into the CLI binary so it
// can be used as a fallback for customer's who have an existing Compute project and
// are simply upgrading their CLI version and might not be familiar with the
// changes in the 4.0.0 release with regards to how build logic has moved to the
// fastly.toml manifest.
//
// NOTE: In the 5.x CLI releases we persisted the default to the fastly.toml
// We no longer do that. In 6.x we use the default and just inform the user.
// This makes the experience less confusing as users didn't expect file changes.
const JsDefaultBuildCommand = "npm exec js-compute-runtime ./src/index.js ./bin/main.wasm"

// JsDefaultBuildCommandForWebpack is a build command compiled into the CLI
// binary so it can be used as a fallback for customer's who have an existing
// Compute project using the 'default' JS Starter Kit, and are simply upgrading
// their CLI version and might not be familiar with the changes in the 4.0.0
// release with regards to how build logic has moved to the fastly.toml manifest.
//
// NOTE: For this variation of the build script to be added to the user's
// fastly.toml will require a successful check for the webpack dependency.
const JsDefaultBuildCommandForWebpack = "npm exec webpack && npm exec js-compute-runtime ./bin/index.js ./bin/main.wasm"

// JsSourceDirectory represents the source code directory.                                               │                                                           │
const JsSourceDirectory = "src"

// NewJavaScript constructs a new JavaScript toolchain.
func NewJavaScript(
	fastlyManifest *manifest.File,
	globals *global.Data,
	flags Flags,
	in io.Reader,
	manifestFilename string,
	out io.Writer,
	spinner text.Spinner,
) *JavaScript {
	return &JavaScript{
		Shell: Shell{},

		autoYes:          globals.Flags.AutoYes,
		build:            fastlyManifest.Scripts.Build,
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

// JavaScript implements a Toolchain for the JavaScript language.
type JavaScript struct {
	Shell

	// autoYes is the --auto-yes flag.
	autoYes bool
	// build is a shell command defined in fastly.toml using [scripts.build].
	build string
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
func (j *JavaScript) DefaultBuildScript() bool {
	return j.defaultBuild
}

// Dependencies returns all dependencies used by the project.
func (j *JavaScript) Dependencies() map[string]string {
	deps := make(map[string]string)
	return deps
}

// Build compiles the user's source code into a Wasm binary.
func (j *JavaScript) Build() error {
	if j.build == "" {
		j.build = JsDefaultBuildCommand
		j.defaultBuild = true

		usesWebpack, err := j.checkForWebpack()
		if err != nil {
			return err
		}
		if usesWebpack {
			j.build = JsDefaultBuildCommandForWebpack
		}
	}

	if j.defaultBuild && j.verbose {
		text.Info(j.output, "No [scripts.build] found in %s. The following default build command for JavaScript will be used: `%s`\n\n", j.manifestFilename, j.build)
	}

	bt := BuildToolchain{
		autoYes:          j.autoYes,
		buildFn:          j.Shell.Build,
		buildScript:      j.build,
		env:              j.env,
		errlog:           j.errlog,
		in:               j.input,
		manifestFilename: j.manifestFilename,
		nonInteractive:   j.nonInteractive,
		out:              j.output,
		postBuild:        j.postBuild,
		spinner:          j.spinner,
		timeout:          j.timeout,
		verbose:          j.verbose,
	}

	return bt.Build()
}

func (j JavaScript) checkForWebpack() (bool, error) {
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

// search recurses up the directory tree looking for the given file.
func search(filename, wd, home string) (found bool, path string, err error) {
	parent := filepath.Dir(wd)

	var noManifest bool
	path = filepath.Join(wd, filename)
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		noManifest = true
	}

	// We've found the manifest.
	if !noManifest {
		return true, path, nil
	}

	// NOTE: The first condition catches if we reach the user's 'root' directory.
	if wd != parent && wd != home {
		return search(filename, parent, home)
	}

	return false, "", nil
}

// NPMPackage represents a package.json manifest and its dependencies.
type NPMPackage struct {
	DevDependencies map[string]string `json:"devDependencies"`
	Dependencies    map[string]string `json:"dependencies"`
}
