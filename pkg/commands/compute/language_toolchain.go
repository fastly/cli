package compute

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	fsterr "github.com/fastly/cli/pkg/errors"
	fstexec "github.com/fastly/cli/pkg/exec"
	"github.com/fastly/cli/pkg/filesystem"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	toml "github.com/pelletier/go-toml"
)

// SDKErrMessageFormat is a format string that can be used by the
// ToolchainValidator and any other language files that need to implement custom
// validation.
const SDKErrMessageFormat = "failed to find SDK '%s' in the '%s' manifest"

// Toolchain abstracts a Compute@Edge source language toolchain.
type Toolchain interface {
	// Initialize handles any non-build related set-up.
	Initialize(out io.Writer) error

	// Verify ensures the user's environment has all the required resources/tools.
	Verify(out io.Writer) error

	// Build compiles the user's source code into a Wasm binary.
	Build(out io.Writer, progress text.Progress, verbose bool, callback func() error) error
}

// ToolchainValidator represents required tools and files that need to exist.
type ToolchainValidator struct {
	// Compilation is a language specific compilation target that converts the
	// language code into a Wasm binary (e.g. wasm32-wasi, tinygo).
	Compilation string

	// CompilationDirectPath is a language specific function type that returns the
	// direct path to a binary to be looked up rather than the $PATH environment
	// variable. This is typically used for looking up installed NPM binaries that
	// don't exist in the $PATH but the `npm` binary can internally resolve them.
	CompilationDirectPath func() (string, error)

	// CompilationIntegrated is a language specific indicator that the
	// compilation target is integrated with the toolchain and not an external
	// tool (e.g. Rust's wasm32-wasi target).
	CompilationIntegrated bool

	// CompilationCommandRemediation is a language specific shell command a user
	// can execute to fix the missing compilation target.
	CompilationCommandRemediation string

	// CompilationSkipVersion is a language specific indicator that the
	// compilation target does not need to have its version checked against a
	// constraint defined in the CLI's application configuration.
	CompilationSkipVersion bool

	// CompilationTargetCommand is a language specific shell command that returns
	// the Compilation target.
	CompilationTargetCommand string

	// CompilationTargetPattern is a language specific regular expression that
	// validates the compilation target is installed.
	CompilationTargetPattern *regexp.Regexp

	// CompilationURL is a language specific homepage for the compilation target.
	CompilationURL string

	// Constraints is a language specific set of supported toolchain and
	// compilation versions. Two keys are supported: "toolchain" and
	// "compilation", with the latter being optional as not all language
	// compilation steps are separate tools from the toolchain itself.
	Constraints map[string]string

	// DefaultBuildCommand is a build command compiled into the CLI binary so it
	// can be used as a fallback for customer's who have an existing C@E project and
	// are simply upgrading their CLI version and might not be familiar with the
	// changes in the 4.0.0 release with regards to how build logic has moved to the
	// fastly.toml manifest.
	DefaultBuildCommand string

	// ErrLog is used to log any errors to the user's local error log file, which
	// is also persisted to Sentry for Fastly error tracking/managment.
	ErrLog fsterr.LogInterface

	// FastlyManifestFile is a reference to the in-memory manifest.File data
	// structure. The reference is needed so the ToolchainValidator can update the
	// manifest file.
	FastlyManifestFile *manifest.File

	// Installer is a language specific command to install the dependencies
	// defined within the language manifest (e.g. go mod download, npm install).
	Installer string

	// Manifest is a language specific manifest file for defining project
	// configuration (e.g. package.json, Cargo.toml, go.mod).
	Manifest string

	// ManifestRemediation is a language specific error remediation.
	ManifestRemediation string

	// Output is the output buffer to write messages to (typically io.Stdout)
	Output io.Writer

	// SDK is a language specific Compute@Edge compatible SDK (e.g.
	// @fastly/js-compute, compute-sdk-go).
	SDK string

	// SDKCustomValidator allows a supported language to define their own method
	// for how to validate if their manifest contains the required SDK dependency.
	SDKCustomValidator func(name string, bs []byte) error

	// Toolchain is a language specific executable responsible for managing
	// dependencies (e.g. npm, cargo, go).
	Toolchain string

	// ToolchainCommandRemediation is a language specific shell command a user
	// can execute to fix the missing compilation target.
	ToolchainCommandRemediation string

	// ToolchainSkipVersion is a language specific indicator that the
	// toolchain does not need to have its version checked against a constraint
	// defined in the CLI's application configuration.
	ToolchainSkipVersion bool

	// ToolchainURL is a language specific homepage for the toolchain.
	ToolchainURL string

	// ToolchainVersionCommand is a language specific shell command that returns
	// the Toolchain version.
	ToolchainVersionCommand string

	// ToolchainVersionPattern is a language specific regular expression that matches the
	// version of the language Toolchain version command.
	ToolchainVersionPattern *regexp.Regexp
}

