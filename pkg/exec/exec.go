package exec

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
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
	Args     []string
	Command  string
	Env      []string
	Output   io.Writer
	Process  *os.Process
	SignalCh chan os.Signal
	Timeout  time.Duration
	Verbose  bool
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
	if s.Verbose {
		text.Break(s.Output)
		text.Description(s.Output, "Executing command", fmt.Sprintf("%s %s", s.Command, strings.Join(s.Args, " ")))
	}

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

	// Pipe the child process stdout and stderr to our own output writer.
	var stdoutBuf, stderrBuf threadsafe.Buffer

	// We discard the build output unless we're in verbose mode.
	output := io.Discard

	if s.Verbose {
		output = s.Output
		text.Info(output, "Command output:")
		text.Output(output, divider)
	}

	cmd.Stdout = io.MultiWriter(output, &stdoutBuf)
	cmd.Stderr = io.MultiWriter(output, &stderrBuf)

	if err := cmd.Start(); err != nil {
		if s.Verbose {
			text.Output(output, divider)
			text.Break(output)
		}
		return err
	}

	// Store off os.Process so it can be killed by signal listener.
	//
	// NOTE: cmd.Process is nil until exec.Start() returns successfully.
	s.Process = cmd.Process

	if err := cmd.Wait(); err != nil {
		var ctx string
		if stderrBuf.Len() > 0 {
			if !s.Verbose {
				ctx = fmt.Sprintf(":\n\n%s", strings.TrimSpace(stderrBuf.String()))
			}
		} else {
			// NOTE: Viceroy doesn't send errors to stderr but to stdout.
			//
			// We want to ensure the compilation errors sent to stdout are displayed
			// regardless of whether the user has --verbose set.
			//
			// If --verbose is set, then all errors will be seen anyway.
			// If --verbose isn't set, then we use text.Progress to constrain output.
			//
			// Meaning: only display stdoutBuf if NOT in verbose mode already.
			var cmdOutput string
			if !s.Verbose {
				cmdOutput = "\n" + text.WrapIndent(stdoutBuf.String(), text.DefaultTextWidth, 5)
			}
			ctx = fmt.Sprintf(":%s\n\n%s", cmdOutput, err)
		}
		if s.Verbose {
			text.Output(output, divider)
			text.Break(output)
		}
		return fmt.Errorf("error during execution process%s", ctx)
	}

	if s.Verbose {
		text.Output(output, divider)
		text.Break(output)
	}
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
