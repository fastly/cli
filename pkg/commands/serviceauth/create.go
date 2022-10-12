package serviceauth

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v6/fastly"
)

// Permissions is a list of supported permission values.
// https://developer.fastly.com/reference/api/account/service-authorization/#data-model
var Permissions = []string{"full", "read_only", "purge_select", "purge_all"}

// CreateCommand calls the Fastly API to create a service authorization.
type CreateCommand struct {
	cmd.Base
	input       fastly.CreateServiceAuthorizationInput
	manifest    manifest.Data
	serviceName cmd.OptionalServiceNameID
	userID      string
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *CreateCommand {
	var c CreateCommand
	c.Globals = globals
	c.manifest = data
	c.CmdClause = parent.Command("create", "Create service authorization").Alias("add")

	// NOTE: We default to 'read_only' for security reasons.
	// The API otherwise defaults to 'full' permissions!
	c.CmdClause.Flag("permissions", "The permission the user has in relation to the service (default: read_only)").HintOptions(Permissions...).Default("read_only").Short('p').EnumVar(&c.input.Permission, Permissions...)

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
	c.CmdClause.Flag("user-id", "Alphanumeric string identifying the user").Required().Short('u').StringVar(&c.userID)
	return &c
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	serviceID, source, flag, err := cmd.ServiceID(c.serviceName, c.manifest, c.Globals.APIClient, c.Globals.ErrLog)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":   c.manifest.Flag.ServiceID,
			"Service Name": c.serviceName.Value,
		})
		return err
	}
	if c.Globals.Flag.Verbose {
		cmd.DisplayServiceID(serviceID, flag, source, out)
	}

	c.input.Service = &fastly.SAService{
		ID: serviceID,
	}
	c.input.User = &fastly.SAUser{
		ID: c.userID,
	}

	s, err := c.Globals.APIClient.CreateServiceAuthorization(&c.input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID": serviceID,
			"Flag":       flag,
		})
		return err
	}

	text.Success(out, "Created service authorization %s", s.ID)
	return nil
}
