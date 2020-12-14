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
	"github.com/fastly/go-fastly/v2/fastly"
)

// UpdateCommand calls the Fastly API to update a dictionary.
type UpdateCommand struct {
	common.Base
	manifest manifest.Data
	input    fastly.UpdateDictionaryInput

	newname   common.OptionalString
	writeOnly common.OptionalString
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent common.Registerer, globals *config.Data) *UpdateCommand {
	var c UpdateCommand
	c.Globals = globals
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("update", "Update name of dictionary on a Fastly service version").Alias("get")
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.input.ServiceVersion)
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
