package dictionary

import (
	"io"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// CreateCommand calls the Fastly API to create a service.
type CreateCommand struct {
	argparser.Base

	// Required.
	serviceVersion argparser.OptionalServiceVersion

	// Optional.
	autoClone   argparser.OptionalAutoClone
	name        argparser.OptionalString
	serviceName argparser.OptionalServiceNameID
	writeOnly   argparser.OptionalBool
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent argparser.Registerer, g *global.Data) *CreateCommand {
	c := CreateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("create", "Create a Fastly edge dictionary on a Fastly service version")

	// Required.
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagVersionName,
		Description: argparser.FlagVersionDesc,
		Dst:         &c.serviceVersion.Value,
		Required:    true,
	})

	// Optional.
	c.RegisterAutoCloneFlag(argparser.AutoCloneFlagOpts{
		Action: c.autoClone.Set,
		Dst:    &c.autoClone.Value,
	})
	c.CmdClause.Flag("name", "Name of Dictionary").Short('n').Action(c.name.Set).StringVar(&c.name.Value)
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
	c.CmdClause.Flag("write-only", "Whether to mark this dictionary as write-only").Action(c.writeOnly.Set).BoolVar(&c.writeOnly.Value)
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

	serviceVersionNumber := fastly.ToValue(serviceVersion.Number)

	input := fastly.CreateDictionaryInput{
		ServiceID:      serviceID,
		ServiceVersion: serviceVersionNumber,
	}
	if c.name.WasSet {
		input.Name = &c.name.Value
	}
	if c.writeOnly.WasSet {
		input.WriteOnly = fastly.ToPointer(fastly.Compatibool(c.writeOnly.Value))
	}

	d, err := c.Globals.APIClient.CreateDictionary(&input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": serviceVersionNumber,
		})
		return err
	}

	var writeOnlyOutput string
	if fastly.ToValue(d.WriteOnly) {
		writeOnlyOutput = "as write-only "
	}

	text.Success(out, "Created dictionary %s %s(id %s, service %s, version %d)", fastly.ToValue(d.Name), writeOnlyOutput, fastly.ToValue(d.DictionaryID), fastly.ToValue(d.ServiceID), fastly.ToValue(d.ServiceVersion))
	return nil
}
