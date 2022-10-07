package serviceauth

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/cli/pkg/time"
	"github.com/fastly/go-fastly/v6/fastly"
)

// ListCommand calls the Fastly API to list service authorizations.
type ListCommand struct {
	cmd.Base
	input fastly.ListServiceAuthorizationsInput
	json  bool
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent cmd.Registerer, globals *config.Data) *ListCommand {
	var c ListCommand
	c.Globals = globals
	c.CmdClause = parent.Command("list", "List service authorizations")
	c.RegisterFlagBool(cmd.BoolFlagOpts{
		Name:        cmd.FlagJSONName,
		Description: cmd.FlagJSONDesc,
		Dst:         &c.json,
		Short:       'j',
	})
	c.CmdClause.Flag("page", "Page number of data set to fetch").IntVar(&c.input.PageNumber)
	c.CmdClause.Flag("per-page", "Number of records per page").IntVar(&c.input.PageSize)
	return &c
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.json {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	resp, err := c.Globals.APIClient.ListServiceAuthorizations(&c.input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Page Number": c.input.PageNumber,
			"Page Size":   c.input.PageSize,
		})
		return err
	}

	if !c.Globals.Verbose() {
		if c.json {
			data, err := json.Marshal(resp)
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

		if len(resp.Items) > 0 {
			tw := text.NewTable(out)
			tw.AddHeader("AUTH ID", "USER ID", "SERVICE ID", "PERMISSION")

			for _, s := range resp.Items {
				tw.AddLine(s.ID, s.User.ID, s.Service.ID, s.Permission)
			}
			tw.Print()

			return nil
		}
	}

	for _, s := range resp.Items {
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
		if s.DeltedAt != nil {
			fmt.Fprintf(out, "Deleted (UTC): %s\n", s.DeltedAt.UTC().Format(time.Format))
		}
	}

	return nil
}
