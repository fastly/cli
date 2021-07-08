package edgedictionary

import (
	"fmt"
	"io"
	"strconv"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v3/fastly"
)

// UpdateCommand calls the Fastly API to update a dictionary.
type UpdateCommand struct {
	cmd.Base
	manifest manifest.Data
	// TODO: make input consistent across commands (most are title case)
	input          fastly.UpdateDictionaryInput
	serviceVersion cmd.OptionalServiceVersion
	autoClone      cmd.OptionalAutoClone

	newname   cmd.OptionalString
	writeOnly cmd.OptionalString
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent cmd.Registerer, globals *config.Data) *UpdateCommand {
	var c UpdateCommand
	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("update", "Update name of dictionary on a Fastly service version").Alias("get")
	c.RegisterServiceIDFlag(&c.manifest.Flag.ServiceID)
	c.RegisterServiceVersionFlag(cmd.ServiceVersionFlagOpts{
		Dst: &c.serviceVersion.Value,
	})
	c.RegisterAutoCloneFlag(cmd.AutoCloneFlagOpts{
		Action: c.autoClone.Set,
		Dst:    &c.autoClone.Value,
	})
	c.CmdClause.Flag("name", "Old name of Dictionary").Short('n').Required().StringVar(&c.input.Name)
	c.CmdClause.Flag("new-name", "New name of Dictionary").Action(c.newname.Set).StringVar(&c.newname.Value)
	c.CmdClause.Flag("write-only", "Whether to mark this dictionary as write-only. Can be true or false (defaults to false)").Action(c.writeOnly.Set).StringVar(&c.writeOnly.Value)
	return &c
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, serviceVersion, err := cmd.ServiceDetails(cmd.ServiceDetailsOpts{
		AutoCloneFlag:      c.autoClone,
		Client:             c.Globals.Client,
		Manifest:           c.manifest,
		Out:                out,
		ServiceVersionFlag: c.serviceVersion,
		VerboseMode:        c.Globals.Flag.Verbose,
	})
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	c.input.ServiceID = serviceID
	c.input.ServiceVersion = serviceVersion.Number

	if !c.newname.WasSet && !c.writeOnly.WasSet {
		return errors.RemediationError{Inner: fmt.Errorf("error parsing arguments: required flag --new-name or --write-only not provided"), Remediation: "To fix this error, provide at least one of the aforementioned flags"}
	}

	if c.newname.WasSet {
		c.input.NewName = &c.newname.Value
	}

	if c.writeOnly.WasSet {
		writeOnly, err := strconv.ParseBool(c.writeOnly.Value)
		if err != nil {
			c.Globals.ErrLog.Add(err)
			return err
		}
		c.input.WriteOnly = fastly.CBool(writeOnly)
	}

	d, err := c.Globals.Client.UpdateDictionary(&c.input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	text.Success(out, "Updated dictionary %s (service %s version %d)", d.Name, d.ServiceID, d.ServiceVersion)

	if c.Globals.Verbose() {
		text.Output(out, "Service ID: %s", d.ServiceID)
		text.Output(out, "Version: %d", d.ServiceVersion)
		text.PrintDictionary(out, "", d)
	}

	return nil
}
