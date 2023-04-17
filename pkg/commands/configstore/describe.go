package configstore

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v8/fastly"
)

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent cmd.Registerer, g *global.Data, m manifest.Data) *DescribeCommand {
	c := DescribeCommand{
		Base: cmd.Base{
			Globals: g,
		},
		manifest: m,
	}

	c.CmdClause = parent.Command("describe", "Retrieve a single config store").Alias("get")

	// Required.
	c.RegisterFlag(cmd.StoreIDFlag(&c.input.ID)) // --store-id

	// Optional.
	c.RegisterFlagBool(c.JSONFlag()) // --json
	c.RegisterFlagBool(cmd.BoolFlagOpts{
		Name:        "metadata",
		Short:       'm',
		Description: "Include config store metadata",
		Dst:         &c.metadata,
	})

	return &c
}

// DescribeCommand calls the Fastly API to describe an appropriate resource.
type DescribeCommand struct {
	cmd.Base
	cmd.JSONOutput

	input    fastly.GetConfigStoreInput
	manifest manifest.Data
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
			ID: c.input.ID,
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
