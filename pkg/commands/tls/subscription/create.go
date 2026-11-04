package subscription

import (
	"context"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/text"
)

const emptyString = ""

var certAuth = []string{"certainly", "lets-encrypt", "globalsign"}

// CreateCommand calls the Fastly API to create an appropriate resource.
type CreateCommand struct {
	argparser.Base

	// Required.
	Domains []string `name:"domain" required:"" sep:"none" help:"Domain(s) to add to the TLS certificates generated for the subscription (set flag once per domain)"`

	// Optional.
	CertAuth   string `default:"" enum:"certainly,lets-encrypt,globalsign," hints:"certainly,lets-encrypt,globalsign" help:"The entity that issues and certifies the TLS certificates for your subscription. Valid values are certainly, lets-encrypt, and globalsign"`
	CommonName string `help:"The domain name associated with the subscription. Default to the first domain specified by --domain"`
	Config     string `help:"Alphanumeric string identifying a TLS configuration"`
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	input := c.constructInput()

	r, err := c.Globals.APIClient.CreateTLSSubscription(context.TODO(), input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"TLS Domains":               c.Domains,
			"TLS Common Name":           c.CommonName,
			"TLS Configuration ID":      c.Config,
			"TLS Certificate Authority": c.CertAuth,
		})
		return err
	}

	text.Success(out, "Created TLS Subscription '%s' (Authority: %s, Common Name: %s)", r.ID, r.CertificateAuthority, r.CommonName.ID)
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *CreateCommand) constructInput() *fastly.CreateTLSSubscriptionInput {
	var input fastly.CreateTLSSubscriptionInput

	domains := make([]*fastly.TLSDomain, len(c.Domains))
	for i, v := range c.Domains {
		domains[i] = &fastly.TLSDomain{ID: v}
	}
	input.Domains = domains

	if c.CommonName != emptyString {
		input.CommonName = &fastly.TLSDomain{ID: c.CommonName}
	}
	if c.CertAuth != emptyString {
		input.CertificateAuthority = c.CertAuth
	}
	if c.Config != emptyString {
		input.Configuration = &fastly.TLSConfiguration{ID: c.Config}
	}

	return &input
}
