package configstore

import (
	"io"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent argparser.Registerer, g *global.Data) *DescribeCommand {
	c := DescribeCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}

	c.CmdClause = parent.Command("describe", "Retrieve a single config store").Alias("get")

	// Required.
	c.RegisterFlag(argparser.StoreIDFlag(&c.input.StoreID)) // --store-id

	// Optional.
	c.RegisterFlagBool(c.JSONFlag()) // --json
	c.RegisterFlagBool(argparser.BoolFlagOpts{
		Name:        "metadata",
		Short:       'm',
		Description: "Include config store metadata",
		Dst:         &c.metadata,
	})

	return &c
}

// DescribeCommand calls the Fastly API to describe an appropriate resource.
type DescribeCommand struct {
	argparser.Base
	argparser.JSONOutput
	input    fastly.GetConfigStoreInput
	metadata bool
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	cs, err := c.Globals.APIClient.GetConfigStore(&c.input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	var csm *fastly.ConfigStoreMetadata
	if c.metadata {
		csm, err = c.Globals.APIClient.GetConfigStoreMetadata(&fastly.GetConfigStoreMetadataInput{
			StoreID: c.input.StoreID,
		})
		if err != nil {
			c.Globals.ErrLog.Add(err)
			return err
		}
	}

	if c.JSONOutput.Enabled {
		// Create an ad-hoc structure for JSON representation of the config store
		// and its metadata.
		data := struct {
			*fastly.ConfigStore
			Metadata *fastly.ConfigStoreMetadata `json:"metadata,omitempty"`
		}{
			ConfigStore: cs,
			Metadata:    csm,
		}

		if ok, err := c.WriteJSON(out, data); ok {
			return err
		}
	}

	text.PrintConfigStore(out, cs, csm)

	return nil
}
