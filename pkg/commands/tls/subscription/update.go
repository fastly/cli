package subscription

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v7/fastly"
)

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent cmd.Registerer, g *global.Data, m manifest.Data) *UpdateCommand {
	var c UpdateCommand
	c.CmdClause = parent.Command("update", "Change the TLS domains or common name associated with this subscription, or update the TLS configuration for this set of domains")
	c.Globals = g
	c.manifest = m

	// required
	c.CmdClause.Flag("id", "Alphanumeric string identifying a TLS subscription").Required().StringVar(&c.id)

	// optional
	c.CmdClause.Flag("common-name", "The domain name associated with the subscription").StringVar(&c.commonName)
	c.CmdClause.Flag("config", "Alphanumeric string identifying a TLS configuration").StringVar(&c.config)
	c.CmdClause.Flag("domain", "Domain(s) to add to the TLS certificates generated for the subscription (set flag once per domain)").StringsVar(&c.domains)
	c.CmdClause.Flag("force", "A flag that allows you to edit and delete a subscription with active domains").Action(c.force.Set).BoolVar(&c.force.Value)

	return &c
}

// UpdateCommand calls the Fastly API to update an appropriate resource.
type UpdateCommand struct {
	cmd.Base

	commonName string
	config     string
	domains    []string
	force      cmd.OptionalBool
	id         string
	manifest   manifest.Data
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	input := c.constructInput()

	r, err := c.Globals.APIClient.UpdateTLSSubscription(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"TLS Subscription ID": c.id,
			"Force":               c.force.Value,
		})
		return err
	}

	text.Success(out, "Updated TLS Subscription '%s' (Authority: %s, Common Name: %s)", r.ID, r.CertificateAuthority, r.CommonName.ID)
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) constructInput() *fastly.UpdateTLSSubscriptionInput {
	var input fastly.UpdateTLSSubscriptionInput

	input.ID = c.id

	domains := make([]*fastly.TLSDomain, len(c.domains))
	for i, v := range c.domains {
		domains[i] = &fastly.TLSDomain{ID: v}
	}
	input.Domains = domains

	if c.commonName != "" {
		input.CommonName = &fastly.TLSDomain{ID: c.commonName}
	}
	if c.config != "" {
		input.Configuration = &fastly.TLSConfiguration{ID: c.config}
	}
	if c.force.WasSet {
		input.Force = c.force.Value
	}

	return &input
}
