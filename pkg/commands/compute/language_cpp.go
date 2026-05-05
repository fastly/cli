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
	"github.com/fastly/cli/pkg/text"
)

// CPPDefaultBuildCommand is a build command compiled into the CLI binary so it
// can be used as a fallback for customers who have an existing Compute project and
// are simply upgrading their CLI version and might not be familiar with the
// changes in the 4.0.0 release with regards to how build logic has moved to the
// fastly.toml manifest.
//
// NOTE: In the 5.x CLI releases we persisted the default to the fastly.toml
// We no longer do that. In 6.x we use the default and just inform the user.
// This makes the experience less confusing as users didn't expect file changes.
const CPPDefaultBuildCommand = "clang++ -O3 --target=%s -o %s main.cpp"

// CPPDefaultWasmWasiTarget is the expected C++ WasmWasi build target.
const CPPDefaultWasmWasiTarget = "wasm32-wasip1"

// CPPSourceDirectory represents the source code directory.
const CPPSourceDirectory = "."

// NewCPP constructs a new C++ toolchain.
func NewCPP(
	c *BuildCommand,
	in io.Reader,
	manifestFilename string,
	out io.Writer,
	spinner text.Spinner,
) *CPP {
	return &CPP{
		Shell: Shell{},

		autoYes:               c.Globals.Flags.AutoYes,
		build:                 c.Globals.Manifest.File.Scripts.Build,
		config:                c.Globals.Config.Language.CPP,
		env:                   c.Globals.Manifest.File.Scripts.EnvVars,
		errlog:                c.Globals.ErrLog,
		input:                 in,
		manifestFilename:      manifestFilename,
		metadataFilterEnvVars: c.MetadataFilterEnvVars,
		nonInteractive:        c.Globals.Flags.NonInteractive,
		output:                out,
		postBuild:             c.Globals.Manifest.File.Scripts.PostBuild,
		spinner:               spinner,
		timeout:               c.Flags.Timeout,
		verbose:               c.Globals.Verbose(),
	}
}

// CPP implements a Toolchain for the C++ language.
type CPP struct {
	Shell

	// autoYes is the --auto-yes flag.
	autoYes bool
	// build is a shell command defined in fastly.toml using [scripts.build].
	build string
	// config is the C++ specific application configuration.
	config config.CPP
	// defaultBuild indicates if the default build script was used.
	defaultBuild bool
	// env is environment variables to be set.
	env []string
	// errlog is an abstraction for recording errors to disk.
	errlog fsterr.LogInterface
	// input is the user's terminal stdin stream.
	input io.Reader
	// manifestFilename is the name of the manifest file.
	manifestFilename string
	// metadataFilterEnvVars is a comma-separated list of user defined env vars.
	metadataFilterEnvVars string
	// nonInteractive is the --non-interactive flag.
	nonInteractive bool
	// output is the user's terminal stdout stream.
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
func (cpp *CPP) DefaultBuildScript() bool {
	return cpp.defaultBuild
}

// Dependencies returns all dependencies used by the project.
func (cpp *CPP) Dependencies() map[string]string {
	// For C++, dependencies are typically managed through various systems
	// (CMake, Conan, vcpkg, etc.). For now, return an empty map.
	// This could be extended in the future to parse CMakeLists.txt or other files.
	return make(map[string]string)
}

// Build compiles the user's source code into a Wasm binary.
func (cpp *CPP) Build() error {
	if cpp.build == "" {
		cpp.build = fmt.Sprintf(CPPDefaultBuildCommand, CPPDefaultWasmWasiTarget, binWasmPath)
		cpp.defaultBuild = true
		if !cpp.verbose {
			text.Break(cpp.output)
		}
		text.Info(cpp.output, "No [scripts.build] found in %s. Visit https://www.fastly.com/documentation/guides/compute/ to learn more about building C++ projects.\n\n", cpp.manifestFilename)
		text.Description(cpp.output, "The following default build command for C++ will be used", cpp.build)
	}

	cpp.toolchainConstraint(
		"clang++", `clang version (?P<version>\d+\.\d+\.\d+)`, cpp.config.ToolchainConstraint,
	)

	wasmWasiTarget := cpp.config.WasmWasiTarget
	if wasmWasiTarget != "" && wasmWasiTarget != CPPDefaultWasmWasiTarget {
		return fmt.Errorf("the default build in .fastly/config.toml should produce a %s binary, but was instead set to produce a %s binary", CPPDefaultWasmWasiTarget, wasmWasiTarget)
	}

	bt := BuildToolchain{
		autoYes:               cpp.autoYes,
		buildFn:               cpp.Shell.Build,
		buildScript:           cpp.build,
		env:                   cpp.env,
		errlog:                cpp.errlog,
		in:                    cpp.input,
		manifestFilename:      cpp.manifestFilename,
		metadataFilterEnvVars: cpp.metadataFilterEnvVars,
		nonInteractive:        cpp.nonInteractive,
		out:                   cpp.output,
		postBuild:             cpp.postBuild,
		spinner:               cpp.spinner,
		timeout:               cpp.timeout,
		verbose:               cpp.verbose,
	}

	return bt.Build()
}

// toolchainConstraint warns the user if the required constraint is not met.
//
// NOTE: We don't stop the build as their toolchain may compile successfully.
// The warning is to help a user know something isn't quite right and gives them
// the opportunity to do something about it if they choose.
func (cpp *CPP) toolchainConstraint(toolchain, pattern, constraint string) {
	if constraint == "" {
		return
	}

	if cpp.verbose {
		text.Info(cpp.output, "The Fastly CLI build step requires a %s version '%s'.\n\n", toolchain, constraint)
	}

	versionCommand := fmt.Sprintf("%s --version", toolchain)
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

	valid, errs := c.Validate(v)
	if !valid {
		text.Warning(cpp.output, "The %s version requirement was not satisfied: %v", toolchain, errs)
	}
}
