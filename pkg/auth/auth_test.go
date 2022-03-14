package auth_test

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/testutil"
)

// m represents a basic fastly.toml manifest.
const m = `name = "authentication"
manifest_version = 2`

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

// TestAuthSuccess validates we're able to initialise the authentication flow
// and to get a new access token returned and persisted to disk.
func TestAuthSuccess(t *testing.T) {
	wd, root := createTestEnvironment(t)
	defer os.RemoveAll(root)
	defer os.Chdir(wd)

	var stdout bytes.Buffer
	args := testutil.Args("pops")
	opts := testutil.NewRunOpts(args, &stdout)
	opts.Stdin = strings.NewReader("y") // Authorise opening of web browser.

	err := app.Run(opts)
	if err != nil {
		t.Log(stdout.String())
		t.Fatal(err)
	}
}

// TODO: Validate --token isn't persisted.
// TODO: Validate migration scenario (e.g. invalidate an old long-lived token by checking its expiry)
