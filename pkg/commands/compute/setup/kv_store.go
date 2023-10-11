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

// KVStore represents the configuration parameters for creating an
// kv store via the API client.
type KVStore struct {
	Name  string
	Items []KVStoreItem
}

// KVStoreItem represents the configuration parameters for creating
// kv store items via the API client.
type KVStoreItem struct {
	Key   string
	Value string
	Body  fastly.LengthReader
}

// Configure prompts the user for specific values related to the service resource.
func (o *KVStores) Configure() error {
	for name, settings := range o.Setup {
		if !o.AcceptDefaults && !o.NonInteractive {
			text.Break(o.Stdout)
			text.Output(o.Stdout, "Configuring kv store '%s'", name)
			if settings.Description != "" {
				text.Output(o.Stdout, settings.Description)
			}
		}

		var items []KVStoreItem

		for key, item := range settings.Items {
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
				text.Break(o.Stdout)
				text.Output(o.Stdout, "Create a kv store key called '%s'", key)
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
				abs, err := filepath.Abs(item.File)
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
			Name:  name,
			Items: items,
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
		err := o.Spinner.Start()
		if err != nil {
			return err
		}
		msg := fmt.Sprintf("Creating kv store '%s'", kvStore.Name)
		o.Spinner.Message(msg + "...")

		store, err := o.APIClient.CreateKVStore(&fastly.CreateKVStoreInput{
			Name: kvStore.Name,
		})
		if err != nil {
			err = fmt.Errorf("error creating kv store: %w", err)
			o.Spinner.StopFailMessage(msg)
			spinErr := o.Spinner.StopFail()
			if spinErr != nil {
				return fmt.Errorf(text.SpinnerErrWrapper, spinErr, err)
			}
			return err
		}

		o.Spinner.StopMessage(msg)
		err = o.Spinner.Stop()
		if err != nil {
			return err
		}

		if len(kvStore.Items) > 0 {
			for _, item := range kvStore.Items {
				err := o.Spinner.Start()
				if err != nil {
					return err
				}
				msg := fmt.Sprintf("Creating kv store key '%s'...", item.Key)
				o.Spinner.Message(msg)

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
					err = fmt.Errorf("error creating kv store key: %w", err)
					o.Spinner.StopFailMessage(msg)
					spinErr := o.Spinner.StopFail()
					if spinErr != nil {
						return fmt.Errorf(text.SpinnerErrWrapper, spinErr, err)
					}
					return err
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
		msg = fmt.Sprintf("Creating resource link between service and kv store '%s'...", kvStore.Name)
		o.Spinner.Message(msg)

		// IMPORTANT: We need to link the kv store to the C@E Service.
		_, err = o.APIClient.CreateResource(&fastly.CreateResourceInput{
			ServiceID:      o.ServiceID,
			ServiceVersion: o.ServiceVersion,
			Name:           fastly.String(store.Name),
			ResourceID:     fastly.String(store.ID),
		})
		if err != nil {
			err = fmt.Errorf("error creating resource link between the service '%s' and the kv store '%s': %w", o.ServiceID, store.Name, err)
			o.Spinner.StopFailMessage(msg)
			spinErr := o.Spinner.StopFail()
			if spinErr != nil {
				return fmt.Errorf(text.SpinnerErrWrapper, spinErr, err)
			}
			return err
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
func (o *KVStores) Predefined() bool {
	return len(o.Setup) > 0
}
