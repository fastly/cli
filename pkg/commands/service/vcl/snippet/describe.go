package snippet

import (
	"context"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *DescribeCommand) AfterApply() error {
	c.serviceVersion.Value = c.Version
	if c.ServiceID != nil {
		c.Globals.Manifest.Flag.ServiceID = *c.ServiceID
	}
	return nil
}

// DescribeCommand calls the Fastly API to describe an appropriate resource.
type DescribeCommand struct {
	argparser.Base

	// Required.
	Version string `required:"" help:"'latest', 'active', 'staged', or the number of a specific Fastly service version"`

	// Optional.
	Content argparser.OptionalBool `help:"Outputs the raw content of the identified snippet"`
	Dynamic argparser.OptionalBool `help:"Whether the VCL snippet is dynamic or versioned"`
	argparser.JSONOutput

	NameFlag       string                          `name:"name" help:"The name of the VCL snippet"`
	ServiceID      *string                         `short:"s" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`
	ServiceName    argparser.OptionalServiceNameID `help:"The name of the service"`
	serviceVersion argparser.OptionalServiceVersion
	SnippetID      string `help:"Alphanumeric string identifying a VCL Snippet"`
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}
	// Ensure that the --content flag is not used
	// with --verbose or --json
	if c.Globals.Verbose() && c.Content.WasSet {
		return fsterr.ErrInvalidContentOutputCombo
	}
	if c.JSONOutput.Enabled && c.Content.WasSet {
		return fsterr.ErrInvalidContentOutputCombo
	}

	serviceID, serviceVersion, err := argparser.ServiceDetails(argparser.ServiceDetailsOpts{
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

		o, err := c.Globals.APIClient.GetDynamicSnippet(context.TODO(), input)
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
				"Service ID":      serviceID,
				"Service Version": serviceVersionNumber,
			})
			return err
		}

		if ok, err := c.WriteJSON(out, o); ok {
			return err
		}

		return c.printDynamic(out, o)
	}

	input, err := c.constructInput(serviceID, serviceVersionNumber)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": serviceVersionNumber,
		})
		return err
	}

	o, err := c.Globals.APIClient.GetSnippet(context.TODO(), input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": serviceVersionNumber,
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

	input.SnippetID = c.SnippetID
	input.ServiceID = serviceID

	if c.SnippetID == "" {
		return nil, fmt.Errorf("error parsing arguments: must provide --snippet-id with a dynamic VCL snippet")
	}

	return &input, nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *DescribeCommand) constructInput(serviceID string, serviceVersion int) (*fastly.GetSnippetInput, error) {
	var input fastly.GetSnippetInput

	input.Name = c.NameFlag
	input.ServiceID = serviceID
	input.ServiceVersion = serviceVersion

	if c.NameFlag == "" {
		return nil, fmt.Errorf("error parsing arguments: must provide --name with a versioned VCL snippet")
	}

	return &input, nil
}

// print displays the 'dynamic' information returned from the API.
func (c *DescribeCommand) printDynamic(out io.Writer, ds *fastly.DynamicSnippet) error {
	// If the --content flag is set, output only the raw VCL content.
	if c.Content.WasSet {
		fmt.Fprint(out, fastly.ToValue(ds.Content))
		return nil
	}

	fmt.Fprintf(out, "\nService ID: %s\n", fastly.ToValue(ds.ServiceID))
	fmt.Fprintf(out, "ID: %s\n", fastly.ToValue(ds.SnippetID))
	fmt.Fprintf(out, "Content: \n%s\n", text.SanitizeTerminalOutput(fastly.ToValue(ds.Content)))
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
	// If the --content flag is set, output only the raw VCL content.
	if c.Content.WasSet {
		fmt.Fprint(out, fastly.ToValue(s.Content))
		return nil
	}

	if !c.Globals.Verbose() {
		fmt.Fprintf(out, "\nService ID: %s\n", fastly.ToValue(s.ServiceID))
	}
	fmt.Fprintf(out, "Service Version: %d\n", fastly.ToValue(s.ServiceVersion))
	fmt.Fprintf(out, "\nName: %s\n", fastly.ToValue(s.Name))
	fmt.Fprintf(out, "ID: %s\n", fastly.ToValue(s.SnippetID))
	fmt.Fprintf(out, "Priority: %s\n", fastly.ToValue(s.Priority))
	fmt.Fprintf(out, "Dynamic: %t\n", argparser.IntToBool(fastly.ToValue(s.Dynamic)))
	fmt.Fprintf(out, "Type: %s\n", fastly.ToValue(s.Type))
	fmt.Fprintf(out, "Content: \n%s\n", text.SanitizeTerminalOutput(fastly.ToValue(s.Content)))
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
