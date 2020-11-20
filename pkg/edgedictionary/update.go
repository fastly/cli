package edgedictionary

import (
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/fastly"
)

// UpdateCommand calls the Fastly API to describe a service.
type UpdateCommand struct {
	common.Base
	manifest manifest.Data
	Input    fastly.UpdateDictionaryInput
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent common.Registerer, globals *config.Data) *UpdateCommand {
	var c UpdateCommand
	c.Globals = globals
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("update", "Update name of dictionary on a Fastly service version").Alias("get")
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.Input.Version)
	c.CmdClause.Flag("name", "Old name of Dictionary").Short('n').Required().StringVar(&c.Input.Name)
	c.CmdClause.Flag("new-name", "New name of Dictionary").Required().StringVar(&c.Input.NewName)
	return &c
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return errors.ErrNoServiceID
	}
	c.Input.Service = serviceID

	d, err := c.Globals.Client.UpdateDictionary(&c.Input)
	if err != nil {
		return err
	}

	text.Success(out, "Updated dictionary %s to %s (service %s version %d)", c.Input.Name, d.Name, d.ServiceID, d.Version)

	if c.Globals.Verbose() {
		text.Output(out, "Service ID: %s", d.ServiceID)
		text.Output(out, "Version: %d", d.Version)
		text.PrintDictionary(out, "", d)
	}

	return nil
}