// Validate ensures the user's local environment has all required resources.
func (tv ToolchainValidator) Validate() error {
	if err := tv.toolchain(); err != nil {
		return err
	}
	if err := tv.manifestFile(); err != nil {
		return err
	}
	if err := tv.sdk(); err != nil {
		return err
	}
	if err := tv.dependencies(); err != nil {
		return err
	}
	if err := tv.compilation(); err != nil {
		return err
	}
	if err := tv.buildScript(); err != nil {
		return err
	}
	return tv.binDir()
}

// toolchain validates the toolchain is installed.
func (tv ToolchainValidator) toolchain() error {
	fmt.Fprintf(tv.Output, "\nChecking if '%s' is installed...\n", tv.Toolchain)

	bin, err := exec.LookPath(tv.Toolchain)
	if err != nil {
		tv.ErrLog.Add(err)

		return fsterr.RemediationError{
			Inner:       fmt.Errorf("'%s' not found in $PATH", tv.Toolchain),
			Remediation: tv.visitURLRemediation(tv.Toolchain, tv.ToolchainURL),
		}
	}

	fmt.Fprintf(tv.Output, "Found '%s' at %s\n", tv.Toolchain, bin)

	if tv.ToolchainSkipVersion {
		return nil
	}
	return tv.toolchainVersion()
}

// toolchainVersion validates the toolchain/compilation meets the defined
// constraints.
func (tv ToolchainValidator) toolchainVersion() error {
	args := strings.Split(tv.ToolchainVersionCommand, " ")

	// gosec flagged this:
	// G204 (CWE-78): Subprocess launched with function call as argument or cmd arguments
	// Disabling as we trust the source of the variable.
	/* #nosec */
	cmd := exec.Command(args[0], args[1:]...)
	stdoutStderr, err := cmd.CombinedOutput()
	output := string(stdoutStderr)
	if err != nil {
		if len(stdoutStderr) > 0 {
			err = fmt.Errorf("%w: %s", err, strings.TrimSpace(output))
		}
		tv.ErrLog.Add(err)
		return fmt.Errorf("failed to execute command '%s': %w", tv.ToolchainVersionCommand, err)
	}

	match := tv.ToolchainVersionPattern.FindStringSubmatch(output)
	if len(match) < 2 { // We expect a pattern with one capture group.
		err := fmt.Errorf("failed to parse the toolchain version: '%s'", tv.ToolchainVersionPattern)
		tv.ErrLog.Add(err)
		return err
	}
	version := match[1]

	v, err := semver.NewVersion(version)
	if err != nil {
		err = fmt.Errorf("error parsing version output %s into a semver: %w", version, err)
		tv.ErrLog.Add(err)
		return err
	}

	constraint, ok := tv.Constraints["toolchain"]
	if !ok {
		err := fmt.Errorf("failed to lookup the toolchain constraint: '%s'", tv.Constraints)
		tv.ErrLog.Add(err)
		return err
	}

	c, err := semver.NewConstraint(constraint)
	if err != nil {
		err = fmt.Errorf("error parsing toolchain constraint %s into a semver: %w", constraint, err)
		tv.ErrLog.Add(err)
		return err
	}

	if !c.Check(v) {
		err := fsterr.RemediationError{
			Inner:       fmt.Errorf("toolchain version %s didn't meet the constraint %s", version, constraint),
			Remediation: tv.commandRemediation(tv.Toolchain, tv.ToolchainURL, tv.ToolchainCommandRemediation),
		}
		tv.ErrLog.Add(err)
		return err
	}

	return nil
}

