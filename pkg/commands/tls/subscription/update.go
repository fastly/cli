package subscription

import (
	"context"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/text"
)

// UpdateCommand calls the Fastly API to update an appropriate resource.
type UpdateCommand struct {
	argparser.Base

	// Required.
	ID string `required:"" help:"Alphanumeric string identifying a TLS subscription"`

	// Optional.
	CommonName string                 `help:"The domain name associated with the subscription"`
	Config     string                 `help:"Alphanumeric string identifying a TLS configuration"`
	Domains    []string               `name:"domain" sep:"none" help:"Domain(s) to add to the TLS certificates generated for the subscription (set flag once per domain)"`
	Force      argparser.OptionalBool `help:"A flag that allows you to edit and delete a subscription with active domains"`
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	input := c.constructInput()

	r, err := c.Globals.APIClient.UpdateTLSSubscription(context.TODO(), input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"TLS Subscription ID": c.ID,
			"Force":               c.Force.Value,
		})
		return err
	}

	text.Success(out, "Updated TLS Subscription '%s' (Authority: %s, Common Name: %s)", r.ID, r.CertificateAuthority, r.CommonName.ID)
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) constructInput() *fastly.UpdateTLSSubscriptionInput {
	var input fastly.UpdateTLSSubscriptionInput

	input.ID = c.ID

	domains := make([]*fastly.TLSDomain, len(c.Domains))
	for i, v := range c.Domains {
		domains[i] = &fastly.TLSDomain{ID: v}
	}
	input.Domains = domains

	if c.CommonName != "" {
		input.CommonName = &fastly.TLSDomain{ID: c.CommonName}
	}
	if c.Config != "" {
		input.Configuration = &fastly.TLSConfiguration{ID: c.Config}
	}
	if c.Force.WasSet {
		input.Force = c.Force.Value
	}

	return &input
}
