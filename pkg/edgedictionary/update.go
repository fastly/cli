package edgedictionary

import (
	"fmt"
	"io"
	"strconv"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v3/fastly"
)

// UpdateCommand calls the Fastly API to update a dictionary.
type UpdateCommand struct {
	common.Base
	manifest manifest.Data
	// TODO: make input consistent across commands (most are title case)
	input          fastly.UpdateDictionaryInput
	serviceVersion common.OptionalServiceVersion
	autoClone      common.OptionalAutoClone

	newname   common.OptionalString
	writeOnly common.OptionalString
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent common.Registerer, globals *config.Data) *UpdateCommand {
	var c UpdateCommand
	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("update", "Update name of dictionary on a Fastly service version").Alias("get")
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.NewServiceVersionFlag(common.ServiceVersionFlagOpts{Dst: &c.serviceVersion.Value})
	c.NewAutoCloneFlag(c.autoClone.Set, &c.autoClone.Value)
	c.CmdClause.Flag("name", "Old name of Dictionary").Short('n').Required().StringVar(&c.input.Name)
	c.CmdClause.Flag("new-name", "New name of Dictionary").Action(c.newname.Set).StringVar(&c.newname.Value)
	c.CmdClause.Flag("write-only", "Whether to mark this dictionary as write-only. Can be true or false (defaults to false)").Action(c.writeOnly.Set).StringVar(&c.writeOnly.Value)
	return &c
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return errors.ErrNoServiceID
	}
	c.input.ServiceID = serviceID

	v, err := c.serviceVersion.Parse(c.input.ServiceID, c.Globals.Client)
	if err != nil {
		return err
	}
	v, err = c.autoClone.Parse(v, c.input.ServiceID, c.Globals.Client)
	if err != nil {
		return err
	}
	c.input.ServiceVersion = v.Number

	if !c.newname.WasSet && !c.writeOnly.WasSet {
		return errors.RemediationError{Inner: fmt.Errorf("error parsing arguments: required flag --new-name or --write-only not provided"), Remediation: "To fix this error, provide at least one of the aforementioned flags"}
	}

	if c.newname.WasSet {
		c.input.NewName = &c.newname.Value
	}

	if c.writeOnly.WasSet {
		writeOnly, err := strconv.ParseBool(c.writeOnly.Value)
		if err != nil {
			return err
		}
		c.input.WriteOnly = fastly.CBool(writeOnly)
	}

	d, err := c.Globals.Client.UpdateDictionary(&c.input)
	if err != nil {
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
