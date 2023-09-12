package testutil

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/auth"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/runtime"
)

var argsPattern = regexp.MustCompile("`.+`")

// Args is a simple wrapper function designed to accept a CLI command
// (including flags) and return it as a slice for consumption by app.Run().
//
// NOTE: One test file (TestBigQueryCreate) passes RSA content inline into the
// args string which means it has to escape the double quotes (used to infer
// the content should be considered a single argument) with a backtick. This
// causes problems when trying to split the args string by a space (as the RSA
// content has spaces) and so we need to be able to identify when backticks are
// used and ensure the backtick argument is considered a single argument (i.e.
// don't incorrectly split by the spaces within the RSA content when converting
// the arg string into a slice).
//
// The logic checks for backticks, and then replaces the content that is
// surrounded by backticks with --- and then splits the resulting string by
// spaces. Afterwards if there was a backtick matched, then we re-insert the
// backticked content into the slice where --- is found.
func Args(args string) []string {
	var backtickMatch []string

	if strings.Contains(args, "`") {
		backtickMatch = argsPattern.FindStringSubmatch(args)
		args = argsPattern.ReplaceAllString(args, "---")
	}
	s := strings.Split(args, " ")

	if len(backtickMatch) > 0 {
		for i, v := range s {
			if v == "---" {
				s[i] = backtickMatch[0]
			}
		}
	}

	return s
}

// NewRunOpts returns a struct that can be used to populate a call to app.Run()
// while the majority of fields will be pre-populated and only those fields
// commonly changed for testing purposes will need to be provided.
func NewRunOpts(args []string, stdout io.Writer) app.RunOpts {
	var md manifest.Data
	md.File.Args = args
	md.File.SetErrLog(errors.Log)
	md.File.SetOutput(stdout)
	_ = md.File.Read(manifest.Filename)

	configPath := "/dev/null"
	if runtime.Windows {
		configPath = "NUL"
	}

	return app.RunOpts{
		Args:       args,
		APIClient:  mock.APIClient(mock.API{}),
		AuthServer: &auth.Server{},
		ConfigFile: config.File{
			Profiles: TokenProfile(),
		},
		ConfigPath: configPath,
		Env:        config.Environment{},
		ErrLog:     errors.Log,
		ExecuteWasmTools: func(bin string, args []string) error {
			return nil
		},
		HTTPClient: &http.Client{Timeout: time.Second * 5},
		Manifest:   &md,
		Opener: func(input string) error {
			fmt.Printf("open url: %s\n", input)
			return nil // no-op
		},
		Stdout: stdout,
	}
}

func TokenProfile() config.Profiles {
	return config.Profiles{
		// IMPORTANT: Tests mock the token to prevent runtime panics.
		//
		// Tokens are now interactively handled unless a token is provided
		// directly via the --token flag or the FASTLY_API_TOKEN env variable.
		//
		// We force the CLI to skip the interactive prompts by setting a default
		// user profile and making sure the timestamp is not expired.
		"user": &config.Profile{
			AccessTokenCreated: 9999999999, // Year: 2286
			Default:            true,
			Email:              "test@example.com",
			Token:              "mock-token",
		},
	}
}
