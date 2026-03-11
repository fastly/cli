package app_test

import (
	"bytes"
	stderrors "errors"
	"io"
	"strings"
	"testing"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/testutil"
)

func TestAuthGuideBlock(t *testing.T) {
	cases := []struct {
		name      string
		args      string
		wantGuide bool
	}{
		{
			name:      "auth --help includes AUTH GUIDE",
			args:      "auth --help",
			wantGuide: true,
		},
		{
			name:      "auth login --help excludes AUTH GUIDE",
			args:      "auth login --help",
			wantGuide: false,
		},
		{
			name:      "help auth includes AUTH GUIDE",
			args:      "help auth",
			wantGuide: true,
		},
		{
			name:      "service --help excludes AUTH GUIDE",
			args:      "service --help",
			wantGuide: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var stdout bytes.Buffer
			args := testutil.SplitArgs(tc.args)

			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				return testutil.MockGlobalData(args, &stdout), nil
			}

			err := app.Run(args, nil)

			var output string
			if err != nil {
				var re errors.RemediationError
				if stderrors.As(err, &re) {
					output = re.Prefix
				}
			}
			output += stdout.String()

			if tc.wantGuide && !strings.Contains(output, "AUTH GUIDE") {
				t.Errorf("expected AUTH GUIDE in output, got:\n%s", output)
			}
			if tc.wantGuide && !strings.Contains(output, "--sso --token") {
				t.Errorf("expected AUTH GUIDE to contain '--sso --token' quick-start example, got:\n%s", output)
			}
			if !tc.wantGuide && strings.Contains(output, "AUTH GUIDE") {
				t.Errorf("did not expect AUTH GUIDE in output, got:\n%s", output)
			}
		})
	}
}
