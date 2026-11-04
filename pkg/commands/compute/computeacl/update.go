package computeacl

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/go-fastly/v19/fastly/computeacls"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// UpdateCommand calls the Fastly API to update a compute ACL.
type UpdateCommand struct {
	argparser.Base

	// Required.
	ComputeACLID string `name:"acl-id" required:"" help:"Alphanumeric string identifying a compute ACL"`

	// Optional.
	File          argparser.OptionalString `help:"Batch update JSON file passed as file path or content, e.g. $(< batch.json)"`
	OperationFlag string                   `name:"operation" default:"" enum:"create,update," hints:"create,update" help:"Indicating that this entry is to be added to/updated in the ACL"`
	operation     argparser.OptionalString
	Prefix        argparser.OptionalString `help:"An IP prefix defined in Classless Inter-Domain Routing (CIDR) format, i.e. a valid IP address (v4 or v6) followed by a forward slash (/) and a prefix length (0-32 or 0-128, depending on address family)"`
	action        argparser.OptionalString
	ActionFlag    string `name:"action" default:"" enum:"BLOCK,ALLOW," hints:"BLOCK,ALLOW" help:"The action taken on the IP address"`
}

// operations is a list of supported operation options.
var operations = []string{"create", "update"}

// actions is a list of supported action options.
var actions = []string{"BLOCK", "ALLOW"}

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *UpdateCommand) AfterApply() error {
	c.operation.Value = c.OperationFlag
	c.action.Value = c.ActionFlag
	return nil
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	if c.File.WasSet {
		input, err := c.constructBatchInput()
		if err != nil {
			return err
		}

		err = computeacls.Update(context.TODO(), fc, input)
		if err != nil {
			c.Globals.ErrLog.Add(err)
			return err
		}

		text.Success(out, "Updated %d compute ACL entries (id: %s)", len(input.Entries), c.ComputeACLID)
		return nil
	}

	input, err := c.constructInput()
	if err != nil {
		return err
	}

	err = computeacls.Update(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	text.Success(out, "Updated compute ACL entry (prefix: %s, id: %s)", c.Prefix.Value, c.ComputeACLID)
	return nil
}

// constructBatchInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) constructBatchInput() (*computeacls.UpdateInput, error) {
	var input computeacls.UpdateInput

	input.ComputeACLID = &c.ComputeACLID

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
			Remediation: "Consult the API documentation for the JSON format: https://www.fastly.com/documentation/reference/api/acls/acls/#compute-acl-update-acls",
		}
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"File": string(bs),
		})
		return nil, err
	}

	return &input, nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) constructInput() (*computeacls.UpdateInput, error) {
	var input computeacls.UpdateInput

	if c.operation.Value == "" || c.Prefix.Value == "" || c.action.Value == "" {
		return nil, fsterr.ErrInvalidComputeACLCombo
	}

	input.ComputeACLID = &c.ComputeACLID
	input.Entries = []*computeacls.BatchComputeACLEntry{
		{
			Prefix:    &c.Prefix.Value,
			Action:    &c.action.Value,
			Operation: &c.operation.Value,
		},
	}

	return &input, nil
}
