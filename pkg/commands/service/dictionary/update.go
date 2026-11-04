package dictionary

import (
	"context"
	"fmt"
	"io"
	"strconv"

	"github.com/fastly/go-fastly/v19/fastly"

	"4d63.com/optional"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// UpdateCommand calls the Fastly API to update a dictionary.
type UpdateCommand struct {
	argparser.Base

	// TODO: make input consistent across commands (most are title case)
	input          fastly.UpdateDictionaryInput
	ServiceName    argparser.OptionalServiceNameID `help:"The name of the service"`
	serviceVersion argparser.OptionalServiceVersion

	// Optional.
	AutoClone argparser.OptionalAutoClone `name:"autoclone" help:"If the selected service version is not editable, clone it and use the clone."`

	Newname   argparser.OptionalString `name:"new-name" help:"New name of Dictionary"`
	WriteOnly argparser.OptionalString `help:"Whether to mark this dictionary as write-only. Can be true or false (defaults to false)"`

	// Required.
	NameFlag string `name:"name" short:"n" required:"" help:"Old name of Dictionary"`
	Version  string `required:"" help:"'latest', 'active', 'staged', or the number of a specific Fastly service version"`

	ServiceID *string `short:"s" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`
}

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *UpdateCommand) AfterApply() error {
	c.input.Name = c.NameFlag
	c.serviceVersion.Value = c.Version
	if c.ServiceID != nil {
		c.Globals.Manifest.Flag.ServiceID = *c.ServiceID
	}
	return nil
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
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
			"Service Version": fsterr.ServiceVersion(serviceVersion),
		})
		return err
	}

	serviceVersionNumber := fastly.ToValue(serviceVersion.Number)

	c.input.ServiceID = serviceID
	c.input.ServiceVersion = serviceVersionNumber

	if !c.Newname.WasSet && !c.WriteOnly.WasSet {
		return fsterr.RemediationError{Inner: fmt.Errorf("error parsing arguments: required flag --new-name or --write-only not provided"), Remediation: "To fix this error, provide at least one of the aforementioned flags"}
	}
	if c.Newname.WasSet {
		c.input.NewName = &c.Newname.Value
	}

	if c.WriteOnly.WasSet {
		writeOnly, err := strconv.ParseBool(c.WriteOnly.Value)
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
				"Service ID":      serviceID,
				"Service Version": serviceVersionNumber,
			})
			return err
		}
		c.input.WriteOnly = new(fastly.Compatibool(writeOnly))
	}

	d, err := c.Globals.APIClient.UpdateDictionary(context.TODO(), &c.input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": serviceVersionNumber,
		})
		return err
	}

	text.Success(out, "Updated dictionary %s (service %s version %d)", fastly.ToValue(d.Name), fastly.ToValue(d.ServiceID), fastly.ToValue(d.ServiceVersion))
	if c.Globals.Verbose() {
		text.Output(out, "\nVersion: %d\n", fastly.ToValue(d.ServiceVersion))
		text.PrintDictionary(out, "", d)
	}
	return nil
}

// FlagOrder lists the flags in the order kingpin showed them in help, which
// the fields do not follow.
func (c *UpdateCommand) FlagOrder() []string {
	return []string{
		"name", "version", "autoclone", "new-name", "service-id",
		"service-name", "write-only",
	}
}
