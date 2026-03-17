package tags

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v13/fastly"
	"github.com/fastly/go-fastly/v13/fastly/apisecurity/operations"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// CreateCommand calls the Fastly API to create an operation tag.
type CreateCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	name        string
	serviceName argparser.OptionalServiceNameID

	// Optional.
	description argparser.OptionalString
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent argparser.Registerer, g *global.Data) *CreateCommand {
	c := CreateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}

	c.CmdClause = parent.Command("create", "Create an operation tag")

	// Required.
	c.CmdClause.Flag("name", "Name of the operation tag").Required().StringVar(&c.name)

	// Optional.
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagServiceIDName,
		Description: argparser.FlagServiceIDDesc,
		Dst:         &g.Manifest.Flag.ServiceID,
	})
	c.RegisterFlag(argparser.StringFlagOpts{
		Action:      c.serviceName.Set,
		Name:        argparser.FlagServiceName,
		Description: argparser.FlagServiceNameDesc,
		Dst:         &c.serviceName.Value,
	})
	c.CmdClause.Flag("description", "Description of the operation tag").Action(c.description.Set).StringVar(&c.description.Value)
	c.RegisterFlagBool(c.JSONFlag())

	return &c
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	serviceID, source, flag, err := argparser.ServiceID(c.serviceName, *c.Globals.Manifest, c.Globals.APIClient, c.Globals.ErrLog)
	if err != nil {
		return err
	}
	if c.Globals.Verbose() {
		argparser.DisplayServiceID(serviceID, flag, source, out)
	}

	if serviceID == "" {
		return errors.New("service-id is required")
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	input := &operations.CreateTagInput{
		ServiceID: &serviceID,
		Name:      &c.name,
	}

	if c.description.WasSet {
		input.Description = &c.description.Value
	}

	tag, err := operations.CreateTag(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, tag); ok {
		return err
	}

	text.Success(out, "Created operation tag '%s' (id: %s)", tag.Name, tag.ID)
	return nil
}
