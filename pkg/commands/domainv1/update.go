package domainv1

import (
	"errors"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v10/fastly"
	v1 "github.com/fastly/go-fastly/v10/fastly/domains/v1"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// UpdateCommand calls the Fastly API to update domains.
type UpdateCommand struct {
	argparser.Base
	domainID  string
	serviceID string
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
	c.CmdClause.Flag("service-id", "The service_id associated with your domain (omit to unset)").StringVar(&c.serviceID)

	return &c
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	input := &v1.UpdateInput{
		DomainID: &c.domainID,
	}
	if c.serviceID != "" {
		input.ServiceID = &c.serviceID
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	d, err := v1.Update(fc, input)
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

	text.Success(out, "Updated domain '%s' (domain-id: %s%s)", d.FQDN, d.DomainID, serviceOutput)
	return nil
}
