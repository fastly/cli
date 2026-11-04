package kvstoreentry

import (
	"context"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// ListCommand calls the Fastly API to list the keys for a given kv store.
type ListCommand struct {
	argparser.Base

	// Required.
	StoreID string `short:"s" required:"" help:"Store ID"`

	// Optional.
	Consistency string `default:"strong" enum:"eventual,strong" hints:"eventual,strong" help:"Determines accuracy of results. i.e. 'eventual' uses caching to improve performance"`
	Prefix      string `help:"Restrict results to items whose keys match this prefix"`
	argparser.JSONOutput
	Input fastly.ListKVStoreKeysInput `kong:"-"`
}

// ConsistencyOptions is a list of allowed consistency values.
var ConsistencyOptions = []string{
	"eventual",
	"strong",
}

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *ListCommand) AfterApply() error {
	c.Input.StoreID = c.StoreID
	return nil
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	var (
		cursor string
		keys   []string
		ok     bool
	)

	c.Input.Cursor = cursor

	spinner, err := text.NewSpinner(out)
	if err != nil {
		return err
	}
	msg := "Getting data"

	// A spinner produces output and is incompatible with JSON expected output.
	if !c.JSONOutput.Enabled {
		err := spinner.Start()
		if err != nil {
			return err
		}
		spinner.Message(msg + "... (this can take a few minutes depending on the number of entries)")
	}

	switch c.Consistency {
	case "eventual":
		c.Input.Consistency = fastly.ConsistencyEventual
	case "strong":
		c.Input.Consistency = fastly.ConsistencyStrong
	}

	if c.Prefix != "" {
		c.Input.Prefix = c.Prefix
	}

	for {
		o, err := c.Globals.APIClient.ListKVStoreKeys(context.TODO(), &c.Input)
		if err != nil {
			c.Globals.ErrLog.Add(err)
			if !c.JSONOutput.Enabled {
				spinner.StopFailMessage(msg)
				spinErr := spinner.StopFail()
				if spinErr != nil {
					return fmt.Errorf(text.SpinnerErrWrapper, spinErr, err)
				}
			}
			return err
		}

		keys = append(keys, o.Data...)

		c.Input.Cursor, ok = o.Meta["next_cursor"]
		if !ok {
			break
		}
	}

	if !c.JSONOutput.Enabled {
		spinner.StopMessage(msg)
		err := spinner.Stop()
		if err != nil {
			return err
		}
	}

	if keys == nil {
		if ok, err := c.WriteJSON(out, []string{}); ok {
			return err
		}
		text.Break(out)
		text.Output(out, "no keys")
		return nil
	}

	if ok, err := c.WriteJSON(out, keys); ok {
		return err
	}

	if c.Globals.Flags.Verbose {
		text.PrintKVStoreKeys(out, "", keys)
		return nil
	}

	for _, k := range keys {
		text.Output(out, k)
	}
	return nil
}
