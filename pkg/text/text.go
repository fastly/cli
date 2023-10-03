package text

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"syscall"

	"github.com/mitchellh/go-wordwrap"
	"golang.org/x/term"

	"github.com/fastly/cli/pkg/sync"
)

// DefaultTextWidth is the width that should be passed to Wrap for most
// general-purpose blocks of text intended for the user.
const DefaultTextWidth = 120

// Wrap a string at word boundaries with a maximum line length of width. Each
// newline-delimited line in the text is trimmed of whitespace before being
// added to the block for wrapping, which means strings can be declared in the
// source code with whatever leading indentation looks best in context. For
// example,
//
//	Wrap(`
//	    Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do
//	    eiusmod tempor incididunt ut labore et dolore magna aliqua. Dolor
//	    sed viverra ipsum nunc aliquet bibendum enim. In massa tempor nec
//	    feugiat.
//	`, 40)
//
// Produces the output string
//
//	Lorem ipsum dolor sit amet, consectetur
//	adipiscing elit, sed do eiusmod tempor
//	incididunt ut labore et dolore magna
//	aliqua. Dolor sed viverra ipsum nunc
//	aliquet bibendum enim. In massa tempor
//	nec feugiat.
func Wrap(text string, width uint) string {
	var b strings.Builder
	s := bufio.NewScanner(strings.NewReader(text))
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" {
			continue
		}
		_, _ = b.WriteString(line + " ")
	}
	return wordwrap.WrapString(strings.TrimSpace(b.String()), width)
}

// WrapIndent a string at word boundaries with a maximum line length of width
// and indenting the lines by a specified number of spaces.
func WrapIndent(s string, limit uint, indent uint) string {
	limit -= indent
	wrapped := wordwrap.WrapString(s, limit)
	var result []string
	for _, line := range strings.Split(wrapped, "\n") {
		result = append(result, strings.Repeat(" ", int(indent))+line)
	}
	return strings.Join(result, "\n")
}

// Indent writes the help text to the writer using WrapIndent with
// DefaultTextWidth, suffixed by a newlines. It's intended to be used to provide
// detailed information, context, or help to the user.
func Indent(w io.Writer, indent uint, format string, args ...any) {
	text := fmt.Sprintf(format, args...)
	fmt.Fprintf(w, "%s\n", WrapIndent(text, DefaultTextWidth, indent))
}

// Output writes the help text to the writer using Wrap with DefaultTextWidth,
// suffixed by a newlines. It's intended to be used to provide detailed
// information, context, or help to the user.
func Output(w io.Writer, format string, args ...any) {
	text := fmt.Sprintf(format, args...)
	fmt.Fprintf(w, "%s\n", Wrap(text, DefaultTextWidth))
}

// Input prints the prefix to the writer, and then reads a single line from the
// reader, trimming writespace. The received line is passed to the validators,
// and if any of them return a non-nil error, the error is printed to the
// writer, and the input process happens again. Otherwise, the line is returned
// to the caller.
//
// Input is intended to be used to take interactive input from the user.
func Input(w io.Writer, prefix string, r io.Reader, validators ...func(string) error) (string, error) {
	s := bufio.NewScanner(r)

outer:
	for {
		fmt.Fprint(w, Bold(prefix))
		if ok := s.Scan(); !ok {
			return "", s.Err()
		}

		line := strings.TrimSpace(s.Text())
		for _, validate := range validators {
			if err := validate(line); err != nil {
				fmt.Fprintln(w, err.Error())
				continue outer
			}
		}

		return line, nil
	}
}

// IsStdin returns true if r is standard input.
func IsStdin(r io.Reader) bool {
	if f, ok := r.(*os.File); ok {
		return f.Fd() == uintptr(syscall.Stdin)
	}
	return false
}

// IsTTY returns true if fd is a terminal. When used in combination
// with IsStdin, it can be used to determine whether standard input
// is being piped data (i.e. IsStdin == true && IsTTY == false).
// Provide STDOUT as a way to determine whether formatting and/or
// prompting is acceptable output.
func IsTTY(fd any) bool {
	if s, ok := fd.(*sync.Writer); ok {
		// STDOUT is commonly wrapped in a sync.Writer, so here
		// we unwrap it to gain access to the underlying Writer/STDOUT.
		fd = s.W
	}
	if f, ok := fd.(*os.File); ok {
		return term.IsTerminal(int(f.Fd()))
	}
	return false
}

