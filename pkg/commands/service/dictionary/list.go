package dictionary

import (
	"context"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// ListCommand calls the Fastly API to list dictionaries.
type ListCommand struct {
	argparser.Base

	// Required.
	Version string `required:"" help:"'latest', 'active', 'staged', or the number of a specific Fastly service version"`

	// Optional.
	argparser.JSONOutput
	ServiceID *string `short:"s" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`

	Input          fastly.ListDictionariesInput    `kong:"-"`
	ServiceName    argparser.OptionalServiceNameID `help:"The name of the service"`
	serviceVersion argparser.OptionalServiceVersion
}

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *ListCommand) AfterApply() error {
	c.serviceVersion.Value = c.Version
	if c.ServiceID != nil {
		c.Globals.Manifest.Flag.ServiceID = *c.ServiceID
	}
	return nil
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}
	serviceID, serviceVersion, err := argparser.ServiceDetails(argparser.ServiceDetailsOpts{
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

	c.Input.ServiceID = serviceID
	c.Input.ServiceVersion = serviceVersionNumber

	o, err := c.Globals.APIClient.ListDictionaries(context.TODO(), &c.Input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": serviceVersionNumber,
		})
		return err
	}

	if ok, err := c.WriteJSON(out, o); ok {
		return err
	}

	if !c.Globals.Verbose() {
		fmt.Fprintf(out, "\nService ID: %s\n", serviceID)
	}
	text.Output(out, "Version: %d", c.Input.ServiceVersion)
	for _, dictionary := range o {
		text.PrintDictionary(out, "", dictionary)
	}

	return nil
}
