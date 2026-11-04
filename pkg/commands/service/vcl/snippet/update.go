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
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// AfterApply checks the values kong cannot check by itself, then hands the
// parsed flags over to the fields the command reads them from.
func (c *UpdateCommand) AfterApply(ctx *kong.Context) error {
	if (c.Location.Value != "" || kongcompat.Given(ctx, "type")) && !slices.Contains([]string{"init", "recv", "hash", "hit", "miss", "pass", "fetch", "error", "deliver", "log", "none"}, c.Location.Value) {
		return fmt.Errorf("enum value must be one of init,recv,hash,hit,miss,pass,fetch,error,deliver,log,none, got '%s'", c.Location.Value)
	}
	c.serviceVersion.Value = c.Version
	if c.ServiceID != nil {
		c.Globals.Manifest.Flag.ServiceID = *c.ServiceID
	}
	return nil
}

// UpdateCommand calls the Fastly API to update an appropriate resource.
type UpdateCommand struct {
	argparser.Base

	// Required.
	Version string `required:"" help:"'latest', 'active', 'staged', or the number of a specific Fastly service version"`

	// Optional.
	AutoClone      argparser.OptionalAutoClone     `name:"autoclone" help:"If the selected service version is not editable, clone it and use the clone."`
	Content        argparser.OptionalString        `help:"VCL snippet passed as file path or content, e.g. $(< snippet.vcl)"`
	Dynamic        argparser.OptionalBool          `help:"Whether the VCL snippet is dynamic or versioned"`
	NameFlag       string                          `name:"name" help:"The name of the VCL snippet to update"`
	NewName        argparser.OptionalString        `help:"New name for the VCL snippet"`
	Priority       argparser.OptionalString        `short:"p" help:"Priority determines execution order. Lower numbers execute first"`
	ServiceID      *string                         `short:"s" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`
	ServiceName    argparser.OptionalServiceNameID `help:"The name of the service"`
	serviceVersion argparser.OptionalServiceVersion
	SnippetID      string `help:"Alphanumeric string identifying a VCL Snippet"`

	// NOTE: Locations is defined in the same snippet package inside create.go
	Location argparser.OptionalString `name:"type" hints:"init,recv,hash,hit,miss,pass,fetch,error,deliver,log,none" help:"The location in generated VCL where the snippet should be placed"`
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	// in the normal case, we do not want to allow 'active' or 'locked' services to be updated,
	// so we require those states to be 'false'
	allowActive := optional.Of(false)
	allowLocked := optional.Of(false)
	if c.Dynamic.WasSet && c.Dynamic.Value {
		// in this case, we will accept all states ('active' and 'inactive', 'locked' and 'unlocked'),
		// so we mark the Optional[bool] fields as 'empty' and they will not be applied as filters
		allowActive = optional.Empty[bool]()
		allowLocked = optional.Empty[bool]()
	}
	serviceID, serviceVersion, err := argparser.ServiceDetails(argparser.ServiceDetailsOpts{
		Active:             allowActive,
		Locked:             allowLocked,
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
			"Service Version": fsterr.ServiceVersion(serviceVersion),
		})
		return err
	}

	serviceVersionNumber := fastly.ToValue(serviceVersion.Number)

	if c.Dynamic.WasSet {
		input, err := c.constructDynamicInput(serviceID, serviceVersionNumber)
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
				"Service ID":      serviceID,
				"Service Version": serviceVersionNumber,
			})
			return err
		}
		v, err := c.Globals.APIClient.UpdateDynamicSnippet(context.TODO(), input)
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
				"Service ID":      serviceID,
				"Service Version": serviceVersionNumber,
			})
			return err
		}
		text.Success(out, "Updated dynamic VCL snippet '%s' (service: %s)", fastly.ToValue(v.SnippetID), fastly.ToValue(v.ServiceID))
		return nil
	}

	input, err := c.constructInput(serviceID, serviceVersionNumber)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": serviceVersionNumber,
		})
		return err
	}
	v, err := c.Globals.APIClient.UpdateSnippet(context.TODO(), input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": serviceVersionNumber,
		})
		return err
	}
	text.Success(out,
		"Updated VCL snippet '%s' (previously: '%s', service: %s, version: %d, type: %v, priority: %s)",
		fastly.ToValue(v.Name),
		input.Name,
		fastly.ToValue(v.ServiceID),
		fastly.ToValue(v.ServiceVersion),
		fastly.ToValue(v.Type),
		fastly.ToValue(v.Priority),
	)
	return nil
}

// constructDynamicInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) constructDynamicInput(serviceID string, _ int) (*fastly.UpdateDynamicSnippetInput, error) {
	var input fastly.UpdateDynamicSnippetInput

	input.SnippetID = c.SnippetID
	input.ServiceID = serviceID

	if c.NewName.WasSet {
		return nil, fmt.Errorf("error parsing arguments: --new-name is not supported when updating a dynamic VCL snippet")
	}
	if c.SnippetID == "" {
		return nil, fmt.Errorf("error parsing arguments: must provide --snippet-id to update a dynamic VCL snippet")
	}
	if c.Content.WasSet {
		input.Content = new(argparser.Content(c.Content.Value))
	}

	return &input, nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) constructInput(serviceID string, serviceVersion int) (*fastly.UpdateSnippetInput, error) {
	var input fastly.UpdateSnippetInput

	input.Name = c.NameFlag
	input.ServiceID = serviceID
	input.ServiceVersion = serviceVersion

	if c.SnippetID != "" {
		return nil, fmt.Errorf("error parsing arguments: --snippet-id is not supported when updating a versioned VCL snippet")
	}
	if c.NameFlag == "" {
		return nil, fmt.Errorf("error parsing arguments: must provide --name to update a versioned VCL snippet")
	}
	if c.NewName.WasSet {
		input.NewName = &c.NewName.Value
	}
	if c.Priority.WasSet {
		input.Priority = &c.Priority.Value
	}
	if c.Content.WasSet {
		input.Content = new(argparser.Content(c.Content.Value))
	}
	if c.Location.WasSet {
		location := fastly.SnippetType(c.Location.Value)
		input.Type = &location
	}

	return &input, nil
}
