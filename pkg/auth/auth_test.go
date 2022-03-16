package auth_test

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/auth"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v6/fastly"
)

// TestAuthSuccess validates we're able to initialise the authentication flow
// and to get a new access token returned and persisted to disk.
func TestAuthSuccess(t *testing.T) {
	wd, root := createTestEnvironment(t)
	defer os.RemoveAll(root)
	defer os.Chdir(wd)

	var stdout bytes.Buffer
	args := testutil.Args("pops")
	opts := testutil.NewRunOpts(args, &stdout)
	opts.ConfigPath = filepath.Join(root, "config.toml")
	opts.ClientFactory = mock.ClientFactory(mockAPISuccess)
	opts.Stdin = strings.NewReader("y") // Authorise opening of web browser.

	endpoint := make(chan string)
	generatedToken := "123"
	auth.Browser = mockBrowser(endpoint, generatedToken)
	opts.AuthService = <-endpoint

	err := app.Run(opts)
	if err != nil {
		t.Log(stdout.String())
		t.Fatal(err)
	}

	wantOutput := "We are about to initialise a new authentication flow"
	testutil.AssertStringContains(t, stdout.String(), wantOutput)

	var f config.File
	err = f.Read(opts.ConfigPath, opts.Stdin, opts.Stdout)
	if err != nil {
		t.Fatal(err)
	}

	wantToken := "123"
	if f.User.Token != wantToken {
		t.Errorf("want token '%s', have '%s'", wantToken, f.User.Token)
	}
	wantEmail := "test@example.com"
	if f.User.Email != wantEmail {
		t.Errorf("want email '%s', have '%s'", wantEmail, f.User.Email)
	}
}

// TestAuthTokenFlag validates the use of --token will result in no
// authentication flow being initialised, as well as the token not being
// persisted to disk.
func TestAuthTokenFlag(t *testing.T) {
	wd, root := createTestEnvironment(t)
	defer os.RemoveAll(root)
	defer os.Chdir(wd)

	var stdout bytes.Buffer
	args := testutil.Args("pops --token 123")
	opts := testutil.NewRunOpts(args, &stdout)
	opts.ConfigPath = filepath.Join(root, "config.toml")
	opts.ClientFactory = mock.ClientFactory(mockAPISuccess)

	endpoint := make(chan string)
	generatedToken := "123"
	auth.Browser = mockBrowser(endpoint, generatedToken)
	opts.AuthService = <-endpoint

	err := app.Run(opts)
	if err != nil {
		t.Log(stdout.String())
		t.Fatal(err)
	}

	wantToIgnore := "We are about to initialise a new authentication flow"
	testutil.AssertStringDoesntContain(t, stdout.String(), wantToIgnore)

	var f config.File
	err = f.Read(opts.ConfigPath, opts.Stdin, opts.Stdout)
	if err != nil {
		t.Fatal(err)
	}

	if f.User.Token != "" {
		t.Errorf("want no token, have '%s'", f.User.Token)
	}
	if f.User.Email != "" {
		t.Errorf("want no email, have '%s'", f.User.Email)
	}
}

// TestAuthMigration validates we're able to initialise the authentication flow
// and to get a new access token returned and persisted to disk, which replaces
// a pre-existing long-lived token (from before the CLI supported OAuth).
func TestAuthMigration(t *testing.T) {
	wd, root := createTestEnvironment(t)
	defer os.RemoveAll(root)
	defer os.Chdir(wd)

	var stdout bytes.Buffer
	args := testutil.Args("pops")
	opts := testutil.NewRunOpts(args, &stdout)
	opts.ConfigFile = config.File{
		User: config.User{
			Token: "123",
		},
	}
	opts.ConfigPath = filepath.Join(root, "config.toml")
	opts.ClientFactory = mock.ClientFactory(mockAPIMigration)
	opts.Stdin = strings.NewReader("y") // Authorise opening of web browser.

	endpoint := make(chan string)
	generatedToken := "456"
	auth.Browser = mockBrowser(endpoint, generatedToken)
	opts.AuthService = <-endpoint

	err := app.Run(opts)
	if err != nil {
		t.Log(stdout.String())
		t.Fatal(err)
	}

	wantOutput := "Your current access token has no expiration and will be replaced with a short-lived token"
	testutil.AssertStringContains(t, stdout.String(), wantOutput)

	wantOutput = "We are about to initialise a new authentication flow"
	testutil.AssertStringContains(t, stdout.String(), wantOutput)

	var f config.File
	err = f.Read(opts.ConfigPath, opts.Stdin, opts.Stdout)
	if err != nil {
		t.Fatal(err)
	}

	wantToken := "456" // token in config was 123 and should be replaced
	if f.User.Token != wantToken {
		t.Errorf("want token '%s', have '%s'", wantToken, f.User.Token)
	}
	wantEmail := "test@example.com"
	if f.User.Email != wantEmail {
		t.Errorf("want email '%s', have '%s'", wantEmail, f.User.Email)
	}
}

