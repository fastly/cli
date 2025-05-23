package setup

import (
	"fmt"
	"io"
	"net"
	"strconv"

	"github.com/fastly/go-fastly/v10/fastly"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/commands/backend"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
)

// Backends represents the service state related to backends defined within the
// fastly.toml [setup] configuration.
//
// NOTE: It implements the setup.Interface interface.
type Backends struct {
	// Public
	APIClient      api.Interface
	AcceptDefaults bool
	NonInteractive bool
	Spinner        text.Spinner
	ServiceID      string
	ServiceVersion int
	Setup          map[string]*manifest.SetupBackend
	Stdin          io.Reader
	Stdout         io.Writer

	// Private
	required []Backend
}

// Backend represents the configuration parameters for creating a backend via
// the API client.
type Backend struct {
	Address         string
	Name            string
	OverrideHost    string
	Port            int
	SSLCertHostname string
	SSLSNIHostname  string
}

// Configure prompts the user for specific values related to the service resource.
func (b *Backends) Configure() error {
	if b.Predefined() {
		return b.checkPredefined()
	}
	return b.promptForBackend()
}

// Create calls the relevant API to create the service resource(s).
func (b *Backends) Create() error {
	if b.Spinner == nil {
		return errors.RemediationError{
			Inner:       fmt.Errorf("internal logic error: no spinner configured for setup.Backends"),
			Remediation: errors.BugRemediation,
		}
	}

	for _, bk := range b.required {
		// Avoids range-loop variable issue (i.e. var is reused across iterations).
		bk := bk

		msg := fmt.Sprintf("Creating backend '%s' (host: %s, port: %d)", bk.Name, bk.Address, bk.Port)

		if !b.isOriginless() {
			err := b.Spinner.Start()
			if err != nil {
				return err
			}
			b.Spinner.Message(msg + "...")
		}

		opts := &fastly.CreateBackendInput{
			ServiceID:      b.ServiceID,
			ServiceVersion: b.ServiceVersion,
			Name:           &bk.Name,
			Address:        &bk.Address,
			Port:           &bk.Port,
		}

		if bk.OverrideHost != "" {
			opts.OverrideHost = &bk.OverrideHost
		}
		if bk.SSLCertHostname != "" {
			opts.SSLCertHostname = &bk.SSLCertHostname
		}
		if bk.SSLSNIHostname != "" {
			opts.SSLSNIHostname = &bk.SSLSNIHostname
		}

		_, err := b.APIClient.CreateBackend(opts)
		if err != nil {
			if !b.isOriginless() {
				err = fmt.Errorf("error creating backend: %w", err)
				b.Spinner.StopFailMessage(msg)
				spinErr := b.Spinner.StopFail()
				if spinErr != nil {
					return fmt.Errorf(text.SpinnerErrWrapper, spinErr, err)
				}
			}
			return fmt.Errorf("error configuring the service: %w", err)
		}

		if !b.isOriginless() {
			b.Spinner.StopMessage(msg)
			err = b.Spinner.Stop()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Predefined indicates if the service resource has been specified within the
// fastly.toml file using a [setup] configuration block.
func (b *Backends) Predefined() bool {
	return len(b.Setup) > 0
}

// isOriginless indicates if the required backend is originless.
func (b *Backends) isOriginless() bool {
	return len(b.required) == 1 && b.required[0].Name == "originless" && b.required[0].Address == "127.0.0.1"
}

// checkPredefined identifies specific backends that are required but missing
// from the user's service (based on the [setup.backends] configuration).
func (b *Backends) checkPredefined() error {
	var i int
	for name, settings := range b.Setup {
		if !b.AcceptDefaults && !b.NonInteractive {
			if i > 0 {
				text.Break(b.Stdout)
			}
			i++
			text.Output(b.Stdout, "Configure a backend called '%s'", name)
			if settings.Description != "" {
				text.Output(b.Stdout, settings.Description)
			}
			text.Break(b.Stdout)
		}

		var (
			addr string
			err  error
		)

		defaultAddress := "127.0.0.1"
		if settings.Address != "" {
			defaultAddress = settings.Address
		}

		prompt := text.Prompt(fmt.Sprintf("Hostname or IP address: [%s] ", defaultAddress))

		if !b.AcceptDefaults && !b.NonInteractive {
			addr, err = text.Input(b.Stdout, prompt, b.Stdin, b.validateAddress)
			if err != nil {
				return fmt.Errorf("error reading prompt input: %w", err)
			}
		}
		if addr == "" {
			addr = defaultAddress
		}

		port := int(443)
		if settings.Port > 0 {
			port = settings.Port
		}
		if !b.AcceptDefaults && !b.NonInteractive {
			input, err := text.Input(b.Stdout, text.Prompt(fmt.Sprintf("Port: [%d] ", port)), b.Stdin)
			if err != nil {
				return fmt.Errorf("error reading prompt input: %w", err)
			}
			if input != "" {
				if i, err := strconv.Atoi(input); err != nil {
					text.Warning(b.Stdout, fmt.Sprintf("error converting prompt input, using default port number (%d)\n\n", port))
				} else {
					port = i
				}
			}
		}

		overrideHost, sslSNIHostname, sslCertHostname := backend.SetBackendHostDefaults(addr)
		b.required = append(b.required, Backend{
			Address:         addr,
			Name:            name,
			OverrideHost:    overrideHost,
			Port:            port,
			SSLCertHostname: sslCertHostname,
			SSLSNIHostname:  sslSNIHostname,
		})
	}

	return nil
}

// promptForBackend issues a prompt requesting one or more Backends that will
// be created within the user's service.
func (b *Backends) promptForBackend() error {
	if b.AcceptDefaults || b.NonInteractive {
		b.required = append(b.required, b.createOriginlessBackend())
		return nil
	}

	var i int
	for {
		if i > 0 {
			text.Break(b.Stdout)
		}
		i++

		addr, err := text.Input(b.Stdout, text.Prompt("Backend (hostname or IP address, or leave blank to stop adding backends): "), b.Stdin, b.validateAddress)
		if err != nil {
			return fmt.Errorf("error reading prompt input %w", err)
		}

		// This block short-circuits the endless prompt loop
		if addr == "" {
			if len(b.required) == 0 {
				b.required = append(b.required, b.createOriginlessBackend())
			}
			return nil
		}

		port := int(443)
		input, err := text.Input(b.Stdout, text.Prompt(fmt.Sprintf("Backend port number: [%d] ", port)), b.Stdin)
		if err != nil {
			return fmt.Errorf("error reading prompt input: %w", err)
		}
		if input != "" {
			if portnumber, err := strconv.Atoi(input); err != nil {
				text.Warning(b.Stdout, fmt.Sprintf("error converting prompt input, using default port number (%d)\n\n", port))
			} else {
				port = portnumber
			}
		}

		defaultName := fmt.Sprintf("backend_%d", i)
		name, err := text.Input(b.Stdout, text.Prompt(fmt.Sprintf("Backend name: [%s] ", defaultName)), b.Stdin)
		if err != nil {
			return fmt.Errorf("error reading prompt input %w", err)
		}
		if name == "" {
			name = defaultName
		}

		overrideHost, sslSNIHostname, sslCertHostname := backend.SetBackendHostDefaults(addr)
		b.required = append(b.required, Backend{
			Address:         addr,
			Name:            name,
			OverrideHost:    overrideHost,
			Port:            port,
			SSLCertHostname: sslCertHostname,
			SSLSNIHostname:  sslSNIHostname,
		})
	}
}

// createOriginlessBackend returns a Backend instance configured to the
// localhost settings expected of an 'originless' backend.
func (b *Backends) createOriginlessBackend() Backend {
	var bk Backend
	bk.Name = "originless"
	bk.Address = "127.0.0.1"
	bk.Port = int(80)
	return bk
}

// validateAddress checks the user entered address is a valid hostname or IP.
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
