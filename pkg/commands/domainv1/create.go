package domainv1

import (
	"errors"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v9/fastly"
	v1 "github.com/fastly/go-fastly/v9/fastly/domains/v1"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// CreateCommand calls the Fastly API to create domains.
type CreateCommand struct {
	argparser.Base

	// Required.
	fqdn      string
	serviceID string
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent argparser.Registerer, g *global.Data) *CreateCommand {
	c := CreateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("create", "Create a domain").Alias("add")

	// Optional.
	c.CmdClause.Flag("fqdn", "The fully qualified domain name").Required().StringVar(&c.fqdn)
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagServiceIDName,
		Description: "The service_id associated with your domain",
		Dst:         &c.serviceID,
		Short:       's',
	})
	return &c
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	input := &v1.CreateInput{
		FQDN: &c.fqdn,
	}
	if c.serviceID != "" {
		input.ServiceID = &c.serviceID
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	d, err := v1.Create(fc, input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"FQDN":       c.fqdn,
			"Service ID": c.serviceID,
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