// TestAuthTokenExpired validates we're able to initialise the authentication
// flow and to get a new access token returned and persisted to disk, which
// replaces a previously generated token that has since expired.
func TestAuthTokenExpired(t *testing.T) {
	wd, root := createTestEnvironment(t)
	defer os.RemoveAll(root)
	defer os.Chdir(wd)

	var stdout bytes.Buffer
	args := testutil.Args("pops")
	opts := testutil.NewRunOpts(args, &stdout)
	opts.ConfigFile = config.File{
		User: config.User{
			Token: "123",
		},
	}
	opts.ConfigPath = filepath.Join(root, "config.toml")
	opts.ClientFactory = mock.ClientFactory(mockAPITokenExpired)
	opts.Stdin = strings.NewReader("y") // Authorise opening of web browser.

	endpoint := make(chan string)
	generatedToken := "456"
	auth.Browser = mockBrowser(endpoint, generatedToken)
	opts.AuthService = <-endpoint

	err := app.Run(opts)
	if err != nil {
		t.Log(stdout.String())
		t.Fatal(err)
	}

	wantOutput := "Your access token has expired"
	testutil.AssertStringContains(t, stdout.String(), wantOutput)

	wantOutput = "We are about to initialise a new authentication flow"
	testutil.AssertStringContains(t, stdout.String(), wantOutput)

	var f config.File
	err = f.Read(opts.ConfigPath, opts.Stdin, opts.Stdout)
	if err != nil {
		t.Fatal(err)
	}

	wantToken := "456" // token in config was 123 and should be replaced
	if f.User.Token != wantToken {
		t.Errorf("want token '%s', have '%s'", wantToken, f.User.Token)
	}
	wantEmail := "test@example.com"
	if f.User.Email != wantEmail {
		t.Errorf("want email '%s', have '%s'", wantEmail, f.User.Email)
	}
}

// TestAuthError validates that we produce an appropriate error to the user if
// their authentication flow fails.
func TestAuthError(t *testing.T) {
	wd, root := createTestEnvironment(t)
	defer os.RemoveAll(root)
	defer os.Chdir(wd)

	var stdout bytes.Buffer
	args := testutil.Args("pops")
	opts := testutil.NewRunOpts(args, &stdout)
	opts.ConfigPath = filepath.Join(root, "config.toml")
	opts.Stdin = strings.NewReader("y") // Authorise opening of web browser.

	endpoint := make(chan string)
	noToken := "" // i.e. we expect an error to be returned
	auth.Browser = mockBrowser(endpoint, noToken)
	opts.AuthService = <-endpoint

	err := app.Run(opts)
	if err == nil {
		t.Log(stdout.String())
		t.Error("expected error, got nil")
	}

	testutil.AssertErrorContains(t, err, "no token received from authentication service")
}

