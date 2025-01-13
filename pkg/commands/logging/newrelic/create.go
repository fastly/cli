package newrelic

import (
	"io"

	"github.com/fastly/go-fastly/v9/fastly"

	"4d63.com/optional"
	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/commands/logging/common"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// CreateCommand calls the Fastly API to create an appropriate resource.
type CreateCommand struct {
	argparser.Base

	// Required.
	serviceName    argparser.OptionalServiceNameID
	serviceVersion argparser.OptionalServiceVersion

	// Optional.
	autoClone         argparser.OptionalAutoClone
	format            argparser.OptionalString
	formatVersion     argparser.OptionalInt
	key               argparser.OptionalString
	name              argparser.OptionalString
	placement         argparser.OptionalString
	region            argparser.OptionalString
	responseCondition argparser.OptionalString
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent argparser.Registerer, g *global.Data) *CreateCommand {
	c := CreateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("create", "Create an New Relic logging endpoint attached to the specified service version").Alias("add")

	// Required.
	c.CmdClause.Flag("name", "The name for the real-time logging configuration").Action(c.name.Set).StringVar(&c.name.Value)
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
	common.FormatVersion(c.CmdClause, &c.formatVersion)
	c.CmdClause.Flag("key", "The Insert API key from the Account page of your New Relic account").Action(c.key.Set).StringVar(&c.key.Value)
	c.CmdClause.Flag("placement", "Where in the generated VCL the logging call should be placed").Action(c.placement.Set).StringVar(&c.placement.Value)
	c.CmdClause.Flag("region", "The region to which to stream logs").Action(c.region.Set).StringVar(&c.region.Value)
	c.CmdClause.Flag("response-condition", "The name of an existing condition in the configured endpoint").Action(c.responseCondition.Set).StringVar(&c.responseCondition.Value)
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

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
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
			"Service Version": errors.ServiceVersion(serviceVersion),
		})
		return err
	}

	input := c.constructInput(serviceID, fastly.ToValue(serviceVersion.Number))

	l, err := c.Globals.APIClient.CreateNewRelic(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": fastly.ToValue(serviceVersion.Number),
		})
		return err
	}

	text.Success(out,
		"Created New Relic logging endpoint '%s' (service: %s, version: %d)",
		fastly.ToValue(l.Name),
		fastly.ToValue(l.ServiceID),
		fastly.ToValue(l.ServiceVersion),
	)
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *CreateCommand) constructInput(serviceID string, serviceVersion int) *fastly.CreateNewRelicInput {
	var input fastly.CreateNewRelicInput

	if c.name.WasSet {
		input.Name = &c.name.Value
	}
	input.ServiceID = serviceID
	input.ServiceVersion = serviceVersion
	if c.key.WasSet {
		input.Token = &c.key.Value
	}

	if c.format.WasSet {
		input.Format = &c.format.Value
	}

	if c.formatVersion.WasSet {
		input.FormatVersion = &c.formatVersion.Value
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

	return &input
}
