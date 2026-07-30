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

func TestHelpSectionOrder(t *testing.T) {
	tests := []struct {
		name     string
		args     string
		sections []string
	}{
		{
			name:     "compact category help",
			args:     "service --help",
			sections: []string{"USAGE", "COMMANDS", "GLOBAL FLAGS", "SEE ALSO"},
		},
		{
			name:     "compact root help",
			args:     "--help",
			sections: []string{"USAGE", "COMMANDS", "GLOBAL FLAGS", "SEE ALSO"},
		},
		{
			name:     "verbose category help",
			args:     "help service",
			sections: []string{"USAGE", "SUBCOMMANDS", "GLOBAL FLAGS", "SEE ALSO"},
		},
		{
			name:     "verbose root help",
			args:     "help",
			sections: []string{"USAGE", "COMMANDS", "GLOBAL FLAGS", "SEE ALSO"},
		},
		{
			name:     "compact leaf help",
			args:     "compute publish --help",
			sections: []string{"USAGE", "OPTIONAL FLAGS", "GLOBAL FLAGS", "SEE ALSO"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout bytes.Buffer
			args := testutil.SplitArgs(tt.args)

			previousInit := app.Init
			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				return testutil.MockGlobalData(args, &stdout), nil
			}
			t.Cleanup(func() {
				app.Init = previousInit
			})

			err := app.Run(args, nil)

			var remediation errors.RemediationError
			if !stderrors.As(err, &remediation) {
				t.Fatalf("expected help output in a remediation error, got %v", err)
			}

			output := remediation.Prefix + stdout.String()
			previousIndex := -1
			for _, section := range tt.sections {
				index := strings.Index(output, section)
				if index < 0 {
					t.Fatalf("missing section %q:\n%s", section, output)
				}
				if index <= previousIndex {
					t.Fatalf("sections not in order %v:\n%s", tt.sections, output)
				}
				previousIndex = index
			}
		})
	}
}
