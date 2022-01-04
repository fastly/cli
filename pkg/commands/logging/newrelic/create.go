package newrelic

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v5/fastly"
)

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
	})

	// Optional flags
	c.RegisterAutoCloneFlag(cmd.AutoCloneFlagOpts{
		Action: c.autoClone.Set,
		Dst:    &c.autoClone.Value,
	})
	c.CmdClause.Flag("format", "A Fastly log format string. Must produce valid JSON that New Relic Logs can ingest").StringVar(&c.format)
	c.CmdClause.Flag("format-version", "The version of the custom logging format used for the configured endpoint").UintVar(&c.formatVersion)
	c.CmdClause.Flag("placement", "Where in the generated VCL the logging call should be placed").StringVar(&c.placement)
	c.CmdClause.Flag("region", "The region to which to stream logs").StringVar(&c.region)
	c.CmdClause.Flag("response-condition", "The name of an existing condition in the configured endpoint").Action(c.responseCondition.Set).StringVar(&c.responseCondition.Value)
	c.RegisterServiceIDFlag(&c.manifest.Flag.ServiceID)
	c.RegisterServiceNameFlag(c.serviceName.Set, &c.serviceName.Value)

	return &c
}

// CreateCommand calls the Fastly API to create an appropriate resource.
type CreateCommand struct {
	cmd.Base

	autoClone         cmd.OptionalAutoClone
	format            string
	formatVersion     uint
	key               string
	manifest          manifest.Data
	name              string
	placement         string
	region            string
	responseCondition cmd.OptionalString
	serviceName       cmd.OptionalServiceNameID
	serviceVersion    cmd.OptionalServiceVersion
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, serviceVersion, err := cmd.ServiceDetails(cmd.ServiceDetailsOpts{
		AutoCloneFlag:      c.autoClone,
		Client:             c.Globals.Client,
		Manifest:           c.manifest,
		Out:                out,
		ServiceNameFlag:    c.serviceName,
		ServiceVersionFlag: c.serviceVersion,
		VerboseMode:        c.Globals.Flag.Verbose,
	})
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
			"Service ID":      serviceID,
			"Service Version": errors.ServiceVersion(serviceVersion),
		})
		return err
	}

	input := c.constructInput(serviceID, serviceVersion.Number)

	l, err := c.Globals.Client.CreateNewRelic(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
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

	if c.format != "" {
		input.Format = c.format
	}
	if c.formatVersion > 0 {
		input.FormatVersion = c.formatVersion
	}
	if c.placement != "" {
		input.Placement = c.placement
	}
	if c.region != "" {
		input.Region = c.region
	}
	if c.responseCondition.WasSet {
		input.ResponseCondition = c.responseCondition.Value
	}

	return &input
}
