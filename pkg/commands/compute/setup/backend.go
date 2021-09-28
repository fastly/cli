package setup

import (
	"fmt"
	"io"
	"net"
	"strconv"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/commands/compute/manifest"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v5/fastly"
)

// Backends represents the service state related to backends defined within the
// fastly.toml [setup] configuration.
//
// NOTE: It implements the setup.Interface interface.
type Backends struct {
	// Public
	Available      []*fastly.Backend
	APIClient      api.Interface
	Progress       text.Progress
	Required       []Backend
	ServiceID      string
	ServiceVersion int
	Setup          map[string]*manifest.SetupBackend
	AcceptDefaults bool
	Stdin          io.Reader
	Stdout         io.Writer

	// Private
	missing bool
}

// Backend represents the configuration parameters for creating a backend via
// the API client.
type Backend struct {
	Address        string
	Name           string
	OverrideHost   string
	Port           uint
	SSLSNIHostname string
}

// Configure prompts the user (if necessary) for specific values related to the
// service resource.
func (b *Backends) Configure() (err error) {
	if b.Predefined() {
		return b.checkPredefined()
	}
	return b.promptForBackend()
}

// Create calls the relevant API to create the service resource(s).
func (b *Backends) Create() error {
	if b.Progress == nil {
		return errors.RemediationError{
			Inner:       fmt.Errorf("internal logic error: no text.Progress configured for setup.Backends"),
			Remediation: errors.BugRemediation,
		}
	}

	for _, backend := range b.Required {
		if !b.isOriginless() {
			b.Progress.Step(fmt.Sprintf("Creating backend '%s' (host: %s, port: %d)...", backend.Name, backend.Address, backend.Port))
		}

		_, err := b.APIClient.CreateBackend(&fastly.CreateBackendInput{
			ServiceID:      b.ServiceID,
			ServiceVersion: b.ServiceVersion,
			Name:           backend.Name,
			Address:        backend.Address,
			Port:           backend.Port,
			OverrideHost:   backend.OverrideHost,
			SSLSNIHostname: backend.SSLSNIHostname,
		})
		if err != nil {
			b.Progress.Fail()
			if b.isOriginless() {
				return fmt.Errorf("error configuring the service: %w", err)
			}
			return fmt.Errorf("error creating backend: %w", err)
		}
	}

	b.missing = false
	return nil
}

// Missing indicates if there are missing resources that need to be created.
func (b *Backends) Missing() bool {
	return b.missing || len(b.Required) > 0
}

// Predefined indicates if the service resource has been specified within the
// fastly.toml file using a [setup] configuration block.
func (b *Backends) Predefined() bool {
	return len(b.Setup) > 0
}

// Validate checks if the service has the required resources.
//
// NOTE: It should set an internal `missing` field (boolean) accordingly so that
// the Missing() method can report the state of the resource.
func (b *Backends) Validate() (err error) {
	if err = b.available(); err != nil {
		return err
	}

	if b.Predefined() {
		for name, settings := range b.Setup {
			var (
				condition bool
				found     bool
			)

			for _, backend := range b.Available {
				condition = backend.Name == name

				if settings.Address != "" {
					condition = condition && backend.Address == settings.Address
				}
				if settings.Port != 0 {
					condition = condition && backend.Port == settings.Port
				}
				if condition {
					settings.Exists = true
					found = true
				}
			}

			if !found {
				b.missing = true
				break
			}
		}
	}

	return nil
}

// isOriginless indicates if the required backend is originless.
func (b *Backends) isOriginless() bool {
	return len(b.Required) == 1 && b.Required[0].Name == "originless" && b.Required[0].Address == "127.0.0.1"
}

// available sets the Available field with the result of calling the
// ListBackends API.
func (b *Backends) available() (err error) {
	b.Available, err = b.APIClient.ListBackends(&fastly.ListBackendsInput{
		ServiceID:      b.ServiceID,
		ServiceVersion: b.ServiceVersion,
	})
	if err != nil {
		return fmt.Errorf("error fetching service backends: %w", err)
	}
	return nil
}

