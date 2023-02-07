package service

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	"github.com/fastly/cli/pkg/cmd"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/cli/pkg/time"
	"github.com/fastly/go-fastly/v7/fastly"
)

// DescribeCommand calls the Fastly API to describe a service.
type DescribeCommand struct {
	cmd.Base
	manifest    manifest.Data
	Input       fastly.GetServiceInput
	json        bool
	serviceName cmd.OptionalServiceNameID
}

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent cmd.Registerer, g *global.Data, m manifest.Data) *DescribeCommand {
	c := DescribeCommand{
		Base: cmd.Base{
			Globals: g,
		},
		manifest: m,
	}
	c.CmdClause = parent.Command("describe", "Show detailed information about a Fastly service").Alias("get")

	// optional
	c.RegisterFlagBool(cmd.BoolFlagOpts{
		Name:        cmd.FlagJSONName,
		Description: cmd.FlagJSONDesc,
		Dst:         &c.json,
		Short:       'j',
	})
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagServiceIDName,
		Description: cmd.FlagServiceIDDesc,
		Dst:         &c.manifest.Flag.ServiceID,
		Short:       's',
	})
	c.RegisterFlag(cmd.StringFlagOpts{
		Action:      c.serviceName.Set,
		Name:        cmd.FlagServiceName,
		Description: cmd.FlagServiceDesc,
		Dst:         &c.serviceName.Value,
	})
	return &c
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.json {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	serviceID, source, flag, err := cmd.ServiceID(c.serviceName, c.manifest, c.Globals.APIClient, c.Globals.ErrLog)
	if err != nil {
		return err
	}
	if c.Globals.Verbose() {
		cmd.DisplayServiceID(serviceID, flag, source, out)
	}

	if source == manifest.SourceUndefined && !c.serviceName.WasSet {
		err := fsterr.ErrNoServiceID
		c.Globals.ErrLog.Add(err)
		return err
	}

	c.Input.ID = serviceID

	service, err := c.Globals.APIClient.GetServiceDetails(&c.Input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID": serviceID,
		})
		return err
	}

	err = c.print(service, out)
	if err != nil {
		return err
	}
	return nil
}

func (c *DescribeCommand) print(s *fastly.ServiceDetail, out io.Writer) error {
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

	activeVersion := "none"
	if s.ActiveVersion.Active {
		activeVersion = strconv.Itoa(s.ActiveVersion.Number)
	}

	fmt.Fprintf(out, "ID: %s\n", s.ID)
	fmt.Fprintf(out, "Name: %s\n", s.Name)
	fmt.Fprintf(out, "Type: %s\n", s.Type)
	if s.Comment != "" {
		fmt.Fprintf(out, "Comment: %s\n", s.Comment)
	}
	fmt.Fprintf(out, "Customer ID: %s\n", s.CustomerID)
	if s.CreatedAt != nil {
		fmt.Fprintf(out, "Created (UTC): %s\n", s.CreatedAt.UTC().Format(time.Format))
	}
	if s.UpdatedAt != nil {
		fmt.Fprintf(out, "Last edited (UTC): %s\n", s.UpdatedAt.UTC().Format(time.Format))
	}
	if s.DeletedAt != nil {
		fmt.Fprintf(out, "Deleted (UTC): %s\n", s.DeletedAt.UTC().Format(time.Format))
	}
	if s.ActiveVersion.Active {
		fmt.Fprintf(out, "Active version:\n")
		text.PrintVersion(out, "\t", &s.ActiveVersion)
	} else {
		fmt.Fprintf(out, "Active version: %s\n", activeVersion)
	}
	fmt.Fprintf(out, "Versions: %d\n", len(s.Versions))
	for j, version := range s.Versions {
		fmt.Fprintf(out, "\tVersion %d/%d\n", j+1, len(s.Versions))
		text.PrintVersion(out, "\t\t", version)
	}
	return nil
}
