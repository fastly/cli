package configstore

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v7/fastly"
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
func (cmd *DescribeCommand) Exec(_ io.Reader, out io.Writer) error {
	if cmd.Globals.Verbose() && cmd.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	cs, err := cmd.Globals.APIClient.GetConfigStore(&cmd.input)
	if err != nil {
		cmd.Globals.ErrLog.Add(err)
		return err
	}

	var csm *fastly.ConfigStoreMetadata
	if cmd.metadata {
		csm, err = cmd.Globals.APIClient.GetConfigStoreMetadata(&fastly.GetConfigStoreMetadataInput{
			ID: cmd.input.ID,
		})
		if err != nil {
			cmd.Globals.ErrLog.Add(err)
			return err
		}
	}

	if cmd.JSONOutput.Enabled {
		// Create an ad-hoc structure for JSON representation of the config store
		// and its metadata.
		data := struct {
			*fastly.ConfigStore
			Metadata *fastly.ConfigStoreMetadata `json:"metadata,omitempty"`
		}{
			ConfigStore: cs,
			Metadata:    csm,
		}

		if ok, err := cmd.WriteJSON(out, data); ok {
			return err
		}
	}

	text.PrintConfigStore(out, cs, csm)

	return nil
}
