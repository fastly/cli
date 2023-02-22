package compute

import (
	"fmt"
	"io"
	"os"
	"time"

	fsterr "github.com/fastly/cli/pkg/errors"
	fstexec "github.com/fastly/cli/pkg/exec"
	"github.com/fastly/cli/pkg/text"
)

// DefaultBuildErrorRemediation is the message returned to a user when there is
// a build error.
var DefaultBuildErrorRemediation = func() string {
	return fmt.Sprintf(`%s:

- Re-run the fastly command with the --verbose flag to see more information.
- Is the required language toolchain (node/npm, rust/cargo etc) installed correctly?
- Is the required version (if any) of the language toolchain installed/activated?
- Were the required dependencies (package.json, Cargo.toml etc) installed?
- Did the build script (see fastly.toml [scripts.build]) produce a ./bin/main.wasm binary file?
- Was there a configured [scripts.post_build] step that needs to be double-checked?

For more information on fastly.toml configuration settings, refer to https://developer.fastly.com/reference/compute/fastly-toml/`,
		text.BoldYellow("Here are some steps you can follow to debug the issue"))
}()

// Toolchain abstracts a Compute@Edge source language toolchain.
type Toolchain interface {
	// Build compiles the user's source code into a Wasm binary.
	Build(out io.Writer, spinner text.Spinner, verbose bool, callback func() error) error
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
	spinner                   text.Spinner
	timeout                   int
	verbose                   bool
}

// Build compiles the user's source code into a Wasm binary.
func (bt BuildToolchain) Build() error {
	var (
		err error
		msg string
	)

	if bt.verbose {
		text.Break(bt.out)
	}

	err = bt.spinner.Start()
	if err != nil {
		return err
	}
	msg = "Running [scripts.build]..."
	bt.spinner.Message(msg)

	bt.spinner.StopMessage(msg)
	err = bt.spinner.Stop()
	if err != nil {
		return err
	}

	err = bt.execCommand(bt.buildScript)
	if err != nil {
		return bt.handleError(err)
	}

	// NOTE: internalPostBuildCallback is only used by Rust currently.
	// It's not a step that would be configured by a user in their fastly.toml
	// It enables Rust to move the compiled binary to a different location.
	// This has to happen BEFORE the postBuild step.
	if bt.internalPostBuildCallback != nil {
		err := bt.internalPostBuildCallback()
		if err != nil {
			return bt.handleError(err)
		}
	}

	// IMPORTANT: The stat check MUST come after the internalPostBuildCallback.
	// This is because for Rust it needs to move the binary first.
	_, err = os.Stat("./bin/main.wasm")
	if err != nil {
		return bt.handleError(err)
	}

	err = bt.spinner.Start()
	if err != nil {
		return err
	}
	msg = "Running post_build callback..."
	bt.spinner.Message(msg)

	bt.spinner.StopMessage(msg)
	err = bt.spinner.Stop()
	if err != nil {
		return err
	}

	if bt.postBuild != "" {
		if err = bt.postBuildCallback(); err == nil {
			err := bt.execCommand(bt.postBuild)
			if err != nil {
				return bt.handleError(err)
			}
		}
	}

	return nil
}

func (bt BuildToolchain) handleError(err error) error {
	text.Break(bt.out)
	return fsterr.RemediationError{
		Inner:       err,
		Remediation: DefaultBuildErrorRemediation,
	}
}

// execCommand opens a sub shell to execute the language build script.
func (bt BuildToolchain) execCommand(script string) error {
	cmd, args := bt.buildFn(script)

	s := fstexec.Streaming{
		Command: cmd,
		Args:    args,
		Env:     os.Environ(),
		Output:  bt.out,
		Verbose: bt.verbose,
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
