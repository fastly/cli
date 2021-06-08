package domain

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v3/fastly"
)

// CreateCommand calls the Fastly API to create domains.
type CreateCommand struct {
	cmd.Base
	manifest       manifest.Data
	Input          fastly.CreateDomainInput
	serviceVersion cmd.OptionalServiceVersion
	autoClone      cmd.OptionalAutoClone
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent cmd.Registerer, globals *config.Data) *CreateCommand {
	var c CreateCommand
	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("create", "Create a domain on a Fastly service version").Alias("add")
	c.CmdClause.Flag("name", "Domain name").Short('n').Required().StringVar(&c.Input.Name)
	c.CmdClause.Flag("comment", "A descriptive note").StringVar(&c.Input.Comment)
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.SetServiceVersionFlag(cmd.ServiceVersionFlagOpts{
		Dst: &c.serviceVersion.Value,
	})
	c.SetAutoCloneFlag(cmd.AutoCloneFlagOpts{
		Action: c.autoClone.Set,
		Dst:    &c.autoClone.Value,
	})
	return &c
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, serviceVersion, err := cmd.ServiceDetails(cmd.ServiceDetailsOpts{
		Manifest:           c.manifest,
		ServiceVersionFlag: c.serviceVersion,
		AutoCloneFlag:      c.autoClone,
		VerboseMode:        c.Globals.Flag.Verbose,
		Out:                out,
		Client:             c.Globals.Client,
	})
	if err != nil {
		return err
	}

	c.Input.ServiceID = serviceID
	c.Input.ServiceVersion = serviceVersion.Number

	d, err := c.Globals.Client.CreateDomain(&c.Input)
	if err != nil {
		return err
	}

	text.Success(out, "Created domain %s (service %s version %d)", d.Name, d.ServiceID, d.ServiceVersion)
	return nil
}
