package condition

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v8/fastly"
)

// CreateCommand calls the Fastly API to create an appropriate resource.
type CreateCommand struct {
	cmd.Base
	manifest manifest.Data

	// Required.
	serviceVersion cmd.OptionalServiceVersion

	// Optional.
	autoClone     cmd.OptionalAutoClone
	conditionType cmd.OptionalString
	name          cmd.OptionalString
	priority      cmd.OptionalInt
	serviceName   cmd.OptionalServiceNameID
	statement     cmd.OptionalString
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent cmd.Registerer, g *global.Data, m manifest.Data) *CreateCommand {
	c := CreateCommand{
		Base: cmd.Base{
			Globals: g,
		},
		manifest: m,
	}
	c.CmdClause = parent.Command("create", "Create a condtion on a Fastly service version").Alias("add")

	// Required flags
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
	c.CmdClause.Flag("name", "Condition name").Short('n').Action(c.name.Set).StringVar(&c.name.Value)
	c.CmdClause.Flag("priority", "Condition priority").Action(c.priority.Set).IntVar(&c.priority.Value)
	c.CmdClause.Flag("statement", "Condition statement").Action(c.statement.Set).StringVar(&c.statement.Value)
	c.CmdClause.Flag("type", "Condition type").Action(c.conditionType.Set).StringVar(&c.conditionType.Value)
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
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
			"Service ID":      serviceID,
			"Service Version": errors.ServiceVersion(serviceVersion),
		})
		return err
	}

	input := fastly.CreateConditionInput{
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion.Number,
	}

	if c.name.WasSet {
		input.Name = &c.name.Value
	}
	if c.statement.WasSet {
		input.Statement = &c.statement.Value
	}
	if c.conditionType.WasSet {
		input.Type = &c.conditionType.Value
	}
	if c.priority.WasSet {
		input.Priority = &c.priority.Value
	}
	r, err := c.Globals.APIClient.CreateCondition(&input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
			"Service ID":      serviceID,
			"Service Version": serviceVersion.Number,
		})
		return err
	}

	text.Success(out, "Created condition %s (service %s version %d)", r.Name, r.ServiceID, r.ServiceVersion)
	return nil
}
