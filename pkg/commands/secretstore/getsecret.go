package secretstore

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v7/fastly"
)

// NewGetSecretCommand returns a usable command registered under the parent.
func NewGetSecretCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *GetSecretCommand {
	c := GetSecretCommand{
		Base: cmd.Base{
			Globals: globals,
		},
		manifest: data,
	}

	c.CmdClause = parent.Command("get", "Get secret")

	// Required.
	c.RegisterFlag(secretNameFlag(&c.Input.Name)) // --name
	c.RegisterFlag(storeIDFlag(&c.Input.ID))      // --store-id

	// Optional.
	c.RegisterFlagBool(c.jsonFlag()) // --json

	return &c
}

// GetSecretCommand calls the Fastly API to list the available secret stores.
type GetSecretCommand struct {
	cmd.Base
	jsonOutput

	Input    fastly.GetSecretInput
	manifest manifest.Data
}

// Exec invokes the application logic for the command.
func (cmd *GetSecretCommand) Exec(_ io.Reader, out io.Writer) error {
	if cmd.Globals.Verbose() && cmd.jsonOutput.enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	o, err := cmd.Globals.APIClient.GetSecret(&cmd.Input)
	if err != nil {
		cmd.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := cmd.WriteJSON(out, o); ok {
		return err
	}

	text.PrintSecret(out, "", o)

	return nil
}
