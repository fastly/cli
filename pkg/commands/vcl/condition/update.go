package condition

import (
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/v10/pkg/argparser"
	"github.com/fastly/cli/v10/pkg/errors"
	"github.com/fastly/cli/v10/pkg/global"
	"github.com/fastly/cli/v10/pkg/text"
)

// UpdateCommand calls the Fastly API to update an appropriate resource.
type UpdateCommand struct {
	argparser.Base
	input          fastly.UpdateConditionInput
	serviceName    argparser.OptionalServiceNameID
	serviceVersion argparser.OptionalServiceVersion
	autoClone      argparser.OptionalAutoClone

	newName       argparser.OptionalString
	statement     argparser.OptionalString
	conditionType argparser.OptionalString
	priority      argparser.OptionalInt
	comment       argparser.OptionalString
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent argparser.Registerer, g *global.Data) *UpdateCommand {
	var c UpdateCommand
	c.CmdClause = parent.Command("update", "Update a condition on a Fastly service version")
	c.Globals = g

	// Required flags
	c.CmdClause.Flag("name", "Domain name").Short('n').Required().StringVar(&c.input.Name)
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagVersionName,
		Description: argparser.FlagVersionDesc,
		Dst:         &c.serviceVersion.Value,
		Required:    true,
	})

	// Optional flags
	c.RegisterAutoCloneFlag(argparser.AutoCloneFlagOpts{
		Action: c.autoClone.Set,
		Dst:    &c.autoClone.Value,
	})
	c.CmdClause.Flag("new-name", "New condition name").Action(c.newName.Set).StringVar(&c.newName.Value)
	c.CmdClause.Flag("priority", "Condition priority").Action(c.priority.Set).IntVar(&c.priority.Value)
	c.CmdClause.Flag("statement", "Condition statement").Action(c.statement.Set).StringVar(&c.statement.Value)
	c.CmdClause.Flag("type", "Condition type").Action(c.conditionType.Set).StringVar(&c.conditionType.Value)
	c.CmdClause.Flag("comment", "Condition comment").Action(c.comment.Set).StringVar(&c.comment.Value)

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

	c.input.ServiceID = serviceID
	c.input.ServiceVersion = fastly.ToValue(serviceVersion.Number)

	// If no argument are provided, error with useful message.
	if !c.newName.WasSet && !c.priority.WasSet && !c.statement.WasSet && !c.conditionType.WasSet {
		return fmt.Errorf("error parsing arguments: must provide either --new-name, --statement, --type or --priority to update condition")
	}

	if c.newName.WasSet {
		c.input.Name = c.newName.Value
	}
	if c.priority.WasSet {
		c.input.Priority = &c.priority.Value
	}
	if c.conditionType.WasSet {
		c.input.Type = &c.conditionType.Value
	}
	if c.statement.WasSet {
		c.input.Statement = &c.statement.Value
	}
	if c.comment.WasSet {
		c.input.Statement = &c.comment.Value
	}

	r, err := c.Globals.APIClient.UpdateCondition(&c.input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": fastly.ToValue(serviceVersion.Number),
		})
		return err
	}

	text.Success(out,
		"Updated condition %s (service %s version %d)",
		fastly.ToValue(r.Name),
		fastly.ToValue(r.ServiceID),
		fastly.ToValue(r.ServiceVersion),
	)
	return nil
}
