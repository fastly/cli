package aclentry

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *UpdateCommand) AfterApply() error {
	if c.ServiceID != nil {
		c.Globals.Manifest.Flag.ServiceID = *c.ServiceID
	}
	return nil
}

// UpdateCommand calls the Fastly API to update an appropriate resource.
type UpdateCommand struct {
	argparser.Base

	// Required.
	ACLID string `name:"acl-id" required:"" help:"Alphanumeric string identifying a ACL"`

	// Optional.
	Comment     argparser.OptionalString        `help:"A freeform descriptive note"`
	File        argparser.OptionalString        `help:"Batch update json passed as file path or content, e.g. $(< batch.json)"`
	ID          argparser.OptionalString        `help:"Alphanumeric string identifying an ACL Entry"`
	IP          argparser.OptionalString        `help:"An IP address"`
	Negated     argparser.OptionalBool          `help:"Whether to negate the match"`
	ServiceID   *string                         `short:"s" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`
	ServiceName argparser.OptionalServiceNameID `help:"The name of the service"`
	Subnet      argparser.OptionalInt           `help:"Number of bits for the subnet mask applied to the IP address"`
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	serviceID, source, flag, err := argparser.ServiceID(c.ServiceName, *c.Globals.Manifest, c.Globals.APIClient, c.Globals.ErrLog)
	if err != nil {
		return err
	}
	if c.Globals.Verbose() {
		argparser.DisplayServiceID(serviceID, flag, source, out)
	}

	if c.File.WasSet {
		input, err := c.constructBatchInput(serviceID)
		if err != nil {
			return err
		}

		err = c.Globals.APIClient.BatchModifyACLEntries(context.TODO(), input)
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

	a, err := c.Globals.APIClient.UpdateACLEntry(context.TODO(), input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID": serviceID,
		})
		return err
	}

	text.Success(out, "Updated ACL entry '%s' (ip: %s, service: %s)", fastly.ToValue(a.EntryID), fastly.ToValue(a.IP), fastly.ToValue(a.ServiceID))
	return nil
}

// constructBatchInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) constructBatchInput(serviceID string) (*fastly.BatchModifyACLEntriesInput, error) {
	var input fastly.BatchModifyACLEntriesInput

	input.ACLID = c.ACLID
	input.ServiceID = serviceID

	s := argparser.Content(c.File.Value)
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
			Inner:       fmt.Errorf("missing 'entries' %s", c.File.Value),
			Remediation: "Consult the API documentation for the JSON format: https://www.fastly.com/documentation/reference/api/acls/acl-entry#bulk-update-acl-entries",
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

	if !c.ID.WasSet {
		return nil, fsterr.ErrNoID
	}

	input.ACLID = c.ACLID
	input.EntryID = c.ID.Value
	input.ServiceID = serviceID

	if c.Comment.WasSet {
		input.Comment = &c.Comment.Value
	}
	if c.IP.WasSet {
		input.IP = &c.IP.Value
	}
	if c.Negated.WasSet {
		input.Negated = new(fastly.Compatibool(c.Negated.Value))
	}
	if c.Subnet.WasSet {
		input.Subnet = &c.Subnet.Value
	}

	return &input, nil
}
