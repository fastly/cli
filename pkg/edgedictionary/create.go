package edgedictionary

import (
	"io"
	"strconv"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v3/fastly"
)

// CreateCommand calls the Fastly API to create a service.
type CreateCommand struct {
	cmd.Base
	manifest       manifest.Data
	Input          fastly.CreateDictionaryInput
	serviceVersion cmd.OptionalServiceVersion
	autoClone      cmd.OptionalAutoClone

	writeOnly cmd.OptionalString
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent cmd.Registerer, globals *config.Data) *CreateCommand {
	var c CreateCommand
	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("create", "Create a Fastly edge dictionary on a Fastly service version")
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.SetServiceVersionFlag(cmd.ServiceVersionFlagOpts{
		Dst: &c.serviceVersion.Value,
	})
	c.SetAutoCloneFlag(cmd.AutoCloneFlagOpts{
		Action: c.autoClone.Set,
		Dst:    &c.autoClone.Value,
	})
	c.CmdClause.Flag("name", "Name of Dictionary").Short('n').Required().StringVar(&c.Input.Name)
	c.CmdClause.Flag("write-only", "Whether to mark this dictionary as write-only. Can be true or false (defaults to false)").Action(c.writeOnly.Set).StringVar(&c.writeOnly.Value)
	return &c
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, serviceVersion, err := cmd.ServiceDetails(c.manifest, c.serviceVersion, c.autoClone, c.Globals.Flag.Verbose, out, c.Globals.Client)
	if err != nil {
		return err
	}

	c.Input.ServiceID = serviceID
	c.Input.ServiceVersion = serviceVersion.Number

	if c.writeOnly.WasSet {
		writeOnly, err := strconv.ParseBool(c.writeOnly.Value)
		if err != nil {
			return err
		}
		c.Input.WriteOnly = fastly.Compatibool(writeOnly)
	}

	d, err := c.Globals.Client.CreateDictionary(&c.Input)
	if err != nil {
		return err
	}

	var writeOnlyOutput string
	if d.WriteOnly {
		writeOnlyOutput = "as write-only "
	}

	text.Success(out, "Created dictionary %s %s(service %s version %d)", d.Name, writeOnlyOutput, d.ServiceID, d.ServiceVersion)
	return nil
}
