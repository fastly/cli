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

// ActivateCommand calls the Fastly API to activate a service version.
type ActivateCommand struct {
	common.Base
	manifest       manifest.Data
	Input          fastly.ActivateVersionInput
	serviceVersion common.OptionalServiceVersion
	autoClone      common.OptionalAutoClone
}

// NewActivateCommand returns a usable command registered under the parent.
func NewActivateCommand(parent common.Registerer, globals *config.Data) *ActivateCommand {
	var c ActivateCommand
	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("activate", "Activate a Fastly service version")
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.NewServiceVersionFlag(common.ServiceVersionFlagOpts{Dst: &c.serviceVersion.Value})
	c.NewAutoCloneFlag(c.autoClone.Set, &c.autoClone.Value)
	return &c
}

// Exec invokes the application logic for the command.
func (c *ActivateCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return errors.ErrNoServiceID
	}
	c.Input.ServiceID = serviceID

	v, err := c.serviceVersion.Parse(c.Input.ServiceID, c.Globals.Client)
	if err != nil {
		return err
	}
	v, err = c.autoClone.Parse(v, c.Input.ServiceID, c.Globals.Client)
	if err != nil {
		return err
	}
	c.Input.ServiceVersion = v.Number

	ver, err := c.Globals.Client.ActivateVersion(&c.Input)
	if err != nil {
		return err
	}

	text.Success(out, "Activated service %s version %d", ver.ServiceID, c.Input.ServiceVersion)
	return nil
}