// manifestFile validates the language manifestFile can be found.
func (tv ToolchainValidator) manifestFile() error {
	fmt.Fprintf(tv.Output, "\nChecking if manifest '%s' exists...\n", tv.Manifest)

	m, err := filepath.Abs(tv.Manifest)
	if err != nil {
		tv.ErrLog.Add(err)
		return fmt.Errorf("failed to construct path to '%s': %w", tv.Manifest, err)
	}

	if !filesystem.FileExists(m) {
		msg := fmt.Sprintf(fsterr.FormatTemplate, text.Bold(tv.ManifestRemediation))
		remediation := fmt.Sprintf("%s\n\nThen execute:\n\n\t$ fastly compute build", msg)
		err := fsterr.RemediationError{
			Inner:       fmt.Errorf("%s not found", tv.Manifest),
			Remediation: remediation,
		}
		tv.ErrLog.Add(err)
		return err
	}

	fmt.Fprintf(tv.Output, "Found '%s' at %s\n", tv.Manifest, m)

	return nil
}

// sdk validates the relevant Compute@Edge SDK is defined in the language
// specific manifest.
func (tv ToolchainValidator) sdk() error {
	fmt.Fprintf(tv.Output, "\nChecking if '%s' is defined in '%s'...\n", tv.SDK, tv.Manifest)

	m, err := filepath.Abs(tv.Manifest)
	if err != nil {
		tv.ErrLog.Add(err)
		return fmt.Errorf("failed to construct path to '%s': %w", tv.Manifest, err)
	}

	// gosec flagged this:
	// G204 (CWE-78): Subprocess launched with function call as argument or cmd arguments
	// Disabling as we trust the source of the variable.
	/* #nosec */
	bs, err := os.ReadFile(m)
	if err != nil {
		err = fmt.Errorf("failed to read file '%s': %w", m, err)
		tv.ErrLog.Add(err)
		return err
	}

	success := fmt.Sprintf("Found '%s' in '%s'\n", tv.SDK, tv.Manifest)

	if tv.SDKCustomValidator != nil {
		if err := tv.SDKCustomValidator(tv.SDK, bs); err != nil {
			return err
		}
		fmt.Fprint(tv.Output, success)
		return nil
	}

	if !bytes.Contains(bs, []byte(tv.SDK)) {
		err = fmt.Errorf(SDKErrMessageFormat, tv.SDK, m)
		tv.ErrLog.Add(err)
		return err
	}

	fmt.Fprint(tv.Output, success)
	return nil
}

// dependencies will download the language dependencies if a command is provided.
func (tv ToolchainValidator) dependencies() error {
	if tv.Installer != "" {
		fmt.Fprintf(tv.Output, "\nInstalling package dependencies...\n")
		installer := strings.Split(tv.Installer, " ")
		cmd := fstexec.Streaming{
			Command: installer[0],
			Args:    installer[1:],
			Env:     os.Environ(),
			Output:  tv.Output,
		}
		return cmd.Exec()
	}

	return nil
}

// compilation validates the compilation target/tool is installed.
//
// NOTE: Some languages use an external command for compilation, while other
// languages might have the Wasm compilation target integrated to their
// toolchain. If an external command, we lookup the command on the $PATH and
// check its version meets any defined constraints.
func (tv ToolchainValidator) compilation() error {
	fmt.Fprintf(tv.Output, "\nChecking if '%s' is installed...\n", tv.Compilation)

	// Lookup the compilation target as an executable binary.
	if !tv.CompilationIntegrated {
		bin := tv.Compilation

		if tv.CompilationDirectPath != nil {
			var err error
			bin, err = tv.CompilationDirectPath()
			if err != nil {
				return err
			}
		}

		bin, err := exec.LookPath(bin)
		if err != nil {
			tv.ErrLog.Add(err)

			remediation := tv.visitURLRemediation(tv.Compilation, tv.CompilationURL)

			if tv.CompilationCommandRemediation != "" {
				remediation = tv.commandRemediation(tv.Compilation, tv.CompilationURL, tv.CompilationCommandRemediation)
			}

			return fsterr.RemediationError{
				Inner:       fmt.Errorf("'%s' not found", tv.Compilation),
				Remediation: remediation,
			}
		}

		fmt.Fprintf(tv.Output, "Found '%s' at %s\n", tv.Compilation, bin)
	}

	// Some languages (JavaScript/AssemblyScript) don't need their compilation
	// tool version checked so we allow the check to be skipped.
	if tv.CompilationSkipVersion {
		return nil
	}

	args := strings.Split(tv.CompilationTargetCommand, " ")

	// gosec flagged this:
	// G204 (CWE-78): Subprocess launched with function call as argument or cmd arguments
	// Disabling as we trust the source of the variable.
	/* #nosec */
	cmd := exec.Command(args[0], args[1:]...)
	stdoutStderr, err := cmd.CombinedOutput()
	output := string(stdoutStderr)
	if err != nil {
		if len(stdoutStderr) > 0 {
			err = fmt.Errorf("%w: %s", err, strings.TrimSpace(output))
		}
		tv.ErrLog.Add(err)
		return fmt.Errorf("failed to execute command '%s': %w", tv.CompilationTargetCommand, err)
	}

	match := tv.CompilationTargetPattern.FindStringSubmatch(output)
	if len(match) < 2 { // We expect a pattern with one capture group.
		err := fmt.Errorf("failed to find '%s' with the pattern '%s'", tv.Compilation, tv.CompilationTargetPattern)
		tv.ErrLog.Add(err)
		return err
	}

	if tv.CompilationIntegrated {
		fmt.Fprintf(tv.Output, "Found '%s'\n", tv.Compilation)
	}

	// If dealing with an executable binary, check the version constraints.
	if !tv.CompilationIntegrated {
		version := match[1]
		return tv.compilationVersion(version)
	}

	return nil
}

