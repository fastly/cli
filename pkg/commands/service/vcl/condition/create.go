package condition

import (
	"context"
	"fmt"
	"io"
	"slices"

	"github.com/alecthomas/kong"

	"github.com/fastly/go-fastly/v19/fastly"

	"4d63.com/optional"

	"github.com/fastly/cli/internal/kongcompat"
	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// ConditionTypes are the allowed input values for the --type flag.
// Reference: https://www.fastly.com/documentation/reference/api/vcl-services/condition
var ConditionTypes = []string{"REQUEST", "CACHE", "RESPONSE", "PREFETCH"}

// CreateCommand calls the Fastly API to create an appropriate resource.
type CreateCommand struct {
	argparser.Base

	// Required.
	serviceVersion argparser.OptionalServiceVersion

	// Optional flags

	// Optional.
	AutoClone     argparser.OptionalAutoClone     `name:"autoclone" help:"If the selected service version is not editable, clone it and use the clone."`
	ConditionType argparser.OptionalString        `name:"type" hints:"REQUEST,CACHE,RESPONSE,PREFETCH" help:"Condition type"`
	NameFlag      argparser.OptionalString        `name:"name" short:"n" help:"Condition name"`
	Priority      argparser.OptionalInt           `help:"Condition priority"`
	ServiceName   argparser.OptionalServiceNameID `help:"The name of the service"`
	Statement     argparser.OptionalString        `help:"Condition statement"`

	// Required flags
	Version string `required:"" help:"'latest', 'active', 'staged', or the number of a specific Fastly service version"`

	ServiceID *string `short:"s" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`
}

// AfterApply checks the values kong cannot check by itself, then hands the
// parsed flags over to the fields the command reads them from.
func (c *CreateCommand) AfterApply(ctx *kong.Context) error {
	if (c.ConditionType.Value != "" || kongcompat.Given(ctx, "type")) && !slices.Contains([]string{"REQUEST", "CACHE", "RESPONSE", "PREFETCH"}, c.ConditionType.Value) {
		return fmt.Errorf("enum value must be one of REQUEST,CACHE,RESPONSE,PREFETCH, got '%s'", c.ConditionType.Value)
	}
	c.serviceVersion.Value = c.Version
	if c.ServiceID != nil {
		c.Globals.Manifest.Flag.ServiceID = *c.ServiceID
	}
	return nil
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
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

	input := fastly.CreateConditionInput{
		ServiceID:      serviceID,
		ServiceVersion: fastly.ToValue(serviceVersion.Number),
	}

	if c.NameFlag.WasSet {
		input.Name = &c.NameFlag.Value
	}
	if c.Statement.WasSet {
		input.Statement = &c.Statement.Value
	}
	if c.ConditionType.WasSet {
		input.Type = &c.ConditionType.Value
	}
	if c.Priority.WasSet {
		input.Priority = &c.Priority.Value
	}
	r, err := c.Globals.APIClient.CreateCondition(context.TODO(), &input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": fastly.ToValue(serviceVersion.Number),
		})
		return err
	}

	text.Success(out,
		"Created condition %s (service %s version %d)",
		fastly.ToValue(r.Name),
		fastly.ToValue(r.ServiceID),
		fastly.ToValue(r.ServiceVersion),
	)
	return nil
}

// FlagOrder lists the flags in the order kingpin showed them in help, which
// the fields do not follow.
func (c *CreateCommand) FlagOrder() []string {
	return []string{
		"version", "autoclone", "name", "priority", "statement", "type",
		"service-id", "service-name",
	}
}
