package subscription

import (
	"context"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
)

var include = []string{"tls_authorizations", "tls_authorizations.globalsign_email_challenge"}

// DescribeCommand calls the Fastly API to describe an appropriate resource.
type DescribeCommand struct {
	argparser.Base

	// Required.
	ID string `required:"" help:"Alphanumeric string identifying a TLS subscription"`

	// Optional.
	Include string `default:"" enum:"tls_authorizations,tls_authorizations.globalsign_email_challenge," hints:"tls_authorizations,tls_authorizations.globalsign_email_challenge" help:"Include related objects (comma-separated values)"`
	argparser.JSONOutput
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input := c.constructInput()

	o, err := c.Globals.APIClient.GetTLSSubscription(context.TODO(), input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"TLS Subscription ID": c.ID,
			"Include":             c.Include,
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

	input.ID = c.ID

	if c.Include != "" {
		input.Include = &c.Include
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
