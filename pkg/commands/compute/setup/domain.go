package setup

import (
	"fmt"
	"io"
	"math/rand"
	"regexp"
	"time"

	petname "github.com/dustinkirkland/golang-petname"
	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v5/fastly"
)

const defaultTopLevelDomain = "edgecompute.app"

var domainNameRegEx = regexp.MustCompile(`(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z0-9][a-z0-9-]{0,61}[a-z0-9]`)

// Domains represents the service state related to domains defined within the
// fastly.toml [setup] configuration.
//
// NOTE: It implements the setup.Interface interface.
type Domains struct {
	// Public
	Available      []*fastly.Domain
	APIClient      api.Interface
	PackageDomain  string
	Progress       text.Progress
	Required       []Domain
	ServiceID      string
	ServiceVersion int
	AcceptDefaults bool
	Stdin          io.Reader
	Stdout         io.Writer

	// Private
	missing bool
}

// Domain represents the configuration parameters for creating a domain via the
// API client.
type Domain struct {
	Name string
}

// Configure prompts the user for specific values related to the service resource.
//
// NOTE: If --domain flag is used we'll use that as the domain to create.
func (d *Domains) Configure() error {
	// PackageDomain is the --domain flag value.
	if d.PackageDomain != "" {
		d.Required = append(d.Required, Domain{
			Name: d.PackageDomain,
		})
		return nil
	}

	rand.Seed(time.Now().UnixNano())
	defaultDomain := fmt.Sprintf("%s.%s", petname.Generate(3, "-"), defaultTopLevelDomain)

	var (
		domain string
		err    error
	)
	if !d.AcceptDefaults {
		domain, err = text.Input(d.Stdout, fmt.Sprintf("Domain: [%s] ", defaultDomain), d.Stdin, d.validateDomain)
		if err != nil {
			return fmt.Errorf("error reading input %w", err)
		}
		text.Break(d.Stdout)
	}

	if domain == "" {
		d.Required = append(d.Required, Domain{
			Name: defaultDomain,
		})
		return nil
	}

	d.Required = append(d.Required, Domain{
		Name: domain,
	})
	return nil
}

// Create calls the relevant API to create the service resource(s).
func (d *Domains) Create() error {
	if d.Progress == nil {
		return errors.RemediationError{
			Inner:       fmt.Errorf("internal logic error: no text.Progress configured for setup.Domains"),
			Remediation: errors.BugRemediation,
		}
	}

	for _, domain := range d.Required {
		d.Progress.Step(fmt.Sprintf("Creating domain '%s'...", domain.Name))

		_, err := d.APIClient.CreateDomain(&fastly.CreateDomainInput{
			ServiceID:      d.ServiceID,
			ServiceVersion: d.ServiceVersion,
			Name:           domain.Name,
		})
		if err != nil {
			d.Progress.Fail()
			return fmt.Errorf("error creating domain: %w", err)
		}
	}

	return nil
}

// Missing indicates if there are missing resources that need to be created.
func (d *Domains) Missing() bool {
	return d.missing || len(d.Required) > 0
}

// Predefined indicates if the service resource has been specified within the
// fastly.toml file using a [setup] configuration block.
//
// NOTE: Domains are not configurable via the fastly.toml [setup] and so this
// becomes a no-op function that returned a canned response.
func (d *Domains) Predefined() bool {
	return false
}

// Validate checks if the service has the required resources.
//
// NOTE: It should set an internal `missing` field (boolean) accordingly so that
// the Missing() method can report the state of the resource.
func (d *Domains) Validate() error {
	var err error
	d.Available, err = d.APIClient.ListDomains(&fastly.ListDomainsInput{
		ServiceID:      d.ServiceID,
		ServiceVersion: d.ServiceVersion,
	})
	if err != nil {
		return fmt.Errorf("error fetching service domains: %w", err)
	}

	if len(d.Available) < 1 {
		d.missing = true
	}
	return nil
}

// validateDomain checks the user entered domain is valid.
//
// NOTE: An empty value is allowed so that a default domain can be utilised.
func (d *Domains) validateDomain(input string) error {
	if input == "" {
		return nil
	}
	if !domainNameRegEx.MatchString(input) {
		return fmt.Errorf("must be valid domain name")
	}
	return nil
}
