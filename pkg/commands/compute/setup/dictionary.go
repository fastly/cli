package setup

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v6/fastly"
)

// Dictionaries represents the service state related to dictionaries defined
// within the fastly.toml [setup] configuration.
//
// NOTE: It implements the setup.Interface interface.
type Dictionaries struct {
	// Public
	APIClient      api.Interface
	AcceptDefaults bool
	Progress       text.Progress
	ServiceID      string
	ServiceVersion int
	Setup          map[string]*manifest.SetupDictionary
	Stdin          io.Reader
	Stdout         io.Writer

	// Private
	required []Dictionary
}

// Dictionary represents the configuration parameters for creating a dictionary
// via the API client.
//
// NOTE: WriteOnly (i.e. private) dictionaries not supported.
type Dictionary struct {
	Name  string
	Items []DictionaryItem
}

// DictionaryItem represents the configuration parameters for creating dictionary
// items via the API client.
type DictionaryItem struct {
	Key   string
	Value string
}

// Configure prompts the user for specific values related to the service resource.
func (d *Dictionaries) Configure() error {
	for name, settings := range d.Setup {
		if !d.AcceptDefaults {
			text.Break(d.Stdout)
			text.Output(d.Stdout, "Configuring dictionary '%s'", name)
			if settings.Description != "" {
				text.Output(d.Stdout, settings.Description)
			}
		}

		var items []DictionaryItem

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

			if !d.AcceptDefaults {
				text.Break(d.Stdout)
				text.Output(d.Stdout, "Create a dictionary key called '%s'", key)
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

			items = append(items, DictionaryItem{
				Key:   key,
				Value: value,
			})
		}

		d.required = append(d.required, Dictionary{
			Name:  name,
			Items: items,
		})
	}

	return nil
}

// Create calls the relevant API to create the service resource(s).
func (d *Dictionaries) Create() error {
	if d.Progress == nil {
		return errors.RemediationError{
			Inner:       fmt.Errorf("internal logic error: no text.Progress configured for setup.Dictionaries"),
			Remediation: errors.BugRemediation,
		}
	}

	for _, dictionary := range d.required {
		d.Progress.Step(fmt.Sprintf("Creating dictionary '%s'...", dictionary.Name))

		dict, err := d.APIClient.CreateDictionary(&fastly.CreateDictionaryInput{
			ServiceID:      d.ServiceID,
			ServiceVersion: d.ServiceVersion,
			Name:           dictionary.Name,
		})
		if err != nil {
			d.Progress.Fail()
			return fmt.Errorf("error creating dictionary: %w", err)
		}

		if len(dictionary.Items) > 0 {
			for _, item := range dictionary.Items {
				d.Progress.Step(fmt.Sprintf("Creating dictionary item '%s'...", item.Key))

				_, err := d.APIClient.CreateDictionaryItem(&fastly.CreateDictionaryItemInput{
					ServiceID:    d.ServiceID,
					DictionaryID: dict.ID,
					ItemKey:      item.Key,
					ItemValue:    item.Value,
				})
				if err != nil {
					d.Progress.Fail()
					return fmt.Errorf("error creating dictionary item: %w", err)
				}
			}
		}
	}

	return nil
}

// Predefined indicates if the service resource has been specified within the
// fastly.toml file using a [setup] configuration block.
func (d *Dictionaries) Predefined() bool {
	return len(d.Setup) > 0
}
