package version

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/go-fastly/fastly"
)

var (
	// AppVersion is the semver for this version of the client, or
	// "v0.0.0-unknown". Set by `make release`.
	AppVersion string

	// GitRevision is the short git SHA associated with this build, or
	// "unknown". Set by `make release`.
	GitRevision string

	// GoVersion is the output of `go version` associated with this build, or
	// "go version unknown". Set by `make release`.
	GoVersion string

	// UserAgent is the user agent which we report in all HTTP requests to the
	// API via go-fastly.
	UserAgent string
)

func init() {
	if AppVersion == "" {
		AppVersion = "v0.0.0-unknown"
	}
	if GitRevision == "" {
		GitRevision = "unknown"
	}
	if GoVersion == "" {
		GoVersion = "go version unknown"
	}

	UserAgent = fmt.Sprintf("FastlyCLI/%s", AppVersion)

	// Override the go-fastly UserAgent value by prepending the CLI version.
	//
	// Results in a header similar too:
	// User-Agent: FastlyCLI/0.1.0, FastlyGo/1.5.0 (1.13.0)
	fastly.UserAgent = fmt.Sprintf("%s, %s", UserAgent, fastly.UserAgent)
}

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	common.Base
}

// NewRootCommand returns a new command registered in the parent.
func NewRootCommand(parent common.Registerer) *RootCommand {
	var c RootCommand
	c.CmdClause = parent.Command("version", "Display version information for the Fastly CLI")
	return &c
}

// Exec implements the command interface.
func (c *RootCommand) Exec(in io.Reader, out io.Writer) error {
	fmt.Fprintf(out, "Fastly CLI version %s (%s)\n", AppVersion, GitRevision)
	fmt.Fprintf(out, "Built with %s\n", GoVersion)
	return nil
}
