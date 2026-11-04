package domain

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"
	"github.com/fastly/go-fastly/v19/fastly/domainmanagement/v1/domains"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/text"
)

// UpdateCommand calls the Fastly API to update domains.
type UpdateCommand struct {
	argparser.Base

	// Required.
	DomainID string `required:"" help:"The Domain Identifier (UUID)"`

	// Optional
	Description     argparser.OptionalString `help:"The description for the domain"`
	ServiceID       string                   `help:"The service_id associated with your domain (omit to unset)"`
	RoutingConfigID argparser.OptionalString `help:"The Routing Config Identifier to associate with the domain. Pass 'nil' to remove the association"`
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	input := &domains.UpdateInput{
		DomainID: &c.DomainID,
	}
	if c.ServiceID != "" {
		input.ServiceID = &c.ServiceID
	}

	if c.Description.WasSet {
		input.Description = &c.Description.Value
	}

	if c.RoutingConfigID.WasSet {
		if c.RoutingConfigID.Value == "nil" {
			input.RoutingConfigurationID = fastly.NullValue[string]()
		} else {
			input.RoutingConfigurationID = fastly.NewNullable(c.RoutingConfigID.Value)
		}
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	d, err := domains.Update(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Domain ID":  c.DomainID,
			"Service ID": c.ServiceID,
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
