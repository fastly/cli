package serviceversion

import (
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v10/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// UpdateCommand calls the Fastly API to update a service version.
type UpdateCommand struct {
	argparser.Base
	input          fastly.UpdateVersionInput
	serviceName    argparser.OptionalServiceNameID
	serviceVersion argparser.OptionalServiceVersion
	autoClone      argparser.OptionalAutoClone

	comment argparser.OptionalString
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent argparser.Registerer, g *global.Data) *UpdateCommand {
	c := UpdateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("update", "Update a Fastly service version")
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
	c.RegisterAutoCloneFlag(argparser.AutoCloneFlagOpts{
		Action: c.autoClone.Set,
		Dst:    &c.autoClone.Value,
	})

	// TODO(integralist):
	// Make 'comment' field mandatory once we roll out a new release of Go-Fastly
	// which will hopefully have better/more correct consistency as far as which
	// fields are supposed to be optional and which should be 'required'.
	//
	c.CmdClause.Flag("comment", "Human-readable comment").Action(c.comment.Set).StringVar(&c.comment.Value)
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

	c.input.ServiceID = serviceID
	c.input.ServiceVersion = fastly.ToValue(serviceVersion.Number)
	if !c.comment.WasSet {
		return fmt.Errorf("error parsing arguments: required flag --comment not provided")
	}
	c.input.Comment = &c.comment.Value

	ver, err := c.Globals.APIClient.UpdateVersion(&c.input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": fastly.ToValue(serviceVersion.Number),
			"Comment":         c.comment.Value,
		})
		return err
	}

	text.Success(out, "Updated service %s version %d", fastly.ToValue(ver.ServiceID), c.input.ServiceVersion)
	return nil
}
