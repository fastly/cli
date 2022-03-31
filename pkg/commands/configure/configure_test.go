package configure_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v6/fastly"
)

func TestConfigure(t *testing.T) {
	var (
		goodToken = func() (*fastly.Token, error) { return &fastly.Token{}, nil }
		badToken  = func() (*fastly.Token, error) { return nil, errors.New("bad token") }
		goodUser  = func(*fastly.GetUserInput) (*fastly.User, error) {
			return &fastly.User{
				Login: "test@example.com",
			}, nil
		}
		badUser = func(*fastly.GetUserInput) (*fastly.User, error) { return nil, errors.New("bad user") }
		args    = testutil.Args
	)

	for _, testcase := range []struct {
		name           string
		args           []string
		env            config.Environment
		file           config.File
		api            mock.API
		configFileData string
		stdin          []string
		wantError      string
		wantOutput     []string
		wantFile       string
	}{
		{
			name: "endpoint already in file should be replaced by flag",
			args: args("configure --endpoint=http://staging.dev"),
			configFileData: `[fastly]
	api_endpoint = "https://api.fastly.com"`,
			api: mock.API{
				GetTokenSelfFn: goodToken,
				GetUserFn:      goodUser,
			},
			stdin: []string{"foo", "123456"}, // expect token to be persisted
			wantOutput: []string{
				"Validating token...",
				"Persisting configuration...",
				"You can find your configuration file at",
			},
			wantFile: `config_version = 0

[cli]
  last_checked = ""
  remote_config = ""
  ttl = ""
  version = ""

[fastly]
  api_endpoint = "http://staging.dev"

[language]

  [language.rust]
    fastly_sys_constraint = ""
    rustup_constraint = ""
    toolchain_constraint = ""
    toolchain_version = ""
    wasm_wasi_target = ""

[profile]

  [profile.foo]
    default = true
    email = "test@example.com"
    token = "123456"

[starter-kits]

[user]
  email = ""
  token = ""

[viceroy]
  last_checked = ""
  latest_version = ""
  ttl = ""
`,
		},
		{
			name: "token from flag should be persisted and no token prompt",
			args: args("configure --token=abcdef"),
			api: mock.API{
				GetTokenSelfFn: goodToken,
				GetUserFn:      goodUser,
			},
			stdin: []string{"foo"},
			wantOutput: []string{
				"Validating token...",
				"Persisting configuration...",
				"You can find your configuration file at",
			},
			wantFile: `config_version = 0

[cli]
  last_checked = ""
  remote_config = ""
  ttl = ""
  version = ""

[fastly]
  api_endpoint = "https://api.fastly.com"

[language]

  [language.rust]
    fastly_sys_constraint = ""
    rustup_constraint = ""
    toolchain_constraint = ""
    toolchain_version = ""
    wasm_wasi_target = ""

[profile]

  [profile.foo]
    default = true
    email = "test@example.com"
    token = "abcdef"

[starter-kits]

[user]
  email = ""
  token = ""

[viceroy]
  last_checked = ""
  latest_version = ""
  ttl = ""
`,
		},
		{
			name:  "token from interactive input",
			args:  args("configure"),
			stdin: []string{"foo", "1234"},
			api: mock.API{
				GetTokenSelfFn: goodToken,
				GetUserFn:      goodUser,
			},
			wantOutput: []string{
				"An API token is used to authenticate requests to the Fastly API. To create a token, visit",
				"https://manage.fastly.com/account/personal/tokens",
				"Fastly API token: ",
				"Validating token...",
				"Persisting configuration...",
				"You can find your configuration file at",
			},
			wantFile: `config_version = 0

[cli]
  last_checked = ""
  remote_config = ""
  ttl = ""
  version = ""

[fastly]
  api_endpoint = "https://api.fastly.com"

[language]

  [language.rust]
    fastly_sys_constraint = ""
    rustup_constraint = ""
    toolchain_constraint = ""
    toolchain_version = ""
    wasm_wasi_target = ""

[profile]

  [profile.foo]
    default = true
    email = "test@example.com"
    token = "1234"

[starter-kits]

[user]
  email = ""
  token = ""

[viceroy]
  last_checked = ""
  latest_version = ""
  ttl = ""
`,
		},
		{
			name:  "token from environment",
			args:  args("configure"),
			env:   config.Environment{Token: "hello"},
			stdin: []string{"foo"},
			api: mock.API{
				GetTokenSelfFn: goodToken,
				GetUserFn:      goodUser,
			},
			wantOutput: []string{
				"Validating token...",
				"Persisting configuration...",
				"You can find your configuration file at",
			},
			wantFile: `config_version = 0

[cli]
  last_checked = ""
  remote_config = ""
  ttl = ""
  version = ""

[fastly]
  api_endpoint = "https://api.fastly.com"

[language]

  [language.rust]
    fastly_sys_constraint = ""
    rustup_constraint = ""
    toolchain_constraint = ""
    toolchain_version = ""
    wasm_wasi_target = ""

[profile]

  [profile.foo]
    default = true
    email = "test@example.com"
    token = "hello"

[starter-kits]

[user]
  email = ""
  token = ""

[viceroy]
  last_checked = ""
  latest_version = ""
  ttl = ""
`,
		},
		{
			name: "new default profile created even if another profile already exists",
			args: args("configure"),
			file: config.File{
				// Due to how the test environment skips the main function, it means we
				// don't actually read a configuration file into memory, so we have to
				// manually construct one to be passed through.
				Profiles: map[string]*config.Profile{
					"foo": {
						Default: true,
						Email:   "foo@example.com",
						Token:   "something",
					},
				},
			},
			stdin: []string{"bar", "y", "user_token_given"}, // y == we want the new profile to become default
			api: mock.API{
				GetTokenSelfFn: goodToken,
				GetUserFn:      goodUser,
			},
			wantOutput: []string{
				"An API token is used to authenticate requests to the Fastly API. To create a token, visit",
				"https://manage.fastly.com/account/personal/tokens",
				"Fastly API token: ",
				"Validating token...",
				"Persisting configuration...",
				"You can find your configuration file at",
			},
			wantFile: `config_version = 0

[cli]
  last_checked = ""
  remote_config = ""
  ttl = ""
  version = ""

[fastly]
  api_endpoint = "https://api.fastly.com"

[language]

  [language.rust]
    fastly_sys_constraint = ""
    rustup_constraint = ""
    toolchain_constraint = ""
    toolchain_version = ""
    wasm_wasi_target = ""

[profile]

  [profile.bar]
    default = true
    email = "test@example.com"
    token = "user_token_given"

  [profile.foo]
    default = false
    email = "foo@example.com"
    token = "something"

[starter-kits]

[user]
  email = ""
  token = ""

[viceroy]
  last_checked = ""
  latest_version = ""
  ttl = ""
`,
		},
		{
			name: "new non-default profile created even if another profile already exists",
			args: args("configure"),
			file: config.File{
				// Due to how the test environment skips the main function, it means we
				// don't actually read a configuration file into memory, so we have to
				// manually construct one to be passed through.
				Profiles: map[string]*config.Profile{
					"foo": {
						Default: true,
						Email:   "foo@example.com",
						Token:   "something",
					},
				},
			},
			stdin: []string{"bar", "N", "user_token_given"}, // N == we don't want the new profile to become default
			api: mock.API{
				GetTokenSelfFn: goodToken,
				GetUserFn:      goodUser,
			},
			wantOutput: []string{
				"An API token is used to authenticate requests to the Fastly API. To create a token, visit",
				"https://manage.fastly.com/account/personal/tokens",
				"Fastly API token: ",
				"Validating token...",
				"Persisting configuration...",
				"You can find your configuration file at",
			},
			wantFile: `config_version = 0

[cli]
  last_checked = ""
  remote_config = ""
  ttl = ""
  version = ""

[fastly]
  api_endpoint = "https://api.fastly.com"

[language]

  [language.rust]
    fastly_sys_constraint = ""
    rustup_constraint = ""
    toolchain_constraint = ""
    toolchain_version = ""
    wasm_wasi_target = ""

[profile]

  [profile.bar]
    default = false
    email = "test@example.com"
    token = "user_token_given"

  [profile.foo]
    default = true
    email = "foo@example.com"
    token = "something"

[starter-kits]

[user]
  email = ""
  token = ""

[viceroy]
  last_checked = ""
  latest_version = ""
  ttl = ""
`,
		},
		{
			name: "validate same user can't be created multiple times",
			args: args("configure"),
			file: config.File{
				// Due to how the test environment skips the main function, it means we
				// don't actually read a configuration file into memory, so we have to
				// manually construct one to be passed through.
				Profiles: map[string]*config.Profile{
					"foo": {
						Default: true,
						Email:   "foo@example.com",
						Token:   "something",
					},
				},
			},
			stdin: []string{"foo", "user_token_given"},
			api: mock.API{
				GetTokenSelfFn: goodToken,
				GetUserFn:      goodUser,
			},
			wantError: "profile 'foo' already exists",
		},
		{
			name:      "profile name is required",
			args:      args("configure"),
			wantError: "required profile name missing",
		},
		{
			name: "invalid token",
			args: args("configure --token=abcdef"),
			api: mock.API{
				GetTokenSelfFn: badToken,
				GetUserFn:      badUser,
			},
			stdin: []string{"foo", "user_token_given"},
			wantOutput: []string{
				"Validating token...",
			},
			wantError: "error validating token: bad token",
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			configFilePath := testutil.MakeTempFile(t, testcase.configFileData)
			defer os.RemoveAll(configFilePath)

			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.args, &stdout)
			opts.APIClient = mock.APIClient(testcase.api)
			opts.ConfigFile = testcase.file
			opts.ConfigPath = configFilePath
			opts.Env = testcase.env

			var err error
			if len(testcase.stdin) > 1 {
				// To handle multiple prompt input from the user we need to do some
				// coordination around io pipes to mimic the required user behaviour.
				stdin, prompt := io.Pipe()
				opts.Stdin = stdin

				// Wait for user input and write it to the prompt
				inputc := make(chan string)
				go func() {
					for input := range inputc {
						fmt.Fprintln(prompt, input)
					}
				}()

				// We need a channel so we wait for `run()` to complete
				done := make(chan bool)

				// Call `app.Run()` and wait for response
				go func() {
					err = app.Run(opts)
					done <- true
				}()

				// User provides input
				//
				// NOTE: Must provide as much input as is expected to be waited on by `run()`.
				//       For example, if `run()` calls `input()` twice, then provide two messages.
				//       Otherwise the select statement will trigger the timeout error.
				for _, input := range testcase.stdin {
					inputc <- input
				}

				select {
				case <-done:
					// Wait for app.Run() to finish
				case <-time.After(time.Second):
					t.Fatalf("unexpected timeout waiting for mocked prompt inputs to be processed")
				}
			} else {
				stdin := ""
				if len(testcase.stdin) > 0 {
					stdin = testcase.stdin[0]
				}
				opts.Stdin = strings.NewReader(stdin)
				err = app.Run(opts)
			}

			t.Log(stdout.String())

			testutil.AssertErrorContains(t, err, testcase.wantError)
			for _, s := range testcase.wantOutput {
				testutil.AssertStringContains(t, stdout.String(), s)
			}
			if testcase.wantError == "" {
				p, err := os.ReadFile(configFilePath)
				testutil.AssertNoError(t, err)
				testutil.AssertString(t, testcase.wantFile, string(p))
			}
		})
	}
}
