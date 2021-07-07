package exec

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
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
