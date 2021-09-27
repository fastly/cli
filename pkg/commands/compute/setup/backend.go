package setup

import (
	"fmt"
	"io"
	"net"
	"strconv"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/commands/compute/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/cli/pkg/undo"
	"github.com/fastly/go-fastly/v5/fastly"
)

// Backends represents the service state related to backends defined within the
// fastly.toml [setup] configuration.
type Backends struct {
	Available      []*fastly.Backend
	APIClient      api.Interface
	Progress       text.Progress
	Required       []Backend
	ServiceID      string
	ServiceVersion int
	Setup          map[string]manifest.SetupBackend
	AcceptDefaults bool
	Stdin          io.Reader
	Stdout         io.Writer
	UndoStack      undo.Stacker
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

// Predefined implements the setup interface.
func (b *Backends) Predefined() bool {
	return len(b.Setup) > 0
}

// Missing implements the setup interface.
//
// NOTE: We return 'false' if the user has ended up with an originless backend.
// This is because we don't want to reveal the fact a backend is still
// currently required even when dealing with the C@E platform.
func (b *Backends) Missing() bool {
	if b.isOriginless() {
		return false
	}
	if len(b.Required) > 0 {
		return true
	}
	return false
}

// isOriginless indicates if the required backend is originless.
func (b *Backends) isOriginless() bool {
	return len(b.Required) == 1 && b.Required[0].Name == "originless" && b.Required[0].Address == "127.0.0.1"
}

// Configure implements the setup interface.
func (b *Backends) Configure() error {
	//

	if b.Predefined() {
		return b.checkPredefined()
	} else {
		return b.promptForBackend()
	}
}

// checkPredefined identifies required backends that are missing from the
// user's service.
func (b *Backends) checkPredefined() (err error) {
	for name, settings := range b.Setup {
		prompt := settings.Prompt
		if settings.Prompt == "" {
			prompt = fmt.Sprintf("Backend for '%s'", name)
		}

		if settings.Address == "" {
			settings.Address, err = text.Input(b.Stdout, prompt, b.Stdin, b.validateAddress)
			if err != nil || settings.Address == "" {
				return fmt.Errorf("error reading prompt input: %w", err)
			}
		}

		if settings.Port == 0 {
			port := uint(80)

			if !b.AcceptDefaults {
				input, err := text.Input(b.Stdout, fmt.Sprintf("Backend port number: [%d]", port), b.Stdin)
				if err != nil {
					return fmt.Errorf("error reading prompt input: %w", err)
				}
				if i, err := strconv.Atoi(input); err != nil {
					text.Warning(b.Stdout, fmt.Sprintf("error converting prompt input, using default port number (%d)", port))
				} else {
					port = uint(i)
				}
			}

			settings.Port = port
		}
		fmt.Printf("\n\nPredefined settings updated:\n%+v\n\n", settings)

		var found bool
		for _, backend := range b.Available {
			if backend.Name == name && backend.Address == settings.Address && backend.Port == settings.Port {
				found = true
				break
			}
		}

		if !found {
			overrideHost, sslSNIHostname := b.setBackendHost(settings.Address)
			b.Required = append(b.Required, Backend{
				Name:           name,
				Address:        settings.Address,
				OverrideHost:   overrideHost,
				Port:           settings.Port,
				SSLSNIHostname: sslSNIHostname,
			})
		}
	}

	fmt.Printf("\n\nPredefined b.Required:\n%+v\n\n", b.Required)
	return nil
}

// promptForBackend issues a prompt requesting one or more Backends.
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
		input, err := text.Input(b.Stdout, fmt.Sprintf("Backend port number: [%d]", port), b.Stdin)
		if err != nil {
			return fmt.Errorf("error reading prompt input: %w", err)
		}
		if i, err := strconv.Atoi(input); err != nil {
			text.Warning(b.Stdout, fmt.Sprintf("error converting prompt input, using default port number (%d)", port))
		} else {
			port = uint(i)
		}

		defaultName := fmt.Sprintf("backend_%d", i+1)
		name, err := text.Input(b.Stdout, fmt.Sprintf("Backend name: [%s]", defaultName), b.Stdin)
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
// NOTE: An empty value is allowed because it allows the caller to
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

// Create implements the setup interface.
func (b *Backends) Create() error {
	for _, backend := range b.Required {
		if !b.isOriginless() {
			b.Progress.Step(fmt.Sprintf("Creating backend '%s' (host: %s, port: %d)...", backend.Name, backend.Address, backend.Port))
		}

		b.UndoStack.Push(func() error {
			return b.APIClient.DeleteBackend(&fastly.DeleteBackendInput{
				ServiceID:      b.ServiceID,
				ServiceVersion: b.ServiceVersion,
				Name:           backend.Address,
			})
		})

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
			if b.isOriginless() {
				return fmt.Errorf("error configuring the service: %w", err)
			}
			return fmt.Errorf("error creating backend: %w", err)
		}
	}
	return nil
}
