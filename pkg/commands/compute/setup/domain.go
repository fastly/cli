package setup

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"time"

	petname "github.com/dustinkirkland/golang-petname"
	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v7/fastly"
)

const defaultTopLevelDomain = "edgecompute.app"

var domainNameRegEx = regexp.MustCompile(`(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z0-9][a-z0-9-]{0,61}[a-z0-9]`)

// Domains represents the service state related to domains.
//
// NOTE: It implements the setup.Interface interface.
type Domains struct {
	// Public
	APIClient      api.Interface
	AcceptDefaults bool
	NonInteractive bool
	PackageDomain  string
	Progress       text.Progress
	ServiceID      string
	ServiceVersion int
	Stdin          io.Reader
	Stdout         io.Writer

	// Private
	available []*fastly.Domain
	missing   bool
	required  []Domain
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
		d.required = append(d.required, Domain{
			Name: d.PackageDomain,
		})
		return nil
	}

	defaultDomain := generateDomainName()

	var (
		domain string
		err    error
	)
	if !d.AcceptDefaults && !d.NonInteractive {
		domain, err = text.Input(d.Stdout, text.BoldYellow(fmt.Sprintf("Domain: [%s] ", defaultDomain)), d.Stdin, d.validateDomain)
		if err != nil {
			return fmt.Errorf("error reading input %w", err)
		}
		text.Break(d.Stdout)
	}

	if domain == "" {
		domain = defaultDomain
	}
	d.required = append(d.required, Domain{
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

OUTER:
	for _, domain := range d.required {
		d.Progress.Step(fmt.Sprintf("Creating domain '%s'...", domain.Name))

		_, err := d.APIClient.CreateDomain(&fastly.CreateDomainInput{
			ServiceID:      d.ServiceID,
			ServiceVersion: d.ServiceVersion,
			Name:           &domain.Name,
		})
		if err != nil {
			d.Progress.Fail()

			if e, ok := err.(*fastly.HTTPError); ok {
				if e.StatusCode == http.StatusBadRequest {
					for _, he := range e.Errors {
						// NOTE: In case the domain is already used by another customer.
						// We'll give the user one additional chance to correct the domain.
						if strings.Contains(he.Detail, "by another customer") {
							var domain string
							defaultDomain := generateDomainName()
							if !d.AcceptDefaults && !d.NonInteractive {
								text.Break(d.Stdout)
								domain, err = text.Input(d.Stdout, text.BoldYellow(fmt.Sprintf("Domain already taken, please choose another: [%s] ", defaultDomain)), d.Stdin, d.validateDomain)
								if err != nil {
									return fmt.Errorf("error reading input %w", err)
								}
								text.Break(d.Stdout)
							}
							if domain == "" {
								domain = defaultDomain
							}
							if _, err = d.APIClient.CreateDomain(&fastly.CreateDomainInput{
								ServiceID:      d.ServiceID,
								ServiceVersion: d.ServiceVersion,
								Name:           &domain,
							}); err == nil {
								continue OUTER
							}
						}
					}
				}
			}

			return fmt.Errorf("error creating domain: %w", err)
		}
	}

	return nil
}

// Missing indicates if there are missing resources that need to be created.
func (d *Domains) Missing() bool {
	return d.missing || len(d.required) > 0
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
	available, err := d.APIClient.ListDomains(&fastly.ListDomainsInput{
		ServiceID:      d.ServiceID,
		ServiceVersion: d.ServiceVersion,
	})
	if err != nil {
		return fmt.Errorf("error fetching service domains: %w", err)
	}
	d.available = available

	if len(d.available) == 0 {
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

func generateDomainName() string {
	// IMPORTANT: go1.20 deprecates rand.Seed
	// The global random number generator (RNG) is now automatically seeded.
	// If not seeded, the same domain name is repeated on each run.
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%s.%s", petname.Generate(3, "-"), defaultTopLevelDomain)
}
