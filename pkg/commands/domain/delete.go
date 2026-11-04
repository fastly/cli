package domain

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"
	"github.com/fastly/go-fastly/v19/fastly/domainmanagement/v1/domains"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/text"
)

// DeleteCommand calls the Fastly API to delete domains.
type DeleteCommand struct {
	argparser.Base

	// Required.
	DomainID string `required:"" help:"The Domain Identifier (UUID)"`
}

// Exec invokes the application logic for the command.
func (c *DeleteCommand) Exec(_ io.Reader, out io.Writer) error {
	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	input := &domains.DeleteInput{
		DomainID: &c.DomainID,
	}

	err := domains.Delete(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Domain ID": c.DomainID,
		})
		return err
	}

	text.Success(out, "Deleted domain (domain-id: %s)", c.DomainID)
	return nil
}
