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

// UpdateCommand calls the Fastly API to update an operation tag.
type UpdateCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	tagID       string
	name        string
	description string
	serviceName argparser.OptionalServiceNameID
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent argparser.Registerer, g *global.Data) *UpdateCommand {
	c := UpdateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}

	c.CmdClause = parent.Command("update", "Update an operation tag")

	// Required.
	c.CmdClause.Flag("tag-id", "Tag ID").Required().StringVar(&c.tagID)
	c.CmdClause.Flag("name", "Updated name of the operation tag").Required().StringVar(&c.name)
	c.CmdClause.Flag("description", "Updated description of the operation tag").Required().StringVar(&c.description)

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

	c.RegisterFlagBool(c.JSONFlag())

	return &c
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
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

	if c.name == "" {
		return errors.New("--name is required")
	}

	if c.description == "" {
		return errors.New("--description is required")
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	input := &operations.UpdateTagInput{
		Description: &c.description,
		Name:        &c.name,
		ServiceID:   &serviceID,
		TagID:       &c.tagID,
	}

	tag, err := operations.UpdateTag(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, tag); ok {
		return err
	}

	text.Success(out, "Updated operation tag '%s' (id: %s)", tag.Name, tag.ID)
	return nil
}
