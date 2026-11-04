package kvstoreentry

import (
	"context"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
)

// DescribeCommand calls the Fastly API to fetch the value of a key from a kv store.
type DescribeCommand struct {
	argparser.Base

	Input fastly.GetKVStoreItemInput `kong:"-"`

	// Required.
	Key     string `short:"k" required:"" help:"Key name"`
	StoreID string `short:"s" required:"" help:"Store ID"`

	// Optional.
	argparser.JSONOutput
}

// BeforeApply sets the command up before kong applies the command line.
func (c *DescribeCommand) BeforeApply() error {
	// This argument suppresses the 'Fastly API' output from the global verbose command.
	c.SuppressVerbose = true
	return nil
}

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *DescribeCommand) AfterApply() error {
	c.Input.Key = c.Key
	c.Input.StoreID = c.StoreID
	return nil
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	if c.Globals.Flags.Verbose {
		// We won't be supporting a --verbose flag here as there wouldn't be any additional output to provide.
		return fmt.Errorf("the 'describe' command does not support the --verbose flag")
	}

	item, err := c.Globals.APIClient.GetKVStoreItem(context.TODO(), &c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if c.JSONOutput.Enabled {
		o := map[string]interface{}{
			"key":        c.Input.Key,
			"generation": fmt.Sprintf("%d", item.Generation),
			"metadata":   item.Metadata,
		}
		if ok, err := c.WriteJSON(out, o); ok {
			return err
		}
		return nil
	}

	// IMPORTANT: Don't use `text` package as binary data can be messed up.
	// Print the key attributes.
	fmt.Fprintf(out, "Key: %s\n", c.Input.Key)
	fmt.Fprintf(out, "Generation: %d\n", item.Generation)
	fmt.Fprintf(out, "Metadata: %s\n", item.Metadata)
	return nil
}
