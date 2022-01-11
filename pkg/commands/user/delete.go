package user

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v5/fastly"
)

// NewDeleteCommand returns a usable command registered under the parent.
func NewDeleteCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *DeleteCommand {
	var c DeleteCommand
	c.CmdClause = parent.Command("delete", "Delete a user of the Fastly API and web interface").Alias("remove")
	c.Globals = globals
	c.manifest = data
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
func (c *DeleteCommand) Exec(in io.Reader, out io.Writer) error {
	_, s := c.Globals.Token()
	if s == config.SourceUndefined {
		return errors.ErrNoToken
	}

	input := c.constructInput()

	err := c.Globals.Client.DeleteUser(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
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
