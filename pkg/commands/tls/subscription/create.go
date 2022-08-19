package subscription

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v6/fastly"
)

const certAuth = []string{"lets-encrypt", "globalsign"}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *CreateCommand {
	var c CreateCommand
	c.CmdClause = parent.Command("create", "Create a new TLS subscription").Alias("add")
	c.Globals = globals
	c.manifest = data

	// Required flags
	c.CmdClause.Flag("domain", "Domain(s) to add to the TLS certificates generated for the subscription (set flag once per domain)").Required().StringsVar(&c.domains)

	// Optional flags
	c.CmdClause.Flag("cert-auth", "The entity that issues and certifies the TLS certificates for your subscription. Valid values are lets-encrypt or globalsign").HintOptions(certAuth...).EnumVar(&c.certAuth, certAuth...)
	c.CmdClause.Flag("common-name", "The domain name associated with the subscription. Default to the first domain specified by --domain").StringVar(&c.commonName)
	c.CmdClause.Flag("config", "Alphanumeric string identifying a TLS configuration").StringVar(&c.config)

	return &c
}

// CreateCommand calls the Fastly API to create an appropriate resource.
type CreateCommand struct {
	cmd.Base

	certAuth   string
	commonName string
	config     string
	domains    []string
	manifest   manifest.Data
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(in io.Reader, out io.Writer) error {
	input := c.constructInput()

	r, err := c.Globals.APIClient.CreateTLSSubscription(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
			"TLS Domains":               c.domains,
			"TLS Common Name":           c.commonName,
			"TLS Configuration ID":      c.config,
			"TLS Certificate Authority": c.certAuth,
		})
		return err
	}

	text.Success(out, "Created TLS Subscription '%s' (Authority: %s, Common Name: %s)", r.ID, r.CertificateAuthority, r.CommonName)
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *CreateCommand) constructInput() *fastly.CreateTLSSubscriptionInput {
	var input fastly.CreateTLSSubscriptionInput

	domains := make([]*fastly.TLSDomain, len(c.domains))
	for i, v := range c.domains {
		domains[i] = &fastly.TLSDomain{ID: v}
	}
	input.Domains = domains

	if c.commonName != "" {
		input.CommonName = &fastly.TLSDomain{ID: c.commonName}
	}
	if c.certAuth != "" {
		input.CertificateAuthority = c.certAuth
	}
	if c.config != "" {
		input.Configuration = &fastly.TLSConfiguration{ID: c.config}
	}

	return &input
}
