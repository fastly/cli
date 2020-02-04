package text_test

import (
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/fastly/cli/pkg/text"
)

func TestProgress(t *testing.T) {
	for _, testcase := range []struct {
		name        string
		constructor func(io.Writer) text.Progress
	}{
		{
			name:        "quiet",
			constructor: func(w io.Writer) text.Progress { return text.NewQuietProgress(w) },
		},
		{
			name:        "verbose",
			constructor: func(w io.Writer) text.Progress { return text.NewVerboseProgress(w) },
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			p := testcase.constructor(os.Stdout)
			for _, f := range []func(){
				func() { fmt.Fprintf(p, "Alpha\n") },
				func() { p.Step("Step one...") },
				func() { fmt.Fprintf(p, "Beta\n") },
				func() { fmt.Fprintf(p, "Delta\n") },
				func() { p.Step("Step two...") },
				func() { fmt.Fprintf(p, "Iota\n") },
				func() { fmt.Fprintf(p, "Kappa\n") },
				func() { fmt.Fprintf(p, "Gamma\n") },
				func() { fmt.Fprintf(p, "Omicron\n") },
				func() { p.Step("Step three...") },
				func() { fmt.Fprintf(p, "Nü\n") },
				func() { p.Done() },
			} {
				f()
				time.Sleep(250 * time.Millisecond)
			}
		})
	}
}

func TestLastFullLine(t *testing.T) {
	for _, testcase := range []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "empty",
			input: "",
			want:  "",
		},
		{
			name:  "no newline",
			input: "abc def ghi",
			want:  "",
		},
		{
			name:  "one newline at end",
			input: "abc def ghi\n",
			want:  "abc def ghi",
		},
		{
			name:  "one full line and a partial",
			input: "foo bar\nbaz quux",
			want:  "foo bar",
		},
		{
			name:  "multiple lines",
			input: "alpha beta\ndelta kappa\ngamma iota\nomicron nu",
			want:  "gamma iota",
		},
		{
			name:  "multiple newlines at end",
			input: "alpha beta\n\n\ndelta kappa\n\ngamma iota\n\nomicron nu\n\n",
			want:  "",
		},
		{
			name:  "multiple newlines in middle",
			input: "alpha beta\n\n\ndelta kappa\n\ngamma iota\n\nomicron nu",
			want:  "",
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			if want, have := testcase.want, text.LastFullLine(testcase.input); want != have {
				t.Fatalf("want %q, have %q", want, have)
			}
		})
	}
}
