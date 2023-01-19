package setup

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v7/fastly"
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
	Progress       text.Progress
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
func (d *ObjectStores) Configure() error {
	for name, settings := range d.Setup {
		if !d.AcceptDefaults && !d.NonInteractive {
			text.Break(d.Stdout)
			text.Output(d.Stdout, "Configuring object store '%s'", name)
			if settings.Description != "" {
				text.Output(d.Stdout, settings.Description)
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

			if !d.AcceptDefaults && !d.NonInteractive {
				text.Break(d.Stdout)
				text.Output(d.Stdout, "Create an object store key called '%s'", key)
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

			items = append(items, ObjectStoreItem{
				Key:   key,
				Value: value,
			})
		}

		d.required = append(d.required, ObjectStore{
			Name:  name,
			Items: items,
		})
	}

	return nil
}

// Create calls the relevant API to create the service resource(s).
func (d *ObjectStores) Create() error {
	if d.Progress == nil {
		return errors.RemediationError{
			Inner:       fmt.Errorf("internal logic error: no text.Progress configured for setup.ObjectStores"),
			Remediation: errors.BugRemediation,
		}
	}

	for _, objectStore := range d.required {
		d.Progress.Step(fmt.Sprintf("Creating object store '%s'...", objectStore.Name))

		store, err := d.APIClient.CreateObjectStore(&fastly.CreateObjectStoreInput{
			Name: objectStore.Name,
		})
		if err != nil {
			d.Progress.Fail()
			return fmt.Errorf("error creating object store: %w", err)
		}

		if len(objectStore.Items) > 0 {
			for _, item := range objectStore.Items {
				d.Progress.Step(fmt.Sprintf("Creating object store key '%s'...", item.Key))

				err := d.APIClient.InsertObjectStoreKey(&fastly.InsertObjectStoreKeyInput{
					ID:    store.ID,
					Key:   item.Key,
					Value: item.Value,
				})
				if err != nil {
					d.Progress.Fail()
					return fmt.Errorf("error creating object store key: %w", err)
				}
			}
		}

		d.Progress.Step(fmt.Sprintf("Creating resource link between service and object store '%s'...", objectStore.Name))

		// IMPORTANT: We need to link the object store to the C@E Service.
		_, err = d.APIClient.CreateResource(&fastly.CreateResourceInput{
			ServiceID:      d.ServiceID,
			ServiceVersion: d.ServiceVersion,
			Name:           fastly.String(store.Name),
			ResourceID:     fastly.String(store.ID),
		})
		if err != nil {
			d.Progress.Fail()
			return fmt.Errorf("error creating resource link between the service '%s' and the object store '%s': %w", d.ServiceID, store.Name, err)
		}
	}

	return nil
}

// Predefined indicates if the service resource has been specified within the
// fastly.toml file using a [setup] configuration block.
func (d *ObjectStores) Predefined() bool {
	return len(d.Setup) > 0
}
