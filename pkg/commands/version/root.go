package version

import (
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/github"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/revision"
	"github.com/fastly/cli/pkg/useragent"
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
	argparser.Base
}

// NewRootCommand returns a new command registered in the parent.
func NewRootCommand(parent argparser.Registerer, g *global.Data) *RootCommand {
	c := RootCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("version", "Display version information for the Fastly CLI")
	return &c
}

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, out io.Writer) error {
	fmt.Fprintf(out, "Fastly CLI version %s (%s)\n", revision.AppVersion, revision.GitCommit)
	fmt.Fprintf(out, "Built with %s (%s)\n", revision.GoVersion, Now().Format("2006-01-02"))

	viceroy := filepath.Join(github.InstallDir, c.Globals.Versioners.Viceroy.BinaryName())
	// gosec flagged this:
	// G204 (CWE-78): Subprocess launched with variable
	// Disabling as we lookup the binary in a trusted location. For this to be a
	// concern the user would need to have an already compromised system where an
	// attacker could swap the actual viceroy executable for something malicious.
	/* #nosec */
	// nosemgrep
	command := exec.Command(viceroy, "--version")
	if stdoutStderr, err := command.CombinedOutput(); err == nil {
		fmt.Fprintf(out, "Viceroy version: %s", stdoutStderr)
	}

	return nil
}

// IsPreRelease determines if the given app version is a pre-release.
//
// NOTE: this is indicated by the presence of a hyphen, e.g. `v1.0.0-rc.1`.
func IsPreRelease(version string) bool {
	return strings.Contains(version, "-")
}

// Now is exposed so that we may mock it from our test file.
var Now = time.Now
