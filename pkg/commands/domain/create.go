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

// CreateCommand calls the Fastly API to create domains.
type CreateCommand struct {
	argparser.Base

	// Required.
	Fqdn      string `required:"" help:"The fully qualified domain name"`
	ServiceID string `short:"s" help:"The service_id associated with your domain"`

	// Optional.
	Description argparser.OptionalString `help:"The description for the domain"`
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	input := &domains.CreateInput{
		FQDN: &c.Fqdn,
	}
	if c.ServiceID != "" {
		input.ServiceID = &c.ServiceID
	}

	if c.Description.WasSet {
		input.Description = &c.Description.Value
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	d, err := domains.Create(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"FQDN":       c.Fqdn,
			"Service ID": c.ServiceID,
		})
		return err
	}

	serviceOutput := ""
	if d.ServiceID != nil {
		serviceOutput = fmt.Sprintf(", service-id: %s", *d.ServiceID)
	}

	text.Success(out, "Created domain '%s' (domain-id: %s%s)", d.FQDN, d.DomainID, serviceOutput)
	return nil
}

// FlagOrder lists the flags in the order kingpin showed them in help, which
// the fields do not follow.
func (c *CreateCommand) FlagOrder() []string {
	return []string{
		"description", "fqdn", "service-id",
	}
}
