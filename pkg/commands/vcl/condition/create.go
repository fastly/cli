package condition

import (
	"io"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/v10/pkg/argparser"
	"github.com/fastly/cli/v10/pkg/errors"
	"github.com/fastly/cli/v10/pkg/global"
	"github.com/fastly/cli/v10/pkg/text"
)

// ConditionTypes are the allowed input values for the --type flag.
// Reference: https://developer.fastly.com/reference/api/vcl-services/condition/
var ConditionTypes = []string{"REQUEST", "CACHE", "RESPONSE", "PREFETCH"}

// CreateCommand calls the Fastly API to create an appropriate resource.
type CreateCommand struct {
	argparser.Base

	// Required.
	serviceVersion argparser.OptionalServiceVersion

	// Optional.
	autoClone     argparser.OptionalAutoClone
	conditionType argparser.OptionalString
	name          argparser.OptionalString
	priority      argparser.OptionalInt
	serviceName   argparser.OptionalServiceNameID
	statement     argparser.OptionalString
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent argparser.Registerer, g *global.Data) *CreateCommand {
	c := CreateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("create", "Create a condition on a Fastly service version").Alias("add")

	// Required flags
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
	c.CmdClause.Flag("name", "Condition name").Short('n').Action(c.name.Set).StringVar(&c.name.Value)
	c.CmdClause.Flag("priority", "Condition priority").Action(c.priority.Set).IntVar(&c.priority.Value)
	c.CmdClause.Flag("statement", "Condition statement").Action(c.statement.Set).StringVar(&c.statement.Value)
	c.CmdClause.Flag("type", "Condition type").HintOptions(ConditionTypes...).Action(c.conditionType.Set).EnumVar(&c.conditionType.Value, ConditionTypes...)
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
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
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

	input := fastly.CreateConditionInput{
		ServiceID:      serviceID,
		ServiceVersion: fastly.ToValue(serviceVersion.Number),
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
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": fastly.ToValue(serviceVersion.Number),
		})
		return err
	}

	text.Success(out,
		"Created condition %s (service %s version %d)",
		fastly.ToValue(r.Name),
		fastly.ToValue(r.ServiceID),
		fastly.ToValue(r.ServiceVersion),
	)
	return nil
}
