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
	opts.ConfigPath = filepath.Join(root, manifest.Filename)
	opts.ClientFactory = mock.ClientFactory(mockAPI)
	opts.Stdin = strings.NewReader("y") // Authorise opening of web browser.

	endpoint := make(chan string)
	auth.Browser = mockBrowser(endpoint)
	opts.AuthService = <-endpoint

	err := app.Run(opts)
	if err != nil {
		t.Log(stdout.String())
		t.Fatal(err)
	}

	data, err := os.ReadFile(opts.ConfigPath)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("data: %+v\n", string(data))
}

// TODO: Validate --token isn't persisted.
// TODO: Validate migration scenario (e.g. invalidate an old long-lived token by checking its expiry)

// createTestEnvironment creates a temp directory to run our integration tests
// within, along with a simple fastly.toml manifest.
func createTestEnvironment(t *testing.T) (wd, root string) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	root = testutil.NewEnv(testutil.EnvOpts{
		T: t,
		Write: []testutil.FileIO{
			{Src: m, Dst: manifest.Filename},
		},
	})

	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}

	return wd, root
}

// m represents a basic fastly.toml manifest.
const m = `name = "authentication"
manifest_version = 2`

// mockAPI represents the expected API calls to be made.
var mockAPI = mock.API{
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
func mockBrowser(endpoint chan string) auth.Opener {
	go listenAndServe(endpoint)

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
func listenAndServe(endpoint chan string) {
	local := "127.0.0.1"
	l, err := net.Listen("tcp", local+":0")
	if err != nil {
		log.Fatal(err)
	}

	endpoint <- fmt.Sprintf("http://%s:%d/auth/login", local, l.Addr().(*net.TCPAddr).Port)

	defer l.Close()
	http.Serve(l, authServer{})
}

// authServer represents a mock Authentication HTTP Server.
//
// NOTE: The handler will redirect to the CLI local server, which will cause
// the CLI flow to unblock as a token will be passed through via a query param.
type authServer struct{}

func (s authServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/auth/login":
		uri := r.URL.Query().Get("redirect_uri")
		url := fmt.Sprintf("%s?access_token=123", uri)
		http.Redirect(w, r, url, http.StatusFound)
	}
}