// InputSecure is like Input but doesn't echo input back to the terminal,
// if and only if r is os.Stdin.
func InputSecure(w io.Writer, prefix string, r io.Reader, validators ...func(string) error) (string, error) {
	if !IsStdin(r) {
		return Input(w, prefix, r, validators...)
	}

	read := func() (string, error) {
		fmt.Fprint(w, Bold(prefix))
		p, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return "", err
		}
		return string(p), nil
	}

outer:
	for {
		line, err := read()
		if err != nil {
			return "", err
		}
		line = strings.TrimSpace(line)

		for _, validate := range validators {
			if err := validate(line); err != nil {
				fmt.Fprintln(w, err.Error())
				continue outer
			}
		}

		return line, nil
	}
}

// AskYesNo is similar to Input, but the line read is coerced to
// one of true (yes and its variants) or false (no, its variants and
// anything else) on success.
func AskYesNo(w io.Writer, prompt string, r io.Reader) (bool, error) {
	answer, err := Input(w, prompt, r)
	if err != nil {
		return false, fmt.Errorf("error reading input %w", err)
	}
	answer = strings.ToLower(answer)
	if answer == "y" || answer == "yes" {
		return true, nil
	}
	return false, nil
}

// Break simply writes a newline to the writer. It's intended to be used between
// blocks of text that would otherwise be adjacent, a sort of semantic markup.
func Break(w io.Writer) {
	fmt.Fprintln(w)
}

// BreakN writes n newlines to the writer. It's intended to be used between
// blocks of text that would otherwise be adjacent, a sort of semantic markup.
func BreakN(w io.Writer, n int) {
	if n == 0 {
		return
	}
	for i := 1; i <= n; i++ {
		fmt.Fprintln(w)
	}
}

// Deprecated is a wrapper for fmt.Fprintf with a bold red "DEPRECATED: " prefix.
// Two line breaks are inserted at the end to provide visual spacing.
func Deprecated(w io.Writer, format string, args ...any) {
	format = strings.TrimRight(format, "\r\n") + "\n"
	fmt.Fprintf(w, Wrap(BoldRed("DEPRECATED: ")+format, DefaultTextWidth)+"\n\n", args...)
}

// Error is a wrapper for fmt.Fprintf with a bold red "ERROR: " prefix.
// Two line breaks are inserted at the end to provide visual spacing.
func Error(w io.Writer, format string, args ...any) {
	format = strings.TrimRight(format, "\r\n") + "\n"
	fmt.Fprintf(w, Wrap(BoldRed("ERROR: ")+format, DefaultTextWidth)+"\n\n", args...)
}

// Info is a wrapper for fmt.Fprintf with a bold "INFO: " prefix.
// Two line breaks are inserted at the end to provide visual spacing.
func Info(w io.Writer, format string, args ...any) {
	format = strings.TrimRight(format, "\r\n") + "\n"
	fmt.Fprintf(w, Wrap(Bold("INFO: ")+format, DefaultTextWidth)+"\n\n", args...)
}

// Success is a wrapper for fmt.Fprintf with a bold green "SUCCESS: " prefix.
// Two line breaks are inserted at the end to provide visual spacing.
func Success(w io.Writer, format string, args ...any) {
	format = strings.TrimRight(format, "\r\n") + "\n"
	fmt.Fprintf(w, Wrap(BoldGreen("SUCCESS: ")+format, DefaultTextWidth)+"\n\n", args...)
}

// Warning is a wrapper for fmt.Fprintf with a bold yellow "WARNING: " prefix.
// Two line breaks are inserted at the end to provide visual spacing.
func Warning(w io.Writer, format string, args ...any) {
	format = strings.TrimRight(format, "\r\n") + "\n"
	fmt.Fprintf(w, Wrap(BoldYellow("WARNING: ")+format, DefaultTextWidth)+"\n\n", args...)
}

// Description formats the output of a description item. A description item
// consists of a `intro` and a `description`. Emphasis is placed on the
// `description` using Bold(). For example:
//
//	To compile the package, run:
//	    fastly compute build
func Description(w io.Writer, intro, description string) {
	fmt.Fprintf(w, "%s:\n\t%s\n\n", intro, Bold(description))
}
