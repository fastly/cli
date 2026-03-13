package discoveredoperations

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/fastly/go-fastly/v13/fastly"
	"github.com/fastly/go-fastly/v13/fastly/apisecurity/operations"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// ListCommand calls the Fastly API to list discovered API operations.
type ListCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	input       operations.ListDiscoveredInput
	serviceName argparser.OptionalServiceNameID

	// Optional.
	domain  argparser.OptionalString
	method  argparser.OptionalString
	path    argparser.OptionalString
	status  argparser.OptionalString
	page    argparser.OptionalInt
	perPage argparser.OptionalInt
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent argparser.Registerer, g *global.Data) *ListCommand {
	c := ListCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("list", "List discovered operations")

	// Required.
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagServiceIDName,
		Description: argparser.FlagServiceIDDesc,
		Dst:         &g.Manifest.Flag.ServiceID,
		Short:       's',
		Required:    true,
	})

	// Optional.
	c.CmdClause.Flag("status", "Filters operations by status. Valid values are: discovered, saved, ignored").Action(c.status.Set).StringVar(&c.status.Value)
	c.CmdClause.Flag("domain", "The domain for the operation").Action(c.domain.Set).StringVar(&c.domain.Value)
	c.CmdClause.Flag("method", "Filters operations by HTTP method (e.g., GET, POST, PUT)").Action(c.method.Set).StringVar(&c.method.Value)
	c.CmdClause.Flag("path", "Filters operations by path (exact match)").Action(c.path.Set).StringVar(&c.path.Value)
	c.CmdClause.Flag("page", "Page number for pagination (0-indexed)").Action(c.page.Set).IntVar(&c.page.Value)
	c.CmdClause.Flag("per-page", "Number of items per page (default: 100)").Action(c.perPage.Set).IntVar(&c.perPage.Value)
	c.RegisterFlagBool(c.JSONFlag()) // --json

	return &c
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	serviceID, source, flag, err := argparser.ServiceID(c.serviceName, *c.Globals.Manifest, c.Globals.APIClient, c.Globals.ErrLog)
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
	if c.status.WasSet {
		switch c.status.Value {
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
			err := fmt.Errorf("invalid status: %s. Valid options: 'discovered', 'saved', 'ignored'", c.status.Value)
			c.Globals.ErrLog.Add(err)
			return err
		}
	}
	if c.domain.WasSet {
		c.input.Domain = []string{c.domain.Value}
	}
	if c.method.WasSet {
		c.input.Method = []string{c.method.Value}
	}
	if c.path.WasSet {
		c.input.Path = &c.path.Value
	}
	if c.page.WasSet {
		c.input.Page = &c.page.Value
	}
	if c.perPage.WasSet {
		c.input.Limit = &c.perPage.Value
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	o, err := operations.ListDiscovered(context.TODO(), fc, &c.input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID": serviceID,
			"Domain":     c.domain.Value,
			"Method":     c.method.Value,
			"Status":     c.status.Value,
			"Path":       c.path.Value,
			"Page":       c.page.Value,
			"Per Page":   c.perPage.Value,
		})
		return err
	}

	if o == nil {
		o = &operations.DiscoveredOperations{
			Data: []operations.DiscoveredOperation{},
		}
	}

	if ok, err := c.WriteJSON(out, o.Data); ok {
		return err
	}

	if !c.Globals.Verbose() {
		return c.printSummary(out, o.Data)
	}

	return c.printVerbose(out, o.Data)
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
