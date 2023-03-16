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

// ConfigStoreWithMetadata combines ConfigStore and ConfigStoreMetadata
// for rendering as JSON and text.
// The included methods allow for the text package to define a matching
// interface, which eliminates the circular dependency.
type ConfigStoreWithMetadata struct {
	*fastly.ConfigStore
	Metdata *fastly.ConfigStoreMetadata `json:"metadata,omitempty"`
}

// GetConfigStore returns the ConfigStore.
func (c ConfigStoreWithMetadata) GetConfigStore() *fastly.ConfigStore {
	return c.ConfigStore
}

// GetConfigStoreMetadata returns the ConfigStoreMetadata, which may be nil.
func (c ConfigStoreWithMetadata) GetConfigStoreMetadata() *fastly.ConfigStoreMetadata {
	return c.Metdata
}

// Exec invokes the application logic for the command.
func (cmd *DescribeCommand) Exec(_ io.Reader, out io.Writer) error {
	if cmd.Globals.Verbose() && cmd.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	var (
		o   ConfigStoreWithMetadata
		err error
	)

	o.ConfigStore, err = cmd.Globals.APIClient.GetConfigStore(&cmd.input)
	if err != nil {
		cmd.Globals.ErrLog.Add(err)
		return err
	}

	if cmd.metadata {
		o.Metdata, err = cmd.Globals.APIClient.GetConfigStoreMetadata(&fastly.GetConfigStoreMetadataInput{
			ID: cmd.input.ID,
		})
		if err != nil {
			cmd.Globals.ErrLog.Add(err)
			return err
		}
	}

	if ok, err := cmd.WriteJSON(out, o); ok {
		return err
	}

	text.PrintConfigStore(out, o)

	return nil
}
