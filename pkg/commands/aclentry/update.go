package aclentry

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/cmd"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
)

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent cmd.Registerer, g *global.Data, m manifest.Data) *UpdateCommand {
	c := UpdateCommand{
		Base: cmd.Base{
			Globals: g,
		},
		manifest: m,
	}
	c.CmdClause = parent.Command("update", "Update an ACL entry for a specified ACL")

	// Required.
	c.CmdClause.Flag("acl-id", "Alphanumeric string identifying a ACL").Required().StringVar(&c.aclID)

	// Optional.
	c.CmdClause.Flag("comment", "A freeform descriptive note").Action(c.comment.Set).StringVar(&c.comment.Value)
	c.CmdClause.Flag("file", "Batch update json passed as file path or content, e.g. $(< batch.json)").Action(c.file.Set).StringVar(&c.file.Value)
	c.CmdClause.Flag("id", "Alphanumeric string identifying an ACL Entry").Action(c.id.Set).StringVar(&c.id.Value)
	c.CmdClause.Flag("ip", "An IP address").Action(c.ip.Set).StringVar(&c.ip.Value)
	c.CmdClause.Flag("negated", "Whether to negate the match").Action(c.negated.Set).BoolVar(&c.negated.Value)
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
	c.CmdClause.Flag("subnet", "Number of bits for the subnet mask applied to the IP address").Action(c.subnet.Set).IntVar(&c.subnet.Value)

	return &c
}

// UpdateCommand calls the Fastly API to update an appropriate resource.
type UpdateCommand struct {
	cmd.Base

	aclID       string
	comment     cmd.OptionalString
	file        cmd.OptionalString
	id          cmd.OptionalString
	ip          cmd.OptionalString
	manifest    manifest.Data
	negated     cmd.OptionalBool
	serviceName cmd.OptionalServiceNameID
	subnet      cmd.OptionalInt
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	serviceID, source, flag, err := cmd.ServiceID(c.serviceName, c.manifest, c.Globals.APIClient, c.Globals.ErrLog)
	if err != nil {
		return err
	}
	if c.Globals.Verbose() {
		cmd.DisplayServiceID(serviceID, flag, source, out)
	}

	if c.file.WasSet {
		input, err := c.constructBatchInput(serviceID)
		if err != nil {
			return err
		}

		err = c.Globals.APIClient.BatchModifyACLEntries(input)
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
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

	a, err := c.Globals.APIClient.UpdateACLEntry(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
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
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"File": s,
		})
		return nil, err
	}

	if len(input.Entries) == 0 {
		err := fsterr.RemediationError{
			Inner:       fmt.Errorf("missing 'entries' %s", c.file.Value),
			Remediation: "Consult the API documentation for the JSON format: https://developer.fastly.com/reference/api/acls/acl-entry/#bulk-update-acl-entries",
		}
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
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
		return nil, fsterr.ErrNoID
	}

	input.ACLID = c.aclID
	input.ID = c.id.Value
	input.ServiceID = serviceID

	if c.comment.WasSet {
		input.Comment = &c.comment.Value
	}
	if c.ip.WasSet {
		input.IP = &c.ip.Value
	}
	if c.negated.WasSet {
		input.Negated = fastly.CBool(c.negated.Value)
	}
	if c.subnet.WasSet {
		input.Subnet = &c.subnet.Value
	}

	return &input, nil
}
