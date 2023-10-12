package setup

import (
	"errors"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/api"
	fsterrors "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
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
	Name              string
	Entries           []SecretStoreEntry
	LinkExistingStore bool
	ExistingStoreID   string
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
	var (
		cursor         string
		existingStores []fastly.SecretStore
	)

	for {
		o, err := s.APIClient.ListSecretStores(&fastly.ListSecretStoresInput{
			Cursor: cursor,
		})
		if err != nil {
			return err
		}
		if o != nil {
			existingStores = append(existingStores, o.Data...)
			if o.Meta.NextCursor != "" {
				cursor = o.Meta.NextCursor
				continue
			}
			break
		}
	}

	for name, settings := range s.Setup {
		var (
			existingStoreID   string
			linkExistingStore bool
		)

		for _, store := range existingStores {
			if store.Name == name {
				if s.AcceptDefaults || s.NonInteractive {
					linkExistingStore = true
					existingStoreID = store.ID
				} else {
					text.Warning(s.Stdout, "\nA Secret Store called '%s' already exists\n\n", name)
					prompt := text.BoldYellow("Use a different store name (or leave empty to use the existing store): ")
					value, err := text.Input(s.Stdout, prompt, s.Stdin)
					if err != nil {
						return fmt.Errorf("error reading prompt input: %w", err)
					}
					if value == "" {
						linkExistingStore = true
						existingStoreID = store.ID
					} else {
						name = value
					}
				}
			}
		}

		if !s.AcceptDefaults && !s.NonInteractive {
			text.Output(s.Stdout, "\nConfiguring Secret Store '%s'", name)
			if settings.Description != "" {
				text.Output(s.Stdout, settings.Description)
			}
		}

		store := SecretStore{
			Name:              name,
			Entries:           make([]SecretStoreEntry, 0, len(settings.Entries)),
			LinkExistingStore: linkExistingStore,
			ExistingStoreID:   existingStoreID,
		}

		for key, entry := range settings.Entries {
			var (
				value string
				err   error
			)

			if !s.AcceptDefaults && !s.NonInteractive {
				text.Output(s.Stdout, "\nCreate a Secret Store entry called '%s'", key)
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
			Inner:       fmt.Errorf("internal logic error: no spinner configured for setup.SecretStores"),
			Remediation: fsterrors.BugRemediation,
		}
	}

	for _, secretStore := range s.required {
		var (
			err   error
			store *fastly.SecretStore
		)

		if secretStore.LinkExistingStore {
			err = s.Spinner.Process(fmt.Sprintf("Retrieving existing Secret Store '%s'", secretStore.Name), func(_ *text.SpinnerWrapper) error {
				store, err = s.APIClient.GetSecretStore(&fastly.GetSecretStoreInput{
					ID: secretStore.ExistingStoreID,
				})
				if err != nil {
					return fmt.Errorf("failed to get existing store '%s': %w", secretStore.Name, err)
				}
				return nil
			})
			if err != nil {
				return err
			}
		} else {
			err = s.Spinner.Process(fmt.Sprintf("Creating Secret Store '%s'", secretStore.Name), func(_ *text.SpinnerWrapper) error {
				store, err = s.APIClient.CreateSecretStore(&fastly.CreateSecretStoreInput{
					Name: secretStore.Name,
				})
				if err != nil {
					return fmt.Errorf("error creating Secret Store %q: %w", secretStore.Name, err)
				}
				return nil
			})
			if err != nil {
				return err
			}
		}

		for _, entry := range secretStore.Entries {
			err = s.Spinner.Process(fmt.Sprintf("Creating Secret Store entry '%s'...", entry.Name), func(_ *text.SpinnerWrapper) error {
				_, err = s.APIClient.CreateSecret(&fastly.CreateSecretInput{
					ID:     store.ID,
					Name:   entry.Name,
					Secret: []byte(entry.Secret),
				})
				if err != nil {
					return fmt.Errorf("error creating Secret Store entry %q: %w", entry.Name, err)
				}
				return nil
			})
			if err != nil {
				return err
			}
		}

		err = s.Spinner.Process(fmt.Sprintf("Creating resource link between service and Secret Store '%s'...", store.Name), func(_ *text.SpinnerWrapper) error {
			// We need to link the secret store to the C@E Service, otherwise the service
			// will not have access to the store.
			_, err = s.APIClient.CreateResource(&fastly.CreateResourceInput{
				ServiceID:      s.ServiceID,
				ServiceVersion: s.ServiceVersion,
				Name:           fastly.String(store.Name),
				ResourceID:     fastly.String(store.ID),
			})
			if err != nil {
				return fmt.Errorf("error creating resource link between the service %q and the Secret Store %q: %w", s.ServiceID, store.Name, err)
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	return nil
}
