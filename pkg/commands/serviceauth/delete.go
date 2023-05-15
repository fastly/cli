package serviceauth

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v8/fastly"
)

// DeleteCommand calls the Fastly API to delete service authorizations.
type DeleteCommand struct {
	cmd.Base
	manifest manifest.Data
	Input    fastly.DeleteServiceAuthorizationInput
}

// NewDeleteCommand returns a usable command registered under the parent.
func NewDeleteCommand(parent cmd.Registerer, g *global.Data, m manifest.Data) *DeleteCommand {
	c := DeleteCommand{
		Base: cmd.Base{
			Globals: g,
		},
		manifest: m,
	}
	c.CmdClause = parent.Command("delete", "Delete service authorization").Alias("remove")

	// Required.
	c.CmdClause.Flag("id", "ID of the service authorization to delete").Required().StringVar(&c.Input.ID)
	return &c
}

// Exec invokes the application logic for the command.
func (c *DeleteCommand) Exec(_ io.Reader, out io.Writer) error {
	if err := c.Globals.APIClient.DeleteServiceAuthorization(&c.Input); err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service Authorization ID": c.Input.ID,
		})
		return err
	}

	text.Success(out, "Deleted service authorization %s", c.Input.ID)
	return nil
}
