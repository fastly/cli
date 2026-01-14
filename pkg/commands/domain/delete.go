package domain

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v12/fastly"
	"github.com/fastly/go-fastly/v12/fastly/domainmanagement/v1/domains"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// DeleteCommand calls the Fastly API to delete domains.
type DeleteCommand struct {
	argparser.Base
	domainID string
}

// NewDeleteCommand returns a usable command registered under the parent.
func NewDeleteCommand(parent argparser.Registerer, g *global.Data) *DeleteCommand {
	c := DeleteCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("delete", "Delete a domain").Alias("remove")

	// Required.
	c.CmdClause.Flag("domain-id", "The Domain Identifier (UUID)").Required().StringVar(&c.domainID)

	return &c
}

// Exec invokes the application logic for the command.
func (c *DeleteCommand) Exec(_ io.Reader, out io.Writer) error {
	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	input := &domains.DeleteInput{
		DomainID: &c.domainID,
	}

	err := domains.Delete(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Domain ID": c.domainID,
		})
		return err
	}

	text.Success(out, "Deleted domain (domain-id: %s)", c.domainID)
	return nil
}
