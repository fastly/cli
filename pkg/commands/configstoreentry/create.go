package configstoreentry

import (
	"context"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *CreateCommand) AfterApply() error {
	c.input.Key = c.Key
	c.input.StoreID = c.StoreID
	c.input.Value = c.Value
	return nil
}

// CreateCommand calls the Fastly API to create an appropriate resource.
type CreateCommand struct {
	argparser.Base
	input fastly.CreateConfigStoreItemInput

	// Required.
	Key     string `short:"k" required:"" help:"Item name"`
	StoreID string `short:"s" required:"" help:"Store ID"`

	// One of these must be set.
	STDIN bool   `help:"Read item value from STDIN. If set, --value will be ignored"`
	Value string `help:"Item value. Required unless --stdin is set"`

	// Optional.
	argparser.JSONOutput
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(in io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	if c.STDIN {
		// Determine if 'in' has data available.
		if in == nil || text.IsTTY(in) {
			return errNoSTDINData
		}

		// Must read one past limit, since LimitReader returns EOF
		// once it reads its limited number of bytes.
		value, err := io.ReadAll(io.LimitReader(in, maxValueLen+1))
		if err != nil {
			return err
		}

		c.input.Value = string(value)
	} else if c.input.Value == "" {
		return errNoValue
	}

	if len(c.input.Key) > maxKeyLen {
		return errMaxKeyLen
	}
	if len(c.input.Value) > maxValueLen {
		return errMaxValueLen
	}

	o, err := c.Globals.APIClient.CreateConfigStoreItem(context.TODO(), &c.input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, o); ok {
		return err
	}

	text.Success(out, "Created key '%s' in Config Store '%s'", o.Key, o.StoreID)
	return nil
}
