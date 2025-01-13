package custom

import (
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v9/fastly"

	"4d63.com/optional"
	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent argparser.Registerer, g *global.Data) *UpdateCommand {
	c := UpdateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("update", "Update the uploaded VCL for a particular service and version")

	// Required.
	c.CmdClause.Flag("name", "The name of the VCL to update").Required().StringVar(&c.name)
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagVersionName,
		Description: argparser.FlagVersionDesc,
		Dst:         &c.serviceVersion.Value,
		Required:    true,
	})

	// Optional.
	c.RegisterAutoCloneFlag(argparser.AutoCloneFlagOpts{
		Action: c.autoClone.Set,
		Dst:    &c.autoClone.Value,
	})
	c.CmdClause.Flag("new-name", "New name for the VCL").Action(c.newName.Set).StringVar(&c.newName.Value)
	c.CmdClause.Flag("content", "VCL passed as file path or content, e.g. $(< main.vcl)").Action(c.content.Set).StringVar(&c.content.Value)
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagServiceIDName,
		Description: argparser.FlagServiceIDDesc,
		Dst:         &g.Manifest.Flag.ServiceID,
		Short:       's',
	})
	c.RegisterFlag(argparser.StringFlagOpts{
		Action:      c.serviceName.Set,
		Name:        argparser.FlagServiceName,
		Description: argparser.FlagServiceNameDesc,
		Dst:         &c.serviceName.Value,
	})

	return &c
}

// UpdateCommand calls the Fastly API to update an appropriate resource.
type UpdateCommand struct {
	argparser.Base

	autoClone      argparser.OptionalAutoClone
	content        argparser.OptionalString
	name           string
	newName        argparser.OptionalString
	serviceName    argparser.OptionalServiceNameID
	serviceVersion argparser.OptionalServiceVersion
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	serviceID, serviceVersion, err := argparser.ServiceDetails(argparser.ServiceDetailsOpts{
		Active:             optional.Of(false),
		Locked:             optional.Of(false),
		AutoCloneFlag:      c.autoClone,
		APIClient:          c.Globals.APIClient,
		Manifest:           *c.Globals.Manifest,
		Out:                out,
		ServiceNameFlag:    c.serviceName,
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

	input, err := c.constructInput(serviceID, fastly.ToValue(serviceVersion.Number))
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": fastly.ToValue(serviceVersion.Number),
		})
		return err
	}

	v, err := c.Globals.APIClient.UpdateVCL(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": fastly.ToValue(serviceVersion.Number),
		})
		return err
	}

	if input.NewName != nil && *input.NewName != "" {
		text.Success(out,
			"Updated custom VCL '%s' (previously: '%s', service: %s, version: %d)",
			fastly.ToValue(v.Name),
			input.Name,
			fastly.ToValue(v.ServiceID),
			fastly.ToValue(v.ServiceVersion),
		)
	} else {
		text.Success(out,
			"Updated custom VCL '%s' (service: %s, version: %d)",
			fastly.ToValue(v.Name),
			fastly.ToValue(v.ServiceID),
			fastly.ToValue(v.ServiceVersion),
		)
	}
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) constructInput(serviceID string, serviceVersion int) (*fastly.UpdateVCLInput, error) {
	var input fastly.UpdateVCLInput

	input.Name = c.name
	input.ServiceID = serviceID
	input.ServiceVersion = serviceVersion

	if !c.newName.WasSet && !c.content.WasSet {
		return nil, fmt.Errorf("error parsing arguments: must provide either --new-name or --content to update the VCL")
	}
	if c.newName.WasSet {
		input.NewName = &c.newName.Value
	}
	if c.content.WasSet {
		input.Content = fastly.ToPointer(argparser.Content(c.content.Value))
	}

	return &input, nil
}
