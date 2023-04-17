package setup

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v8/fastly"
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
	Name  string
	Items []ConfigStoreItem
}

// ConfigStoreItem represents the configuration parameters for creating config
// store items via the API client.
type ConfigStoreItem struct {
	Key   string
	Value string
}

// Configure prompts the user for specific values related to the service resource.
func (o *ConfigStores) Configure() error {
	for name, settings := range o.Setup {
		if !o.AcceptDefaults && !o.NonInteractive {
			text.Break(o.Stdout)
			text.Output(o.Stdout, "Configuring config store '%s'", name)
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
			prompt := text.BoldYellow(fmt.Sprintf("Value: [%s] ", dv))

			var (
				value string
				err   error
			)

			if !o.AcceptDefaults && !o.NonInteractive {
				text.Break(o.Stdout)
				text.Output(o.Stdout, "Create a config store key called '%s'", key)
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
			Name:  name,
			Items: items,
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

	for _, store := range o.required {
		err := o.Spinner.Start()
		if err != nil {
			return err
		}
		msg := fmt.Sprintf("Creating config store '%s'", store.Name)
		o.Spinner.Message(msg + "...")

		cs, err := o.APIClient.CreateConfigStore(&fastly.CreateConfigStoreInput{
			Name: store.Name,
		})
		if err != nil {
			o.Spinner.StopFailMessage(msg)
			err := o.Spinner.StopFail()
			if err != nil {
				return err
			}
			return fmt.Errorf("error creating config store: %w", err)
		}

		o.Spinner.StopMessage(msg)
		err = o.Spinner.Stop()
		if err != nil {
			return err
		}

		if len(store.Items) > 0 {
			for _, item := range store.Items {
				err := o.Spinner.Start()
				if err != nil {
					return err
				}
				msg := fmt.Sprintf("Creating config store item '%s'", item.Key)
				o.Spinner.Message(msg + "...")

				_, err = o.APIClient.CreateConfigStoreItem(&fastly.CreateConfigStoreItemInput{
					StoreID: cs.ID,
					Key:     item.Key,
					Value:   item.Value,
				})
				if err != nil {
					o.Spinner.StopFailMessage(msg)
					err := o.Spinner.StopFail()
					if err != nil {
						return err
					}
					return fmt.Errorf("error creating config store item: %w", err)
				}

				o.Spinner.StopMessage(msg)
				err = o.Spinner.Stop()
				if err != nil {
					return err
				}
			}
		}

		err = o.Spinner.Start()
		if err != nil {
			return err
		}
		msg = fmt.Sprintf("Creating resource link between service and config store '%s'...", cs.Name)
		o.Spinner.Message(msg)

		// IMPORTANT: We need to link the config store to the C@E Service.
		_, err = o.APIClient.CreateResource(&fastly.CreateResourceInput{
			ServiceID:      o.ServiceID,
			ServiceVersion: o.ServiceVersion,
			Name:           fastly.String(cs.Name),
			ResourceID:     fastly.String(cs.ID),
		})
		if err != nil {
			o.Spinner.StopFailMessage(msg)
			err := o.Spinner.StopFail()
			if err != nil {
				return err
			}
			return fmt.Errorf("error creating resource link between the service '%s' and the config store '%s': %w", o.ServiceID, store.Name, err)
		}

		o.Spinner.StopMessage(msg)
		err = o.Spinner.Stop()
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
