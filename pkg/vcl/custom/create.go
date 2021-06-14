package custom

import (
	"io"
	"path/filepath"
	"strings"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v3/fastly"
)

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent cmd.Registerer, globals *config.Data) *CreateCommand {
	var c CreateCommand
	c.CmdClause = parent.Command("create", "Upload a VCL for a particular service and version").Alias("add")
	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)

	// Required flags
	c.CmdClause.Flag("file", "The path to the VCL").Required().StringVar(&c.file)
	c.SetServiceVersionFlag(cmd.ServiceVersionFlagOpts{
		Dst: &c.serviceVersion.Value,
	})

	// Optional flags
	c.SetAutoCloneFlag(cmd.AutoCloneFlagOpts{
		Action: c.autoClone.Set,
		Dst:    &c.autoClone.Value,
	})
	c.CmdClause.Flag("main", "Whether the VCL is the 'main' entrypoint").Action(c.main.Set).BoolVar(&c.main.Value)
	c.CmdClause.Flag("name", "The name of the VCL (defaults to --file filename without extension)").Action(c.name.Set).StringVar(&c.name.Value)
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)

	return &c
}

// CreateCommand calls the Fastly API to create an appropriate resource.
type CreateCommand struct {
	cmd.Base

	autoClone      cmd.OptionalAutoClone
	file           string
	main           cmd.OptionalBool
	manifest       manifest.Data
	name           cmd.OptionalString
	serviceVersion cmd.OptionalServiceVersion
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, serviceVersion, err := cmd.ServiceDetails(cmd.ServiceDetailsOpts{
		AutoCloneFlag:      c.autoClone,
		Client:             c.Globals.Client,
		Manifest:           c.manifest,
		Out:                out,
		ServiceVersionFlag: c.serviceVersion,
		VerboseMode:        c.Globals.Flag.Verbose,
	})
	if err != nil {
		return err
	}

	input := c.createInput(serviceID, serviceVersion.Number)

	v, err := c.Globals.Client.CreateVCL(input)
	if err != nil {
		return err
	}

	text.Success(out, "Created custom VCL %s (service %s version %d main %t)", v.Name, v.ServiceID, v.ServiceVersion, v.Main)
	return nil
}

// createInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *CreateCommand) createInput(serviceID string, serviceVersion int) *fastly.CreateVCLInput {
	var input fastly.CreateVCLInput

	input.ServiceID = serviceID
	input.ServiceVersion = serviceVersion
	input.Content = c.file

	if c.main.WasSet {
		input.Main = c.main.Value
	}
	if c.name.WasSet {
		input.Name = c.name.Value
	} else {
		basename := filepath.Base(c.file)
		input.Name = strings.TrimSuffix(basename, filepath.Ext(basename))
	}

	return &input
}
