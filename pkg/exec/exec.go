package exec

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

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
}

// MonitorSignals spawns a goroutine that configures signal handling so that
// the long running subprocess can be killed using SIGINT/SIGTERM.
func (s *Streaming) MonitorSignals() {
	go s.MonitorSignalsAsync()
}

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
		/* #nosec */
		cmd = exec.CommandContext(ctx, s.Command, s.Args...)
	} else {
		// gosec flagged this:
		// G204 (CWE-78): Subprocess launched with variable
		// Disabling as the variables come from trusted sources.
		/* #nosec */
		cmd = exec.Command(s.Command, s.Args...)
	}
	cmd.Env = append(os.Environ(), s.Env...)

	// Pipe the child process stdout and stderr to our own output writer.
	var stderrBuf bytes.Buffer
	cmd.Stdout = s.Output
	cmd.Stderr = io.MultiWriter(s.Output, &stderrBuf)

	if err := cmd.Start(); err != nil {
		return err
	}

	// Store off os.Process so it can be killed by signal listener.
	//
	// NOTE: cmd.Process is nil until exec.Start() returns successfully.
	//
	//lint:ignore SA4005 because it doesn't fail on macOS but does when run in CI.
	s.Process = cmd.Process

	if err := cmd.Wait(); err != nil {
		var ctx string
		if stderrBuf.Len() > 0 {
			ctx = fmt.Sprintf(":\n%s", strings.TrimSpace(stderrBuf.String()))
		} else {
			ctx = fmt.Sprintf(":\n%s", err)
		}
		return fmt.Errorf("error during execution process%s", ctx)
	}

	return nil
}

// Signal enables spawned subprocess to accept given signal.
func (s *Streaming) Signal(signal os.Signal) error {
	if s.Process != nil {
		err := s.Process.Signal(signal)
		if err != nil {
			return err
		}
	}
	return nil
}
