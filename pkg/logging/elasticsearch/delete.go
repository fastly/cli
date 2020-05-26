package elasticsearch

import (
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/fastly"
)

// DeleteCommand calls the Fastly API to delete Elasticsearch logging endpoints.
type DeleteCommand struct {
	common.Base
	manifest manifest.Data
	Input    fastly.DeleteElasticsearchInput
}

// NewDeleteCommand returns a usable command registered under the parent.
func NewDeleteCommand(parent common.Registerer, globals *config.Data) *DeleteCommand {
	var c DeleteCommand
	c.Globals = globals
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("delete", "Delete an Elasticsearch logging endpoint on a Fastly service version").Alias("remove")
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.Input.Version)
	c.CmdClause.Flag("name", "The name of the Elasticsearch logging object").Short('n').Required().StringVar(&c.Input.Name)
	return &c
}

// Exec invokes the application logic for the command.
func (c *DeleteCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return errors.ErrNoServiceID
	}
	c.Input.Service = serviceID

	if err := c.Globals.Client.DeleteElasticsearch(&c.Input); err != nil {
		return err
	}

	text.Success(out, "Deleted Elasticsearch logging endpoint %s (service %s version %d)", c.Input.Name, c.Input.Service, c.Input.Version)
	return nil
}
