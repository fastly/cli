package configstoreentry

import (
	"io"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent argparser.Registerer, g *global.Data) *UpdateCommand {
	c := UpdateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}

	c.CmdClause = parent.Command("update", "Update a config store item")

	// Required.
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        "key",
		Short:       'k',
		Description: "Item name",
		Dst:         &c.input.Key,
		Required:    true,
	})
	c.RegisterFlag(argparser.StoreIDFlag(&c.input.StoreID)) // --store-id

	// One of these must be set.
	c.RegisterFlagBool(argparser.BoolFlagOpts{
		Name:        "stdin",
		Description: "Read item value from STDIN. If set, --value will be ignored",
		Dst:         &c.stdin,
		Required:    false,
	})
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        "value",
		Description: "Item value. Required unless --stdin is set",
		Dst:         &c.input.Value,
		Required:    false,
	})

	// Optional.
	c.RegisterFlagBool(c.JSONFlag()) // --json
	c.RegisterFlagBool(argparser.BoolFlagOpts{
		Name:        "upsert",
		Short:       'u',
		Description: "If true, insert or update an entry in a config store. Otherwise, only update",
		Dst:         &c.input.Upsert,
	})

	return &c
}

// UpdateCommand calls the Fastly API to update an appropriate resource.
type UpdateCommand struct {
	argparser.Base
	argparser.JSONOutput
	input fastly.UpdateConfigStoreItemInput
	stdin bool
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(in io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	if c.stdin {
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

	o, err := c.Globals.APIClient.UpdateConfigStoreItem(&c.input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, o); ok {
		return err
	}

	var action string
	if c.input.Upsert {
		// The Fastly API does not provide a way to determine if
		// an item was created or updated when using 'upsert' operation.
		action = "Created or updated"
	} else {
		action = "Updated"
	}

	text.Success(out, "%s config store item %s in store %s", action, o.Key, o.StoreID)
	return nil
}
