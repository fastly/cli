package domain

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/commands/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/go-fastly/v5/fastly"
)

// NewValidateCommand returns a usable command registered under the parent.
func NewValidateCommand(parent cmd.Registerer, globals *config.Data) *ValidateCommand {
	var c ValidateCommand
	c.CmdClause = parent.Command("validate", "Checks the status of a specific domain's DNS record for a Service Version")
	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)

	// Required flags
	c.RegisterServiceVersionFlag(cmd.ServiceVersionFlagOpts{
		Dst: &c.serviceVersion.Value,
	})

	// Optional flags
	c.CmdClause.Flag("all", "Checks the status of all domains' DNS records for a Service Version").Short('a').BoolVar(&c.all)
	c.CmdClause.Flag("name", "The name of the domain associated with this service").Short('n').Action(c.name.Set).StringVar(&c.name.Value)
	c.RegisterServiceIDFlag(&c.manifest.Flag.ServiceID)

	return &c
}

// ValidateCommand calls the Fastly API to describe an appropriate resource.
type ValidateCommand struct {
	cmd.Base

	all            bool
	manifest       manifest.Data
	name           cmd.OptionalString
	serviceVersion cmd.OptionalServiceVersion
}

// Exec invokes the application logic for the command.
func (c *ValidateCommand) Exec(in io.Reader, out io.Writer) error {
	// Exit early if no token configured.
	_, s := c.Globals.Token()
	if s == config.SourceUndefined {
		return errors.ErrNoToken
	}

	serviceID, serviceVersion, err := cmd.ServiceDetails(cmd.ServiceDetailsOpts{
		AllowActiveLocked:  true,
		Client:             c.Globals.Client,
		Manifest:           c.manifest,
		Out:                out,
		ServiceVersionFlag: c.serviceVersion,
		VerboseMode:        c.Globals.Flag.Verbose,
	})
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
			"Service ID":      serviceID,
			"Service Version": errors.ServiceVersion(serviceVersion),
		})
		return err
	}

	if c.all {
		input := c.constructInputAll(serviceID, serviceVersion.Number)

		r, err := c.Globals.Client.ValidateAllDomains(input)
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
				"Service ID":      serviceID,
				"Service Version": serviceVersion.Number,
			})
			return err
		}

		c.printAll(out, r)
		return nil
	}

	input, err := c.constructInput(serviceID, serviceVersion.Number)
	if err != nil {
		return err
	}

	r, err := c.Globals.Client.ValidateDomain(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
			"Service ID":      serviceID,
			"Service Version": serviceVersion.Number,
			"Domain Name":     c.name,
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

	if !c.name.WasSet {
		return nil, errors.RemediationError{
			Inner:       fmt.Errorf("error parsing arguments: must provide --name flag"),
			Remediation: "Alternatively pass --all to validate all domains.",
		}
	}
	input.Name = c.name.Value

	return &input, nil
}

// print displays the information returned from the API.
func (c *ValidateCommand) print(out io.Writer, r *fastly.DomainValidationResult) {
	fmt.Fprintf(out, "\nService ID: %s\n", r.Metadata.ServiceID)
	fmt.Fprintf(out, "Service Version: %d\n\n", r.Metadata.ServiceVersion)
	fmt.Fprintf(out, "Name: %s\n", r.Metadata.Name)
	fmt.Fprintf(out, "Valid: %t\n", r.Valid)

	if r.CName != "" {
		fmt.Fprintf(out, "CNAME: %s\n", r.CName)
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
			fmt.Fprintf(out, "\nService ID: %s\n", r.Metadata.ServiceID)
			fmt.Fprintf(out, "Service Version: %d\n\n", r.Metadata.ServiceVersion)
		}
		fmt.Fprintf(out, "Name: %s\n", r.Metadata.Name)
		fmt.Fprintf(out, "Valid: %t\n", r.Valid)

		if r.CName != "" {
			fmt.Fprintf(out, "CNAME: %s\n", r.CName)
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
