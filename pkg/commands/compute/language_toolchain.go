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

// BuildToolchain enables a language toolchain to compile their build script.
type BuildToolchain struct {
	buildFn                   func(string) (string, []string)
	buildScript               string
	errlog                    fsterr.LogInterface
	internalPostBuildCallback func() error
	out                       io.Writer
	postBuild                 string
	postBuildCallback         func() error
	progress                  text.Progress
	timeout                   int
	verbose                   bool
}

func (bt BuildToolchain) Build() error {
	err := bt.execCommand(bt.buildScript)
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
	if bt.internalPostBuildCallback != nil {
		err := bt.internalPostBuildCallback()
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
	bt.progress.Done()

	if bt.postBuild != "" {
		if err = bt.postBuildCallback(); err == nil {
			err := bt.execCommand(bt.postBuild)
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
func (bt BuildToolchain) execCommand(script string) error {
	cmd, args := bt.buildFn(script)

	s := fstexec.Streaming{
		Command:  cmd,
		Args:     args,
		Env:      os.Environ(),
		Output:   bt.out,
		Progress: bt.progress,
		Verbose:  bt.verbose,
	}
	if bt.timeout > 0 {
		s.Timeout = time.Duration(bt.timeout) * time.Second
	}
	if err := s.Exec(); err != nil {
		bt.errlog.Add(err)
		return err
	}
	return nil
}
