package serviceversion

import (
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/fastly"
)

// CloneCommand calls the Fastly API to clone a service version.
type CloneCommand struct {
	common.Base
	manifest manifest.Data
	Input    fastly.CloneVersionInput
}

// NewCloneCommand returns a usable command registered under the parent.
func NewCloneCommand(parent common.Registerer, globals *config.Data) *CloneCommand {
	var c CloneCommand
	c.Globals = globals
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("clone", "Clone a Fastly service version")
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("version", "Number of version you wish to clone").Required().IntVar(&c.Input.Version)
	return &c
}

// Exec invokes the application logic for the command.
func (c *CloneCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return errors.ErrNoServiceID
	}
	c.Input.Service = serviceID

	v, err := c.Globals.Client.CloneVersion(&c.Input)
	if err != nil {
		return err
	}

	text.Success(out, "Cloned service %s version %d to version %d", v.ServiceID, c.Input.Version, v.Number)
	return nil
}
