package version

import (
	"context"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// UpdateCommand calls the Fastly API to update a service version.
type UpdateCommand struct {
	argparser.Base
	input          fastly.UpdateVersionInput
	ServiceID      *string                         `short:"s" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`
	ServiceName    argparser.OptionalServiceNameID `help:"The name of the service"`
	Version        string                          `required:"" help:"'latest', 'active', 'staged', or the number of a specific Fastly service version"`
	serviceVersion argparser.OptionalServiceVersion
	AutoClone      argparser.OptionalAutoClone `name:"autoclone" help:"If the selected service version is not editable, clone it and use the clone."`

	// TODO(integralist):
	// Make 'comment' field mandatory once we roll out a new release of Go-Fastly
	// which will hopefully have better/more correct consistency as far as which
	// fields are supposed to be optional and which should be 'required'.
	//
	Comment argparser.OptionalString `help:"Human-readable comment"`
}

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *UpdateCommand) AfterApply() error {
	if c.ServiceID != nil {
		c.Globals.Manifest.Flag.ServiceID = *c.ServiceID
	}
	c.serviceVersion.Value = c.Version
	return nil
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	serviceID, serviceVersion, err := argparser.ServiceDetails(argparser.ServiceDetailsOpts{
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

	c.input.ServiceID = serviceID
	c.input.ServiceVersion = fastly.ToValue(serviceVersion.Number)
	if !c.Comment.WasSet {
		return fmt.Errorf("error parsing arguments: required flag --comment not provided")
	}
	c.input.Comment = &c.Comment.Value

	ver, err := c.Globals.APIClient.UpdateVersion(context.TODO(), &c.input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": fastly.ToValue(serviceVersion.Number),
			"Comment":         c.Comment.Value,
		})
		return err
	}

	text.Success(out, "Updated service %s version %d", fastly.ToValue(ver.ServiceID), c.input.ServiceVersion)
	return nil
}
