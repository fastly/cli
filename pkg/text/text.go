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
)

// DefaultTextWidth is the width that should be passed to Wrap for most
// general-purpose blocks of text intended for the user.
const DefaultTextWidth = 90

// Wrap a string at word boundaries with a maximum line length of width. Each
// newline-delimited line in the text is trimmed of whitespace before being
// added to the block for wrapping, which means strings can be declared in the
// source code with whatever leading indentation looks best in context. For
// example,
//
//     Wrap(`
//         Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do
//         eiusmod tempor incididunt ut labore et dolore magna aliqua. Dolor
//         sed viverra ipsum nunc aliquet bibendum enim. In massa tempor nec
//         feugiat.
//     `, 40)
//
// Produces the output string
//
//      Lorem ipsum dolor sit amet, consectetur
//      adipiscing elit, sed do eiusmod tempor
//      incididunt ut labore et dolore magna
//      aliqua. Dolor sed viverra ipsum nunc
//      aliquet bibendum enim. In massa tempor
//      nec feugiat.
//
func Wrap(text string, width uint) string {
	var b strings.Builder
	s := bufio.NewScanner(strings.NewReader(text))
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" {
			continue
		}
		b.WriteString(line + " ")
	}
	return wordwrap.WrapString(strings.TrimSpace(b.String()), width)
}

// Output writes the help text to the writer using Wrap with DefaultTextWidth,
// suffixed by a newlines. It's intended to be used to provide detailed
// information, context, or help to the user.
func Output(w io.Writer, format string, args ...interface{}) {
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

// InputSecure is like Input but doesn't echo input back to the terminal,
// if and only if r is os.Stdin.
func InputSecure(w io.Writer, prefix string, r io.Reader, validators ...func(string) error) (string, error) {
	var (
		f, ok   = r.(*os.File)
		isStdin = ok && uintptr(f.Fd()) == uintptr(syscall.Stdin)
	)
	if !isStdin {
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

// Break simply writes a newline to the writer. It's intended to be used between
// blocks of text that would otherwise be adjacent, a sort of semantic markup.
func Break(w io.Writer) {
	fmt.Fprintln(w)
}

// Error is a wrapper for fmt.Fprintf with a bold red "ERROR: " prefix.
func Error(w io.Writer, format string, args ...interface{}) {
	format = strings.TrimRight(format, "\r\n") + "\n"
	fmt.Fprintf(w, "\n"+BoldRed("ERROR: ")+format, args...)
}

// Warning is a wrapper for fmt.Fprintf with a bold yellow "WARNING: " prefix.
func Warning(w io.Writer, format string, args ...interface{}) {
	format = strings.TrimRight(format, "\r\n") + "\n"
	fmt.Fprintf(w, "\n"+BoldYellow("WARNING: ")+format, args...)
}

// Info is a wrapper for fmt.Fprintf with a bold "INFO: " prefix.
func Info(w io.Writer, format string, args ...interface{}) {
	format = strings.TrimRight(format, "\r\n") + "\n"
	fmt.Fprintf(w, "\n"+Bold("INFO: ")+format, args...)
}

// Success is a wrapper for fmt.Fprintf with a bold green "SUCCESS: " prefix.
func Success(w io.Writer, format string, args ...interface{}) {
	format = strings.TrimRight(format, "\r\n") + "\n"
	fmt.Fprintf(w, "\n"+BoldGreen("SUCCESS: ")+format, args...)
}

// Description formats the output of a description item. A description item
// consists of a `term` and a `description`. Emphasis is placed on the
// `description` using Bold(). For example:
//
//     To compile the package, run:
//         fastly compute build
//
func Description(w io.Writer, term, description string) {
	fmt.Fprintf(w, "%s:\n\t%s\n\n", term, Bold(description))
}
