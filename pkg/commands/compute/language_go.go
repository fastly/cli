package compute

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/Masterminds/semver/v3"
	"golang.org/x/mod/modfile"

	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// TinyGoDefaultBuildCommand is a build command compiled into the CLI binary so it
// can be used as a fallback for customer's who have an existing Compute project and
// are simply upgrading their CLI version and might not be familiar with the
// changes in the 4.0.0 release with regards to how build logic has moved to the
// fastly.toml manifest.
//
// NOTE: In the 5.x CLI releases we persisted the default to the fastly.toml
// We no longer do that. In 6.x we use the default and just inform the user.
// This makes the experience less confusing as users didn't expect file changes.
var TinyGoDefaultBuildCommand = fmt.Sprintf("tinygo build -target=wasi -gc=conservative -o %s ./", binWasmPath)

// GoSourceDirectory represents the source code directory.
const GoSourceDirectory = "."

// NewGo constructs a new Go toolchain.
func NewGo(
	c *BuildCommand,
	in io.Reader,
	manifestFilename string,
	out io.Writer,
	spinner text.Spinner,
) *Go {
	return &Go{
		Shell: Shell{},

		autoYes:               c.Globals.Flags.AutoYes,
		build:                 c.Globals.Manifest.File.Scripts.Build,
		config:                c.Globals.Config.Language.Go,
		env:                   c.Globals.Manifest.File.Scripts.EnvVars,
		errlog:                c.Globals.ErrLog,
		input:                 in,
		manifestFilename:      manifestFilename,
		metadataFilterEnvVars: c.MetadataFilterEnvVars,
		nonInteractive:        c.Globals.Flags.NonInteractive,
		output:                out,
		postBuild:             c.Globals.Manifest.File.Scripts.PostBuild,
		serveFile:             c.ServeFile,
		spinner:               spinner,
		timeout:               c.Flags.Timeout,
		verbose:               c.Globals.Verbose(),
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

	// autoYes is the --auto-yes flag.
	autoYes bool
	// build is a shell command defined in fastly.toml using [scripts.build].
	build string
	// config is the Go specific application configuration.
	config config.Go
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
	// metadataFilterEnvVars is a comma-separated list of user defined env vars.
	metadataFilterEnvVars string
	// nonInteractive is the --non-interactive flag.
	nonInteractive bool
	// output is the users terminal stdout stream
	output io.Writer
	// postBuild is a custom script executed after the build but before the Wasm
	// binary is added to the .tar.gz archive.
	postBuild string
	// serveFile indicates if --file was passed as part of `compute serve`.
	serveFile bool
	// spinner is a terminal progress status indicator.
	spinner text.Spinner
	// timeout is the build execution threshold.
	timeout int
	// verbose indicates if the user set --verbose
	verbose bool
}

// DefaultBuildScript indicates if a custom build script was used.
func (g *Go) DefaultBuildScript() bool {
	return g.defaultBuild
}

// Dependencies returns all dependencies used by the project.
func (g *Go) Dependencies() map[string]string {
	deps := make(map[string]string)
	data, err := os.ReadFile("go.mod")
	if err != nil {
		return deps
	}
	f, err := modfile.ParseLax("go.mod", data, nil)
	if err != nil {
		return deps
	}
	for _, req := range f.Require {
		if req.Indirect {
			continue
		}
		deps[req.Mod.Path] = req.Mod.Version
	}
	return deps
}

// Build compiles the user's source code into a Wasm binary.
func (g *Go) Build() error {
	var (
		tinygoToolchain     bool
		toolchainConstraint string
	)

	if g.build == "" {
		g.build = TinyGoDefaultBuildCommand
		g.defaultBuild = true
		tinygoToolchain = true
		toolchainConstraint = g.config.ToolchainConstraintTinyGo
		if !g.verbose {
			text.Break(g.output)
		}
		text.Info(g.output, "No [scripts.build] found in %s. Visit https://developer.fastly.com/learning/compute/go/ to learn how to target standard Go vs TinyGo.\n\n", g.manifestFilename)
		text.Description(g.output, "The following default build command for TinyGo will be used", g.build)
	}

	if g.build != "" {
		// IMPORTANT: All Fastly starter-kits for Go/TinyGo will have build script.
		//
		// So we'll need to parse the build script to identify if TinyGo is used so
		// we can set the constraints appropriately.
		if strings.Contains(g.build, "tinygo build") {
			tinygoToolchain = true
			toolchainConstraint = g.config.ToolchainConstraintTinyGo
		} else {
			toolchainConstraint = g.config.ToolchainConstraint
		}
	}

	// IMPORTANT: The Go SDK 0.2.0 bumps the tinygo requirement to 0.28.1
	//
	// This means we need to check the go.mod of the user's project for
	// `compute-sdk-go` and then parse the version and identify if it's less than
	// 0.2.0 version. If it less than, change the TinyGo constraint to 0.26.0
	tinygoConstraint := identifyTinyGoConstraint(g.config.TinyGoConstraint, g.config.TinyGoConstraintFallback)

	g.toolchainConstraint(
		"go", `go version go(?P<version>\d[^\s]+)`, toolchainConstraint,
	)

	if tinygoToolchain {
		g.toolchainConstraint(
			"tinygo", `tinygo version (?P<version>\d[^\s]+)`, tinygoConstraint,
		)
	}

	bt := BuildToolchain{
		autoYes:               g.autoYes,
		buildFn:               g.Shell.Build,
		buildScript:           g.build,
		env:                   g.env,
		errlog:                g.errlog,
		in:                    g.input,
		manifestFilename:      g.manifestFilename,
		metadataFilterEnvVars: g.metadataFilterEnvVars,
		nonInteractive:        g.nonInteractive,
		out:                   g.output,
		postBuild:             g.postBuild,
		serveFile:             g.serveFile,
		spinner:               g.spinner,
		timeout:               g.timeout,
		verbose:               g.verbose,
	}

	return bt.Build()
}

// identifyTinyGoConstraint checks the compute-sdk-go version used by the
// project and if it's less than 0.2.0 we'll change the TinyGo constraint to be
// version 0.26.0
//
// We do this because the 0.2.0 release of the compute-sdk-go bumps the TinyGo
// version requirement to 0.28.1 and we want to avoid any scenarios where a
// bump in SDK version causes the user's build to break (which would happen for
// users with a pre-existing project who happen to update their CLI version: the
// new CLI version would have a TinyGo constraint that would be higher than
// before and would stop their build from working).
//
// NOTE: The `configConstraint` is the latest CLI application config version.
// If there are any errors trying to parse the go.mod we'll default to the
// config constraint.
func identifyTinyGoConstraint(configConstraint, fallback string) string {
	moduleName := "github.com/fastly/compute-sdk-go"
	version := ""

	f, err := os.Open("go.mod")
	if err != nil {
		return configConstraint
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)

		// go.mod has two separate definition possibilities:
		//
		// 1.
		// require github.com/fastly/compute-sdk-go v0.1.7
		//
		// 2.
		// require (
		//   github.com/fastly/compute-sdk-go v0.1.7
		// )
		if len(parts) >= 2 {
			// 1. require [github.com/fastly/compute-sdk-go] v0.1.7
			if parts[1] == moduleName {
				version = strings.TrimPrefix(parts[2], "v")
				break
			}
			// 2. [github.com/fastly/compute-sdk-go] v0.1.7
			if parts[0] == moduleName {
				version = strings.TrimPrefix(parts[1], "v")
				break
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return configConstraint
	}

	if version == "" {
		return configConstraint
	}

	gomodVersion, err := semver.NewVersion(version)
	if err != nil {
		return configConstraint
	}

	// 0.2.0 introduces the break by bumping the TinyGo minimum version to 0.28.1
	breakingSDKVersion, err := semver.NewVersion("0.2.0")
	if err != nil {
		return configConstraint
	}

	if gomodVersion.LessThan(breakingSDKVersion) {
		return fallback
	}

	return configConstraint
}

// toolchainConstraint warns the user if the required constraint is not met.
//
// NOTE: We don't stop the build as their toolchain may compile successfully.
// The warning is to help a user know something isn't quite right and gives them
// the opportunity to do something about it if they choose.
func (g *Go) toolchainConstraint(toolchain, pattern, constraint string) {
	if g.verbose {
		text.Info(g.output, "The Fastly CLI build step requires a %s version '%s'.\n\n", toolchain, constraint)
	}

	versionCommand := fmt.Sprintf("%s version", toolchain)
	args := strings.Split(versionCommand, " ")

	// gosec flagged this:
	// G204 (CWE-78): Subprocess launched with function call as argument or cmd arguments
	// Disabling as we trust the source of the variable.
	// #nosec
	// nosemgrep
	cmd := exec.Command(args[0], args[1:]...)
	stdoutStderr, err := cmd.CombinedOutput()
	output := string(stdoutStderr)
	if err != nil {
		return
	}

	versionPattern := regexp.MustCompile(pattern)
	match := versionPattern.FindStringSubmatch(output)
	if len(match) < 2 { // We expect a pattern with one capture group.
		return
	}
	version := match[1]

	v, err := semver.NewVersion(version)
	if err != nil {
		return
	}

	c, err := semver.NewConstraint(constraint)
	if err != nil {
		return
	}

	if !c.Check(v) {
		text.Warning(g.output, "The %s version '%s' didn't meet the constraint '%s'\n\n", toolchain, version, constraint)
	}
}
