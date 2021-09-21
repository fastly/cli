package aclentry

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/commands/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v4/fastly"
)

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent cmd.Registerer, globals *config.Data) *UpdateCommand {
	var c UpdateCommand
	c.CmdClause = parent.Command("update", "Update an ACL entry for a specified ACL")
	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)

	// Required flags
	c.CmdClause.Flag("acl-id", "Alphanumeric string identifying a ACL").Required().StringVar(&c.aclID)

	// Optional flags
	c.CmdClause.Flag("comment", "A freeform descriptive note").Action(c.comment.Set).StringVar(&c.comment.Value)
	c.CmdClause.Flag("file", "Batch update json passed as file path or content, e.g. $(< batch.json)").Action(c.file.Set).StringVar(&c.file.Value)
	c.CmdClause.Flag("id", "Alphanumeric string identifying an ACL Entry").Action(c.id.Set).StringVar(&c.id.Value)
	c.CmdClause.Flag("ip", "An IP address").Action(c.ip.Set).StringVar(&c.ip.Value)
	c.CmdClause.Flag("negated", "Whether to negate the match").Action(c.negated.Set).BoolVar(&c.negated.Value)
	c.RegisterServiceIDFlag(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("subnet", "Number of bits for the subnet mask applied to the IP address").Action(c.subnet.Set).IntVar(&c.subnet.Value)

	return &c
}

// UpdateCommand calls the Fastly API to update an appropriate resource.
type UpdateCommand struct {
	cmd.Base

	aclID    string
	comment  cmd.OptionalString
	file     cmd.OptionalString
	id       cmd.OptionalString
	ip       cmd.OptionalString
	manifest manifest.Data
	negated  cmd.OptionalBool
	subnet   cmd.OptionalInt
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, source := c.manifest.ServiceID()
	if c.Globals.Verbose() {
		cmd.DisplayServiceID(serviceID, source, out)
	}
	if source == manifest.SourceUndefined {
		err := errors.ErrNoServiceID
		c.Globals.ErrLog.Add(err)
		return err
	}

	if c.file.WasSet {
		input, err := c.constructBatchInput(serviceID)
		if err != nil {
			return err
		}

		err = c.Globals.Client.BatchModifyACLEntries(input)
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
				"Service ID": serviceID,
			})
			return err
		}

		text.Success(out, "Updated %d ACL entries (service: %s)", len(input.Entries), serviceID)
		return nil
	}

	input, err := c.constructInput(serviceID)
	if err != nil {
		return err
	}

	a, err := c.Globals.Client.UpdateACLEntry(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
			"Service ID": serviceID,
		})
		return err
	}

	text.Success(out, "Updated ACL entry '%s' (ip: %s, service: %s)", a.ID, a.IP, a.ServiceID)
	return nil
}

// constructBatchInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) constructBatchInput(serviceID string) (*fastly.BatchModifyACLEntriesInput, error) {
	var input fastly.BatchModifyACLEntriesInput

	input.ACLID = c.aclID
	input.ServiceID = serviceID

	s := cmd.Content(c.file.Value)
	bs := []byte(s)

	err := json.Unmarshal(bs, &input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
			"File": s,
		})
		return nil, err
	}

	if len(input.Entries) == 0 {
		err := errors.RemediationError{
			Inner:       fmt.Errorf("missing 'entries' %s", c.file.Value),
			Remediation: "Consult the API documentation for the JSON format: https://developer.fastly.com/reference/api/acls/acl-entry/#bulk-update-acl-entries",
		}
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
			"File": string(bs),
		})
		return nil, err
	}

	return &input, nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) constructInput(serviceID string) (*fastly.UpdateACLEntryInput, error) {
	var input fastly.UpdateACLEntryInput

	if !c.id.WasSet {
		return nil, errors.ErrNoID
	}

	input.ACLID = c.aclID
	input.ID = c.id.Value
	input.ServiceID = serviceID

	if c.comment.WasSet {
		input.Comment = fastly.String(c.comment.Value)
	}
	if c.ip.WasSet {
		input.IP = fastly.String(c.ip.Value)
	}
	if c.negated.WasSet {
		input.Negated = fastly.Bool(c.negated.Value)
	}
	if c.subnet.WasSet {
		input.Subnet = fastly.Int(c.subnet.Value)
	}

	return &input, nil
}
