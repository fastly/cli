package compute

import (
	"io"
	"os"
	"time"

	fsterr "github.com/fastly/cli/pkg/errors"
	fstexec "github.com/fastly/cli/pkg/exec"
	"github.com/fastly/cli/pkg/text"
)

// DefaultBuildErrorRemediation is the message returned to a user when there is
// a build error.
const DefaultBuildErrorRemediation = `There was an error building your project.

Here are some steps you can follow to debug the issue:

- Re-run the fastly subcommand with the --verbose flag to see more information.
- Is the required language toolchain (node/npm, rust/cargo etc) installed correctly?
- Is the required version (if any) of the language toolchain installed/activated?
- Were the required dependencies (package.json, Cargo.toml etc) installed?
- Did the build script (see fastly.toml [scripts.build]) produce a ./bin/main.wasm binary file?
- Was there a configured [scripts.post_build] step that needs to be double-checked?

For more information on fastly.toml configuration settings, refer to https://developer.fastly.com/reference/compute/fastly-toml/`

// Toolchain abstracts a Compute@Edge source language toolchain.
type Toolchain interface {
	// Build compiles the user's source code into a Wasm binary.
	Build(out io.Writer, progress text.Progress, verbose bool, callback func() error) error
}

// buildOpts enables reducing the number of arguments passed to `build()`.
//
// NOTE: We're unable to make the build function generic.
// The generics support in Go1.18 doesn't include accessing struct fields.
type buildOpts struct {
	buildScript string
	buildFn     func(string) (string, []string)
	errlog      fsterr.LogInterface
	postBuild   string
	timeout     int
}

// build compiles the user's source code into a Wasm binary.
func build(
	opts buildOpts,
	out io.Writer,
	progress text.Progress,
	verbose bool,
	internalPostBuildCallback func() error,
	postBuildCallback func() error,
) error {
	cmd, args := opts.buildFn(opts.buildScript)

	err := execCommand(cmd, args, out, progress, verbose, opts.timeout, opts.errlog)
	if err != nil {
		return fsterr.RemediationError{
			Inner:       err,
			Remediation: DefaultBuildErrorRemediation,
		}
	}

	// NOTE: internalPostBuildCallback is only used by Rust currently.
	// It's not a step that would be configured by a user in their fastly.toml
	// It enables Rust to move the compiled binary to a different location.
	// This has to happen BEFORE the postBuild step.
	if internalPostBuildCallback != nil {
		err := internalPostBuildCallback()
		if err != nil {
			return fsterr.RemediationError{
				Inner:       err,
				Remediation: DefaultBuildErrorRemediation,
			}
		}
	}

	// IMPORTANT: The stat check MUST come after the internalPostBuildCallback.
	// This is because for Rust it needs to move the binary first.
	_, err = os.Stat("./bin/main.wasm")
	if err != nil {
		return fsterr.RemediationError{
			Inner:       err,
			Remediation: DefaultBuildErrorRemediation,
		}
	}

	// NOTE: We set the progress indicator to Done() so that any output we now
	// print via the post_build callback doesn't get hidden by the progress status.
	// The progress is 'reset' inside the main build controller `build.go`.
	progress.Done()

	if opts.postBuild != "" {
		if err = postBuildCallback(); err == nil {
			cmd, args := opts.buildFn(opts.postBuild)
			err := execCommand(cmd, args, out, progress, verbose, opts.timeout, opts.errlog)
			if err != nil {
				return fsterr.RemediationError{
					Inner:       err,
					Remediation: DefaultBuildErrorRemediation,
				}
			}
		}
	}

	return nil
}

// execCommand opens a sub shell to execute the language build script.
func execCommand(
	cmd string,
	args []string,
	out, progress io.Writer,
	verbose bool,
	timeout int,
	errlog fsterr.LogInterface,
) error {
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
