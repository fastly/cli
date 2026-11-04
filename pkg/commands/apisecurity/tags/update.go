package tags

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"
	"github.com/fastly/go-fastly/v19/fastly/apisecurity/operations"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// UpdateCommand calls the Fastly API to update an operation tag.
type UpdateCommand struct {
	argparser.Base

	// Required.
	TagID       string `required:"" help:"Tag ID"`
	NameFlag    string `name:"name" required:"" help:"Updated name of the operation tag"`
	Description string `required:"" help:"Updated description of the operation tag"`

	// Optional.
	ServiceID   *string                         `help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`
	ServiceName argparser.OptionalServiceNameID `help:"The name of the service"`

	argparser.JSONOutput
}

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *UpdateCommand) AfterApply() error {
	if c.ServiceID != nil {
		c.Globals.Manifest.Flag.ServiceID = *c.ServiceID
	}
	return nil
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	serviceID, source, flag, err := argparser.ServiceID(c.ServiceName, *c.Globals.Manifest, c.Globals.APIClient, c.Globals.ErrLog)
	if err != nil {
		return err
	}
	if c.Globals.Verbose() {
		argparser.DisplayServiceID(serviceID, flag, source, out)
	}

	if serviceID == "" {
		return errors.New("service-id is required")
	}

	if c.NameFlag == "" {
		return errors.New("--name is required")
	}

	if c.Description == "" {
		return errors.New("--description is required")
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	input := &operations.UpdateTagInput{
		Description: &c.Description,
		Name:        &c.NameFlag,
		ServiceID:   &serviceID,
		TagID:       &c.TagID,
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
