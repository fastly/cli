package testutil

import (
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/mock"
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
	return app.RunOpts{
		ConfigPath: "/dev/null",
		Args:       args,
		APIClient:  mock.APIClient(mock.API{}),
		Env:        config.Environment{},
		ErrLog:     errors.Log,
		ConfigFile: config.File{},
		HTTPClient: http.DefaultClient,
		Stdout:     stdout,
	}
}
