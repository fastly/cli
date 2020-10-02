package common

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
)

// StreamingExec models a generic command execution that consumers can use to
// execute commands and stream their output to an io.Writer. For example a
// compute commands can use this to standardizec the flow control for each
// compiler toolchain.
type StreamingExec struct {
	command string
	args    []string
	env     []string
	verbose bool
	output  io.Writer
}

// NewStreamingExec constructs a new StreamingExec instance.
func NewStreamingExec(cmd string, args, env []string, verbose bool, out io.Writer) *StreamingExec {
	return &StreamingExec{
		cmd,
		args,
		env,
		verbose,
		out,
	}
}

// Exec executes the compiler command and pipes the child process stdout and
// stderr output to the supplied io.Writer, it waits for the command to exit
// cleanly or returns an error.
func (s StreamingExec) Exec() error {
	//Constrcut the command with given arguments and environment.
	//
	// gosec flagged this:
	// G204 (CWE-78): Subprocess launched with variable
	// Disabling as the variables come from trusted sources.
	/* #nosec */
	cmd := exec.Command(s.command, s.args...)
	cmd.Env = append(os.Environ(), s.env...)

	// Pipe the child process stdout and stderr to our own output writer.
	var stdoutBuf, stderrBuf bytes.Buffer
	stdoutIn, _ := cmd.StdoutPipe()
	stderrIn, _ := cmd.StderrPipe()
	stdout := io.MultiWriter(s.output, &stdoutBuf)
	stderr := io.MultiWriter(s.output, &stderrBuf)

	// Start the command.
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start execution process: %w", err)
	}

	var errStdout, errStderr error
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		_, errStdout = io.Copy(stdout, stdoutIn)
		wg.Done()
	}()

	_, errStderr = io.Copy(stderr, stderrIn)
	wg.Wait()

	if errStdout != nil {
		return fmt.Errorf("error streaming stdout output from child process: %w", errStdout)
	}
	if errStderr != nil {
		return fmt.Errorf("error streaming stderr output from child process: %w", errStderr)
	}

	// Wait for the command to exit.
	if err := cmd.Wait(); err != nil {
		// If we're not in verbose mode return the bufferred stderr output
		// from cargo as the error.
		if !s.verbose && stderrBuf.Len() > 0 {
			return fmt.Errorf("error during execution process:\n%s", strings.TrimSpace(stderrBuf.String()))
		}
		return fmt.Errorf("error during execution process")
	}

	return nil
}
