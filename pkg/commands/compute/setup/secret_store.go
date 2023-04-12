package setup

import (
	"errors"
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/api"
	fsterrors "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v8/fastly"
)

// SecretStores represents the service state related to secret stores defined
// within the fastly.toml [setup] configuration.
//
// NOTE: It implements the setup.Interface interface.
type SecretStores struct {
	// Public
	APIClient      api.Interface
	AcceptDefaults bool
	NonInteractive bool
	Spinner        text.Spinner
	ServiceID      string
	ServiceVersion int
	Setup          map[string]*manifest.SetupSecretStore
	Stdin          io.Reader
	Stdout         io.Writer

	// Private
	required []SecretStore
}

// SecretStore represents the configuration parameters for creating a
// secret store via the API client.
type SecretStore struct {
	Name    string
	Entries []SecretStoreEntry
}

// SecretStoreEntry represents the configuration parameters for creating
// secret store items via the API client.
type SecretStoreEntry struct {
	Name   string
	Secret string
}

// Predefined indicates if the service resource has been specified within the
// fastly.toml file using a [setup] configuration block.
func (s *SecretStores) Predefined() bool {
	return len(s.Setup) > 0
}

// Configure prompts the user for specific values related to the service resource.
func (s *SecretStores) Configure() error {
	for name, settings := range s.Setup {
		if !s.AcceptDefaults && !s.NonInteractive {
			text.Break(s.Stdout)
			text.Output(s.Stdout, "Configuring secret store '%s'", name)
			if settings.Description != "" {
				text.Output(s.Stdout, settings.Description)
			}
		}

		store := SecretStore{
			Name:    name,
			Entries: make([]SecretStoreEntry, 0, len(settings.Entries)),
		}

		for key, entry := range settings.Entries {
			var (
				value string
				err   error
			)

			if !s.AcceptDefaults && !s.NonInteractive {
				text.Break(s.Stdout)
				text.Output(s.Stdout, "Create a secret store entry called '%s'", key)
				if entry.Description != "" {
					text.Output(s.Stdout, entry.Description)
				}
				text.Break(s.Stdout)

				prompt := text.BoldYellow("Value: ")
				value, err = text.InputSecure(s.Stdout, prompt, s.Stdin)
				if err != nil {
					return fmt.Errorf("error reading prompt input: %w", err)
				}
			}

			if value == "" {
				return errors.New("value cannot be blank")
			}

			store.Entries = append(store.Entries, SecretStoreEntry{
				Name:   key,
				Secret: value,
			})
		}

		s.required = append(s.required, store)
	}

	return nil
}

// Create calls the relevant API to create the service resource(s).
func (s *SecretStores) Create() error {
	if s.Spinner == nil {
		return fsterrors.RemediationError{
			Inner:       fmt.Errorf("internal logic error: no text.Progress configured for setup.SecretStores"),
			Remediation: fsterrors.BugRemediation,
		}
	}

	for _, secretStore := range s.required {
		if err := s.Spinner.Start(); err != nil {
			return err
		}
		msg := fmt.Sprintf("Creating secret store '%s'", secretStore.Name)
		s.Spinner.Message(msg + "...")

		store, err := s.APIClient.CreateSecretStore(&fastly.CreateSecretStoreInput{
			Name: secretStore.Name,
		})
		if err != nil {
			s.Spinner.StopFailMessage(msg)
			if serr := s.Spinner.StopFail(); serr != nil {
				return serr
			}
			return fmt.Errorf("error creating secret store %q: %w", secretStore.Name, err)
		}
		s.Spinner.StopMessage(msg)
		if err = s.Spinner.Stop(); err != nil {
			return err
		}

		for _, entry := range secretStore.Entries {
			if err = s.Spinner.Start(); err != nil {
				return err
			}
			msg = fmt.Sprintf("Creating secret store entry '%s'...", entry.Name)
			s.Spinner.Message(msg)

			_, err = s.APIClient.CreateSecret(&fastly.CreateSecretInput{
				ID:     store.ID,
				Name:   entry.Name,
				Secret: []byte(entry.Secret),
			})
			if err != nil {
				s.Spinner.StopFailMessage(msg)
				if serr := s.Spinner.StopFail(); serr != nil {
					return serr
				}
				return fmt.Errorf("error creating secret store entry %q: %w", entry.Name, err)
			}

			s.Spinner.StopMessage(msg)
			if err = s.Spinner.Stop(); err != nil {
				return err
			}
		}

		if err = s.Spinner.Start(); err != nil {
			return err
		}
		msg = fmt.Sprintf("Creating resource link between service and secret store '%s'...", store.Name)
		s.Spinner.Message(msg)

		// We need to link the secret store to the C@E Service, otherwise the service
		// will not have access to the store.
		_, err = s.APIClient.CreateResource(&fastly.CreateResourceInput{
			ServiceID:      s.ServiceID,
			ServiceVersion: s.ServiceVersion,
			Name:           fastly.String(store.Name),
			ResourceID:     fastly.String(store.ID),
		})
		if err != nil {
			s.Spinner.StopFailMessage(msg)
			if serr := s.Spinner.StopFail(); serr != nil {
				return serr
			}
			return fmt.Errorf("error creating resource link between the service %q and the secret store %q: %w", s.ServiceID, store.Name, err)
		}

		s.Spinner.StopMessage(msg)
		if err = s.Spinner.Stop(); err != nil {
			return err
		}
	}

	return nil
}
