package condition

import (
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
)

// UpdateCommand calls the Fastly API to update an appropriate resource.
type UpdateCommand struct {
	cmd.Base
	manifest       manifest.Data
	input          fastly.UpdateConditionInput
	serviceName    cmd.OptionalServiceNameID
	serviceVersion cmd.OptionalServiceVersion
	autoClone      cmd.OptionalAutoClone

	newName       cmd.OptionalString
	statement     cmd.OptionalString
	conditionType cmd.OptionalString
	priority      cmd.OptionalInt
	comment       cmd.OptionalString
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent cmd.Registerer, globals *global.Data, data manifest.Data) *UpdateCommand {
	var c UpdateCommand
	c.CmdClause = parent.Command("update", "Update a condition on a Fastly service version")
	c.Globals = globals
	c.manifest = data

	// Required flags
	c.CmdClause.Flag("name", "Domain name").Short('n').Required().StringVar(&c.input.Name)
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
	c.CmdClause.Flag("new-name", "New condition name").Action(c.newName.Set).StringVar(&c.newName.Value)
	c.CmdClause.Flag("priority", "Condition priority").Action(c.priority.Set).IntVar(&c.priority.Value)
	c.CmdClause.Flag("statement", "Condition statement").Action(c.statement.Set).StringVar(&c.statement.Value)
	c.CmdClause.Flag("type", "Condition type").Action(c.conditionType.Set).StringVar(&c.conditionType.Value)
	c.CmdClause.Flag("comment", "Condition comment").Action(c.comment.Set).StringVar(&c.comment.Value)

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
	c.input.ServiceVersion = serviceVersion.Number

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
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
			"Service ID":      serviceID,
			"Service Version": serviceVersion.Number,
		})
		return err
	}

	text.Success(out, "Updated condition %s (service %s version %d)", r.Name, r.ServiceID, r.ServiceVersion)
	return nil
}
