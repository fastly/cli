package serviceversion

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v3/fastly"
)

// ActivateCommand calls the Fastly API to activate a service version.
type ActivateCommand struct {
	cmd.Base
	manifest manifest.Data
	Input    fastly.ActivateVersionInput
}

// NewActivateCommand returns a usable command registered under the parent.
func NewActivateCommand(parent cmd.Registerer, globals *config.Data) *ActivateCommand {
	var c ActivateCommand
	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("activate", "Activate a Fastly service version")
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("version", "Number of version you wish to activate").Required().IntVar(&c.Input.ServiceVersion)
	return &c
}

// Exec invokes the application logic for the command.
func (c *ActivateCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return errors.ErrNoServiceID
	}
	c.Input.ServiceID = serviceID

	v, err := c.Globals.Client.ActivateVersion(&c.Input)
	if err != nil {
		return err
	}

	text.Success(out, "Activated service %s version %d", v.ServiceID, c.Input.ServiceVersion)
	return nil
}
