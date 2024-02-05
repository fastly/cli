package snippet

import (
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent argparser.Registerer, g *global.Data) *UpdateCommand {
	c := UpdateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("update", "Update a VCL snippet for a particular service and version")

	// Required.
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagVersionName,
		Description: argparser.FlagVersionDesc,
		Dst:         &c.serviceVersion.Value,
		Required:    true,
	})

	// Optional.
	c.RegisterAutoCloneFlag(argparser.AutoCloneFlagOpts{
		Action: c.autoClone.Set,
		Dst:    &c.autoClone.Value,
	})
	c.CmdClause.Flag("content", "VCL snippet passed as file path or content, e.g. $(< snippet.vcl)").Action(c.content.Set).StringVar(&c.content.Value)
	c.CmdClause.Flag("dynamic", "Whether the VCL snippet is dynamic or versioned").Action(c.dynamic.Set).BoolVar(&c.dynamic.Value)
	c.CmdClause.Flag("name", "The name of the VCL snippet to update").StringVar(&c.name)
	c.CmdClause.Flag("new-name", "New name for the VCL snippet").Action(c.newName.Set).StringVar(&c.newName.Value)
	c.CmdClause.Flag("priority", "Priority determines execution order. Lower numbers execute first").Short('p').Action(c.priority.Set).IntVar(&c.priority.Value)
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagServiceIDName,
		Description: argparser.FlagServiceIDDesc,
		Dst:         &g.Manifest.Flag.ServiceID,
		Short:       's',
	})
	c.RegisterFlag(argparser.StringFlagOpts{
		Action:      c.serviceName.Set,
		Name:        argparser.FlagServiceName,
		Description: argparser.FlagServiceDesc,
		Dst:         &c.serviceName.Value,
	})
	c.CmdClause.Flag("snippet-id", "Alphanumeric string identifying a VCL Snippet").StringVar(&c.snippetID)

	// NOTE: Locations is defined in the same snippet package inside create.go
	c.CmdClause.Flag("type", "The location in generated VCL where the snippet should be placed").HintOptions(Locations...).Action(c.location.Set).EnumVar(&c.location.Value, Locations...)

	return &c
}

// UpdateCommand calls the Fastly API to update an appropriate resource.
type UpdateCommand struct {
	argparser.Base

	autoClone      argparser.OptionalAutoClone
	content        argparser.OptionalString
	dynamic        argparser.OptionalBool
	location       argparser.OptionalString
	name           string
	newName        argparser.OptionalString
	priority       argparser.OptionalInt
	serviceName    argparser.OptionalServiceNameID
	serviceVersion argparser.OptionalServiceVersion
	snippetID      string
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	serviceID, serviceVersion, err := argparser.ServiceDetails(argparser.ServiceDetailsOpts{
		AllowActiveLocked:  c.dynamic.WasSet && c.dynamic.Value,
		AutoCloneFlag:      c.autoClone,
		APIClient:          c.Globals.APIClient,
		Manifest:           *c.Globals.Manifest,
		Out:                out,
		ServiceNameFlag:    c.serviceName,
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

	if c.dynamic.WasSet {
		input, err := c.constructDynamicInput(serviceID, serviceVersionNumber)
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
				"Service ID":      serviceID,
				"Service Version": serviceVersionNumber,
			})
			return err
		}
		v, err := c.Globals.APIClient.UpdateDynamicSnippet(input)
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
	v, err := c.Globals.APIClient.UpdateSnippet(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": serviceVersionNumber,
		})
		return err
	}
	text.Success(out,
		"Updated VCL snippet '%s' (previously: '%s', service: %s, version: %d, type: %v, priority: %d)",
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

	input.SnippetID = c.snippetID
	input.ServiceID = serviceID

	if c.newName.WasSet {
		return nil, fmt.Errorf("error parsing arguments: --new-name is not supported when updating a dynamic VCL snippet")
	}
	if c.snippetID == "" {
		return nil, fmt.Errorf("error parsing arguments: must provide --snippet-id to update a dynamic VCL snippet")
	}
	if c.content.WasSet {
		input.Content = fastly.ToPointer(argparser.Content(c.content.Value))
	}

	return &input, nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) constructInput(serviceID string, serviceVersion int) (*fastly.UpdateSnippetInput, error) {
	var input fastly.UpdateSnippetInput

	input.Name = c.name
	input.ServiceID = serviceID
	input.ServiceVersion = serviceVersion

	if c.snippetID != "" {
		return nil, fmt.Errorf("error parsing arguments: --snippet-id is not supported when updating a versioned VCL snippet")
	}
	if c.name == "" {
		return nil, fmt.Errorf("error parsing arguments: must provide --name to update a versioned VCL snippet")
	}
	if c.newName.WasSet {
		input.NewName = &c.newName.Value
	}
	if c.priority.WasSet {
		input.Priority = &c.priority.Value
	}
	if c.content.WasSet {
		input.Content = fastly.ToPointer(argparser.Content(c.content.Value))
	}
	if c.location.WasSet {
		location := fastly.SnippetType(c.location.Value)
		input.Type = &location
	}

	return &input, nil
}
