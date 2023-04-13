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
func (d *ConfigStores) Configure() error {
	for name, settings := range d.Setup {
		if !d.AcceptDefaults && !d.NonInteractive {
			text.Break(d.Stdout)
			text.Output(d.Stdout, "Configuring config store '%s'", name)
			if settings.Description != "" {
				text.Output(d.Stdout, settings.Description)
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

			if !d.AcceptDefaults && !d.NonInteractive {
				text.Break(d.Stdout)
				text.Output(d.Stdout, "Create a config store key called '%s'", key)
				if item.Description != "" {
					text.Output(d.Stdout, item.Description)
				}
				text.Break(d.Stdout)

				value, err = text.Input(d.Stdout, prompt, d.Stdin)
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

		d.required = append(d.required, ConfigStore{
			Name:  name,
			Items: items,
		})
	}

	return nil
}

// Create calls the relevant API to create the service resource(s).
func (d *ConfigStores) Create() error {
	if d.Spinner == nil {
		return errors.RemediationError{
			Inner:       fmt.Errorf("internal logic error: no spinner configured for setup.ConfigStores"),
			Remediation: errors.BugRemediation,
		}
	}

	for _, store := range d.required {
		err := d.Spinner.Start()
		if err != nil {
			return err
		}
		msg := fmt.Sprintf("Creating config store '%s'", store.Name)
		d.Spinner.Message(msg + "...")

		cs, err := d.APIClient.CreateConfigStore(&fastly.CreateConfigStoreInput{
			Name: store.Name,
		})
		if err != nil {
			d.Spinner.StopFailMessage(msg)
			err := d.Spinner.StopFail()
			if err != nil {
				return err
			}
			return fmt.Errorf("error creating config store: %w", err)
		}

		d.Spinner.StopMessage(msg)
		err = d.Spinner.Stop()
		if err != nil {
			return err
		}

		if len(store.Items) > 0 {
			for _, item := range store.Items {
				err := d.Spinner.Start()
				if err != nil {
					return err
				}
				msg := fmt.Sprintf("Creating config store item '%s'", item.Key)
				d.Spinner.Message(msg + "...")

				_, err = d.APIClient.CreateConfigStoreItem(&fastly.CreateConfigStoreItemInput{
					StoreID: cs.ID,
					Key:     item.Key,
					Value:   item.Value,
				})
				if err != nil {
					d.Spinner.StopFailMessage(msg)
					err := d.Spinner.StopFail()
					if err != nil {
						return err
					}
					return fmt.Errorf("error creating config store item: %w", err)
				}

				d.Spinner.StopMessage(msg)
				err = d.Spinner.Stop()
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// Predefined indicates if the service resource has been specified within the
// fastly.toml file using a [setup] configuration block.
func (d *ConfigStores) Predefined() bool {
	return len(d.Setup) > 0
}
