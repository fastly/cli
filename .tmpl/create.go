package ${CLI_PACKAGE}

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v4/fastly"
)

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent argparser.Registerer, globals *config.Data, data manifest.Data) *CreateCommand {
	var c CreateCommand
	c.CmdClause = parent.Command("create", "<...>").Alias("add")
	c.Globals = globals
	c.manifest = data

	// Required flags
	// c.CmdClause.Flag("<...>", "<...>").Required().StringVar(&c.<...>)
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagVersionName,
		Description: argparser.FlagVersionDesc,
		Dst:         &c.serviceVersion.Value,
		Required:    true,
	})

	// Optional flags
	// c.CmdClause.Flag("<...>", "<...>").Action(c.<...>.Set).StringVar(&c.<...>.Value)
	c.RegisterAutoCloneFlag(argparser.AutoCloneFlagOpts{
		Action: c.autoClone.Set,
		Dst:    &c.autoClone.Value,
	})
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagServiceIDName,
		Description: argparser.FlagServiceIDDesc,
		Dst:         &c.manifest.Flag.ServiceID,
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

// CreateCommand calls the Fastly API to create an appropriate resource.
type CreateCommand struct {
	argparser.Base

	autoClone      argparser.OptionalAutoClone
	manifest manifest.Data
	serviceName    argparser.OptionalServiceNameID
	serviceVersion argparser.OptionalServiceVersion
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, serviceVersion, err := argparser.ServiceDetails(argparser.ServiceDetailsOpts{
		Active:             optional.Of(false),
		Locked:             optional.Of(false),
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

	r, err := c.Globals.Client.Create${CLI_API}(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
			"Service ID": serviceID,
			"Service Version": serviceVersion.Number,
		})
		return err
	}

	text.Success(out, "Created <...> '%s' (service: %s, version: %d)", r.<...>, r.ServiceID, r.ServiceVersion)
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *CreateCommand) constructInput(serviceID string, serviceVersion int) *fastly.Create${CLI_API}Input {
	var input fastly.Create${CLI_API}Input

	input.ServiceID = serviceID
	input.ServiceVersion = serviceVersion

	// if c.<...>.WasSet {
	// 	input.<...> = c.<...>.Value
	// }

	return &input
}
