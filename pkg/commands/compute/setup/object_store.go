package setup

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v7/fastly"
	"github.com/theckman/yacspin"
)

// ObjectStores represents the service state related to object stores defined
// within the fastly.toml [setup] configuration.
//
// NOTE: It implements the setup.Interface interface.
type ObjectStores struct {
	// Public
	APIClient      api.Interface
	AcceptDefaults bool
	NonInteractive bool
	Spinner        *yacspin.Spinner
	ServiceID      string
	ServiceVersion int
	Setup          map[string]*manifest.SetupObjectStore
	Stdin          io.Reader
	Stdout         io.Writer

	// Private
	required []ObjectStore
}

// ObjectStore represents the configuration parameters for creating an
// object store via the API client.
type ObjectStore struct {
	Name  string
	Items []ObjectStoreItem
}

// ObjectStoreItem represents the configuration parameters for creating
// object store items via the API client.
type ObjectStoreItem struct {
	Key   string
	Value string
}

// Configure prompts the user for specific values related to the service resource.
func (o *ObjectStores) Configure() error {
	for name, settings := range o.Setup {
		if !o.AcceptDefaults && !o.NonInteractive {
			text.Break(o.Stdout)
			text.Output(o.Stdout, "Configuring object store '%s'", name)
			if settings.Description != "" {
				text.Output(o.Stdout, settings.Description)
			}
		}

		var items []ObjectStoreItem

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
				text.Output(o.Stdout, "Create an object store key called '%s'", key)
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

			items = append(items, ObjectStoreItem{
				Key:   key,
				Value: value,
			})
		}

		o.required = append(o.required, ObjectStore{
			Name:  name,
			Items: items,
		})
	}

	return nil
}

// Create calls the relevant API to create the service resource(s).
func (o *ObjectStores) Create() error {
	if o.Spinner == nil {
		return errors.RemediationError{
			Inner:       fmt.Errorf("internal logic error: no text.Progress configured for setup.ObjectStores"),
			Remediation: errors.BugRemediation,
		}
	}

	for _, objectStore := range o.required {
		err := o.Spinner.Start()
		if err != nil {
			return err
		}
		msg := fmt.Sprintf("Creating object store '%s'...", objectStore.Name)
		o.Spinner.Message(msg)

		store, err := o.APIClient.CreateObjectStore(&fastly.CreateObjectStoreInput{
			Name: objectStore.Name,
		})
		if err != nil {
			o.Spinner.StopFailMessage(msg)
			err := o.Spinner.StopFail()
			if err != nil {
				return err
			}
			return fmt.Errorf("error creating object store: %w", err)
		}

		o.Spinner.StopMessage(msg)
		err = o.Spinner.Stop()
		if err != nil {
			return err
		}

		if len(objectStore.Items) > 0 {
			for _, item := range objectStore.Items {
				err := o.Spinner.Start()
				if err != nil {
					return err
				}
				msg := fmt.Sprintf("Creating object store key '%s'...", item.Key)
				o.Spinner.Message(msg)

				err = o.APIClient.InsertObjectStoreKey(&fastly.InsertObjectStoreKeyInput{
					ID:    store.ID,
					Key:   item.Key,
					Value: item.Value,
				})
				if err != nil {
					o.Spinner.StopFailMessage(msg)
					err := o.Spinner.StopFail()
					if err != nil {
						return err
					}
					return fmt.Errorf("error creating object store key: %w", err)
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
		msg = fmt.Sprintf("Creating resource link between service and object store '%s'...", objectStore.Name)
		o.Spinner.Message(msg)

		// IMPORTANT: We need to link the object store to the C@E Service.
		_, err = o.APIClient.CreateResource(&fastly.CreateResourceInput{
			ServiceID:      o.ServiceID,
			ServiceVersion: o.ServiceVersion,
			Name:           fastly.String(store.Name),
			ResourceID:     fastly.String(store.ID),
		})
		if err != nil {
			o.Spinner.StopFailMessage(msg)
			err := o.Spinner.StopFail()
			if err != nil {
				return err
			}
			return fmt.Errorf("error creating resource link between the service '%s' and the object store '%s': %w", o.ServiceID, store.Name, err)
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
func (o *ObjectStores) Predefined() bool {
	return len(o.Setup) > 0
}
