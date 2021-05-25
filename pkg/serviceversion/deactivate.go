package serviceversion

import (
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v3/fastly"
)

// DeactivateCommand calls the Fastly API to deactivate a service version.
type DeactivateCommand struct {
	common.Base
	manifest       manifest.Data
	Input          fastly.DeactivateVersionInput
	serviceVersion common.OptionalServiceVersion
}

// NewDeactivateCommand returns a usable command registered under the parent.
func NewDeactivateCommand(parent common.Registerer, globals *config.Data) *DeactivateCommand {
	var c DeactivateCommand
	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("deactivate", "Deactivate a Fastly service version")
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.NewServiceVersionFlag(common.ServiceVersionFlagOpts{Dst: &c.serviceVersion.Value})
	return &c
}

// Exec invokes the application logic for the command.
func (c *DeactivateCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return errors.ErrNoServiceID
	}
	c.Input.ServiceID = serviceID

	v, err := c.serviceVersion.Parse(c.Input.ServiceID, c.Globals.Client)
	if err != nil {
		return err
	}
	c.Input.ServiceVersion = v.Number

	ver, err := c.Globals.Client.DeactivateVersion(&c.Input)
	if err != nil {
		return err
	}

	text.Success(out, "Deactivated service %s version %d", ver.ServiceID, c.Input.ServiceVersion)
	return nil
}
