package domain

import (
	"context"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/errors"
)

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *ValidateCommand) AfterApply() error {
	c.serviceVersion.Value = c.Version
	if c.ServiceID != nil {
		c.Globals.Manifest.Flag.ServiceID = *c.ServiceID
	}
	return nil
}

// ValidateCommand calls the Fastly API to describe an appropriate resource.
type ValidateCommand struct {
	argparser.Base

	// Required.
	Version string `required:"" help:"'latest', 'active', 'staged', or the number of a specific Fastly service version"`

	// Optional.
	All            bool                            `short:"a" help:"Checks the status of all domains' DNS records for a Service Version"`
	NameFlag       argparser.OptionalString        `name:"name" short:"n" help:"The name of the domain associated with this service"`
	ServiceID      *string                         `short:"s" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`
	ServiceName    argparser.OptionalServiceNameID `help:"The name of the service"`
	serviceVersion argparser.OptionalServiceVersion
}

// Exec invokes the application logic for the command.
func (c *ValidateCommand) Exec(_ io.Reader, out io.Writer) error {
	serviceID, serviceVersion, err := argparser.ServiceDetails(argparser.ServiceDetailsOpts{
		APIClient:          c.Globals.APIClient,
		Manifest:           *c.Globals.Manifest,
		Out:                out,
		ServiceNameFlag:    c.ServiceName,
		ServiceVersionFlag: c.serviceVersion,
		VerboseMode:        c.Globals.Flags.Verbose,
	})
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": errors.ServiceVersion(serviceVersion),
		})
		return err
	}

	serviceVersionNumber := fastly.ToValue(serviceVersion.Number)

	if c.All {
		input := c.constructInputAll(serviceID, serviceVersionNumber)

		r, err := c.Globals.APIClient.ValidateAllDomains(context.TODO(), input)
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
				"Service ID":      serviceID,
				"Service Version": serviceVersionNumber,
			})
			return err
		}

		c.printAll(out, r)
		return nil
	}

	input, err := c.constructInput(serviceID, serviceVersionNumber)
	if err != nil {
		return err
	}

	r, err := c.Globals.APIClient.ValidateDomain(context.TODO(), input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": serviceVersionNumber,
			"Domain Name":     c.NameFlag,
		})
		return err
	}

	c.print(out, r)
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *ValidateCommand) constructInput(serviceID string, serviceVersion int) (*fastly.ValidateDomainInput, error) {
	var input fastly.ValidateDomainInput

	input.ServiceID = serviceID
	input.ServiceVersion = serviceVersion

	if !c.NameFlag.WasSet {
		return nil, errors.RemediationError{
			Inner:       fmt.Errorf("error parsing arguments: must provide --name flag"),
			Remediation: "Alternatively pass --all to validate all domains.",
		}
	}
	input.Name = c.NameFlag.Value

	return &input, nil
}

// print displays the information returned from the API.
func (c *ValidateCommand) print(out io.Writer, r *fastly.DomainValidationResult) {
	fmt.Fprintf(out, "\nService ID: %s\n", fastly.ToValue(r.Metadata.ServiceID))
	fmt.Fprintf(out, "Service Version: %d\n\n", fastly.ToValue(r.Metadata.ServiceVersion))
	fmt.Fprintf(out, "Name: %s\n", fastly.ToValue(r.Metadata.Name))
	fmt.Fprintf(out, "Valid: %t\n", fastly.ToValue(r.Valid))

	if r.CName != nil {
		fmt.Fprintf(out, "CNAME: %s\n", *r.CName)
	}
	if r.Metadata.CreatedAt != nil {
		fmt.Fprintf(out, "Created at: %s\n", r.Metadata.CreatedAt)
	}
	if r.Metadata.UpdatedAt != nil {
		fmt.Fprintf(out, "Updated at: %s\n", r.Metadata.UpdatedAt)
	}
	if r.Metadata.DeletedAt != nil {
		fmt.Fprintf(out, "Deleted at: %s\n", r.Metadata.DeletedAt)
	}
	fmt.Fprintf(out, "\n")
}

// constructInputAll transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *ValidateCommand) constructInputAll(serviceID string, serviceVersion int) *fastly.ValidateAllDomainsInput {
	var input fastly.ValidateAllDomainsInput

	input.ServiceID = serviceID
	input.ServiceVersion = serviceVersion

	return &input
}

// printAll displays all domain validation results returned from the API.
func (c *ValidateCommand) printAll(out io.Writer, rs []*fastly.DomainValidationResult) {
	for i, r := range rs {
		// We only need to print the Service ID/Version once.
		if i == 0 {
			fmt.Fprintf(out, "\nService ID: %s\n", fastly.ToValue(r.Metadata.ServiceID))
			fmt.Fprintf(out, "Service Version: %d\n\n", fastly.ToValue(r.Metadata.ServiceVersion))
		}
		fmt.Fprintf(out, "Name: %s\n", fastly.ToValue(r.Metadata.Name))
		fmt.Fprintf(out, "Valid: %t\n", fastly.ToValue(r.Valid))

		if r.CName != nil {
			fmt.Fprintf(out, "CNAME: %s\n", *r.CName)
		}
		if r.Metadata.CreatedAt != nil {
			fmt.Fprintf(out, "Created at: %s\n", r.Metadata.CreatedAt)
		}
		if r.Metadata.UpdatedAt != nil {
			fmt.Fprintf(out, "Updated at: %s\n", r.Metadata.UpdatedAt)
		}
		if r.Metadata.DeletedAt != nil {
			fmt.Fprintf(out, "Deleted at: %s\n", r.Metadata.DeletedAt)
		}
		fmt.Fprintf(out, "\n")
	}
}
