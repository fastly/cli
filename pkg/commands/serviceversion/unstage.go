package serviceversion

import (
	"io"

	"github.com/fastly/go-fastly/v9/fastly"

	"4d63.com/optional"
	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// UnstageCommand calls the Fastly API to unstage a service version.
type UnstageCommand struct {
	argparser.Base
	Input          fastly.DeactivateVersionInput
	serviceName    argparser.OptionalServiceNameID
	serviceVersion argparser.OptionalServiceVersion
}

// NewUnstageCommand returns a usable command registered under the parent.
func NewUnstageCommand(parent argparser.Registerer, g *global.Data) *UnstageCommand {
	var c UnstageCommand
	c.Globals = g
	c.CmdClause = parent.Command("unstage", "Unstage a Fastly service version")
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
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagVersionName,
		Description: argparser.FlagVersionDesc,
		Dst:         &c.serviceVersion.Value,
		Required:    true,
	})
	return &c
}

// Exec invokes the application logic for the command.
func (c *UnstageCommand) Exec(_ io.Reader, out io.Writer) error {
	serviceID, serviceVersion, err := argparser.ServiceDetails(argparser.ServiceDetailsOpts{
		Staging:            optional.Of(true),
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

	c.Input.ServiceID = serviceID
	c.Input.ServiceVersion = fastly.ToValue(serviceVersion.Number)
	c.Input.Environment = "staging"

	ver, err := c.Globals.APIClient.DeactivateVersion(&c.Input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": fastly.ToValue(serviceVersion.Number),
		})
		return err
	}

	text.Success(out, "Unstaged service %s version %d", fastly.ToValue(ver.ServiceID), c.Input.ServiceVersion)
	return nil
}
