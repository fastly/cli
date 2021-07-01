package testutil

import (
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/update"
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

// NewAppRunArgs returns a struct that can be used to populate a call to
// app.Run() while the majority of fields will be pre-populated and only those
// fields commonly changed for testing purposes will need to be provided.
//
// NOTE: most consumers of NewAppRunArgs() won't need to pass mocked
// implementations, and so we provide helper functions on the receiver to update
// the less commonly modified fields when it comes to the testing environment.
func NewAppRunArgs(args []string, stdout io.Writer) *AppRunArgs {
	return &AppRunArgs{
		AppConfigFile: "/dev/null",
		Args:          args,
		ClientFactory: mock.APIClient(mock.API{}),
		Env:           config.Environment{},
		File:          config.File{},
		HTTPClient:    http.DefaultClient,
		Out:           stdout,
	}
}

// AppRunArgs represents the structure of the args passed into app.Run().
//
// NOTE: in future the app.Run() signature will be updated to accept a struct,
// and that will mean any references to the AppRunArgs testing struct can be
// replaced with the app.Run() struct.
type AppRunArgs struct {
	AppConfigFile string
	Args          []string
	CLIVersioner  update.Versioner
	ClientFactory func(token, endpoint string) (api.Interface, error)
	Env           config.Environment
	File          config.File
	HTTPClient    api.HTTPClient
	In            io.Reader
	Out           io.Writer
}

// SetFile allows setting the application configuration.
func (ara *AppRunArgs) SetFile(file config.File) {
	ara.File = file
}

// SetClient allows setting the HTTP client.
func (ara *AppRunArgs) SetClient(client api.HTTPClient) {
	ara.HTTPClient = client
}

// SetStdin allows setting stdin.
func (ara *AppRunArgs) SetStdin(stdin io.Reader) {
	ara.In = stdin
}

// SetEnv allows setting the environment.
func (ara *AppRunArgs) SetEnv(env config.Environment) {
	ara.Env = env
}

// SetAppConfigFile allows setting the path to the app config file.
func (ara *AppRunArgs) SetAppConfigFile(fpath string) {
	ara.AppConfigFile = fpath
}

// SetClientFactory allows setting the mocked API.
func (ara *AppRunArgs) SetClientFactory(api mock.API) {
	ara.ClientFactory = mock.APIClient(api)
}
