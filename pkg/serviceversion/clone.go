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

// CloneCommand calls the Fastly API to clone a service version.
type CloneCommand struct {
	cmd.Base
	manifest       manifest.Data
	Input          fastly.CloneVersionInput
	serviceVersion cmd.OptionalServiceVersion
}

// NewCloneCommand returns a usable command registered under the parent.
func NewCloneCommand(parent cmd.Registerer, globals *config.Data) *CloneCommand {
	var c CloneCommand
	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("clone", "Clone a Fastly service version")
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.NewServiceVersionFlag(cmd.ServiceVersionFlagOpts{Dst: &c.serviceVersion.Value})
	return &c
}

// Exec invokes the application logic for the command.
func (c *CloneCommand) Exec(in io.Reader, out io.Writer) error {
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

	ver, err := c.Globals.Client.CloneVersion(&c.Input)
	if err != nil {
		return err
	}

	text.Success(out, "Cloned service %s version %d to version %d", ver.ServiceID, c.Input.ServiceVersion, ver.Number)
	return nil
}
