package operations

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/fastly/go-fastly/v14/fastly"
	"github.com/fastly/go-fastly/v14/fastly/apisecurity/operations"

	"github.com/fastly/kingpin"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// UpdateCommand calls the Fastly API to update an operation.
type UpdateCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	serviceName argparser.OptionalServiceNameID
	operationID string

	// Optional.
	description argparser.OptionalString
	tagIDs      argparser.OptionalStringSlice
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent argparser.Registerer, g *global.Data) *UpdateCommand {
	c := UpdateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("update", "Update an operation")

	// Required.
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagServiceIDName,
		Description: argparser.FlagServiceIDDesc,
		Dst:         &g.Manifest.Flag.ServiceID,
		Short:       's',
		Required:    true,
	})
	c.RegisterFlag(argparser.StringFlagOpts{
		Action:      c.serviceName.Set,
		Name:        argparser.FlagServiceName,
		Description: argparser.FlagServiceNameDesc,
		Dst:         &c.serviceName.Value,
	})
	c.CmdClause.Flag("operation-id", "The unique identifier of the operation").Required().StringVar(&c.operationID)

	// Optional.
	c.CmdClause.Flag("description", "Updated description of what the operation does").Action(c.description.Set).StringVar(&c.description.Value)
	c.CmdClause.Flag("tag-ids", "Comma-separated list of tag IDs to associate with the operation").Action(c.tagIDs.Set).StringsVar(&c.tagIDs.Value, kingpin.Separator(","))
	c.RegisterFlagBool(c.JSONFlag()) // --json

	return &c
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	if !c.description.WasSet && !c.tagIDs.WasSet {
		return fmt.Errorf("error parsing arguments: must provide at least one field to update (--description or --tag-ids)")
	}

	serviceID, source, flag, err := argparser.ServiceID(c.serviceName, *c.Globals.Manifest, c.Globals.APIClient, c.Globals.ErrLog)
	if err != nil {
		return err
	}
	if c.Globals.Verbose() {
		argparser.DisplayServiceID(serviceID, flag, source, out)
	}

	input := c.constructInput(serviceID)

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	o, err := operations.Update(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":   serviceID,
			"Operation ID": c.operationID,
		})
		return err
	}

	if ok, err := c.WriteJSON(out, o); ok {
		return err
	}

	text.Success(out, "Updated operation %s %s%s (ID: %s)", strings.ToUpper(o.Method), o.Domain, o.Path, o.ID)

	if c.Globals.Verbose() {
		fmt.Fprintln(out)
		if o.Description != "" {
			fmt.Fprintf(out, "Description: %s\n", o.Description)
		}
		if len(o.TagIDs) > 0 {
			fmt.Fprintf(out, "Tags: %d associated\n", len(o.TagIDs))
		}
		if o.UpdatedAt != "" {
			fmt.Fprintf(out, "Updated At: %s\n", o.UpdatedAt)
		}
	}

	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) constructInput(serviceID string) *operations.UpdateInput {
	input := &operations.UpdateInput{
		ServiceID:   &serviceID,
		OperationID: &c.operationID,
	}

	if c.description.WasSet {
		input.Description = &c.description.Value
	}

	if c.tagIDs.WasSet {
		input.TagIDs = c.tagIDs.Value
	}

	return input
}
