package app_test

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/testutil"
)

func TestShellCompletion(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name: "bash shell complete",
			Args: "--completion-script-bash",
			WantOutput: `
_fastly_bash_autocomplete() {
    local cur prev opts base
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    opts=$( ${COMP_WORDS[0]} --completion-bash ${COMP_WORDS[@]:1:$COMP_CWORD} )
    COMPREPLY=( $(compgen -W "${opts}" -- ${cur}) )
    return 0
}
complete -F _fastly_bash_autocomplete fastly

`,
		},
		{
			Name: "zsh shell complete",
			Args: "--completion-script-zsh",
			WantOutput: `
#compdef fastly
autoload -U compinit && compinit
autoload -U bashcompinit && bashcompinit

_fastly_bash_autocomplete() {
    local cur prev opts base
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    opts=$( ${COMP_WORDS[0]} --completion-bash ${COMP_WORDS[@]:1:$COMP_CWORD} )
    COMPREPLY=( $(compgen -W "${opts}" -- ${cur}) )
    [[ $COMPREPLY ]] && return
    compgen -f
    return 0
}
complete -F _fastly_bash_autocomplete fastly
`,
		},
		// FIXME: Put back `sso` GA.
		{
			Name: "shell evaluate completion options",
			Args: "--completion-bash",
			WantOutput: `help
sso
acl
acl-entry
alerts
auth-token
backend
compute
config
config-store
config-store-entry
dashboard
dictionary
dictionary-entry
domain
healthcheck
install
ip-list
kv-store
kv-store-entry
log-tail
logging
pops
products
profile
purge
rate-limit
resource-link
secret-store
secret-store-entry
service
service-auth
service-version
stats
tls-config
tls-custom
tls-platform
tls-subscription
update
user
vcl
version
whoami
`,
		},
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(testcase.Name, func(t *testing.T) {
			var (
				stdout bytes.Buffer
				stderr bytes.Buffer
			)

			// NOTE: The Kingpin dependency internally overrides our stdout
			// variable when doing shell completion to the os.Stdout variable and so
			// in order for us to verify it contains the shell completion output, we
			// need an os.Pipe so we can copy off anything written to os.Stdout.
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w
			outC := make(chan string)

			go func() {
				var buf bytes.Buffer
				_, _ = io.Copy(&buf, r)
				outC <- buf.String()
			}()

			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				return testutil.MockGlobalData(testutil.SplitArgs(testcase.Args), &stdout), nil
			}
			err := app.Run(testutil.SplitArgs(testcase.Args), nil)
			if err != nil {
				errors.Deduce(err).Print(&stderr)
			}

			w.Close()
			os.Stdout = old
			out := <-outC

			testutil.AssertString(t, testcase.WantOutput, stripTrailingSpace(out))
		})
	}
}

// stripTrailingSpace removes any trailing spaces from the multiline str.
func stripTrailingSpace(str string) string {
	buf := bytes.NewBuffer(nil)

	scan := bufio.NewScanner(strings.NewReader(str))
	for scan.Scan() {
		_, _ = buf.WriteString(strings.TrimRight(scan.Text(), " \t\r\n"))
		_, _ = buf.WriteString("\n")
	}
	return buf.String()
}
