package condition

import (
	"context"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"4d63.com/optional"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// UpdateCommand calls the Fastly API to update an appropriate resource.
type UpdateCommand struct {
	argparser.Base
	input          fastly.UpdateConditionInput
	serviceVersion argparser.OptionalServiceVersion

	// Required flags
	NameFlag string `name:"name" short:"n" required:"" help:"Domain name"`
	Version  string `required:"" help:"'latest', 'active', 'staged', or the number of a specific Fastly service version"`

	// Optional flags
	AutoClone argparser.OptionalAutoClone `name:"autoclone" help:"If the selected service version is not editable, clone it and use the clone."`

	NewName       argparser.OptionalString `help:"New condition name"`
	Priority      argparser.OptionalInt    `help:"Condition priority"`
	Statement     argparser.OptionalString `help:"Condition statement"`
	ConditionType argparser.OptionalString `name:"type" help:"Condition type"`
	Comment       argparser.OptionalString `help:"Condition comment"`

	ServiceID   *string                         `short:"s" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`
	ServiceName argparser.OptionalServiceNameID `help:"The name of the service"`
}

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *UpdateCommand) AfterApply() error {
	c.input.Name = c.NameFlag
	c.serviceVersion.Value = c.Version
	if c.ServiceID != nil {
		c.Globals.Manifest.Flag.ServiceID = *c.ServiceID
	}
	return nil
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	serviceID, serviceVersion, err := argparser.ServiceDetails(argparser.ServiceDetailsOpts{
		Active:             optional.Of(false),
		Locked:             optional.Of(false),
		AutoCloneFlag:      c.AutoClone,
		APIClient:          c.Globals.APIClient,
		Manifest:           *c.Globals.Manifest,
		Out:                out,
		ServiceNameFlag:    c.ServiceName,
		ServiceVersionFlag: c.serviceVersion,
		VerboseMode:        c.Globals.Flags.Verbose,
	})
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": errors.ServiceVersion(serviceVersion),
		})
		return err
	}

	c.input.ServiceID = serviceID
	c.input.ServiceVersion = fastly.ToValue(serviceVersion.Number)

	// If no argument are provided, error with useful message.
	if !c.NewName.WasSet && !c.Priority.WasSet && !c.Statement.WasSet && !c.ConditionType.WasSet && !c.Comment.WasSet {
		return fmt.Errorf("error parsing arguments: must provide either --new-name, --statement, --type, --priority or --comment to update condition")
	}

	if c.NewName.WasSet {
		c.input.Name = c.NewName.Value
	}
	if c.Priority.WasSet {
		c.input.Priority = &c.Priority.Value
	}
	if c.ConditionType.WasSet {
		c.input.Type = &c.ConditionType.Value
	}
	if c.Statement.WasSet {
		c.input.Statement = &c.Statement.Value
	}
	if c.Comment.WasSet {
		c.input.Comment = &c.Comment.Value
	}

	r, err := c.Globals.APIClient.UpdateCondition(context.TODO(), &c.input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": fastly.ToValue(serviceVersion.Number),
		})
		return err
	}

	text.Success(out,
		"Updated condition %s (service %s version %d)",
		fastly.ToValue(r.Name),
		fastly.ToValue(r.ServiceID),
		fastly.ToValue(r.ServiceVersion),
	)
	return nil
}
