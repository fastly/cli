package text_test

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/cli/pkg/text"
	"github.com/google/go-cmp/cmp"
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

func TestPrefixes(t *testing.T) {
	for _, testcase := range []struct {
		name   string
		f      func(io.Writer, string, ...interface{})
		format string
		args   []interface{}
		want   string
	}{
		{
			name:   "Error",
			f:      text.Error,
			format: "Test string %d.",
			args:   []interface{}{123},
			want:   "\nERROR: Test string 123.\n",
		},
		{
			name:   "Success",
			f:      text.Success,
			format: "%s %q %d.",
			args:   []interface{}{"Good", "job", 99},
			want:   "\nSUCCESS: Good \"job\" 99.\n",
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
