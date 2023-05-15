package snippet

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v8/fastly"
)

// Locations is a list of VCL subroutines.
var Locations = []string{"init", "recv", "hash", "hit", "miss", "pass", "fetch", "error", "deliver", "log", "none"}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent cmd.Registerer, g *global.Data, m manifest.Data) *CreateCommand {
	c := CreateCommand{
		Base: cmd.Base{
			Globals: g,
		},
		manifest: m,
	}
	c.CmdClause = parent.Command("create", "Create a snippet for a particular service and version").Alias("add")

	// Required.
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagVersionName,
		Description: cmd.FlagVersionDesc,
		Dst:         &c.serviceVersion.Value,
		Required:    true,
	})

	// Optional.
	c.RegisterAutoCloneFlag(cmd.AutoCloneFlagOpts{
		Action: c.autoClone.Set,
		Dst:    &c.autoClone.Value,
	})
	c.CmdClause.Flag("content", "VCL snippet passed as file path or content, e.g. $(< snippet.vcl)").Action(c.content.Set).StringVar(&c.content.Value)
	c.CmdClause.Flag("dynamic", "Whether the VCL snippet is dynamic or versioned").Action(c.dynamic.Set).BoolVar(&c.dynamic.Value)
	c.CmdClause.Flag("name", "The name of the VCL snippet").Action(c.name.Set).StringVar(&c.name.Value)
	c.CmdClause.Flag("priority", "Priority determines execution order. Lower numbers execute first").Short('p').Action(c.priority.Set).IntVar(&c.priority.Value)

	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagServiceIDName,
		Description: cmd.FlagServiceIDDesc,
		Dst:         &c.manifest.Flag.ServiceID,
		Short:       's',
	})
	c.RegisterFlag(cmd.StringFlagOpts{
		Action:      c.serviceName.Set,
		Name:        cmd.FlagServiceName,
		Description: cmd.FlagServiceDesc,
		Dst:         &c.serviceName.Value,
	})
	c.CmdClause.Flag("type", "The location in generated VCL where the snippet should be placed").Action(c.location.Set).HintOptions(Locations...).EnumVar(&c.location.Value, Locations...)

	return &c
}

// CreateCommand calls the Fastly API to create an appropriate resource.
type CreateCommand struct {
	cmd.Base

	autoClone      cmd.OptionalAutoClone
	content        cmd.OptionalString
	dynamic        cmd.OptionalBool
	location       cmd.OptionalString
	manifest       manifest.Data
	name           cmd.OptionalString
	priority       cmd.OptionalInt
	serviceName    cmd.OptionalServiceNameID
	serviceVersion cmd.OptionalServiceVersion
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	serviceID, serviceVersion, err := cmd.ServiceDetails(cmd.ServiceDetailsOpts{
		AutoCloneFlag:      c.autoClone,
		APIClient:          c.Globals.APIClient,
		Manifest:           c.manifest,
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

	input := c.constructInput(serviceID, serviceVersion.Number)

	v, err := c.Globals.APIClient.CreateSnippet(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": serviceVersion.Number,
		})
		return err
	}

	text.Success(out, "Created VCL snippet '%s' (service: %s, version: %d, dynamic: %t, snippet id: %s, type: %s, priority: %d)", v.Name, v.ServiceID, v.ServiceVersion, c.dynamic.WasSet, v.ID, c.location.Value, v.Priority)
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *CreateCommand) constructInput(serviceID string, serviceVersion int) *fastly.CreateSnippetInput {
	input := fastly.CreateSnippetInput{
		Dynamic:        fastly.Int(0),
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion,
	}
	if c.name.WasSet {
		input.Name = &c.name.Value
	}
	if c.content.WasSet {
		input.Content = fastly.String(cmd.Content(c.content.Value))
	}
	if c.location.WasSet {
		sType := fastly.SnippetType(c.location.Value)
		input.Type = &sType
	}
	if c.dynamic.WasSet {
		input.Dynamic = fastly.Int(1)
	}
	if c.priority.WasSet {
		input.Priority = &c.priority.Value
	}

	return &input
}
