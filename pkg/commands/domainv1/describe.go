package domainv1

import (
	"errors"
	"io"

	"github.com/fastly/go-fastly/v10/fastly"
	v1 "github.com/fastly/go-fastly/v10/fastly/domains/v1"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
)

// DescribeCommand calls the Fastly API to describe a domain.
type DescribeCommand struct {
	argparser.Base
	argparser.JSONOutput
	domainID string
}

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent argparser.Registerer, g *global.Data) *DescribeCommand {
	c := DescribeCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("describe", "Show detailed information about a domain").Alias("get")

	// Required.
	c.CmdClause.Flag("domain-id", "The Domain Identifier (UUID)").Required().StringVar(&c.domainID)

	// Optional.
	c.RegisterFlagBool(c.JSONFlag()) // --json
	return &c
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	input := &v1.GetInput{
		DomainID: &c.domainID,
	}

	d, err := v1.Get(fc, input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Domain ID": c.domainID,
		})
		return err
	}

	if ok, err := c.WriteJSON(out, d); ok {
		return err
	}

	if d != nil {
		cl := []v1.Data{*d}
		if c.Globals.Verbose() {
			printVerbose(out, cl)
		} else {
			printSummary(out, cl)
		}
	}
	return nil
}
