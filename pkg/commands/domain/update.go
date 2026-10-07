package domain

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v17/fastly"
	"github.com/fastly/go-fastly/v17/fastly/domainmanagement/v1/domains"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// UpdateCommand calls the Fastly API to update domains.
type UpdateCommand struct {
	argparser.Base
	domainID             string
	serviceID            string
	description          argparser.OptionalString
	routingConfigID      argparser.OptionalString
	unsetRoutingConfigID bool
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent argparser.Registerer, g *global.Data) *UpdateCommand {
	c := UpdateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("update", "Update a domain")

	// Required.
	c.CmdClause.Flag("domain-id", "The Domain Identifier (UUID)").Required().StringVar(&c.domainID)

	// Optional
	c.CmdClause.Flag("description", "The description for the domain").Action(c.description.Set).StringVar(&c.description.Value)
	c.CmdClause.Flag("service-id", "The service_id associated with your domain (omit to unset)").StringVar(&c.serviceID)
	c.CmdClause.Flag("routing-config-id", "The Routing Config Identifier to associate with the domain. The routing config must be active").Action(c.routingConfigID.Set).StringVar(&c.routingConfigID.Value)
	c.CmdClause.Flag("unset-routing-config-id", "Remove the domain's association with its routing config").BoolVar(&c.unsetRoutingConfigID)

	return &c
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	input := &domains.UpdateInput{
		DomainID: &c.domainID,
	}
	if c.serviceID != "" {
		input.ServiceID = &c.serviceID
	}

	if c.description.WasSet {
		input.Description = &c.description.Value
	}

	if c.unsetRoutingConfigID && c.routingConfigID.WasSet {
		return errors.New("--routing-config-id and --unset-routing-config-id are mutually exclusive")
	}
	if c.unsetRoutingConfigID {
		input.RoutingConfigurationID = fastly.NullValue[string]()
	} else if c.routingConfigID.WasSet {
		input.RoutingConfigurationID = fastly.NewNullable(c.routingConfigID.Value)
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	d, err := domains.Update(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Domain ID":  c.domainID,
			"Service ID": c.serviceID,
		})
		return err
	}

	serviceOutput := ""
	if d.ServiceID != nil {
		serviceOutput = fmt.Sprintf(", service-id: %s", *d.ServiceID)
	}
	routingConfigOutput := ""
	if d.RoutingConfigurationID != nil {
		routingConfigOutput = fmt.Sprintf(", routing-config-id: %s", *d.RoutingConfigurationID)
	}

	text.Success(out, "Updated domain '%s' (domain-id: %s%s%s)", d.FQDN, d.DomainID, serviceOutput, routingConfigOutput)
	return nil
}
