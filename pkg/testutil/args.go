package testutil

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/update"
)

// Args is a simple wrapper function designed to accept a CLI command
// (including flags) and return it as a slice for consumption by app.Run().
func Args(args string) []string {
	return strings.Split(args, " ")
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
	HTTPClient    *http.Client
	In            io.Reader
	Out           io.Writer
}

// NewAppRunArgs returns a struct that can be used to populate a call to
// app.Run() while the majority of fields will be pre-populated and only those
// fields commonly changed for testing purposes will need to be provided.
func NewAppRunArgs(args []string, api mock.API, buf *bytes.Buffer) *AppRunArgs {
	return &AppRunArgs{
		AppConfigFile: "/dev/null",
		Args:          args,
		ClientFactory: mock.APIClient(api),
		Env:           config.Environment{},
		File:          config.File{},
		HTTPClient:    http.DefaultClient,
		Out:           buf,
	}
}
