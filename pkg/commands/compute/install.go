package compute

import (
	"errors"
	"io"
	"runtime"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
)

// InstallCommand installs the Viceroy binary that `compute serve` otherwise
// downloads on first use. It shares the install logic via viceroyInstaller
// (see viceroy.go), and is intended to pre-warm container images so that
// `compute serve` doesn't need network access at runtime.
type InstallCommand struct {
	argparser.Base
}

// NewInstallCommand returns a usable command registered under the parent.
func NewInstallCommand(parent argparser.Registerer, g *global.Data) *InstallCommand {
	var c InstallCommand
	c.Globals = g
	c.CmdClause = parent.Command("install-viceroy", "Download and install the Viceroy binary used by `compute serve`")
	return &c
}

// Exec implements the command interface.
func (c *InstallCommand) Exec(_ io.Reader, out io.Writer) error {
	if runtime.GOARCH == "386" {
		return fsterr.RemediationError{
			Inner:       errors.New("this command doesn't support the '386' architecture"),
			Remediation: "Although the Fastly CLI supports '386', https://github.com/fastly/Viceroy does not.",
		}
	}

	spinner, err := text.NewSpinner(out)
	if err != nil {
		return err
	}

	// The versioner is already seeded from the manifest's viceroy_version at
	// startup, so a pinned version is honored when run inside a project. The
	// manifest path is only used in messages, hence the default filename.
	bin, err := viceroyInstaller{
		Globals:   c.Globals,
		Versioner: c.Globals.Versioners.Viceroy,
	}.get(spinner, out, manifest.Filename)
	if err != nil {
		return err
	}

	text.Success(out, "Installed Viceroy to: %s", bin)
	return nil
}
