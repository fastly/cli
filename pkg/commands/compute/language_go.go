package compute

import (
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strings"

	"github.com/Masterminds/semver/v3"

	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
)

// TinyGoDefaultBuildCommand is a build command compiled into the CLI binary so it
// can be used as a fallback for customer's who have an existing C@E project and
// are simply upgrading their CLI version and might not be familiar with the
// changes in the 4.0.0 release with regards to how build logic has moved to the
// fastly.toml manifest.
//
// NOTE: In the 5.x CLI releases we persisted the default to the fastly.toml
// We no longer do that. In 6.x we use the default and just inform the user.
// This makes the experience less confusing as users didn't expect file changes.
const TinyGoDefaultBuildCommand = "tinygo build -target=wasi -gc=conservative -o bin/main.wasm ./"

// GoSourceDirectory represents the source code directory.                                               │                                                           │
const GoSourceDirectory = "."

// NewGo constructs a new Go toolchain.
func NewGo(
	fastlyManifest *manifest.File,
	globals *global.Data,
	flags Flags,
	in io.Reader,
	out io.Writer,
	spinner text.Spinner,
) *Go {
	return &Go{
		Shell: Shell{},

		autoYes:        globals.Flags.AutoYes,
		build:          fastlyManifest.Scripts.Build,
		config:         globals.Config.Language.Go,
		errlog:         globals.ErrLog,
		input:          in,
		nonInteractive: globals.Flags.NonInteractive,
		output:         out,
		postBuild:      fastlyManifest.Scripts.PostBuild,
		spinner:        spinner,
		timeout:        flags.Timeout,
		toolchain:      fastlyManifest.Scripts.Toolchain,
		verbose:        globals.Verbose(),
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
	// errlog is an abstraction for recording errors to disk.
	errlog fsterr.LogInterface
	// input is the user's terminal stdin stream
	input io.Reader
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
	// toolchain indicates the toolchain used for compiling your program.
	toolchain string
	// verbose indicates if the user set --verbose
	verbose bool
}

// Build compiles the user's source code into a Wasm binary.
func (g *Go) Build() error {
	var (
		tinygoToolchain     bool
		toolchainConstraint string
	)

	if g.build == "" {
		g.build = TinyGoDefaultBuildCommand
		tinygoToolchain = true
		toolchainConstraint = g.config.ToolchainConstraintTinyGo
		text.Info(g.output, "No [scripts.build] found in fastly.toml. Visit https://developer.fastly.com/learning/compute/go/ to learn how to target native Go vs TinyGo.")
		text.Break(g.output)
		text.Description(g.output, "The following default build command for TinyGo will be used:", g.build)
		text.Break(g.output)
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

	g.toolchainConstraint(
		"go", `go version go(?P<version>\d[^\s]+)`, toolchainConstraint,
	)

	if tinygoToolchain {
		g.toolchainConstraint(
			"tinygo", `tinygo version (?P<version>\d[^\s]+)`, g.config.TinyGoConstraint,
		)
	}

	bt := BuildToolchain{
		autoYes:        g.autoYes,
		buildFn:        g.Shell.Build,
		buildScript:    g.build,
		env:            []string{"GOARCH=wasm", "GOOS=wasip1"},
		errlog:         g.errlog,
		in:             g.input,
		nonInteractive: g.nonInteractive,
		out:            g.output,
		postBuild:      g.postBuild,
		spinner:        g.spinner,
		timeout:        g.timeout,
		verbose:        g.verbose,
	}

	return bt.Build()
}

// toolchainConstraint warns the user if the required constraint is not met.
//
// NOTE: We don't stop the build as their toolchain may compile successfully.
// The warning is to help a user know something isn't quite right and gives them
// the opportunity to do something about it if they choose.
func (g *Go) toolchainConstraint(toolchain, pattern, constraint string) {
	if g.verbose {
		text.Info(g.output, "The Fastly CLI requires a %s version '%s'. ", toolchain, constraint)
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
		text.Warning(g.output, "The %s version '%s' didn't meet the constraint '%s'", toolchain, version, constraint)
		text.Break(g.output)
	}
}
