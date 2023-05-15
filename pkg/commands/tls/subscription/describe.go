package subscription

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/go-fastly/v8/fastly"
)

var include = []string{"tls_authorizations", "tls_authorizations.globalsign_email_challenge"}

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent cmd.Registerer, g *global.Data, m manifest.Data) *DescribeCommand {
	var c DescribeCommand
	c.CmdClause = parent.Command("describe", "Show a TLS subscription").Alias("get")
	c.Globals = g
	c.manifest = m

	// Required.
	c.CmdClause.Flag("id", "Alphanumeric string identifying a TLS subscription").Required().StringVar(&c.id)

	// Optional.
	c.CmdClause.Flag("include", "Include related objects (comma-separated values)").HintOptions(include...).EnumVar(&c.include, include...)
	c.RegisterFlagBool(c.JSONFlag()) // --json

	return &c
}

// DescribeCommand calls the Fastly API to describe an appropriate resource.
type DescribeCommand struct {
	cmd.Base
	cmd.JSONOutput

	id       string
	include  string
	manifest manifest.Data
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input := c.constructInput()

	o, err := c.Globals.APIClient.GetTLSSubscription(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"TLS Subscription ID": c.id,
			"Include":             c.include,
		})
		return err
	}

	if ok, err := c.WriteJSON(out, o); ok {
		return err
	}

	return c.print(out, o)
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *DescribeCommand) constructInput() *fastly.GetTLSSubscriptionInput {
	var input fastly.GetTLSSubscriptionInput

	input.ID = c.id

	if c.include != "" {
		input.Include = &c.include
	}

	return &input
}

// print displays the information returned from the API.
func (c *DescribeCommand) print(out io.Writer, r *fastly.TLSSubscription) error {
	fmt.Fprintf(out, "\nID: %s\n", r.ID)
	fmt.Fprintf(out, "Certificate Authority: %s\n", r.CertificateAuthority)
	fmt.Fprintf(out, "State: %s\n", r.State)

	if r.CreatedAt != nil {
		fmt.Fprintf(out, "Created at: %s\n", r.CreatedAt)
	}
	if r.UpdatedAt != nil {
		fmt.Fprintf(out, "Updated at: %s\n", r.UpdatedAt)
	}

	return nil
}
