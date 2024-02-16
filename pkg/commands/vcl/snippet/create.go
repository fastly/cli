package snippet

import (
	"io"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/v10/pkg/argparser"
	"github.com/fastly/cli/v10/pkg/errors"
	"github.com/fastly/cli/v10/pkg/global"
	"github.com/fastly/cli/v10/pkg/text"
)

// Locations is a list of VCL subroutines.
var Locations = []string{"init", "recv", "hash", "hit", "miss", "pass", "fetch", "error", "deliver", "log", "none"}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent argparser.Registerer, g *global.Data) *CreateCommand {
	c := CreateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("create", "Create a snippet for a particular service and version").Alias("add")

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
	c.CmdClause.Flag("name", "The name of the VCL snippet").Action(c.name.Set).StringVar(&c.name.Value)
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
	c.CmdClause.Flag("type", "The location in generated VCL where the snippet should be placed").Action(c.location.Set).HintOptions(Locations...).EnumVar(&c.location.Value, Locations...)

	return &c
}

// CreateCommand calls the Fastly API to create an appropriate resource.
type CreateCommand struct {
	argparser.Base

	autoClone      argparser.OptionalAutoClone
	content        argparser.OptionalString
	dynamic        argparser.OptionalBool
	location       argparser.OptionalString
	name           argparser.OptionalString
	priority       argparser.OptionalInt
	serviceName    argparser.OptionalServiceNameID
	serviceVersion argparser.OptionalServiceVersion
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	serviceID, serviceVersion, err := argparser.ServiceDetails(argparser.ServiceDetailsOpts{
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
			"Service Version": errors.ServiceVersion(serviceVersion),
		})
		return err
	}

	input := c.constructInput(serviceID, fastly.ToValue(serviceVersion.Number))

	v, err := c.Globals.APIClient.CreateSnippet(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": fastly.ToValue(serviceVersion.Number),
		})
		return err
	}

	text.Success(out,
		"Created VCL snippet '%s' (service: %s, version: %d, dynamic: %t, snippet id: %s, type: %s, priority: %d)",
		fastly.ToValue(v.Name),
		fastly.ToValue(v.ServiceID),
		fastly.ToValue(v.ServiceVersion),
		c.dynamic.WasSet,
		fastly.ToValue(v.SnippetID),
		c.location.Value,
		fastly.ToValue(v.Priority),
	)
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *CreateCommand) constructInput(serviceID string, serviceVersion int) *fastly.CreateSnippetInput {
	input := fastly.CreateSnippetInput{
		Dynamic:        fastly.ToPointer(0),
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion,
	}
	if c.name.WasSet {
		input.Name = &c.name.Value
	}
	if c.content.WasSet {
		input.Content = fastly.ToPointer(argparser.Content(c.content.Value))
	}
	if c.location.WasSet {
		sType := fastly.SnippetType(c.location.Value)
		input.Type = &sType
	}
	if c.dynamic.WasSet {
		input.Dynamic = fastly.ToPointer(1)
	}
	if c.priority.WasSet {
		input.Priority = &c.priority.Value
	}

	return &input
}
