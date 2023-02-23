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
	Args           []string
	Command        string
	Env            []string
	ForceOutput    bool
	Output         io.Writer
	Process        *os.Process
	SignalCh       chan os.Signal
	Spinner        text.Spinner
	SpinnerMessage string
	Timeout        time.Duration
	Verbose        bool
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

		if s.Spinner != nil {
			s.Spinner.StopFailMessage(s.SpinnerMessage)
			spinErr := s.Spinner.StopFail()
			if spinErr != nil {
				return spinErr
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
