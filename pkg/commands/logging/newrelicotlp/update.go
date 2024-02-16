package newrelicotlp

import (
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/v10/pkg/argparser"
	"github.com/fastly/cli/v10/pkg/commands/logging/common"
	"github.com/fastly/cli/v10/pkg/errors"
	"github.com/fastly/cli/v10/pkg/global"
	"github.com/fastly/cli/v10/pkg/text"
)

// UpdateCommand calls the Fastly API to update an appropriate resource.
type UpdateCommand struct {
	argparser.Base

	endpointName   string
	serviceName    argparser.OptionalServiceNameID
	serviceVersion argparser.OptionalServiceVersion

	autoClone         argparser.OptionalAutoClone
	format            argparser.OptionalString
	formatVersion     argparser.OptionalInt
	key               argparser.OptionalString
	newName           argparser.OptionalString
	placement         argparser.OptionalString
	region            argparser.OptionalString
	responseCondition argparser.OptionalString
	url               argparser.OptionalString
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent argparser.Registerer, g *global.Data) *UpdateCommand {
	c := UpdateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("update", "Update a New Relic Logs logging object for a particular service and version")

	// Required.
	c.CmdClause.Flag("name", "The name for the real-time logging configuration to update").Required().StringVar(&c.endpointName)
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
	common.Format(c.CmdClause, &c.format)
	c.CmdClause.Flag("format-version", "The version of the custom logging format used for the configured endpoint").Action(c.formatVersion.Set).IntVar(&c.formatVersion.Value)
	c.CmdClause.Flag("key", "The Insert API key from the Account page of your New Relic account").Action(c.key.Set).StringVar(&c.key.Value)
	c.CmdClause.Flag("new-name", "The name for the real-time logging configuration").Action(c.newName.Set).StringVar(&c.newName.Value)
	c.CmdClause.Flag("placement", "Where in the generated VCL the logging call should be placed").Action(c.placement.Set).StringVar(&c.placement.Value)
	c.CmdClause.Flag("region", "The region to which to stream logs").Action(c.region.Set).StringVar(&c.region.Value)
	c.CmdClause.Flag("response-condition", "The name of an existing condition in the configured endpoint").Action(c.responseCondition.Set).StringVar(&c.responseCondition.Value)
	c.CmdClause.Flag("url", "URL of the New Relic Trace Observer, if you are using New Relic Infinite Tracing").Action(c.url.Set).StringVar(&c.url.Value)
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagServiceIDName,
		Description: argparser.FlagServiceIDDesc,
		Dst:         &g.Manifest.Flag.ServiceID,
		Short:       's',
	})
	c.RegisterFlag(argparser.StringFlagOpts{
		Action:      c.serviceName.Set,
		Name:        argparser.FlagServiceName,
		Description: argparser.FlagServiceDesc,
		Dst:         &c.serviceName.Value,
	})

	return &c
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	serviceID, serviceVersion, err := argparser.ServiceDetails(argparser.ServiceDetailsOpts{
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
			"Service Version": errors.ServiceVersion(serviceVersion),
		})
		return err
	}

	input := c.constructInput(serviceID, fastly.ToValue(serviceVersion.Number))

	l, err := c.Globals.APIClient.UpdateNewRelicOTLP(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": fastly.ToValue(serviceVersion.Number),
		})
		return err
	}

	var prev string
	if c.newName.WasSet {
		prev = fmt.Sprintf("previously: %s, ", c.endpointName)
	}

	text.Success(out,
		"Updated New Relic OTLP logging endpoint '%s' (%sservice: %s, version: %d)",
		fastly.ToValue(l.Name),
		prev,
		fastly.ToValue(l.ServiceID),
		fastly.ToValue(l.ServiceVersion),
	)
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) constructInput(serviceID string, serviceVersion int) *fastly.UpdateNewRelicOTLPInput {
	var input fastly.UpdateNewRelicOTLPInput

	input.Name = c.endpointName
	input.ServiceID = serviceID
	input.ServiceVersion = serviceVersion

	if c.format.WasSet {
		input.Format = &c.format.Value
	}
	if c.formatVersion.WasSet {
		input.FormatVersion = &c.formatVersion.Value
	}
	if c.key.WasSet {
		input.Token = &c.key.Value
	}
	if c.newName.WasSet {
		input.NewName = &c.newName.Value
	}
	if c.placement.WasSet {
		input.Placement = &c.placement.Value
	}
	if c.region.WasSet {
		input.Region = &c.region.Value
	}
	if c.responseCondition.WasSet {
		input.ResponseCondition = &c.responseCondition.Value
	}
	if c.url.WasSet {
		input.URL = &c.url.Value
	}

	return &input
}
