package domain

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"
	"github.com/fastly/go-fastly/v19/fastly/domainmanagement/v1/domains"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
)

// DescribeCommand calls the Fastly API to describe a domain.
type DescribeCommand struct {
	argparser.Base

	// Required.
	DomainID string `required:"" help:"The Domain Identifier (UUID)"`

	// Optional.
	argparser.JSONOutput
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	input := &domains.GetInput{
		DomainID: &c.DomainID,
	}

	d, err := domains.Get(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Domain ID": c.DomainID,
		})
		return err
	}

	if ok, err := c.WriteJSON(out, d); ok {
		return err
	}

	if d != nil {
		cl := []domains.Data{*d}
		if c.Globals.Verbose() {
			printVerbose(out, cl)
		} else {
			printSummary(out, cl)
		}
	}
	return nil
}
