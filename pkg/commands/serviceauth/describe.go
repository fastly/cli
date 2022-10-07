package serviceauth

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/time"
	"github.com/fastly/go-fastly/v6/fastly"
)

// DescribeCommand calls the Fastly API to describe a service authorization.
type DescribeCommand struct {
	cmd.Base
	manifest manifest.Data
	Input    fastly.GetServiceAuthorizationInput
	json     bool
}

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *DescribeCommand {
	var c DescribeCommand
	c.Globals = globals
	c.manifest = data
	c.CmdClause = parent.Command("describe", "Show service authorization").Alias("get")
	c.CmdClause.Flag("id", "ID of the service authorization to retrieve").Required().StringVar(&c.Input.ID)
	c.RegisterFlagBool(cmd.BoolFlagOpts{
		Name:        cmd.FlagJSONName,
		Description: cmd.FlagJSONDesc,
		Dst:         &c.json,
		Short:       'j',
	})
	return &c
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.json {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	service, err := c.Globals.APIClient.GetServiceAuthorization(&c.Input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service Authorization ID": c.Input.ID,
		})
		return err
	}

	err = c.print(service, out)
	if err != nil {
		return err
	}
	return nil
}

func (c *DescribeCommand) print(s *fastly.ServiceAuthorization, out io.Writer) error {
	if c.json {
		data, err := json.Marshal(s)
		if err != nil {
			return err
		}
		_, err = out.Write(data)
		if err != nil {
			c.Globals.ErrLog.Add(err)
			return fmt.Errorf("error: unable to write data to stdout: %w", err)
		}
		return nil
	}

	fmt.Fprintf(out, "ID: %s\n", s.ID)
	fmt.Fprintf(out, "User ID: %s\n", s.User.ID)
	fmt.Fprintf(out, "Service ID: %s\n", s.Service.ID)
	fmt.Fprintf(out, "Permission: %s\n", s.Permission)

	if s.CreatedAt != nil {
		fmt.Fprintf(out, "Created (UTC): %s\n", s.CreatedAt.UTC().Format(time.Format))
	}
	if s.UpdatedAt != nil {
		fmt.Fprintf(out, "Last edited (UTC): %s\n", s.UpdatedAt.UTC().Format(time.Format))
	}
	if s.DeltedAt != nil {
		fmt.Fprintf(out, "Deleted (UTC): %s\n", s.DeltedAt.UTC().Format(time.Format))
	}

	return nil
}
