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

// PythonDefaultBuildCommand is the default build command for Python projects.
// Uses uv (modern Python package manager) to run fastly-compute-py build tool.
var PythonDefaultBuildCommand = fmt.Sprintf("uv run fastly-compute-py build -o %s", binWasmPath)

// PythonSourceDirectory represents the source code directory.
// Python projects typically use the root directory rather than a src/ subdirectory.
const PythonSourceDirectory = "."

// PythonManifest is the manifest file for Python projects.
const PythonManifest = "pyproject.toml"

// NewPython constructs a new Python toolchain.
func NewPython(
	c *BuildCommand,
	in io.Reader,
	manifestFilename string,
	out io.Writer,
	spinner text.Spinner,
) *Python {
	return &Python{
		Shell: Shell{},

		autoYes:               c.Globals.Flags.AutoYes,
		build:                 c.Globals.Manifest.File.Scripts.Build,
		config:                c.Globals.Config.Language.Python,
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

// Python implements a Toolchain for the Python language.
type Python struct {
	Shell

	// autoYes is the --auto-yes flag.
	autoYes bool
	// build is a shell command defined in fastly.toml using [scripts.build].
	build string
	// config is the Python language configuration from config.toml.
	config config.Python
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
	// spinner is a terminal progress status indicator.
	spinner text.Spinner
	// timeout is the build execution threshold.
	timeout int
	// verbose indicates if the user set --verbose
	verbose bool
}

// DefaultBuildScript indicates if a custom build script was used.
func (p *Python) DefaultBuildScript() bool {
	return p.defaultBuild
}

// Dependencies returns the project's Python package dependencies.
//
// For Python, fastly-compute-py injects fastly_data directly into the Wasm
// during the build, so the CLI does not need to collect dependencies separately.
// The CLI's AnnotateWasmBinaryLong merges with the existing fastly_data rather
// than overwriting it.
func (p *Python) Dependencies() map[string]string {
	return make(map[string]string)
}

// Build compiles the user's source code into a Wasm binary.
func (p *Python) Build() error {
	// Check Python and UV versions before attempting build
	if err := p.toolchainConstraint(); err != nil {
		return err
	}

	if p.build == "" {
		p.build = PythonDefaultBuildCommand
		p.defaultBuild = true

		if p.verbose {
			text.Info(p.output, "No custom build script found in fastly.toml [scripts.build].\n")
			text.Info(p.output, "Using default build command: %s\n\n", p.build)
		}
	}

	bt := BuildToolchain{
		autoYes:               p.autoYes,
		buildFn:               p.Shell.Build,
		buildScript:           p.build,
		env:                   p.env,
		errlog:                p.errlog,
		in:                    p.input,
		manifestFilename:      p.manifestFilename,
		metadataFilterEnvVars: p.metadataFilterEnvVars,
		nonInteractive:        p.nonInteractive,
		out:                   p.output,
		postBuild:             p.postBuild,
		spinner:               p.spinner,
		timeout:               p.timeout,
		verbose:               p.verbose,
	}

	return bt.Build()
}

// toolchainConstraint validates that the Python and UV toolchains meet version requirements.
func (p *Python) toolchainConstraint() error {
	// Check Python version
	if err := p.checkPythonVersion(); err != nil {
		return err
	}

	// Check UV is installed
	if err := p.checkUVInstalled(); err != nil {
		return err
	}

	return nil
}

// checkPythonVersion validates the Python version meets the minimum requirement.
func (p *Python) checkPythonVersion() error {
	requiredConstraint := p.config.ToolchainConstraint
	if requiredConstraint == "" {
		requiredConstraint = ">= 3.11"
	}

	if p.verbose {
		text.Info(p.output, "Checking Python version (required: %s)...\n\n", requiredConstraint)
	}

	// Try 'python' first, then 'python3'
	var cmd *exec.Cmd
	var stdout []byte
	var err error

	cmd = exec.Command("python", "--version")
	stdout, err = cmd.CombinedOutput()
	if err != nil {
		// Try python3
		cmd = exec.Command("python3", "--version")
		stdout, err = cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("Python not found in PATH. Please install Python %s or later", requiredConstraint)
		}
	}

	output := strings.TrimSpace(string(stdout))

	// Parse "Python 3.12.11" format
	versionPattern := regexp.MustCompile(`Python (?P<version>\d+\.\d+\.\d+)`)
	match := versionPattern.FindStringSubmatch(output)
	if len(match) < 2 {
		return fmt.Errorf("unable to parse Python version from: %s", output)
	}

	versionStr := match[1]
	v, err := semver.NewVersion(versionStr)
	if err != nil {
		return fmt.Errorf("invalid Python version '%s': %w", versionStr, err)
	}

	c, err := semver.NewConstraint(requiredConstraint)
	if err != nil {
		return fmt.Errorf("invalid toolchain constraint '%s': %w", requiredConstraint, err)
	}

	valid, errs := c.Validate(v)
	if !valid {
		var errMsgs []string
		for _, e := range errs {
			errMsgs = append(errMsgs, e.Error())
		}
		return fmt.Errorf("Python version %s does not satisfy constraint %s: %s", v, requiredConstraint, strings.Join(errMsgs, ", "))
	}

	if p.verbose {
		text.Success(p.output, "Python version %s meets requirement %s\n\n", v, requiredConstraint)
	}

	return nil
}

// checkUVInstalled validates that UV package manager is installed.
func (p *Python) checkUVInstalled() error {
	if p.verbose {
		text.Info(p.output, "Checking for UV package manager...\n\n")
	}

	cmd := exec.Command("uv", "--version")
	output, err := cmd.Output()
	if err != nil {
		text.Break(p.output)
		text.Error(p.output, "UV package manager not found in PATH.\n")
		text.Info(p.output, "\nUV is required to build Python applications for Fastly Compute.\n")
		text.Info(p.output, "Install UV: https://docs.astral.sh/uv/\n")
		text.Info(p.output, "\nQuick install:\n")
		text.Info(p.output, "  macOS/Linux: curl -LsSf https://astral.sh/uv/install.sh | sh\n")
		text.Info(p.output, "  Windows:     powershell -c \"irm https://astral.sh/uv/install.ps1 | iex\"\n\n")
		return fmt.Errorf("uv not found in PATH")
	}

	if p.verbose {
		uvVersion := strings.TrimSpace(string(output))
		text.Success(p.output, "UV found: %s\n\n", uvVersion)
	}

	return nil
}
