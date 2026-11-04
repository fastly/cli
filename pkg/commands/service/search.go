package service

import (
	"context"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// SearchCommand calls the Fastly API to describe a service.
type SearchCommand struct {
	argparser.Base

	// Required.
	NameFlag string `name:"name" short:"n" required:"" help:"Service name"`

	// Optional.
	argparser.JSONOutput

	Input fastly.SearchServiceInput `kong:"-"`
}

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *SearchCommand) AfterApply() error {
	c.Input.Name = c.NameFlag
	return nil
}

// Exec invokes the application logic for the command.
func (c *SearchCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	service, err := c.Globals.APIClient.SearchService(context.TODO(), &c.Input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service Name": c.Input.Name,
		})
		return err
	}

	if ok, err := c.WriteJSON(out, service); ok {
		return err
	}

	text.PrintService(out, "", service)
	return nil
}
