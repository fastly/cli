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

// NewCreateStoreCommand returns a usable command registered under the parent.
func NewCreateStoreCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *CreateStoreCommand {
	c := CreateStoreCommand{
		Base: cmd.Base{
			Globals: globals,
		},
		manifest: data,
	}

	c.CmdClause = parent.Command("create", "Create a new secret store")

	// Required.
	c.RegisterFlag(storeNameFlag(&c.Input.Name)) // --name

	// Optional.
	c.RegisterFlagBool(c.jsonFlag()) // --json

	return &c
}

// CreateStoreCommand calls the Fastly API to create an appropriate resource.
type CreateStoreCommand struct {
	cmd.Base
	jsonOutput

	Input    fastly.CreateSecretStoreInput
	manifest manifest.Data
}

// Exec invokes the application logic for the command.
func (cmd *CreateStoreCommand) Exec(_ io.Reader, out io.Writer) error {
	if cmd.Globals.Verbose() && cmd.jsonOutput.enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	o, err := cmd.Globals.APIClient.CreateSecretStore(&cmd.Input)
	if err != nil {
		cmd.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := cmd.WriteJSON(out, o); ok {
		return err
	}

	text.Success(out, "Created secret store %s (name %s)", o.ID, o.Name)

	return nil
}
