package setup

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
)

// KVStores represents the service state related to kv stores defined
// within the fastly.toml [setup] configuration.
//
// NOTE: It implements the setup.Interface interface.
type KVStores struct {
	// Public
	APIClient      api.Interface
	AcceptDefaults bool
	NonInteractive bool
	Spinner        text.Spinner
	ServiceID      string
	ServiceVersion int
	Setup          map[string]*manifest.SetupKVStore
	Stdin          io.Reader
	Stdout         io.Writer

	// Private
	required []KVStore
}

// KVStore represents the configuration parameters for creating a KV Store via
// the API client.
type KVStore struct {
	Name              string
	Items             []KVStoreItem
	LinkExistingStore bool
	ExistingStoreID   string
}

// KVStoreItem represents the configuration parameters for creating KV Store
// items via the API client.
type KVStoreItem struct {
	Key   string
	Value string
	Body  fastly.LengthReader
}

// Configure prompts the user for specific values related to the service resource.
func (o *KVStores) Configure() error {
	for name, settings := range o.Setup {
		var (
			cursor         string
			existingStores []fastly.KVStore
		)

		for {
			kvs, err := o.APIClient.ListKVStores(&fastly.ListKVStoresInput{
				Cursor: cursor,
			})
			if err != nil {
				return err
			}

			if kvs != nil {
				for _, store := range kvs.Data {
					// Avoid gosec loop aliasing check
					store := store
					existingStores = append(existingStores, store)
				}
				if cur, ok := kvs.Meta["next_cursor"]; ok && cur != "" && cur != cursor {
					cursor = cur
					continue
				}
				break
			}
		}

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
					text.Warning(o.Stdout, "\nA KV Store called '%s' already exists\n\n", name)
					prompt := text.BoldYellow("Use a different store name (or leave empty to use the existing store): ")
					value, err := text.Input(o.Stdout, prompt, o.Stdin)
					if err != nil {
						return fmt.Errorf("error reading prompt input: %w", err)
					}
					text.Break(o.Stdout)
					if value == "" {
						linkExistingStore = true
					} else {
						name = value
					}
				}
			}
		}

		if !o.AcceptDefaults && !o.NonInteractive {
			text.Output(o.Stdout, "\nConfiguring kv store '%s'", name)
			if settings.Description != "" {
				text.Output(o.Stdout, settings.Description)
			}
		}

		var items []KVStoreItem

		for key, item := range settings.Items {
			if item.Value != "" && item.File != "" {
				return errors.RemediationError{
					Inner:       fmt.Errorf("invalid config: both 'value' and 'file' were set"),
					Remediation: fmt.Sprintf("Edit the [setup.kv_stores.%s.items.%s] configuration to use either 'value' or 'file', not both", name, key),
				}
			}
			promptMessage := "Value"
			dv := "example"
			if item.Value != "" {
				dv = item.Value
			}
			if item.File != "" {
				promptMessage = "File"
				dv = item.File
			}
			prompt := text.BoldYellow(fmt.Sprintf("%s: [%s] ", promptMessage, dv))

			var (
				value string
				err   error
			)

			if !o.AcceptDefaults && !o.NonInteractive {
				text.Output(o.Stdout, "\nCreate a kv store key called '%s'", key)
				if item.Description != "" {
					text.Output(o.Stdout, item.Description)
				}
				text.Break(o.Stdout)

				value, err = text.Input(o.Stdout, prompt, o.Stdin)
				if err != nil {
					return fmt.Errorf("error reading prompt input: %w", err)
				}
				text.Break(o.Stdout)
			}
			if value == "" {
				value = dv
			}

			var f *os.File
			if item.File != "" {
				abs, err := filepath.Abs(value)
				if err != nil {
					return fmt.Errorf("failed to construct absolute path for '%s': %w", item.File, err)
				}
				// G304 (CWE-22): Potential file inclusion via variable
				// Disabling as we trust the source of the variable.
				// #nosec
				f, err = os.Open(abs)
				if err != nil {
					return fmt.Errorf("failed to open file '%s': %w", abs, err)
				}
			}

			kvsi := KVStoreItem{
				Key: key,
			}
			if item.File != "" && f != nil {
				lr, err := fastly.FileLengthReader(f)
				if err != nil {
					return fmt.Errorf("failed to convert file to a LengthReader: %w", err)
				}
				kvsi.Body = lr
			} else {
				kvsi.Value = value
			}
			items = append(items, kvsi)
		}

		o.required = append(o.required, KVStore{
			Name:              name,
			Items:             items,
			LinkExistingStore: linkExistingStore,
			ExistingStoreID:   existingStoreID,
		})
	}

	return nil
}

// Create calls the relevant API to create the service resource(s).
func (o *KVStores) Create() error {
	if o.Spinner == nil {
		return errors.RemediationError{
			Inner:       fmt.Errorf("internal logic error: no spinner configured for setup.KVStores"),
			Remediation: errors.BugRemediation,
		}
	}

	for _, kvStore := range o.required {
		var (
			err   error
			store *fastly.KVStore
		)

		if kvStore.LinkExistingStore {
			err = o.Spinner.Process(fmt.Sprintf("Retrieving existing kv store '%s'", kvStore.Name), func(_ *text.SpinnerWrapper) error {
				store, err = o.APIClient.GetKVStore(&fastly.GetKVStoreInput{
					ID: kvStore.ExistingStoreID,
				})
				if err != nil {
					return fmt.Errorf("failed to get existing store '%s': %w", kvStore.Name, err)
				}
				return nil
			})
			if err != nil {
				return err
			}
		} else {
			err = o.Spinner.Process(fmt.Sprintf("Creating kv store '%s'", kvStore.Name), func(_ *text.SpinnerWrapper) error {
				store, err = o.APIClient.CreateKVStore(&fastly.CreateKVStoreInput{
					Name: kvStore.Name,
				})
				if err != nil {
					return fmt.Errorf("error creating kv store: %w", err)
				}
				return nil
			})
			if err != nil {
				return err
			}
		}

		if len(kvStore.Items) > 0 {
			for _, item := range kvStore.Items {
				err = o.Spinner.Process(fmt.Sprintf("Creating kv store key '%s'...", item.Key), func(_ *text.SpinnerWrapper) error {
					input := &fastly.InsertKVStoreKeyInput{
						ID:  store.ID,
						Key: item.Key,
					}
					if item.Body != nil {
						input.Body = item.Body
					} else {
						input.Value = item.Value
					}
					err = o.APIClient.InsertKVStoreKey(input)
					if err != nil {
						return fmt.Errorf("error creating kv store key: %w", err)
					}
					return nil
				})
				if err != nil {
					return err
				}
			}
		}

		// IMPORTANT: We need to link the KV Store to the Compute Service.
		err = o.Spinner.Process(fmt.Sprintf("Creating resource link between service and kv store '%s'...", kvStore.Name), func(_ *text.SpinnerWrapper) error {
			_, err = o.APIClient.CreateResource(&fastly.CreateResourceInput{
				ServiceID:      o.ServiceID,
				ServiceVersion: o.ServiceVersion,
				Name:           fastly.String(store.Name),
				ResourceID:     fastly.String(store.ID),
			})
			if err != nil {
				return fmt.Errorf("error creating resource link between the service '%s' and the kv store '%s': %w", o.ServiceID, store.Name, err)
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
func (o *KVStores) Predefined() bool {
	return len(o.Setup) > 0
}