// compilationVersion validates the compilation target version constraints are
// met.
func (tv ToolchainValidator) compilationVersion(version string) error {
	fmt.Fprintf(tv.Output, "\nChecking version constraints for '%s'...\n", tv.Compilation)

	v, err := semver.NewVersion(version)
	if err != nil {
		err = fmt.Errorf("error parsing version output %s into a semver: %w", version, err)
		tv.ErrLog.Add(err)
		return err
	}

	constraint, ok := tv.Constraints["compilation"]
	if !ok {
		err := fmt.Errorf("failed to lookup the compilation constraint: '%s'", tv.Constraints)
		tv.ErrLog.Add(err)
		return err
	}

	c, err := semver.NewConstraint(constraint)
	if err != nil {
		err = fmt.Errorf("error parsing compilation constraint %s into a semver: %w", constraint, err)
		tv.ErrLog.Add(err)
		return err
	}

	if !c.Check(v) {
		err := fsterr.RemediationError{
			Inner:       fmt.Errorf("version %s didn't meet the constraint %s", version, constraint),
			Remediation: tv.visitURLRemediation(tv.Compilation, tv.CompilationURL),
		}
		tv.ErrLog.Add(err)
		return err
	}

	fmt.Fprintf(tv.Output, "Version constraints for '%s' are met\n", tv.Compilation)

	return nil
}

// buildScript validates the Fastly manifest contains a scripts.build value.
func (tv ToolchainValidator) buildScript() error {
	fmt.Fprintf(tv.Output, "\nChecking manifest '%s' contains a build script...\n", manifest.Filename)

	m, err := filepath.Abs(manifest.Filename)
	if err != nil {
		tv.ErrLog.Add(err)
		return fmt.Errorf("failed to construct path to '%s': %w", manifest.Filename, err)
	}

	// gosec flagged this:
	// G204 (CWE-78): Subprocess launched with function call as argument or cmd arguments
	// Disabling as we trust the source of the variable.
	/* #nosec */
	bs, err := os.ReadFile(m)
	if err != nil {
		err = fmt.Errorf("failed to read file '%s': %w", m, err)
		tv.ErrLog.Add(err)
		return err
	}

	tree, err := toml.LoadBytes(bs)
	if err != nil {
		err = fmt.Errorf("failed to parse toml file '%s': %w", m, err)
		tv.ErrLog.Add(err)
		return err
	}

	if v, ok := tree.GetArray("scripts").(*toml.Tree); ok {
		if script, ok := v.Get("build").(string); ok && script != "" {
			fmt.Fprintf(tv.Output, "Found [scripts.build] '%s'\n", script)
			return nil
		}
	}

	err = fmt.Errorf("failed to find a [scripts] section with a 'build' property in the '%s' manifest", m)
	tv.ErrLog.Add(err)

	// We failed to find a [scripts.build] which is required, so we'll set a
	// default value for the user.
	tree.Set("scripts.build", tv.DefaultBuildCommand)

	data, err := tree.Marshal()
	if err != nil {
		return fmt.Errorf("error updating fastly.toml with a default [scripts.build]: %w", err)
	}

	err = tv.FastlyManifestFile.Load(data)
	if err != nil {
		return fsterr.RemediationError{
			Inner:       err,
			Remediation: fmt.Sprintf(fsterr.ComputeBuildRemediation, tv.DefaultBuildCommand),
		}
	}

	fmt.Fprintf(tv.Output, "Found default [scripts.build] '%s'\n", tv.DefaultBuildCommand)
	return nil
}

