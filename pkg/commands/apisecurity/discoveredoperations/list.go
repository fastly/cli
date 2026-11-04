package discoveredoperations

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/fastly/go-fastly/v19/fastly"
	"github.com/fastly/go-fastly/v19/fastly/apisecurity/operations"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// ListCommand calls the Fastly API to list discovered API operations.
type ListCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	input       operations.ListDiscoveredInput
	ServiceName argparser.OptionalServiceNameID `help:"The name of the service"`

	// Optional.
	Domain argparser.OptionalString `help:"The domain for the operation"`
	Method argparser.OptionalString `help:"Filters operations by HTTP method (e.g., GET, POST, PUT)"`
	Path   argparser.OptionalString `help:"Filters operations by path (exact match)"`
	Status argparser.OptionalString `help:"Filters operations by status. Valid values are: discovered, saved, ignored"`

	// Required.
	ServiceID *string `short:"s" required:"" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`
}

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *ListCommand) AfterApply() error {
	if c.ServiceID != nil {
		c.Globals.Manifest.Flag.ServiceID = *c.ServiceID
	}
	return nil
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(_ io.Reader, out io.Writer) error {
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

	c.input.ServiceID = &serviceID

	// The API only accepts uppercase values for 'status',
	// so we are handling accordingly here and allowing
	// end users to still use the normal lowercase pattern
	// for input in the CLI.
	if c.Status.WasSet {
		switch c.Status.Value {
		case "discovered":
			status := "DISCOVERED"
			c.input.Status = &status
		case "saved":
			status := "SAVED"
			c.input.Status = &status
		case "ignored":
			status := "IGNORED"
			c.input.Status = &status
		default:
			err := fmt.Errorf("invalid status: %s. Valid options: 'discovered', 'saved', 'ignored'", c.Status.Value)
			c.Globals.ErrLog.Add(err)
			return err
		}
	}
	if c.Domain.WasSet {
		c.input.Domain = []string{c.Domain.Value}
	}
	if c.Method.WasSet {
		c.input.Method = []string{c.Method.Value}
	}
	if c.Path.WasSet {
		c.input.Path = &c.Path.Value
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	// Auto-paginate through all results
	var allOperations []operations.DiscoveredOperation
	page := 0
	limit := 100

	for {
		c.input.Page = &page
		c.input.Limit = &limit

		o, err := operations.ListDiscovered(context.TODO(), fc, &c.input)
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
				"Service ID": serviceID,
				"Domain":     c.Domain.Value,
				"Method":     c.Method.Value,
				"Status":     c.Status.Value,
				"Path":       c.Path.Value,
				"Page":       page,
			})
			return err
		}

		if o == nil || len(o.Data) == 0 {
			break
		}

		allOperations = append(allOperations, o.Data...)

		// Check if we've fetched all results
		if len(allOperations) >= o.Meta.Total {
			break
		}

		page++
	}

	if ok, err := c.WriteJSON(out, allOperations); ok {
		return err
	}

	if !c.Globals.Verbose() {
		return c.printSummary(out, allOperations)
	}

	return c.printVerbose(out, allOperations)
}

// printSummary displays the discovered operations in a table format.
func (c *ListCommand) printSummary(out io.Writer, o []operations.DiscoveredOperation) error {
	tw := text.NewTable(out)
	tw.AddHeader("METHOD", "DOMAIN", "PATH", "STATUS", "RPS", "LAST SEEN")
	for _, op := range o {
		tw.AddLine(
			strings.ToUpper(op.Method),
			op.Domain,
			op.Path,
			op.Status,
			fmt.Sprintf("%.2f", op.RPS),
			op.LastSeenAt,
		)
	}
	tw.Print()

	return nil
}

// printVerbose displays detailed information for each discovered operation.
func (c *ListCommand) printVerbose(out io.Writer, o []operations.DiscoveredOperation) error {
	for i, op := range o {
		fmt.Fprintf(out, "\nDiscovered Operation %d/%d\n", i+1, len(o))
		fmt.Fprintf(out, "\tID: %s\n", op.ID)
		fmt.Fprintf(out, "\tMethod: %s\n", strings.ToUpper(op.Method))
		fmt.Fprintf(out, "\tDomain: %s\n", op.Domain)
		fmt.Fprintf(out, "\tPath: %s\n", op.Path)
		fmt.Fprintf(out, "\tStatus: %s\n", op.Status)
		fmt.Fprintf(out, "\tRPS: %.2f\n", op.RPS)
		if op.LastSeenAt != "" {
			fmt.Fprintf(out, "\tLast Seen: %s\n", op.LastSeenAt)
		}
		if op.UpdatedAt != "" {
			fmt.Fprintf(out, "\tUpdated At: %s\n", op.UpdatedAt)
		}
	}
	fmt.Fprintln(out)

	return nil
}

// FlagOrder lists the flags in the order kingpin showed them in help, which
// the fields do not follow.
func (c *ListCommand) FlagOrder() []string {
	return []string{
		"service-id", "service-name", "status", "domain", "method", "path",
		"json",
	}
}
