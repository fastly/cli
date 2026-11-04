package dictionary

import (
	"context"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"4d63.com/optional"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// CreateCommand calls the Fastly API to create a service.
type CreateCommand struct {
	argparser.Base

	// Required.
	serviceVersion argparser.OptionalServiceVersion
	Version        string `required:"" help:"'latest', 'active', 'staged', or the number of a specific Fastly service version"`

	// Optional.
	AutoClone   argparser.OptionalAutoClone     `name:"autoclone" help:"If the selected service version is not editable, clone it and use the clone."`
	NameFlag    argparser.OptionalString        `name:"name" short:"n" help:"Name of Dictionary"`
	ServiceID   *string                         `short:"s" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`
	ServiceName argparser.OptionalServiceNameID `help:"The name of the service"`
	WriteOnly   argparser.OptionalBool          `help:"Whether to mark this dictionary as write-only"`
}

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *CreateCommand) AfterApply() error {
	c.serviceVersion.Value = c.Version
	if c.ServiceID != nil {
		c.Globals.Manifest.Flag.ServiceID = *c.ServiceID
	}
	return nil
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	serviceID, serviceVersion, err := argparser.ServiceDetails(argparser.ServiceDetailsOpts{
		Active:             optional.Of(false),
		Locked:             optional.Of(false),
		AutoCloneFlag:      c.AutoClone,
		APIClient:          c.Globals.APIClient,
		Manifest:           *c.Globals.Manifest,
		Out:                out,
		ServiceNameFlag:    c.ServiceName,
		ServiceVersionFlag: c.serviceVersion,
		VerboseMode:        c.Globals.Flags.Verbose,
	})
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": errors.ServiceVersion(serviceVersion),
		})
		return err
	}

	serviceVersionNumber := fastly.ToValue(serviceVersion.Number)

	input := fastly.CreateDictionaryInput{
		ServiceID:      serviceID,
		ServiceVersion: serviceVersionNumber,
	}
	if c.NameFlag.WasSet {
		input.Name = &c.NameFlag.Value
	}
	if c.WriteOnly.WasSet {
		input.WriteOnly = new(fastly.Compatibool(c.WriteOnly.Value))
	}

	d, err := c.Globals.APIClient.CreateDictionary(context.TODO(), &input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": serviceVersionNumber,
		})
		return err
	}

	var writeOnlyOutput string
	if fastly.ToValue(d.WriteOnly) {
		writeOnlyOutput = "as write-only "
	}

	text.Success(out, "Created dictionary %s %s(id %s, service %s, version %d)", fastly.ToValue(d.Name), writeOnlyOutput, fastly.ToValue(d.DictionaryID), fastly.ToValue(d.ServiceID), fastly.ToValue(d.ServiceVersion))
	return nil
}
