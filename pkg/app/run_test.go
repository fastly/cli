package app_test

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/testutil"
)

// If you add a Short flag and this test starts failing, it could be due to the same short flag existing at the global level.
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
		{
			Name: "shell evaluate completion options",
			Args: "--completion-bash",
			WantOutput: `help
auth
apisecurity
compute
config
config-store
config-store-entry
dashboard
domain
install
ip-list
kv-store
kv-store-entry
log-tail
ngwaf
object-storage
pops
products
secret-store
secret-store-entry
service
stats
tls-config
tls-custom
tls-platform
tls-subscription
tools
update
user
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
			err := app.Run(testutil.SplitArgs(testcase.Args), nil, nil)
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

// TestExecQuietSuppressesExpiryWarning exercises the full Exec path to verify
// that --quiet suppresses the expiration warning end-to-end. (--json also sets
// Quiet=true at run.go:204, but config doesn't accept --json; the unit test
// TestCheckTokenExpirationWarningSuppression covers the Quiet flag directly.)
func TestExecQuietSuppressesExpiryWarning(t *testing.T) {
	var stdout bytes.Buffer

	args := testutil.SplitArgs("config -l --quiet")
	app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
		data := testutil.MockGlobalData(args, &stdout)
		// Set the default token to expire soon so a warning would fire without --quiet.
		data.Config.Auth.Tokens["user"].APITokenExpiresAt = time.Now().Add(3 * 24 * time.Hour).Format(time.RFC3339)
		return data, nil
	}
	err := app.Run(args, nil, nil)
	if err != nil {
		t.Fatalf("app.Run returned unexpected error: %v", err)
	}

	output := stdout.String()
	if strings.Contains(output, "expires in") {
		t.Errorf("--quiet should suppress expiry warning, but got: %s", output)
	}
}

// TestExecConfigShowsExpiryWarning is a companion test verifying the warning
// does appear for a non-quiet, non-auth command when the token is expiring.
func TestExecConfigShowsExpiryWarning(t *testing.T) {
	var stdout bytes.Buffer

	args := testutil.SplitArgs("config -l")
	app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
		data := testutil.MockGlobalData(args, &stdout)
		data.Config.Auth.Tokens["user"].APITokenExpiresAt = time.Now().Add(3 * 24 * time.Hour).Format(time.RFC3339)
		return data, nil
	}
	err := app.Run(args, nil, nil)
	if err != nil {
		t.Fatalf("app.Run returned unexpected error: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "expires in") {
		t.Errorf("expected expiry warning for config command, got: %s", output)
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