// checkPredefined identifies specific backends that are required but missing
// from the user's service.
func (b *Backends) checkPredefined() error {
	var i int
	for name, settings := range b.Setup {
		if settings.Exists {
			continue
		}

		if i > 0 {
			text.Break(b.Stdout)
		}
		i++
		text.Output(b.Stdout, "%s %s", text.Bold("Backend name:"), name)

		var defaultAddress string
		if settings.Address != "" {
			defaultAddress = fmt.Sprintf(": [%s]", settings.Address)
		}

		prompt := fmt.Sprintf("%s%s ", settings.Prompt, defaultAddress)
		if settings.Prompt == "" {
			prompt = fmt.Sprintf("Backend address%s ", defaultAddress)
		}

		addr, err := text.Input(b.Stdout, prompt, b.Stdin, b.validateAddress)
		if err != nil {
			return fmt.Errorf("error reading prompt input: %w", err)
		}
		if addr == "" {
			if settings.Address == "" {
				return fmt.Errorf("error reading prompt input: backend address is required")
			}
			addr = settings.Address
		}

		port := uint(80)
		if settings.Port > 0 {
			port = settings.Port
		}
		if !b.AcceptDefaults {
			input, err := text.Input(b.Stdout, fmt.Sprintf("Backend port number: [%d] ", port), b.Stdin)
			if err != nil {
				return fmt.Errorf("error reading prompt input: %w", err)
			}
			if input != "" {
				if i, err := strconv.Atoi(input); err != nil {
					text.Warning(b.Stdout, fmt.Sprintf("error converting prompt input, using default port number (%d)", port))
				} else {
					port = uint(i)
				}
			}
		}

		var found bool
		for _, backend := range b.Available {
			if backend.Name == name && backend.Address == addr && backend.Port == port {
				text.Break(b.Stdout)
				text.Output(b.Stdout, "We wont attempt to create the backend '%s' as it looks to already exist.", name)
				found = true
				break
			}
		}

		if !found {
			overrideHost, sslSNIHostname := b.setBackendHost(addr)
			b.Required = append(b.Required, Backend{
				Name:           name,
				Address:        addr,
				OverrideHost:   overrideHost,
				Port:           port,
				SSLSNIHostname: sslSNIHostname,
			})
		}
	}

	return nil
}

// promptForBackend issues a prompt requesting one or more Backends that will
// be created within the user's service.
func (b *Backends) promptForBackend() error {
	if b.AcceptDefaults {
		b.Required = append(b.Required, b.createOriginlessBackend())
		return nil
	}

	var i int
	for {
		addr, err := text.Input(b.Stdout, "Backend (hostname or IP address, or leave blank to stop adding backends): ", b.Stdin, b.validateAddress)
		if err != nil {
			return fmt.Errorf("error reading prompt input %w", err)
		}

		// This block short-circuits the endless prompt loop
		if addr == "" {
			if len(b.Required) == 0 {
				b.Required = append(b.Required, b.createOriginlessBackend())
			}
			return nil
		}

		port := uint(80)
		input, err := text.Input(b.Stdout, fmt.Sprintf("Backend port number: [%d] ", port), b.Stdin)
		if err != nil {
			return fmt.Errorf("error reading prompt input: %w", err)
		}
		if input != "" {
			if i, err := strconv.Atoi(input); err != nil {
				text.Warning(b.Stdout, fmt.Sprintf("error converting prompt input, using default port number (%d)", port))
			} else {
				port = uint(i)
			}
		}

		defaultName := fmt.Sprintf("backend_%d", i+1)
		name, err := text.Input(b.Stdout, fmt.Sprintf("Backend name: [%s] ", defaultName), b.Stdin)
		if err != nil {
			return fmt.Errorf("error reading prompt input %w", err)
		}
		if name == "" {
			name = defaultName
		}

		overrideHost, sslSNIHostname := b.setBackendHost(addr)
		b.Required = append(b.Required, Backend{
			Name:           name,
			Address:        addr,
			OverrideHost:   overrideHost,
			Port:           port,
			SSLSNIHostname: sslSNIHostname,
		})
	}
}

// createOriginlessBackend returns a Backend instance configured to the
// localhost settings expected of an 'originless' backend.
func (b *Backends) createOriginlessBackend() Backend {
	var backend Backend
	backend.Name = "originless"
	backend.Address = "127.0.0.1"
	backend.Port = uint(80)
	return backend
}

// setBackendHost configures the OverrideHost and SSLSNIHostname field values.
//
// By default we set the override_host and ssl_sni_hostname properties of the
// Backend object to the hostname, unless the given input is an IP.
func (b *Backends) setBackendHost(address string) (overrideHost, sslSNIHostname string) {
	if _, err := net.LookupAddr(address); err != nil {
		overrideHost = address
	}
	if overrideHost != "" {
		sslSNIHostname = overrideHost
	}
	return
}

// validateAddress checks the user entered address is a valid hostname or IP.
//
// NOTE: An empty value can be allowed because it enables the caller to
// short-circuit logic related to whether the user is prompted endlessly.
func (b *Backends) validateAddress(input string) error {
	var isHost bool
	if _, err := net.LookupHost(input); err == nil {
		isHost = true
	}
	var isAddr bool
	if _, err := net.LookupAddr(input); err == nil {
		isAddr = true
	}
	isEmpty := input == ""
	if !isEmpty && !isHost && !isAddr {
		return fmt.Errorf(`must be a valid hostname, IPv4, or IPv6 address`)
	}
	return nil
}