// TestAuthSkipAuthCommands validates that some commands don't require
// authentication because they don't use a token. This means skipping the OAuth
// flow entirely.
func TestAuthSkipAuthCommands(t *testing.T) {
	wd, root := createTestEnvironment(t)
	defer os.RemoveAll(root)
	defer os.Chdir(wd)

	tests := []struct {
		args    string
		skip    bool
		persist bool
	}{
		{args: "configure --token 123", skip: true, persist: true},
		{args: "ip-list", skip: true},
		{args: "update", skip: true},
		{args: "version", skip: true},
		{args: "compute build", skip: true},
		{args: "compute deploy"},
		{args: "compute init", skip: true},
		{args: "compute pack --wasm-binary ./foo.wasm", skip: true},
		{args: "compute publish"},
		{args: "compute serve", skip: true},
		{args: "compute update --version latest --package ./foo.tar.gz"},
		{args: "compute validate --package ./foo.tar.gz", skip: true},
	}

	for _, tc := range tests {
		t.Run(tc.args, func(t *testing.T) {
			var stdout bytes.Buffer
			args := testutil.Args(tc.args)
			opts := testutil.NewRunOpts(args, &stdout)
			opts.ConfigFile = config.File{
				StarterKits: config.StarterKitLanguages{
					Rust: []config.StarterKit{
						{
							Name:   "Default",
							Path:   "https://github.com/fastly/compute-starter-kit-rust-default",
							Branch: "main",
						},
					},
				},
			}
			opts.ConfigPath = filepath.Join(root, "config.toml")
			opts.ClientFactory = mock.ClientFactory(mockAPISkipCommandsSuccess)

			// Clean-up the fastly.toml after each test case.
			os.Remove(filepath.Join(root, "config.toml"))
			os.Remove(filepath.Join(root, manifest.Filename))

			// Required for the `version` command.
			opts.Versioners = app.Versioners{
				Viceroy: mock.Versioner{Version: "0.0.0"},
				CLI:     mock.Versioner{Version: "0.0.0"},
			}

			endpoint := make(chan string)
			generatedToken := "123"
			auth.Browser = mockBrowser(endpoint, generatedToken)
			opts.AuthService = <-endpoint

			opts.Stdin = strings.NewReader("") // Some commands like `init` might try to read a prompt.
			if !tc.skip {
				opts.Stdin = strings.NewReader("y") // Authorise opening of web browser.
			}

			app.Run(opts)

			output := "We are about to initialise a new authentication flow"
			if tc.skip {
				testutil.AssertStringDoesntContain(t, stdout.String(), output)
			} else {
				testutil.AssertStringContains(t, stdout.String(), output)
			}

			var f config.File
			err := f.Read(opts.ConfigPath, opts.Stdin, opts.Stdout)
			if err != nil {
				t.Fatal(err)
			}

			if !tc.persist && tc.skip {
				if f.User.Token != "" {
					t.Errorf("want no token, have '%s'", f.User.Token)
				}
				if f.User.Email != "" {
					t.Errorf("want no email, have '%s'", f.User.Email)
				}
			} else {
				wantToken := "123"
				if f.User.Token != wantToken {
					t.Errorf("want token '%s', have '%s'", wantToken, f.User.Token)
				}
				wantEmail := "test@example.com"
				if f.User.Email != wantEmail {
					t.Errorf("want email '%s', have '%s'", wantEmail, f.User.Email)
				}
			}
		})
	}
}

// createTestEnvironment creates a temp directory to run our integration tests.
func createTestEnvironment(t *testing.T) (wd, root string) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	root = testutil.NewEnv(testutil.EnvOpts{
		T: t,
	})

	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}

	return wd, root
}

// mockAPISuccess represents the expected API calls to be made for a
// successful token generation using the new CLI OAuth flow.
var mockAPISuccess = mock.API{
	GetTokenSelfFn: func() (*fastly.Token, error) {
		return &fastly.Token{UserID: "123"}, nil
	},
	GetUserFn: func(*fastly.GetUserInput) (*fastly.User, error) {
		return &fastly.User{
			Login: "test@example.com",
		}, nil
	},
	AllDatacentersFn: func() ([]fastly.Datacenter, error) {
		return []fastly.Datacenter{}, nil
	},
}

// mockAPIMigration represents the expected API calls to be made for a
// successful migration from a long-lived token to short-lived tokens generated
// via the new CLI OAuth flow.
//
// NOTE: GetTokenSelf is effectively called twice, once at the start of the
// command execution to validate if the current token is valid and again before
// persisting the generated short-lived token to disk. We'll use the same token
// data for both calls but this isn't a problem.
//
// NOTE: For the sake of the 'migration' scenario, we'll not set an Expiry,
// which means the test output will show an appropriate messaging.
var mockAPIMigration = mock.API{
	GetTokenSelfFn: func() (*fastly.Token, error) {
		return &fastly.Token{UserID: "456"}, nil
	},
	GetUserFn: func(*fastly.GetUserInput) (*fastly.User, error) {
		return &fastly.User{
			Login: "test@example.com",
		}, nil
	},
	AllDatacentersFn: func() ([]fastly.Datacenter, error) {
		return []fastly.Datacenter{}, nil
	},
}

