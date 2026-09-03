package eventmapping

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v17/fastly"
	"github.com/fastly/go-fastly/v17/fastly/notifications/v1/eventmappings"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// DescribeCommand calls the Fastly API to describe an audit log event mapping.
type DescribeCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	id string
}

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent argparser.Registerer, g *global.Data) *DescribeCommand {
	c := DescribeCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("describe", "Retrieve a single audit log event mapping").Alias("get")

	// Required.
	c.CmdClause.Flag("id", "The unique identifier of the event mapping").Required().StringVar(&c.id)

	// Optional.
	c.RegisterFlagBool(c.JSONFlag())

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

	em, err := eventmappings.Get(context.TODO(), fc, &eventmappings.GetInput{
		MappingID: &c.id,
	})
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Mapping ID": c.id,
		})
		return err
	}

	if ok, err := c.WriteJSON(out, em); ok {
		return err
	}

	text.PrintEventMapping(out, em)
	return nil
}
