package exec

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/cli/pkg/threadsafe"
)

// divider is used as separator lines around shell output.
const divider = "--------------------------------------------------------------------------------"

// Streaming models a generic command execution that consumers can use to
// execute commands and stream their output to an io.Writer. For example
// compute commands can use this to standardize the flow control for each
// compiler toolchain.
type Streaming struct {
	// Args are the command positional arguments.
	Args []string
	// Command is the command to be executed.
	Command string
	// Env is the environment variables to set.
	Env []string
	// ForceOutput ensures output is displayed (default: only display on error).
	ForceOutput bool
	// Output is where to write output (e.g. stdout)
	Output io.Writer
	// Process is the process to terminal if signal received.
	Process *os.Process
	// SignalCh is a channel handling signal events.
	SignalCh chan os.Signal
	// Spinner is a specific spinner instance.
	Spinner text.Spinner
	// SpinnerMessage is the messaging to use.
	SpinnerMessage string
	// Timeout is the command timeout.
	Timeout time.Duration
	// Verbose outputs additional information.
	Verbose bool
}

// MonitorSignals spawns a goroutine that configures signal handling so that
// the long running subprocess can be killed using SIGINT/SIGTERM.
func (s *Streaming) MonitorSignals() {
	go s.MonitorSignalsAsync()
}

// MonitorSignalsAsync configures the signal notifications.
func (s *Streaming) MonitorSignalsAsync() {
	signals := []os.Signal{
		syscall.SIGINT,
		syscall.SIGTERM,
	}

	signal.Notify(s.SignalCh, signals...)

	<-s.SignalCh
	signal.Stop(s.SignalCh)

	// NOTE: We don't do error handling here because the user might be doing local
	// development with the --watch flag and that workflow will have already
	// killed the process. The reason this line still exists is for users running
	// their application locally without the --watch flag and who then execute
	// Ctrl-C to kill the process.
	s.Signal(os.Kill)
}

// Exec executes the compiler command and pipes the child process stdout and
// stderr output to the supplied io.Writer, it waits for the command to exit
// cleanly or returns an error.
func (s *Streaming) Exec() error {
	// Construct the command with given arguments and environment.
	var cmd *exec.Cmd
	if s.Timeout > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), s.Timeout)
		defer cancel()
		// gosec flagged this:
		// G204 (CWE-78): Subprocess launched with variable
		// Disabling as the variables come from trusted sources.
		// #nosec
		// nosemgrep
		cmd = exec.CommandContext(ctx, s.Command, s.Args...)
	} else {
		// gosec flagged this:
		// G204 (CWE-78): Subprocess launched with variable
		// Disabling as the variables come from trusted sources.
		// #nosec
		// nosemgrep
		cmd = exec.Command(s.Command, s.Args...)
	}
	cmd.Env = append(os.Environ(), s.Env...)

	// We store all output in a buffer to hide it unless there was an error.
	var buf threadsafe.Buffer

	var output io.Writer
	output = &buf

	// We only display the stored output if there is an error.
	// But some commands like `compute serve` expect the full output regardless.
	// So for those scenarios they can force all output.
	if s.ForceOutput {
		output = s.Output
	}

	if !s.Verbose {
		text.Break(output)
	}
	text.Info(output, "Command output:")
	text.Output(output, divider)

	cmd.Stdout = output
	cmd.Stderr = output

	if err := cmd.Start(); err != nil {
		text.Output(output, divider)
		return err
	}

	// Store off os.Process so it can be killed by signal listener.
	//
	// NOTE: cmd.Process is nil until exec.Start() returns successfully.
	s.Process = cmd.Process

	if err := cmd.Wait(); err != nil {
		text.Output(output, divider)

		// If we're in verbose mode, the build output is shown.
		// So in that case we don't want to have a spinner as it'll interweave output.
		// In non-verbose mode we have a spinner running while the build is happening.
		if !s.Verbose {
			if s.Spinner != nil {
				s.Spinner.StopFailMessage(s.SpinnerMessage)
				spinErr := s.Spinner.StopFail()
				if spinErr != nil {
					return spinErr
				}
			}
		}

		// Display the buffer stored output as we have an error.
		fmt.Fprintf(s.Output, "%s", buf.String())

		// IMPORTANT: We MUST wrap the original error.
		// This is because the `compute serve` command requires it for --watch
		// Specifically we need to check the error message for "killed".
		// This enables the watching logic to restart the Viceroy binary.
		return fmt.Errorf("error during execution process (see 'command output' above): %w", err)
	}

	text.Output(output, divider)
	return nil
}

// Signal enables spawned subprocess to accept given signal.
func (s *Streaming) Signal(sig os.Signal) error {
	if s.Process != nil {
		err := s.Process.Signal(sig)
		if err != nil {
			return err
		}
	}
	return nil
}

// CommandOpts are arguments for executing a streaming command.
type CommandOpts struct {
	// Args are the command positional arguments.
	Args []string
	// Command is the command to be executed.
	Command string
	// Env is the environment variables to set.
	Env []string
	// ErrLog provides an interface for recording errors to disk.
	ErrLog fsterr.LogInterface
	// Output is where to write output (e.g. stdout)
	Output io.Writer
	// Spinner is a specific spinner instance.
	Spinner text.Spinner
	// SpinnerMessage is the messaging to use.
	SpinnerMessage string
	// Timeout is the command timeout.
	Timeout int
	// Verbose outputs additional information.
	Verbose bool
}

// Command is an abstraction over a Streaming type. It is used by both the
// `compute init` and `compute build` commands to run post init/build scripts.
func Command(opts CommandOpts) error {
	s := Streaming{
		Command:        opts.Command,
		Args:           opts.Args,
		Env:            opts.Env,
		Output:         opts.Output,
		Spinner:        opts.Spinner,
		SpinnerMessage: opts.SpinnerMessage,
		Verbose:        opts.Verbose,
	}
	if opts.Verbose {
		s.ForceOutput = true
	}
	if opts.Timeout > 0 {
		s.Timeout = time.Duration(opts.Timeout) * time.Second
	}
	if err := s.Exec(); err != nil {
		opts.ErrLog.Add(err)
		return err
	}
	return nil
}
