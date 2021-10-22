package text

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	fstruntime "github.com/fastly/cli/pkg/runtime"
	"github.com/mattn/go-isatty"
)

// Progress is a producer contract, abstracting over the quiet and verbose
// Progress types. Consumers may use a Progress value in their code, and assign
// it based on the presence of a -v, --verbose flag. Callers are expected to
// call Step for each new major step of their procedural code, and Write with
// the verbose or detailed output of those steps. Callers must eventually call
// either Done or Fail, to signal success or failure respectively.
type Progress interface {
	io.Writer
	Tick(rune)
	Step(string)
	Done()
	Fail()
}

// NewProgress returns a Progress based on the given verbosity level or whether
// the current process is running in a terminal environment.
func NewProgress(output io.Writer, verbose bool) Progress {
	var progress Progress
	if verbose {
		progress = NewVerboseProgress(output)
	} else if isTerminal() {
		progress = NewInteractiveProgress(output)
	} else {
		progress = NewQuietProgress(output)
	}
	return progress
}

// isTerminal indicates if the consumer is a modern terminal.
//
// EXAMPLE: If the user is on a standard Windows 'command prompt' the spinner
// output doesn't work, nor does any colour ouput, so we avoid both features.
func isTerminal() bool {
	if isatty.IsTerminal(os.Stdout.Fd()) && !fstruntime.Windows || isatty.IsCygwinTerminal(os.Stdout.Fd()) {
		return true
	}
	return false
}

// Ticker is a small consumer contract for the Spin function,
// capturing part of the Progress interface.
type Ticker interface {
	Tick(r rune)
}

