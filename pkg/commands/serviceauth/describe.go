package serviceauth

import (
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v10/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/time"
)

// DescribeCommand calls the Fastly API to describe a service authorization.
type DescribeCommand struct {
	argparser.Base
	argparser.JSONOutput

	Input fastly.GetServiceAuthorizationInput
}

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent argparser.Registerer, g *global.Data) *DescribeCommand {
	c := DescribeCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("describe", "Show service authorization").Alias("get")

	// Required.
	c.CmdClause.Flag("id", "ID of the service authorization to retrieve").Required().StringVar(&c.Input.ID)

	// Optional.
	c.RegisterFlagBool(c.JSONFlag()) // --json
	return &c
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	o, err := c.Globals.APIClient.GetServiceAuthorization(&c.Input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service Authorization ID": c.Input.ID,
		})
		return err
	}

	if ok, err := c.WriteJSON(out, o); ok {
		return err
	}

	return c.print(o, out)
}

func (c *DescribeCommand) print(s *fastly.ServiceAuthorization, out io.Writer) error {
	fmt.Fprintf(out, "Auth ID: %s\n", s.ID)
	fmt.Fprintf(out, "User ID: %s\n", s.User.ID)
	fmt.Fprintf(out, "Service ID: %s\n", s.Service.ID)
	fmt.Fprintf(out, "Permission: %s\n", s.Permission)

	if s.CreatedAt != nil {
		fmt.Fprintf(out, "Created (UTC): %s\n", s.CreatedAt.UTC().Format(time.Format))
	}
	if s.UpdatedAt != nil {
		fmt.Fprintf(out, "Last edited (UTC): %s\n", s.UpdatedAt.UTC().Format(time.Format))
	}
	if s.DeletedAt != nil {
		fmt.Fprintf(out, "Deleted (UTC): %s\n", s.DeletedAt.UTC().Format(time.Format))
	}

	return nil
}
