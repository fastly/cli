package configstoreentry

import (
	"io"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/v10/pkg/argparser"
	fsterr "github.com/fastly/cli/v10/pkg/errors"
	"github.com/fastly/cli/v10/pkg/global"
	"github.com/fastly/cli/v10/pkg/text"
)

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent argparser.Registerer, g *global.Data) *CreateCommand {
	c := CreateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}

	c.CmdClause = parent.Command("create", "Create a new config store item").Alias("insert")

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

	return &c
}

// CreateCommand calls the Fastly API to create an appropriate resource.
type CreateCommand struct {
	argparser.Base
	argparser.JSONOutput
	input fastly.CreateConfigStoreItemInput
	stdin bool
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(in io.Reader, out io.Writer) error {
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

	o, err := c.Globals.APIClient.CreateConfigStoreItem(&c.input)
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
