package configstoreentry

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v7/fastly"
)

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent cmd.Registerer, g *global.Data, m manifest.Data) *UpdateCommand {
	c := UpdateCommand{
		Base: cmd.Base{
			Globals: g,
		},
		manifest: m,
	}

	c.CmdClause = parent.Command("update", "Update a config store item")

	// Required.
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        "key",
		Short:       'k',
		Description: "Item name",
		Dst:         &c.input.Key,
		Required:    true,
	})
	c.RegisterFlag(cmd.StoreIDFlag(&c.input.StoreID)) // --store-id

	// One of these must be set.
	c.RegisterFlagBool(cmd.BoolFlagOpts{
		Name:        "stdin",
		Description: "Read item value from STDIN. If set, --value will be ignored",
		Dst:         &c.stdin,
		Required:    false,
	})
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        "value",
		Description: "Item value. Required unless --stdin is set",
		Dst:         &c.input.Value,
		Required:    false,
	})

	// Optional.
	c.RegisterFlagBool(c.JSONFlag()) // --json
	c.RegisterFlagBool(cmd.BoolFlagOpts{
		Name:        "upsert",
		Short:       'u',
		Description: "If true, insert or update an entry in a config store. Otherwise, only update",
		Dst:         &c.input.Upsert,
	})

	return &c
}

// UpdateCommand calls the Fastly API to update an appropriate resource.
type UpdateCommand struct {
	cmd.Base
	cmd.JSONOutput

	input    fastly.UpdateConfigStoreItemInput
	stdin    bool
	manifest manifest.Data
}

// Exec invokes the application logic for the command.
func (cmd *UpdateCommand) Exec(in io.Reader, out io.Writer) error {
	if cmd.Globals.Verbose() && cmd.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	if cmd.stdin {
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

		cmd.input.Value = string(value)
	} else if cmd.input.Value == "" {
		return errNoValue
	}

	if len(cmd.input.Key) > maxKeyLen {
		return errMaxKeyLen
	}
	if len(cmd.input.Value) > maxValueLen {
		return errMaxValueLen
	}

	o, err := cmd.Globals.APIClient.UpdateConfigStoreItem(&cmd.input)
	if err != nil {
		cmd.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := cmd.WriteJSON(out, o); ok {
		return err
	}

	var action string
	if cmd.input.Upsert {
		// The Fastly API does not provide a way to determine if
		// an item was created or updated when using 'upsert' operation.
		action = "Created or updated"
	} else {
		action = "Updated"
	}

	text.Success(out, "%s config store item %s in store %s", action, o.Key, o.StoreID)

	return nil
}
