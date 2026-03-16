package operations

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

// ListCommand calls the Fastly API to list operations.
type ListCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	input       operations.ListOperationsInput
	serviceName argparser.OptionalServiceNameID

	// Optional.
	domain  argparser.OptionalString
	method  argparser.OptionalString
	path    argparser.OptionalString
	tagID   argparser.OptionalString
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
	c.CmdClause = parent.Command("list", "List operations")

	// Required.
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagServiceIDName,
		Description: argparser.FlagServiceIDDesc,
		Dst:         &g.Manifest.Flag.ServiceID,
		Short:       's',
		Required:    true,
	})
	c.RegisterFlag(argparser.StringFlagOpts{
		Action:      c.serviceName.Set,
		Name:        argparser.FlagServiceName,
		Description: argparser.FlagServiceNameDesc,
		Dst:         &c.serviceName.Value,
	})

	// Optional.
	c.CmdClause.Flag("domain", "Filters operations by domain (exact match)").Action(c.domain.Set).StringVar(&c.domain.Value)
	c.CmdClause.Flag("method", "Filters operations by HTTP method (e.g., GET, POST, PUT)").Action(c.method.Set).StringVar(&c.method.Value)
	c.CmdClause.Flag("path", "Filters operations by path (exact match)").Action(c.path.Set).StringVar(&c.path.Value)
	c.CmdClause.Flag("tag-id", "Filters operations by tag ID").Action(c.tagID.Set).StringVar(&c.tagID.Value)
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

	if c.domain.WasSet {
		c.input.Domain = []string{c.domain.Value}
	}
	if c.method.WasSet {
		c.input.Method = []string{c.method.Value}
	}
	if c.path.WasSet {
		c.input.Path = &c.path.Value
	}
	if c.tagID.WasSet {
		c.input.TagID = &c.tagID.Value
	}

	// Set pagination parameters
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

	o, err := operations.ListOperations(context.TODO(), fc, &c.input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID": serviceID,
			"Domain":     c.domain.Value,
			"Method":     c.method.Value,
			"Path":       c.path.Value,
			"Tag ID":     c.tagID.Value,
			"Page":       c.page.Value,
			"Per Page":   c.perPage.Value,
		})
		return err
	}

	if o == nil {
		o = &operations.Operations{
			Data: []operations.Operation{},
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

// printSummary displays the operations in a table format.
func (c *ListCommand) printSummary(out io.Writer, o []operations.Operation) error {
	tw := text.NewTable(out)
	tw.AddHeader("ID", "METHOD", "DOMAIN", "PATH", "DESCRIPTION", "TAGS")
	for _, op := range o {
		description := op.Description
		if len(description) > 50 {
			description = description[:47] + "..."
		}
		tags := fmt.Sprintf("%d", len(op.TagIDs))
		tw.AddLine(
			op.ID,
			strings.ToUpper(op.Method),
			op.Domain,
			op.Path,
			description,
			tags,
		)
	}
	tw.Print()

	return nil
}

// printVerbose displays detailed information for each operation.
func (c *ListCommand) printVerbose(out io.Writer, o []operations.Operation) error {
	for i, op := range o {
		fmt.Fprintf(out, "\nOperation %d/%d\n", i+1, len(o))
		fmt.Fprintf(out, "\tID: %s\n", op.ID)
		fmt.Fprintf(out, "\tMethod: %s\n", strings.ToUpper(op.Method))
		fmt.Fprintf(out, "\tDomain: %s\n", op.Domain)
		fmt.Fprintf(out, "\tPath: %s\n", op.Path)
		if op.Description != "" {
			fmt.Fprintf(out, "\tDescription: %s\n", op.Description)
		}
		if op.Status != "" {
			fmt.Fprintf(out, "\tStatus: %s\n", op.Status)
		}
		if len(op.TagIDs) > 0 {
			fmt.Fprintf(out, "\tTag IDs: %s\n", strings.Join(op.TagIDs, ", "))
		}
		if op.RPS > 0 {
			fmt.Fprintf(out, "\tRPS: %.2f\n", op.RPS)
		}
		if op.CreatedAt != "" {
			fmt.Fprintf(out, "\tCreated At: %s\n", op.CreatedAt)
		}
		if op.UpdatedAt != "" {
			fmt.Fprintf(out, "\tUpdated At: %s\n", op.UpdatedAt)
		}
		if op.LastSeenAt != "" {
			fmt.Fprintf(out, "\tLast Seen: %s\n", op.LastSeenAt)
		}
	}
	fmt.Fprintln(out)

	return nil
}
