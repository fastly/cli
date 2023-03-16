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

// NewListServicesCommand returns a usable command registered under the parent.
func NewListServicesCommand(parent cmd.Registerer, g *global.Data, m manifest.Data) *ListServicesCommand {
	c := ListServicesCommand{
		Base: cmd.Base{
			Globals: g,
		},
		manifest: m,
	}

	c.CmdClause = parent.Command("list-services", "List config store's services")

	// Required.
	c.RegisterFlag(cmd.StoreIDFlag(&c.input.ID)) // --store-id

	// Optional.
	c.RegisterFlagBool(c.JSONFlag()) // --json

	return &c
}

// ListServicesCommand calls the Fastly API to list appropriate resources.
type ListServicesCommand struct {
	cmd.Base
	cmd.JSONOutput

	input    fastly.ListConfigStoreServicesInput
	manifest manifest.Data
}

// Exec invokes the application logic for the command.
func (cmd *ListServicesCommand) Exec(_ io.Reader, out io.Writer) error {
	if cmd.Globals.Verbose() && cmd.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	o, err := cmd.Globals.APIClient.ListConfigStoreServices(&cmd.input)
	if err != nil {
		cmd.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := cmd.WriteJSON(out, o); ok {
		return err
	}

	text.PrintConfigStoreServicesTbl(out, o)

	return nil
}
