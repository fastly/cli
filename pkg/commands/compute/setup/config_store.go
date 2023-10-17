package setup

import (
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
)

// ConfigStores represents the service state related to config stores defined
// within the fastly.toml [setup] configuration.
//
// NOTE: It implements the setup.Interface interface.
type ConfigStores struct {
	// Public
	APIClient      api.Interface
	AcceptDefaults bool
	NonInteractive bool
	Spinner        text.Spinner
	ServiceID      string
	ServiceVersion int
	Setup          map[string]*manifest.SetupConfigStore
	Stdin          io.Reader
	Stdout         io.Writer

	// Private
	required []ConfigStore
}

// ConfigStore represents the configuration parameters for creating a config
// store via the API client.
type ConfigStore struct {
	Name              string
	Items             []ConfigStoreItem
	LinkExistingStore bool
	ExistingStoreID   string
}

// ConfigStoreItem represents the configuration parameters for creating config
// store items via the API client.
type ConfigStoreItem struct {
	Key   string
	Value string
}

// Configure prompts the user for specific values related to the service resource.
func (o *ConfigStores) Configure() error {
	existingStores, err := o.APIClient.ListConfigStores()
	if err != nil {
		return err
	}

	for name, settings := range o.Setup {
		var (
			existingStoreID   string
			linkExistingStore bool
		)

		for _, store := range existingStores {
			if store.Name == name {
				if o.AcceptDefaults || o.NonInteractive {
					linkExistingStore = true
					existingStoreID = store.ID
				} else {
					text.Warning(o.Stdout, "\nA Config Store called '%s' already exists. If you use this store, then this implies that any keys defined in your setup configuration will either be newly created or will update an existing one. To avoid updating an existing key, then stop the command now and edit the setup configuration before re-running the deployment process\n\n", name)
					prompt := text.Prompt("Use a different store name (or leave empty to use the existing store): ")
					value, err := text.Input(o.Stdout, prompt, o.Stdin)
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

		if !o.AcceptDefaults && !o.NonInteractive {
			text.Output(o.Stdout, "\nConfiguring config store '%s'", name)
			if settings.Description != "" {
				text.Output(o.Stdout, settings.Description)
			}
		}

		var items []ConfigStoreItem

		for key, item := range settings.Items {
			dv := "example"
			if item.Value != "" {
				dv = item.Value
			}
			prompt := text.Prompt(fmt.Sprintf("Value: [%s] ", dv))

			var (
				value string
				err   error
			)

			if !o.AcceptDefaults && !o.NonInteractive {
				text.Output(o.Stdout, "\nCreate a config store key called '%s'", key)
				if item.Description != "" {
					text.Output(o.Stdout, item.Description)
				}
				text.Break(o.Stdout)

				value, err = text.Input(o.Stdout, prompt, o.Stdin)
				if err != nil {
					return fmt.Errorf("error reading prompt input: %w", err)
				}
			}

			if value == "" {
				value = dv
			}

			items = append(items, ConfigStoreItem{
				Key:   key,
				Value: value,
			})
		}

		o.required = append(o.required, ConfigStore{
			Name:              name,
			Items:             items,
			LinkExistingStore: linkExistingStore,
			ExistingStoreID:   existingStoreID,
		})
	}

	return nil
}

// Create calls the relevant API to create the service resource(s).
func (o *ConfigStores) Create() error {
	if o.Spinner == nil {
		return errors.RemediationError{
			Inner:       fmt.Errorf("internal logic error: no spinner configured for setup.ConfigStores"),
			Remediation: errors.BugRemediation,
		}
	}

	for _, configStore := range o.required {
		var (
			err error
			cs  *fastly.ConfigStore
		)

		if configStore.LinkExistingStore {
			err = o.Spinner.Process(fmt.Sprintf("Retrieving existing Config Store '%s'", configStore.Name), func(_ *text.SpinnerWrapper) error {
				cs, err = o.APIClient.GetConfigStore(&fastly.GetConfigStoreInput{
					ID: configStore.ExistingStoreID,
				})
				if err != nil {
					return fmt.Errorf("failed to get existing store '%s': %w", configStore.Name, err)
				}
				return nil
			})
			if err != nil {
				return err
			}
		} else {
			err = o.Spinner.Process(fmt.Sprintf("Creating config store '%s'", configStore.Name), func(_ *text.SpinnerWrapper) error {
				cs, err = o.APIClient.CreateConfigStore(&fastly.CreateConfigStoreInput{
					Name: configStore.Name,
				})
				if err != nil {
					return fmt.Errorf("error creating config store: %w", err)
				}
				return nil
			})
			if err != nil {
				return err
			}
		}

		if len(configStore.Items) > 0 {
			for _, item := range configStore.Items {
				err = o.Spinner.Process(fmt.Sprintf("Creating config store item '%s'", item.Key), func(_ *text.SpinnerWrapper) error {
					_, err = o.APIClient.UpdateConfigStoreItem(&fastly.UpdateConfigStoreItemInput{
						Upsert:  true, // Use upsert to avoid conflicts when reusing a starter kit.
						StoreID: cs.ID,
						Key:     item.Key,
						Value:   item.Value,
					})
					if err != nil {
						return fmt.Errorf("error creating config store item: %w", err)
					}
					return nil
				})
				if err != nil {
					return err
				}
			}
		}

		// IMPORTANT: We need to link the config store to the Compute Service.
		err = o.Spinner.Process(fmt.Sprintf("Creating resource link between service and config store '%s'...", cs.Name), func(_ *text.SpinnerWrapper) error {
			_, err = o.APIClient.CreateResource(&fastly.CreateResourceInput{
				ServiceID:      o.ServiceID,
				ServiceVersion: o.ServiceVersion,
				Name:           fastly.String(cs.Name),
				ResourceID:     fastly.String(cs.ID),
			})
			if err != nil {
				return fmt.Errorf("error creating resource link between the service '%s' and the config store '%s': %w", o.ServiceID, configStore.Name, err)
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// Predefined indicates if the service resource has been specified within the
// fastly.toml file using a [setup] configuration block.
func (o *ConfigStores) Predefined() bool {
	return len(o.Setup) > 0
}