// mockAPITokenExpired represents the expected API calls to be made for a
// successful update from an existing generated token that has expired to a
// newly generated token via the new CLI OAuth flow.
//
// NOTE: GetTokenSelf is effectively called twice, once at the start of the
// command execution to validate if the current token is valid and again before
// persisting the generated short-lived token to disk. We mock the response to
// change depending on the order it's called so we can see a successful output.
//
// NOTE: For the sake of the 'expiry' scenario, we'll return a fastly.HTTPError
// with the appropriate status code, which means the test output will show an
// appropriate messaging.
var mockAPITokenExpired = mock.API{
	GetTokenSelfFn: statefulTokenSelf(),
	GetUserFn: func(*fastly.GetUserInput) (*fastly.User, error) {
		return &fastly.User{
			Login: "test@example.com",
		}, nil
	},
	AllDatacentersFn: func() ([]fastly.Datacenter, error) {
		return []fastly.Datacenter{}, nil
	},
}

// mockAPISkipCommandsSuccess represents the expected API calls to be made for a
// successful token generation using the new CLI OAuth flow.
var mockAPISkipCommandsSuccess = mock.API{
	GetTokenSelfFn: func() (*fastly.Token, error) {
		return &fastly.Token{UserID: "123"}, nil
	},
	GetUserFn: func(*fastly.GetUserInput) (*fastly.User, error) {
		return &fastly.User{
			Login: "test@example.com",
		}, nil
	},
	AllDatacentersFn: func() ([]fastly.Datacenter, error) {
		return []fastly.Datacenter{}, nil
	},
	AllIPsFn: func() (v4, v6 fastly.IPAddrs, err error) {
		return []string{
				"00.123.45.6/78",
			}, []string{
				"0a12:3b45::/67",
			}, nil
	},
}

var statefulTokenSelf = func() func() (*fastly.Token, error) {
	count := 0
	return func() (*fastly.Token, error) {
		count += 1
		switch count {
		case 1:
			return nil, &fastly.HTTPError{
				StatusCode: http.StatusUnauthorized,
			}
		case 2:
			t := testutil.Date
			return &fastly.Token{
				UserID:    "456",
				ExpiresAt: &t,
			}, nil
		}
		return nil, nil
	}
}

// mockBrowser mocks the behaviour for opening a web browser.
//
// NOTE:
// A local HTTP server that handles /auth/login will be started.
// The auth.App value will point to the local server.
//
// NOTE:
// The returned Opener type will be parsed an arg that will point to the local
// server, along with a callback query param. The function will make a request
// to the URL for which the local server will 302 to the CLI local server's
// /auth-callback endpoint (passing along a mock token).
func mockBrowser(endpoint chan string, token string) auth.Opener {
	go listenAndServe(endpoint, token)

	return func(url string) error {
		go http.Get(url)
		return nil
	}
}

// listenAndServe listens on a random TCP network port and then calls Serve with
// custom handler type to handle requests on incoming connections.
//
// NOTE: This function is expected to be called asynchronously, hence a channel
// is provided for synchronised communication.
func listenAndServe(endpoint chan string, token string) {
	local := "127.0.0.1"
	l, err := net.Listen("tcp", local+":0")
	if err != nil {
		log.Fatal(err)
	}

	endpoint <- fmt.Sprintf("http://%s:%d/auth/login", local, l.Addr().(*net.TCPAddr).Port)

	defer l.Close()
	http.Serve(l, authServer{token: token})
}

// authServer represents a mock Authentication HTTP Server.
//
// NOTE: The handler will redirect to the CLI local server, which will cause
// the CLI flow to unblock as a token will be passed through via a query param.
//
// NOTE: When testing the failure scenario, we replace the access_token query
// parameter with auth_error.
type authServer struct {
	token string
}

func (s authServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/auth/login":
		uri := r.URL.Query().Get("redirect_uri")
		url := fmt.Sprintf("%s?access_token=%s", uri, s.token)
		if s.token == "" {
			url = fmt.Sprintf("%s?auth_error=whoops", uri)
		}
		http.Redirect(w, r, url, http.StatusFound)
	}
}
