package version

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v17/fastly"
	"github.com/fastly/go-fastly/v17/fastly/domainmanagement/v1/routingconfigs/versions"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// ListCommand calls the Fastly API to list a routing config's inactive
// versions.
type ListCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	routingConfigID string

	// Optional.
	limit argparser.OptionalInt
	sort  argparser.OptionalString
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent argparser.Registerer, g *global.Data) *ListCommand {
	c := ListCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("list", "List a routing config's inactive versions")

	// Required.
	c.CmdClause.Flag("routing-config-id", "The Routing Config Identifier").Required().StringVar(&c.routingConfigID)

	// Optional.
	c.RegisterFlagBool(c.JSONFlag()) // --json
	c.CmdClause.Flag("limit", "Limit how many results are returned per page").Action(c.limit.Set).IntVar(&c.limit.Value)
	c.CmdClause.Flag("sort", "The order in which to list the results").Action(c.sort.Set).StringVar(&c.sort.Value)
	return &c
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input := &versions.ListInput{
		RoutingConfigID: &c.routingConfigID,
	}
	if c.limit.WasSet {
		input.Limit = &c.limit.Value
	}
	if c.sort.WasSet {
		input.Sort = &c.sort.Value
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	cl, err := versions.List(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Routing Config ID": c.routingConfigID,
		})
		return err
	}

	if ok, err := c.WriteJSON(out, cl); ok {
		return err
	}

	t := text.NewTable(out)
	t.AddHeader("VERSION ID", "COMMENT", "CREATED AT", "ACTIVATED AT")
	for _, v := range cl {
		var activatedAt string
		if v.ActivatedAt != nil {
			activatedAt = v.ActivatedAt.String()
		}
		t.AddLine(v.VersionID, v.Comment, v.CreatedAt, activatedAt)
	}
	t.Print()
	return nil
}
