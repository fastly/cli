package serviceauth

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v7/fastly"
)

// UpdateCommand calls the Fastly API to update service authorizations.
type UpdateCommand struct {
	cmd.Base

	input    fastly.UpdateServiceAuthorizationInput
	manifest manifest.Data
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *UpdateCommand {
	var c UpdateCommand
	c.Globals = globals
	c.manifest = data
	c.CmdClause = parent.Command("update", "Update service authorization")
	c.CmdClause.Flag("id", "ID of the service authorization to delete").Required().StringVar(&c.input.ID)
	c.CmdClause.Flag("permission", "The permission the user has in relation to the service").Required().HintOptions(Permissions...).Short('p').EnumVar(&c.input.Permission, Permissions...)
	return &c
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	s, err := c.Globals.APIClient.UpdateServiceAuthorization(&c.input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service Authorization ID": c.input.ID,
		})
		return err
	}

	text.Success(out, "Updated service authorization %s", s.ID)
	return nil
}
