package edgedictionary

import (
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v3/fastly"
)

// ListCommand calls the Fastly API to list dictionaries
type ListCommand struct {
	common.Base
	manifest manifest.Data
	Input    fastly.ListDictionariesInput
}

// NewListCommand returns a usable command registered under the parent
func NewListCommand(parent common.Registerer, globals *config.Data) *ListCommand {
	var c ListCommand
	c.Globals = globals
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("list", "List all dictionaries on a Fastly service version")
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.Input.ServiceVersion)
	return &c
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return errors.ErrNoServiceID
	}
	c.Input.ServiceID = serviceID

	dictionaries, err := c.Globals.Client.ListDictionaries(&c.Input)
	if err != nil {
		return err
	}

	text.Output(out, "Service ID: %s", serviceID)
	text.Output(out, "Version: %d", c.Input.ServiceVersion)
	for _, dictionary := range dictionaries {
		text.PrintDictionary(out, "", dictionary)
	}

	return nil
}
