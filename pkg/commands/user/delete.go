package user

import (
	"io"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
)

// NewDeleteCommand returns a usable command registered under the parent.
func NewDeleteCommand(parent cmd.Registerer, globals *global.Data, m manifest.Data) *DeleteCommand {
	var c DeleteCommand
	c.CmdClause = parent.Command("delete", "Delete a user of the Fastly API and web interface").Alias("remove")
	c.Globals = globals
	c.manifest = m
	c.CmdClause.Flag("id", "Alphanumeric string identifying the user").Required().StringVar(&c.id)
	return &c
}

// DeleteCommand calls the Fastly API to delete an appropriate resource.
type DeleteCommand struct {
	cmd.Base

	id       string
	manifest manifest.Data
}

// Exec invokes the application logic for the command.
func (c *DeleteCommand) Exec(_ io.Reader, out io.Writer) error {
	input := c.constructInput()

	err := c.Globals.APIClient.DeleteUser(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"User ID": c.id,
		})
		return err
	}

	text.Success(out, "Deleted user (id: %s)", c.id)
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *DeleteCommand) constructInput() *fastly.DeleteUserInput {
	var input fastly.DeleteUserInput

	input.ID = c.id

	return &input
}
