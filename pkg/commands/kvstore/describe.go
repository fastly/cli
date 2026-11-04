package kvstore

import (
	"context"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// DescribeCommand calls the Fastly API to fetch the value of a key from a kv store.
type DescribeCommand struct {
	argparser.Base

	// Required.
	StoreID string `short:"s" required:"" help:"Store ID"`

	// Optional.
	argparser.JSONOutput

	Input fastly.GetKVStoreInput `kong:"-"`
}

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *DescribeCommand) AfterApply() error {
	c.Input.StoreID = c.StoreID
	return nil
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	o, err := c.Globals.APIClient.GetKVStore(context.TODO(), &c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, o); ok {
		return err
	}

	text.PrintKVStore(out, "", o)
	return nil
}
