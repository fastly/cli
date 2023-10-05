package text_test

import (
	"bytes"
	"errors"
	"io"
	"strconv"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/cli/pkg/text"
)

func TestInput(t *testing.T) {
	for _, testcase := range []struct {
		name       string
		in         string
		prefix     string
		validators []func(string) error
		wantOutput string
		wantResult string
	}{
		{
			name:       "empty",
			in:         "\n",
			prefix:     "Press enter ",
			wantOutput: "Press enter ",
			wantResult: "",
		},
		{
			name:       "single letter",
			in:         "a\n",
			prefix:     "> ",
			wantOutput: "> ",
			wantResult: "a",
		},
		{
			name:       "single letter with whitespace",
			in:         " a  \n",
			prefix:     "> ",
			wantOutput: "> ",
			wantResult: "a",
		},
		{
			name:   "nonempty validator",
			in:     "\n\nFINE\n",
			prefix: "Tell me something: ",
			validators: []func(string) error{
				func(s string) error {
					if s == "" {
						return errors.New("nothing isn't something")
					}
					return nil
				},
			},
			wantOutput: "Tell me something: nothing isn't something\nTell me something: nothing isn't something\nTell me something: ",
			wantResult: "FINE",
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			var buf bytes.Buffer
			result, err := text.Input(&buf, testcase.prefix, strings.NewReader(testcase.in), testcase.validators...)
			testutil.AssertNoError(t, err)
			testutil.AssertString(t, testcase.wantOutput, buf.String())
			testutil.AssertString(t, testcase.wantResult, result)

			buf.Reset()
			result, err = text.InputSecure(&buf, testcase.prefix, strings.NewReader(testcase.in), testcase.validators...)
			testutil.AssertNoError(t, err)
			testutil.AssertString(t, testcase.wantOutput, buf.String())
			testutil.AssertString(t, testcase.wantResult, result)
		})
	}
}

func TestAskYesNo(t *testing.T) {
	for _, testcase := range []struct {
		name       string
		in         string
		wantResult bool
	}{
		{
			name:       "empty",
			in:         "\n",
			wantResult: false,
		},
		{
			name:       "uppercase y",
			in:         "Y\n",
			wantResult: true,
		},
		{
			name:       "lowercase y",
			in:         "y\n",
			wantResult: true,
		},
		{
			name:       "mixed case yes",
			in:         "yEs\n",
			wantResult: true,
		},
		{
			name:       "mixed case no",
			in:         "nO\n",
			wantResult: false,
		},
		{
			name:       "anything else",
			in:         "whatever\n",
			wantResult: false,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			var buf bytes.Buffer
			result, err := text.AskYesNo(&buf, "", strings.NewReader(testcase.in))
			testutil.AssertNoError(t, err)
			testutil.AssertBool(t, testcase.wantResult, result)
		})
	}
}

func TestPrefixes(t *testing.T) {
	for _, testcase := range []struct {
		name   string
		f      func(io.Writer, string, ...any)
		format string
		args   []any
		want   string
	}{
		{
			name:   "Deprecated",
			f:      text.Deprecated,
			format: "Test string %d.",
			args:   []any{123},
			want:   "DEPRECATED: Test string 123.\n",
		},
		{
			name:   "Error",
			f:      text.Error,
			format: "Test string %d.",
			args:   []any{123},
			want:   "ERROR: Test string 123.\n",
		},
		{
			name:   "Info",
			f:      text.Info,
			format: "Test string %d.",
			args:   []any{123},
			want:   "INFO: Test string 123.\n",
		},
		{
			name:   "Success",
			f:      text.Success,
			format: "%s %q %d.",
			args:   []any{"Good", "job", 99},
			want:   "SUCCESS: Good \"job\" 99.\n",
		},
		{
			name:   "Warning",
			f:      text.Warning,
			format: "\nTest string %d.\n\n", // notice inline line breaks override the default single suffix line break
			args:   []any{123},
			want:   "\nWARNING: Test string 123.\n\n",
		},
		{
			name:   "Info with irregular line breaks and tabs placement",
			f:      text.Info,
			format: "\n\nTest string\n\t%s",
			args:   []any{"anything"},
			want:   "\n\nINFO: Test string \tanything\n",
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			var buf bytes.Buffer
			testcase.f(&buf, testcase.format, testcase.args...)
			if want, have := testcase.want, buf.String(); want != have {
				t.Error(cmp.Diff(want, have))
			}
		})
	}
}

func TestWrap(t *testing.T) {
	for i, testcase := range []struct {
		text, want string
		limit      uint
	}{
		{
			text:  "Example text goes here.",
			limit: 2,
			want:  "Example\ntext\ngoes\nhere.", // notice it won't split individual words
		},
		{
			text:  "Example text goes here.",
			limit: 12,
			want:  "Example text\ngoes here.",
		},
		{
			text:  "Example text goes here.",
			limit: 100,
			want:  "Example text goes here.",
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			output := text.Wrap(testcase.text, testcase.limit)
			if want, have := testcase.want, output; want != have {
				t.Error(cmp.Diff(want, have))
			}
		})
	}
}

func TestWrapIndent(t *testing.T) {
	for i, testcase := range []struct {
		text, want    string
		limit, indent uint // internally limit subtracts the indent
	}{
		{
			text:   "Example text goes here.",
			limit:  2,
			indent: 2,
			want:   "  Example text goes here.", // indent causes limit to become zero so we effectively just get an indent.
		},
		{
			text:   "Example text goes here.",
			limit:  20,
			indent: 4,
			want:   "    Example text\n    goes here.",
		},
		{
			text:   "Example text goes here.",
			limit:  100,
			indent: 6,
			want:   "      Example text goes here.",
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			output := text.WrapIndent(testcase.text, testcase.limit, testcase.indent)
			if want, have := testcase.want, output; want != have {
				t.Error(cmp.Diff(want, have))
			}
		})
	}
}

func TestParseBreaks(t *testing.T) {
	for _, testcase := range []struct {
		name   string
		in     string
		prefix int
		suffix int
		txt    string
	}{
		{
			name:   "no line breaks",
			in:     "example",
			prefix: 0,
			suffix: 0,
			txt:    "example",
		},
		{
			name:   "starting line breaks",
			in:     "\n\n\nexample",
			prefix: 3,
			suffix: 0,
			txt:    "example",
		},
		{
			name:   "ending line breaks",
			in:     "example\n\n\n",
			prefix: 0,
			suffix: 3,
			txt:    "example",
		},
		{
			name:   "both ends line breaks",
			in:     "\n\nexample\n\n\n",
			prefix: 2,
			suffix: 3,
			txt:    "example",
		},
		{
			name:   "line breaks in the main text",
			in:     "\n\nexample message with\na line break inside\n\n\n",
			prefix: 2,
			suffix: 3,
			txt:    "example message with a line break inside",
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			prefix, suffix, txt := text.ParseBreaks(testcase.in)
			if prefix != testcase.prefix {
				t.Errorf("want: %d, have: %d", testcase.prefix, prefix)
			}
			if suffix != testcase.suffix {
				t.Errorf("want: %d, have: %d", testcase.suffix, suffix)
			}
			if txt != testcase.txt {
				t.Errorf("want: %s, have: %s", testcase.txt, txt)
			}
		})
	}
}
