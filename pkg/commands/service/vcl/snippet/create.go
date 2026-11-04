package snippet

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

// Locations is a list of VCL subroutines.
var Locations = []string{"init", "recv", "hash", "hit", "miss", "pass", "fetch", "error", "deliver", "log", "none"}

// AfterApply checks the values kong cannot check by itself, then hands the
// parsed flags over to the fields the command reads them from.
func (c *CreateCommand) AfterApply(ctx *kong.Context) error {
	if (c.Location.Value != "" || kongcompat.Given(ctx, "type")) && !slices.Contains([]string{"init", "recv", "hash", "hit", "miss", "pass", "fetch", "error", "deliver", "log", "none"}, c.Location.Value) {
		return fmt.Errorf("enum value must be one of init,recv,hash,hit,miss,pass,fetch,error,deliver,log,none, got '%s'", c.Location.Value)
	}
	c.serviceVersion.Value = c.Version
	if c.ServiceID != nil {
		c.Globals.Manifest.Flag.ServiceID = *c.ServiceID
	}
	return nil
}

// CreateCommand calls the Fastly API to create an appropriate resource.
type CreateCommand struct {
	argparser.Base

	// Required.
	Version string `required:"" help:"'latest', 'active', 'staged', or the number of a specific Fastly service version"`

	// Optional.
	AutoClone argparser.OptionalAutoClone `name:"autoclone" help:"If the selected service version is not editable, clone it and use the clone."`
	Content   argparser.OptionalString    `help:"VCL snippet passed as file path or content, e.g. $(< snippet.vcl)"`
	Dynamic   argparser.OptionalBool      `help:"Whether the VCL snippet is dynamic or versioned"`
	NameFlag  argparser.OptionalString    `name:"name" help:"The name of the VCL snippet"`
	Priority  argparser.OptionalString    `short:"p" help:"Priority determines execution order. Lower numbers execute first"`

	ServiceID      *string                         `short:"s" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`
	ServiceName    argparser.OptionalServiceNameID `help:"The name of the service"`
	Location       argparser.OptionalString        `name:"type" hints:"init,recv,hash,hit,miss,pass,fetch,error,deliver,log,none" help:"The location in generated VCL where the snippet should be placed"`
	serviceVersion argparser.OptionalServiceVersion
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

	input := c.constructInput(serviceID, fastly.ToValue(serviceVersion.Number))

	v, err := c.Globals.APIClient.CreateSnippet(context.TODO(), input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": fastly.ToValue(serviceVersion.Number),
		})
		return err
	}

	text.Success(out,
		"Created VCL snippet '%s' (service: %s, version: %d, dynamic: %t, snippet id: %s, type: %s, priority: %s)",
		fastly.ToValue(v.Name),
		fastly.ToValue(v.ServiceID),
		fastly.ToValue(v.ServiceVersion),
		c.Dynamic.WasSet,
		fastly.ToValue(v.SnippetID),
		c.Location.Value,
		fastly.ToValue(v.Priority),
	)
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *CreateCommand) constructInput(serviceID string, serviceVersion int) *fastly.CreateSnippetInput {
	input := fastly.CreateSnippetInput{
		Dynamic:        new(0),
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion,
	}
	if c.NameFlag.WasSet {
		input.Name = &c.NameFlag.Value
	}
	if c.Content.WasSet {
		input.Content = new(argparser.Content(c.Content.Value))
	}
	if c.Location.WasSet {
		sType := fastly.SnippetType(c.Location.Value)
		input.Type = &sType
	}
	if c.Dynamic.WasSet {
		input.Dynamic = new(1)
	}
	if c.Priority.WasSet {
		input.Priority = &c.Priority.Value
	}

	return &input
}
