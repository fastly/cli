package dictionaryitem

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v6/fastly"
)

// CreateCommand calls the Fastly API to create a dictionary item.
type CreateCommand struct {
	cmd.Base
	manifest    manifest.Data
	Input       fastly.CreateDictionaryItemInput
	serviceName cmd.OptionalServiceNameID
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *CreateCommand {
	var c CreateCommand
	c.Globals = globals
	c.manifest = data
	c.CmdClause = parent.Command("create", "Create a new item on a Fastly edge dictionary")
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagServiceIDName,
		Description: cmd.FlagServiceIDDesc,
		Dst:         &c.manifest.Flag.ServiceID,
		Short:       's',
	})
	c.RegisterFlag(cmd.StringFlagOpts{
		Action:      c.serviceName.Set,
		Name:        cmd.FlagServiceName,
		Description: cmd.FlagServiceDesc,
		Dst:         &c.serviceName.Value,
	})
	c.CmdClause.Flag("dictionary-id", "Dictionary ID").Required().StringVar(&c.Input.DictionaryID)
	c.CmdClause.Flag("key", "Dictionary item key").Required().StringVar(&c.Input.ItemKey)
	c.CmdClause.Flag("value", "Dictionary item value").Required().StringVar(&c.Input.ItemValue)
	return &c
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, source, flag, err := cmd.ServiceID(c.serviceName, c.manifest, c.Globals.APIClient, c.Globals.ErrLog)
	if err != nil {
		return err
	}
	if c.Globals.Verbose() {
		cmd.DisplayServiceID(serviceID, flag, source, out)
	}

	c.Input.ServiceID = serviceID

	_, err = c.Globals.APIClient.CreateDictionaryItem(&c.Input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID": serviceID,
		})
		return err
	}

	text.Success(out, "Created dictionary item %s (service %s, dictionary %s)", c.Input.ItemKey, c.Input.ServiceID, c.Input.DictionaryID)

	return nil
}
