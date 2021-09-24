package version

import (
	"fmt"
	"io"
	"strings"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/revision"
	"github.com/fastly/cli/pkg/useragent"
	"github.com/fastly/go-fastly/v5/fastly"
)

func init() {
	// Override the go-fastly UserAgent value by prepending the CLI version.
	//
	// Results in a header similar too:
	// User-Agent: FastlyCLI/0.1.0, FastlyGo/1.5.0 (1.13.0)
	fastly.UserAgent = fmt.Sprintf("%s, %s", useragent.Name, fastly.UserAgent)
}

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	cmd.Base
}

// NewRootCommand returns a new command registered in the parent.
func NewRootCommand(parent cmd.Registerer) *RootCommand {
	var c RootCommand
	c.CmdClause = parent.Command("version", "Display version information for the Fastly CLI")
	return &c
}

// Exec implements the command interface.
func (c *RootCommand) Exec(in io.Reader, out io.Writer) error {
	fmt.Fprintf(out, "Fastly CLI version %s (%s)\n", revision.AppVersion, revision.GitCommit)
	fmt.Fprintf(out, "Built with %s\n", revision.GoVersion)
	return nil
}

// IsPreRelease determines if the given app version is a pre-release.
//
// NOTE: this is indicated by the presence of a hyphen, e.g. v1.0.0-rc.1
func IsPreRelease(version string) bool {
	return strings.Contains(version, "-")
}