// Spin calls Tick on the target with the relevant frame every interval. It
// returns when context is canceled, so should be called in its own goroutine.
func Spin(ctx context.Context, frames []rune, interval time.Duration, target Ticker) error {
	var (
		cursor = 0
		ticker = time.NewTicker(interval)
	)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			target.Tick(frames[cursor])
			cursor = (cursor + 1) % len(frames)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// InteractiveProgress is an implementation of Progress that includes a spinner at the
// beginning of each Step, and where newline-delimited lines written via Write
// overwrite the current step line in the output.
type InteractiveProgress struct {
	mtx    sync.Mutex
	output io.Writer

	stepHeader     string       // title of current step
	writeBuffer    bytes.Buffer // receives Write calls
	lastBufferLine string       // last full line in writeBuffer
	currentOutput  string       // the content of the current line displayed to user

	cancel func()          // tell Spin to stop
	done   <-chan struct{} // wait for Spin to stop
}

// NewInteractiveProgress returns a InteractiveProgress outputting to the writer.
func NewInteractiveProgress(output io.Writer) *InteractiveProgress {
	p := &InteractiveProgress{
		output:     output,
		stepHeader: "Initializing...",
	}

	var (
		ctx, cancel = context.WithCancel(context.Background())
		done        = make(chan struct{})
	)
	go func() {
		Spin(ctx, []rune{'-', '\\', '|', '/'}, 100*time.Millisecond, p)
		close(done)
	}()
	p.cancel = cancel
	p.done = done

	return p
}

func (p *InteractiveProgress) replaceLine(format string, args ...interface{}) {
	// Clear the current line.
	n := utf8.RuneCountInString(p.currentOutput)
	switch runtime.GOOS {
	case "windows":
		fmt.Fprintf(p.output, "%s\r", strings.Repeat(" ", n))
	default:
		del, _ := hex.DecodeString("7f")
		sequence := fmt.Sprintf("\b%s\b\033[K", del)
		fmt.Fprintf(p.output, "%s\r", strings.Repeat(sequence, n))
	}

	// Generate the new line.
	s := fmt.Sprintf(format, args...)
	p.currentOutput = s
	fmt.Fprint(p.output, p.currentOutput)
}

func (p *InteractiveProgress) getStatus() string {
	if p.lastBufferLine != "" {
		return p.lastBufferLine // takes precedence
	}
	return p.stepHeader
}

// Tick implements the Progress interface.
func (p *InteractiveProgress) Tick(r rune) {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	p.replaceLine("%s %s", string(r), p.getStatus())
}

// Write implements the Progress interface, emitting each incoming byte slice
// to the internal buffer to be written to the terminal on the next tick.
func (p *InteractiveProgress) Write(buf []byte) (int, error) {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	p.writeBuffer.Write(buf)
	p.lastBufferLine = LastFullLine(p.writeBuffer.String())

	return len(buf), nil
}

// Step implements the Progress interface.
func (p *InteractiveProgress) Step(msg string) {
	msg = strings.TrimSpace(msg)

	p.mtx.Lock()
	defer p.mtx.Unlock()

	// Previous step complete.
	p.replaceLine("%s %s", Bold("✓"), p.stepHeader)
	fmt.Fprintln(p.output)

	// Reset all the stepwise state.
	p.stepHeader = msg
	p.writeBuffer.Reset()
	p.lastBufferLine = ""
	p.currentOutput = ""

	// New step beginning.
	p.replaceLine("%s %s", Bold("·"), p.stepHeader)
}

// Done implements the Progress interface.
func (p *InteractiveProgress) Done() {
	// It's important to cancel the Spin goroutine before taking the lock,
	// because otherwise it's possible to generate a deadlock if the output
	// io.Writer is also synchronized.
	p.cancel()
	<-p.done

	p.mtx.Lock()
	defer p.mtx.Unlock()

	p.replaceLine("%s %s", Bold("✓"), p.stepHeader)
	fmt.Fprintln(p.output)
}

// Fail implements the Progress interface.
func (p *InteractiveProgress) Fail() {
	p.cancel()
	<-p.done

	p.mtx.Lock()
	defer p.mtx.Unlock()

	p.replaceLine("%s %s", Bold("✗"), p.stepHeader)
	fmt.Fprintln(p.output)
}

// LastFullLine returns the last full \n delimited line in s. That is, s must
// contain at least one \n for LastFullLine to return anything.
func LastFullLine(s string) string {
	last := strings.LastIndex(s, "\n")
	if last < 0 {
		return ""
	}

	prev := strings.LastIndex(s[:last], "\n")
	if prev < 0 {
		prev = 0
	}

	return strings.TrimSpace(s[prev:last])
}

//
//
//

// QuietProgress is an implementation of Progress that attempts to be quiet in
// it's output. I.e. it only prints each Step as it progresses and discards any
// intermediary writes between steps. No spinners are used, therefore it's
// useful for non-TTY environiments, such as CI.
type QuietProgress struct {
	output     io.Writer
	nullWriter io.Writer
}

// NewQuietProgress returns a QuietProgress outputting to the writer.
func NewQuietProgress(output io.Writer) *QuietProgress {
	qp := &QuietProgress{
		output:     output,
		nullWriter: io.Discard,
	}
	qp.Step("Initializing...")
	return qp
}

// Tick implements the Progress interface. It's a no-op.
func (p *QuietProgress) Tick(r rune) {}

// Tick implements the Progress interface.
func (p *QuietProgress) Write(buf []byte) (int, error) {
	return p.nullWriter.Write(buf)
}

// Step implements the Progress interface.
func (p *QuietProgress) Step(msg string) {
	fmt.Fprintln(p.output, strings.TrimSpace(msg))
}

// Done implements the Progress interface. It's a no-op.
func (p *QuietProgress) Done() {}

// Fail implements the Progress interface. It's a no-op.
func (p *QuietProgress) Fail() {}

//
//
//

// VerboseProgress is an implementation of Progress that treats Step and Write
// more or less the same: it simply pipes all output to the provided Writer. No
// spinners are used.
type VerboseProgress struct {
	output io.Writer
}

// NewVerboseProgress returns a VerboseProgress outputting to the writer.
func NewVerboseProgress(output io.Writer) *VerboseProgress {
	return &VerboseProgress{
		output: output,
	}
}

// Tick implements the Progress interface. It's a no-op.
func (p *VerboseProgress) Tick(r rune) {}

// Tick implements the Progress interface.
func (p *VerboseProgress) Write(buf []byte) (int, error) {
	return p.output.Write(buf)
}

// Step implements the Progress interface.
func (p *VerboseProgress) Step(msg string) {
	fmt.Fprintln(p.output, strings.TrimSpace(msg))
}

// Done implements the Progress interface. It's a no-op.
func (p *VerboseProgress) Done() {}

// Fail implements the Progress interface. It's a no-op.
func (p *VerboseProgress) Fail() {}

//
//
//

// NullProgress is an implementation of Progress which discards everything
// written into it and produces no output.
type NullProgress struct {
	output io.Writer
}

// NewNullProgress returns a NullProgress.
func NewNullProgress() *NullProgress {
	return &NullProgress{
		output: io.Discard,
	}
}

// Tick implements the Progress interface. It's a no-opt
func (p *NullProgress) Tick(r rune) {}

// Tick implements the Progress interface.
func (p *NullProgress) Write(buf []byte) (int, error) {
	return p.output.Write(buf)
}

// Step implements the Progress interface.
func (p *NullProgress) Step(msg string) {}

// Done implements the Progress interface. It's a no-op.
func (p *NullProgress) Done() {}

// Fail implements the Progress interface. It's a no-op.
func (p *NullProgress) Fail() {}
