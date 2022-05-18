package service

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v6/fastly"
)

// UpdateCommand calls the Fastly API to create services.
type UpdateCommand struct {
	cmd.Base

	comment     cmd.OptionalString
	input       fastly.UpdateServiceInput
	manifest    manifest.Data
	name        cmd.OptionalString
	serviceName cmd.OptionalServiceNameID
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *UpdateCommand {
	var c UpdateCommand
	c.Globals = globals
	c.manifest = data
	c.CmdClause = parent.Command("update", "Update a Fastly service")
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
	c.CmdClause.Flag("name", "Service name").Short('n').Action(c.name.Set).StringVar(&c.name.Value)
	c.CmdClause.Flag("comment", "Human-readable comment").Action(c.comment.Set).StringVar(&c.comment.Value)
	return &c
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, source, flag, err := cmd.ServiceID(c.serviceName, c.manifest, c.Globals.APIClient, c.Globals.ErrLog)
	if err != nil {
		return err
	}
	if c.Globals.Verbose() {
		cmd.DisplayServiceID(serviceID, flag, source, out)
	}

	c.input.ServiceID = serviceID

	// TODO(integralist):
	// Validation such as this should become redundant once Go-Fastly is
	// consistently implementing validation (which itself should be redundant
	// once each backend API is 100% confirmed as validating client inputs).
	//
	// As it stands we have multiple clients duplicating logic which should exist
	// (and thus be relied upon) at the API layer.
	//
	// If neither arguments are provided, error with useful message.
	if !c.name.WasSet && !c.comment.WasSet {
		return fmt.Errorf("error parsing arguments: must provide either --name or --comment to update service")
	}

	if c.name.WasSet {
		c.input.Name = fastly.String(c.name.Value)
	}
	if c.comment.WasSet {
		c.input.Comment = fastly.String(c.comment.Value)
	}

	s, err := c.Globals.APIClient.UpdateService(&c.input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
			"Service ID":   serviceID,
			"Service Name": c.name.Value,
			"Comment":      c.comment.Value,
		})
		return err
	}

	text.Success(out, "Updated service %s", s.ID)
	return nil
}
