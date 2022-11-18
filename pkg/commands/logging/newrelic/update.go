package newrelic

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/commands/logging/common"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v7/fastly"
)

// UpdateCommand calls the Fastly API to update an appropriate resource.
type UpdateCommand struct {
	cmd.Base

	endpointName   string
	serviceName    cmd.OptionalServiceNameID
	serviceVersion cmd.OptionalServiceVersion

	autoClone         cmd.OptionalAutoClone
	format            cmd.OptionalString
	formatVersion     cmd.OptionalInt
	key               cmd.OptionalString
	manifest          manifest.Data
	newName           cmd.OptionalString
	placement         cmd.OptionalString
	region            cmd.OptionalString
	responseCondition cmd.OptionalString
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *UpdateCommand {
	c := UpdateCommand{
		Base: cmd.Base{
			Globals: globals,
		},
		manifest: data,
	}
	c.CmdClause = parent.Command("update", "Update a New Relic Logs logging object for a particular service and version")

	// required
	c.CmdClause.Flag("name", "The name for the real-time logging configuration to update").Required().StringVar(&c.endpointName)
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
	c.CmdClause.Flag("format-version", "The version of the custom logging format used for the configured endpoint").Action(c.formatVersion.Set).IntVar(&c.formatVersion.Value)
	c.CmdClause.Flag("key", "The Insert API key from the Account page of your New Relic account").Action(c.key.Set).StringVar(&c.key.Value)
	c.CmdClause.Flag("new-name", "The name for the real-time logging configuration").Action(c.newName.Set).StringVar(&c.newName.Value)
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
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
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

	l, err := c.Globals.APIClient.UpdateNewRelic(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": serviceVersion.Number,
		})
		return err
	}

	var prev string
	if c.newName.WasSet {
		prev = fmt.Sprintf("previously: %s, ", c.endpointName)
	}

	text.Success(out, "Updated New Relic logging endpoint '%s' (%sservice: %s, version: %d)", l.Name, prev, l.ServiceID, l.ServiceVersion)
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) constructInput(serviceID string, serviceVersion int) *fastly.UpdateNewRelicInput {
	var input fastly.UpdateNewRelicInput

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

	return &input
}
