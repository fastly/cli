package activation

import (
	"context"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// ListCommand calls the Fastly API to list appropriate resources.
type ListCommand struct {
	argparser.Base

	// Optional.
	FilterTLSCertID   string `name:"filter-cert" help:"Limit the returned activations to a specific certificate"`
	FilterTLSConfigID string `name:"filter-config" help:"Limit the returned activations to a specific TLS configuration"`
	FilterTLSDomainID string `name:"filter-domain" help:"Limit the returned rules to a specific domain name"`
	Include           string `default:"" enum:"tls_certificate,tls_configuration,tls_domain," hints:"tls_certificate,tls_configuration,tls_domain" help:"Include related objects (comma-separated values)"`
	argparser.JSONOutput
	PageNumber int `name:"page" help:"Page number of data set to fetch"`
	PageSize   int `name:"per-page" help:"Number of records per page"`
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input := c.constructInput()

	o, err := c.Globals.APIClient.ListTLSActivations(context.TODO(), input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Filter TLS Certificate ID":   c.FilterTLSCertID,
			"Filter TLS Configuration ID": c.FilterTLSConfigID,
			"Filter TLS Domain ID":        c.FilterTLSDomainID,
			"Include":                     c.Include,
			"Page Number":                 c.PageNumber,
			"Page Size":                   c.PageSize,
		})
		return err
	}

	if ok, err := c.WriteJSON(out, o); ok {
		return err
	}

	if c.Globals.Verbose() {
		c.printVerbose(out, o)
	} else {
		err = c.printSummary(out, o)
		if err != nil {
			return err
		}
	}
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *ListCommand) constructInput() *fastly.ListTLSActivationsInput {
	var input fastly.ListTLSActivationsInput

	if c.FilterTLSCertID != "" {
		input.FilterTLSCertificateID = c.FilterTLSCertID
	}
	if c.FilterTLSConfigID != "" {
		input.FilterTLSConfigurationID = c.FilterTLSConfigID
	}
	if c.FilterTLSDomainID != "" {
		input.FilterTLSDomainID = c.FilterTLSDomainID
	}
	if c.Include != "" {
		input.Include = c.Include
	}
	if c.PageNumber > 0 {
		input.PageNumber = c.PageNumber
	}
	if c.PageSize > 0 {
		input.PageSize = c.PageSize
	}

	return &input
}

// printVerbose displays the information returned from the API in a verbose
// format.
func (c *ListCommand) printVerbose(out io.Writer, rs []*fastly.TLSActivation) {
	for _, r := range rs {
		fmt.Fprintf(out, "\nID: %s\n", r.ID)

		if r.CreatedAt != nil {
			fmt.Fprintf(out, "Created at: %s\n", r.CreatedAt)
		}

		fmt.Fprintf(out, "\n")
	}
}

// printSummary displays the information returned from the API in a summarised
// format.
func (c *ListCommand) printSummary(out io.Writer, rs []*fastly.TLSActivation) error {
	t := text.NewTable(out)
	t.AddHeader("ID", "CREATED_AT")
	for _, r := range rs {
		t.AddLine(r.ID, r.CreatedAt)
	}
	t.Print()
	return nil
}
