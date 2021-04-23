package exec

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
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
	Command string
	Args    []string
	Env     []string
	Output  io.Writer
	Timeout time.Duration
	process *os.Process
}

// MonitorSignals spawns a goroutine that configures signal handling so that
// the long running subprocess can be killed using SIGINT/SIGTERM.
func (s *Streaming) MonitorSignals() {
	go monitorSignals(s)
}

func monitorSignals(s *Streaming) {
	sig := make(chan os.Signal, 1)

	signals := []os.Signal{
		syscall.SIGINT,
		syscall.SIGTERM,
	}

	signal.Notify(sig, signals...)

	<-sig
	signal.Stop(sig)

	err := s.Signal(os.Kill)
	if err != nil {
		log.Fatal(err)
	}
}

// Exec executes the compiler command and pipes the child process stdout and
// stderr output to the supplied io.Writer, it waits for the command to exit
// cleanly or returns an error.
func (s Streaming) Exec() error {
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

	// Store off Process so it can be killed by signals
	//lint:ignore SA4005 because it doesn't fail on macOS but does when run in CI.
	s.process = cmd.Process

	// Pipe the child process stdout and stderr to our own output writer.
	var stderrBuf bytes.Buffer
	cmd.Stdout = s.Output
	cmd.Stderr = io.MultiWriter(s.Output, &stderrBuf)

	if err := cmd.Run(); err != nil {
		var ctx string
		if stderrBuf.Len() > 0 {
			ctx = fmt.Sprintf(":\n%s", strings.TrimSpace(stderrBuf.String()))
		}
		return fmt.Errorf("error during execution process%s", ctx)
	}

	return nil
}

// Signal enables spawned subprocess to accept given signal.
func (s Streaming) Signal(signal os.Signal) error {
	if s.process != nil {
		err := s.process.Signal(signal)
		if err != nil {
			return err
		}
	}
	return nil
}
