package dictionary

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v7/fastly"
)

// CreateCommand calls the Fastly API to create a service.
type CreateCommand struct {
	cmd.Base
	manifest manifest.Data

	// required
	serviceVersion cmd.OptionalServiceVersion

	// optional
	autoClone   cmd.OptionalAutoClone
	name        cmd.OptionalString
	serviceName cmd.OptionalServiceNameID
	writeOnly   cmd.OptionalBool
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent cmd.Registerer, g *global.Data, m manifest.Data) *CreateCommand {
	c := CreateCommand{
		Base: cmd.Base{
			Globals: g,
		},
		manifest: m,
	}
	c.CmdClause = parent.Command("create", "Create a Fastly edge dictionary on a Fastly service version")

	// required
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
	c.CmdClause.Flag("name", "Name of Dictionary").Short('n').Action(c.name.Set).StringVar(&c.name.Value)
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
	c.CmdClause.Flag("write-only", "Whether to mark this dictionary as write-only").Action(c.writeOnly.Set).BoolVar(&c.writeOnly.Value)
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
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": errors.ServiceVersion(serviceVersion),
		})
		return err
	}

	input := fastly.CreateDictionaryInput{
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion.Number,
	}

	if c.name.WasSet {
		input.Name = &c.name.Value
	}
	if c.writeOnly.WasSet {
		input.WriteOnly = fastly.CBool(c.writeOnly.Value)
	}

	d, err := c.Globals.APIClient.CreateDictionary(&input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": serviceVersion.Number,
		})
		return err
	}

	var writeOnlyOutput string
	if d.WriteOnly {
		writeOnlyOutput = "as write-only "
	}

	text.Success(out, "Created dictionary %s %s(id %s, service %s, version %d)", d.Name, writeOnlyOutput, d.ID, d.ServiceID, d.ServiceVersion)
	return nil
}
