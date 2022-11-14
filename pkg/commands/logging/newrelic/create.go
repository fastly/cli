package newrelic

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/commands/logging/common"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v7/fastly"
)

// CreateCommand calls the Fastly API to create an appropriate resource.
type CreateCommand struct {
	cmd.Base

	name           string
	key            string
	serviceName    cmd.OptionalServiceNameID
	serviceVersion cmd.OptionalServiceVersion

	autoClone         cmd.OptionalAutoClone
	format            cmd.OptionalString
	formatVersion     cmd.OptionalUint
	manifest          manifest.Data
	placement         cmd.OptionalString
	region            cmd.OptionalString
	responseCondition cmd.OptionalString
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *CreateCommand {
	var c CreateCommand
	c.CmdClause = parent.Command("create", "Create an New Relic logging endpoint attached to the specified service version").Alias("add")
	c.Globals = globals
	c.manifest = data

	// Required flags
	c.CmdClause.Flag("key", "The Insert API key from the Account page of your New Relic account").Required().StringVar(&c.key)
	c.CmdClause.Flag("name", "The name for the real-time logging configuration").Required().StringVar(&c.name)
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagVersionName,
		Description: cmd.FlagVersionDesc,
		Dst:         &c.serviceVersion.Value,
		Required:    true,
	})

	// Optional flags
	c.RegisterAutoCloneFlag(cmd.AutoCloneFlagOpts{
		Action: c.autoClone.Set,
		Dst:    &c.autoClone.Value,
	})
	common.Format(c.CmdClause, &c.format)
	common.FormatVersion(c.CmdClause, &c.formatVersion)
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
		VerboseMode:        c.Globals.Flag.Verbose,
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

	input.Name = c.name
	input.ServiceID = serviceID
	input.ServiceVersion = serviceVersion
	input.Token = c.key

	if c.format.WasSet {
		input.Format = c.format.Value
	}

	if c.formatVersion.WasSet {
		input.FormatVersion = c.formatVersion.Value
	}

	if c.placement.WasSet {
		input.Placement = c.placement.Value
	}

	if c.region.WasSet {
		input.Region = c.region.Value
	}

	if c.responseCondition.WasSet {
		input.ResponseCondition = c.responseCondition.Value
	}

	return &input
}
