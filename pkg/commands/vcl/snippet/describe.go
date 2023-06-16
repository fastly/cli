package snippet

import (
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/cmd"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
)

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent cmd.Registerer, g *global.Data, m manifest.Data) *DescribeCommand {
	c := DescribeCommand{
		Base: cmd.Base{
			Globals: g,
		},
		manifest: m,
	}
	c.CmdClause = parent.Command("describe", "Get the uploaded VCL snippet for a particular service and version").Alias("get")

	// Required.
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagVersionName,
		Description: cmd.FlagVersionDesc,
		Dst:         &c.serviceVersion.Value,
		Required:    true,
	})

	// Optional.
	c.CmdClause.Flag("dynamic", "Whether the VCL snippet is dynamic or versioned").Action(c.dynamic.Set).BoolVar(&c.dynamic.Value)
	c.RegisterFlagBool(c.JSONFlag()) // --json
	c.CmdClause.Flag("name", "The name of the VCL snippet").StringVar(&c.name)
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
	c.CmdClause.Flag("snippet-id", "Alphanumeric string identifying a VCL Snippet").StringVar(&c.snippetID)

	return &c
}

// DescribeCommand calls the Fastly API to describe an appropriate resource.
type DescribeCommand struct {
	cmd.Base
	cmd.JSONOutput

	dynamic        cmd.OptionalBool
	manifest       manifest.Data
	name           string
	serviceName    cmd.OptionalServiceNameID
	serviceVersion cmd.OptionalServiceVersion
	snippetID      string
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	serviceID, serviceVersion, err := cmd.ServiceDetails(cmd.ServiceDetailsOpts{
		AllowActiveLocked:  true,
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
			"Service Version": fsterr.ServiceVersion(serviceVersion),
		})
		return err
	}

	if c.dynamic.WasSet {
		input, err := c.constructDynamicInput(serviceID, serviceVersion.Number)
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
				"Service ID":      serviceID,
				"Service Version": serviceVersion.Number,
			})
			return err
		}

		o, err := c.Globals.APIClient.GetDynamicSnippet(input)
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
				"Service ID":      serviceID,
				"Service Version": serviceVersion.Number,
			})
			return err
		}

		if ok, err := c.WriteJSON(out, o); ok {
			return err
		}

		return c.printDynamic(out, o)
	}

	input, err := c.constructInput(serviceID, serviceVersion.Number)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": serviceVersion.Number,
		})
		return err
	}

	o, err := c.Globals.APIClient.GetSnippet(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": serviceVersion.Number,
		})
		return err
	}

	if ok, err := c.WriteJSON(out, o); ok {
		return err
	}

	return c.print(out, o)
}

// constructDynamicInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *DescribeCommand) constructDynamicInput(serviceID string, _ int) (*fastly.GetDynamicSnippetInput, error) {
	var input fastly.GetDynamicSnippetInput

	input.ID = c.snippetID
	input.ServiceID = serviceID

	if c.snippetID == "" {
		return nil, fmt.Errorf("error parsing arguments: must provide --snippet-id with a dynamic VCL snippet")
	}

	return &input, nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *DescribeCommand) constructInput(serviceID string, serviceVersion int) (*fastly.GetSnippetInput, error) {
	var input fastly.GetSnippetInput

	input.Name = c.name
	input.ServiceID = serviceID
	input.ServiceVersion = serviceVersion

	if c.name == "" {
		return nil, fmt.Errorf("error parsing arguments: must provide --name with a versioned VCL snippet")
	}

	return &input, nil
}

// print displays the 'dynamic' information returned from the API.
func (c *DescribeCommand) printDynamic(out io.Writer, ds *fastly.DynamicSnippet) error {
	fmt.Fprintf(out, "\nService ID: %s\n", ds.ServiceID)
	fmt.Fprintf(out, "ID: %s\n", ds.ID)
	fmt.Fprintf(out, "Content: \n%s\n", ds.Content)
	if ds.CreatedAt != nil {
		fmt.Fprintf(out, "Created at: %s\n", ds.CreatedAt)
	}
	if ds.UpdatedAt != nil {
		fmt.Fprintf(out, "Updated at: %s\n", ds.UpdatedAt)
	}
	return nil
}

// print displays the information returned from the API.
func (c *DescribeCommand) print(out io.Writer, s *fastly.Snippet) error {
	if !c.Globals.Verbose() {
		fmt.Fprintf(out, "\nService ID: %s\n", s.ServiceID)
	}
	fmt.Fprintf(out, "Service Version: %d\n", s.ServiceVersion)

	fmt.Fprintf(out, "\nName: %s\n", s.Name)
	fmt.Fprintf(out, "ID: %s\n", s.ID)
	fmt.Fprintf(out, "Priority: %d\n", s.Priority)
	fmt.Fprintf(out, "Dynamic: %t\n", cmd.IntToBool(s.Dynamic))
	fmt.Fprintf(out, "Type: %s\n", s.Type)
	fmt.Fprintf(out, "Content: \n%s\n", s.Content)
	if s.CreatedAt != nil {
		fmt.Fprintf(out, "Created at: %s\n", s.CreatedAt)
	}
	if s.UpdatedAt != nil {
		fmt.Fprintf(out, "Updated at: %s\n", s.UpdatedAt)
	}
	if s.DeletedAt != nil {
		fmt.Fprintf(out, "Deleted at: %s\n", s.DeletedAt)
	}
	return nil
}
