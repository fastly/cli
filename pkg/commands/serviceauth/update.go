package serviceauth

import (
	"io"

	"github.com/fastly/go-fastly/v10/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// UpdateCommand calls the Fastly API to update service authorizations.
type UpdateCommand struct {
	argparser.Base

	input fastly.UpdateServiceAuthorizationInput
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent argparser.Registerer, g *global.Data) *UpdateCommand {
	c := UpdateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("update", "Update service authorization")

	// Required.
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
