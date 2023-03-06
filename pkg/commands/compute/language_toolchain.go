package compute

import (
	"fmt"
	"io"
	"os"
	"strings"
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
	Build() error
}

// BuildToolchain enables a language toolchain to compile their build script.
type BuildToolchain struct {
	// autoYes is the --auto-yes flag.
	autoYes bool
	// buildFn constructs a `sh -c` command from the buildScript.
	buildFn func(string) (string, []string)
	// buildScript is the [scripts.build] within the fastly.toml manifest.
	buildScript string
	// errlog is an abstraction for recording errors to disk.
	errlog fsterr.LogInterface
	// in is the user's terminal stdin stream
	in io.Reader
	// internalPostBuildCallback is run after the build but before post build.
	internalPostBuildCallback func() error
	// nonInteractive is the --non-interactive flag.
	nonInteractive bool
	// out is the users terminal stdout stream
	out io.Writer
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

// Build compiles the user's source code into a Wasm binary.
func (bt BuildToolchain) Build() error {
	cmd, args := bt.buildFn(bt.buildScript)

	if bt.verbose {
		text.Break(bt.out)
		text.Description(bt.out, "Build script to execute", fmt.Sprintf("%s %s", cmd, strings.Join(args, " ")))
	}

	var err error
	msg := "Running [scripts.build]"

	// If we're in verbose mode, the build output is shown.
	// So in that case we don't want to have a spinner as it'll interweave output.
	// In non-verbose mode we have a spinner running while the build is happening.
	if !bt.verbose {
		err = bt.spinner.Start()
		if err != nil {
			return err
		}
		bt.spinner.Message(msg + "...")
	}

	err = bt.execCommand(cmd, args, msg)
	if err != nil {
		// In verbose mode we'll have the failure status AFTER the error output.
		// But we can't just call StopFailMessage() without first starting the spinner.
		if bt.verbose {
			text.Break(bt.out)
			err := bt.spinner.Start()
			if err != nil {
				return err
			}
			bt.spinner.Message(msg + "...")

			bt.spinner.StopFailMessage(msg)
			spinErr := bt.spinner.StopFail()
			if spinErr != nil {
				return spinErr
			}
		}
		// WARNING: Don't try to add 'StopFailMessage/StopFail' calls here.
		// If we're in non-verbose mode, then the spiner is BEFORE the error output.
		// Also, in non-verbose mode stopping the spinner is handled internally.
		// See the call to StopFailMessage() inside fstexec.Streaming.Exec().
		return bt.handleError(err)
	}

	// In verbose mode we'll have the failure status AFTER the error output.
	// But we can't just call StopMessage() without first starting the spinner.
	if bt.verbose {
		err = bt.spinner.Start()
		if err != nil {
			return err
		}
		bt.spinner.Message(msg + "...")
		text.Break(bt.out)
	}

	bt.spinner.StopMessage(msg)
	err = bt.spinner.Stop()
	if err != nil {
		return err
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

	if bt.postBuild != "" {
		if !bt.autoYes && !bt.nonInteractive {
			err := bt.promptForBuildContinue(CustomPostBuildScriptMessage, bt.postBuild, bt.out, bt.in, bt.verbose)
			if err != nil {
				return err
			}
		}

		err = bt.spinner.Start()
		if err != nil {
			return err
		}
		msg = "Running [scripts.post_build]..."
		bt.spinner.Message(msg)

		cmd, args := bt.buildFn(bt.postBuild)
		err := bt.execCommand(cmd, args, msg)
		if err != nil {
			// WARNING: Don't try to add 'StopFailMessage/StopFail' calls here.
			// It is handled internally by fstexec.Streaming.Exec().
			return bt.handleError(err)
		}

		bt.spinner.StopMessage(msg)
		err = bt.spinner.Stop()
		if err != nil {
			return err
		}
	}

	return nil
}

func (bt BuildToolchain) handleError(err error) error {
	return fsterr.RemediationError{
		Inner:       err,
		Remediation: DefaultBuildErrorRemediation,
	}
}

// execCommand opens a sub shell to execute the language build script.
//
// NOTE: We pass the spinner and associated message to handle error cases.
// This avoids an issue where the spinner is still running when an error occurs.
// When the error occurs the command output is displayed.
// This causes the spinner message to be displayed twice with different status.
// By passing in the spinner and message we can short-circuit the spinner.
func (bt BuildToolchain) execCommand(cmd string, args []string, spinMessage string) error {
	s := fstexec.Streaming{
		Command:        cmd,
		Args:           args,
		Env:            os.Environ(),
		Output:         bt.out,
		Spinner:        bt.spinner,
		SpinnerMessage: spinMessage,
		Verbose:        bt.verbose,
	}
	if bt.verbose {
		s.ForceOutput = true
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

// promptForBuildContinue ensures the user is happy to continue with the build
// when there is either a custom build or post build in the fastly.toml
// manifest file.
func (bt BuildToolchain) promptForBuildContinue(msg, script string, out io.Writer, in io.Reader, verbose bool) error {
	text.Info(out, "%s:\n", msg)
	text.Break(out)
	text.Indent(out, 4, "%s", script)

	label := "\nAre you sure you want to continue with the post build step? [y/N] "
	answer, err := text.AskYesNo(out, label, in)
	if err != nil {
		return err
	}
	if !answer {
		return fsterr.ErrBuildStopped
	}
	text.Break(out)
	return nil
}
