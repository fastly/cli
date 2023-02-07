package newrelic

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/commands/logging/common"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v7/fastly"
)

// CreateCommand calls the Fastly API to create an appropriate resource.
type CreateCommand struct {
	cmd.Base
	manifest manifest.Data

	// required
	serviceName    cmd.OptionalServiceNameID
	serviceVersion cmd.OptionalServiceVersion

	// optional
	autoClone         cmd.OptionalAutoClone
	format            cmd.OptionalString
	formatVersion     cmd.OptionalInt
	key               cmd.OptionalString
	name              cmd.OptionalString
	placement         cmd.OptionalString
	region            cmd.OptionalString
	responseCondition cmd.OptionalString
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent cmd.Registerer, g *global.Data, m manifest.Data) *CreateCommand {
	c := CreateCommand{
		Base: cmd.Base{
			Globals: g,
		},
		manifest: m,
	}
	c.CmdClause = parent.Command("create", "Create an New Relic logging endpoint attached to the specified service version").Alias("add")

	// required
	c.CmdClause.Flag("name", "The name for the real-time logging configuration").Action(c.name.Set).StringVar(&c.name.Value)
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagVersionName,
		Description: cmd.FlagVersionDesc,
		Dst:         &c.serviceVersion.Value,
		Required:    true,
	})

	// optional
	c.RegisterAutoCloneFlag(cmd.AutoCloneFlagOpts{
		Action: c.autoClone.Set,
		Dst:    &c.autoClone.Value,
	})
	common.Format(c.CmdClause, &c.format)
	common.FormatVersion(c.CmdClause, &c.formatVersion)
	c.CmdClause.Flag("key", "The Insert API key from the Account page of your New Relic account").Action(c.key.Set).StringVar(&c.key.Value)
	c.CmdClause.Flag("placement", "Where in the generated VCL the logging call should be placed").Action(c.placement.Set).StringVar(&c.placement.Value)
	c.CmdClause.Flag("region", "The region to which to stream logs").Action(c.region.Set).StringVar(&c.region.Value)
	c.CmdClause.Flag("response-condition", "The name of an existing condition in the configured endpoint").Action(c.responseCondition.Set).StringVar(&c.responseCondition.Value)
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagServiceIDName,
		Description: cmd.FlagServiceIDDesc,
		Dst:         &c.manifest.Flag.ServiceID,
		Short:       's',
	})
	c.RegisterFlag(cmd.StringFlagOpts{
		Action:      c.serviceName.Set,
		Name:        cmd.FlagServiceName,
		Description: cmd.FlagServiceDesc,
		Dst:         &c.serviceName.Value,
	})

	return &c
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	serviceID, serviceVersion, err := cmd.ServiceDetails(cmd.ServiceDetailsOpts{
		AutoCloneFlag:      c.autoClone,
		APIClient:          c.Globals.APIClient,
		Manifest:           c.manifest,
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

	input := c.constructInput(serviceID, serviceVersion.Number)

	l, err := c.Globals.APIClient.CreateNewRelic(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": serviceVersion.Number,
		})
		return err
	}

	text.Success(out, "Created New Relic logging endpoint '%s' (service: %s, version: %d)", l.Name, l.ServiceID, l.ServiceVersion)
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
