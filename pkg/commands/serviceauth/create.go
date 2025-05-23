package serviceauth

import (
	"io"

	"github.com/fastly/go-fastly/v10/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// Permissions is a list of supported permission values.
// https://www.fastly.com/documentation/reference/api/account/service-authorization#data-model
var Permissions = []string{"full", "read_only", "purge_select", "purge_all"}

// CreateCommand calls the Fastly API to create a service authorization.
type CreateCommand struct {
	argparser.Base
	input       fastly.CreateServiceAuthorizationInput
	serviceName argparser.OptionalServiceNameID
	userID      string
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent argparser.Registerer, g *global.Data) *CreateCommand {
	c := CreateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("create", "Create service authorization").Alias("add")

	// Required.
	c.CmdClause.Flag("user-id", "Alphanumeric string identifying the user").Required().Short('u').StringVar(&c.userID)

	// Optional.
	// NOTE: We default to 'read_only' for security reasons.
	// The API otherwise defaults to 'full' permissions!
	c.CmdClause.Flag("permission", "The permission the user has in relation to the service (default: read_only)").HintOptions(Permissions...).Default("read_only").Short('p').EnumVar(&c.input.Permission, Permissions...)
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagServiceIDName,
		Description: argparser.FlagServiceIDDesc,
		Dst:         &g.Manifest.Flag.ServiceID,
		Short:       's',
	})
	c.RegisterFlag(argparser.StringFlagOpts{
		Action:      c.serviceName.Set,
		Name:        argparser.FlagServiceName,
		Description: argparser.FlagServiceNameDesc,
		Dst:         &c.serviceName.Value,
	})
	return &c
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	serviceID, source, flag, err := argparser.ServiceID(c.serviceName, *c.Globals.Manifest, c.Globals.APIClient, c.Globals.ErrLog)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":   c.Globals.Manifest.Flag.ServiceID,
			"Service Name": c.serviceName.Value,
		})
		return err
	}
	if c.Globals.Flags.Verbose {
		argparser.DisplayServiceID(serviceID, flag, source, out)
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