// binDir validates a bin directory is available.
func (tv ToolchainValidator) binDir() error {
	dir, err := os.Getwd()
	if err != nil {
		err = fmt.Errorf("failed to identify the current working directory: %w", err)
		tv.ErrLog.Add(err)
		return err
	}

	binDir := filepath.Join(dir, "bin")
	if err := filesystem.MakeDirectoryIfNotExists(binDir); err != nil {
		err = fmt.Errorf("failed to create 'bin' directory: %w", err)
		tv.ErrLog.Add(err)
		return err
	}

	return nil
}

// visitURLRemediation returns a remediation error that suggests visiting the
// official resource URL.
func (tv ToolchainValidator) visitURLRemediation(resource, resourceURL string) string {
	return fmt.Sprintf(`To fix this error, install '%s' by visiting:

    %s

  Then execute:

    $ fastly compute build`, resource, text.Bold(resourceURL))
}

// commandRemediation returns a remediation error that suggests executing a
// command to resolve the missing resource.
func (tv ToolchainValidator) commandRemediation(resource, resourceURL, resourceCommand string) string {
	return fmt.Sprintf(`To fix this error, install '%s' by executing:
    $ %s

  Visit %s for more information.`, resource, resourceCommand, text.Bold(resourceURL))
}

// getJsToolchainBinPath returns the path to where NPM installs binaries.
func getJsToolchainBinPath(bin string) (string, error) {
	// gosec flagged this:
	// G204 (CWE-78): Subprocess launched with variable
	// Disabling as the variables come from trusted sources:
	// The CLI parser enforces supported values via EnumVar.
	/* #nosec */
	path, err := exec.Command(bin, "bin").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(path)), nil
}

// execCommand opens a sub shell to execute the language build script.
func execCommand(cmd string, args []string, out, progress io.Writer, verbose bool, timeout int, errlog fsterr.LogInterface) error {
	s := fstexec.Streaming{
		Command:  cmd,
		Args:     args,
		Env:      os.Environ(),
		Output:   out,
		Progress: progress,
		Verbose:  verbose,
	}
	if timeout > 0 {
		s.Timeout = time.Duration(timeout) * time.Second
	}
	if err := s.Exec(); err != nil {
		errlog.Add(err)
		return err
	}
	return nil
}

// language enables reducing the number of arguments passed to `build()`.
//
// NOTE: We're unable to make the build function generic.
// The generics support in Go1.18 doesn't include accessing struct fields.
type language struct {
	buildScript string
	buildFn     func(string) (string, []string)
	errlog      fsterr.LogInterface
	pkgName     string
	postBuild   string
	timeout     int
}

// build compiles the user's source code into a Wasm binary.
func build(
	l language,
	out io.Writer,
	progress text.Progress,
	verbose bool,
	optionalLocationProcess func() error,
	postBuildCallback func() error,
) error {
	buildScript := strings.ReplaceAll(l.buildScript, "{name}", l.pkgName)
	cmd, args := l.buildFn(buildScript)

	err := execCommand(cmd, args, out, progress, verbose, l.timeout, l.errlog)
	if err != nil {
		return err
	}

	if optionalLocationProcess != nil {
		err := optionalLocationProcess()
		if err != nil {
			return err
		}
	}

	// NOTE: We set the progress indicator to Done() so that any output we now
	// print via the post_build callback doesn't get hidden by the progress status.
	// The progress is 'reset' inside the main build controller `build.go`.
	progress.Done()

	if l.postBuild != "" {
		if err = postBuildCallback(); err == nil {
			cmd, args := l.buildFn(l.postBuild)
			err := execCommand(cmd, args, out, progress, verbose, l.timeout, l.errlog)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
