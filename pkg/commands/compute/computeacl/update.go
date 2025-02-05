package computeacl

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v9/fastly"
	"github.com/fastly/go-fastly/v9/fastly/computeacls"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// UpdateCommand calls the Fastly API to update a compute ACL.
type UpdateCommand struct {
	argparser.Base

	// Required.
	computeACLID string

	// Optional.
	file      argparser.OptionalString
	operation argparser.OptionalString
	prefix    argparser.OptionalString
	action    argparser.OptionalString
}

// operations is a list of supported operation options.
var operations = []string{"create", "update"}

// actions is a list of supported action options.
var actions = []string{"BLOCK", "ALLOW"}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent argparser.Registerer, g *global.Data) *UpdateCommand {
	c := UpdateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}

	c.CmdClause = parent.Command("update", "Update a compute ACL")

	// Required.
	c.CmdClause.Flag("acl-id", "Alphanumeric string identifying a compute ACL").Required().StringVar(&c.computeACLID)

	// Optional.
	c.CmdClause.Flag("file", "Batch update json passed as file path or content, e.g. $(< batch.json)").Action(c.file.Set).StringVar(&c.file.Value)
	c.CmdClause.Flag("operation", "Indicating that this entry is to be added to/updated in the ACL").HintOptions(operations...).EnumVar(&c.operation.Value, operations...)
	c.CmdClause.Flag("prefix", "An IP prefix defined in Classless Inter-Domain Routing (CIDR) format, i.e. a valid IP address (v4 or v6) followed by a forward slash (/) and a prefix length (0-32 or 0-128, depending on address family)").Action(c.prefix.Set).StringVar(&c.prefix.Value)
	c.CmdClause.Flag("action", "The action taken on the IP address").HintOptions(actions...).EnumVar(&c.action.Value, actions...)

	return &c
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	if c.file.WasSet {
		input, err := c.constructBatchInput()
		if err != nil {
			return err
		}

		err = computeacls.Update(fc, input)
		if err != nil {
			c.Globals.ErrLog.Add(err)
			return err
		}

		text.Success(out, "Updated %d compute ACL entries (id: %s)", len(input.Entries), c.computeACLID)
		return nil
	}

	input, err := c.constructInput()
	if err != nil {
		return err
	}

	err = computeacls.Update(fc, input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	text.Success(out, "Updated compute ACL entry (prefix: %s, id: %s)", c.prefix.Value, c.computeACLID)
	return nil
}

// constructBatchInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) constructBatchInput() (*computeacls.UpdateInput, error) {
	var input computeacls.UpdateInput

	input.ComputeACLID = &c.computeACLID

	s := argparser.Content(c.file.Value)
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

	if c.operation.Value == "" || c.prefix.Value == "" || c.action.Value == "" {
		return nil, fsterr.ErrInvalidComputeACLCombo
	}

	input.ComputeACLID = &c.computeACLID
	input.Entries = []*computeacls.BatchComputeACLEntry{
		{
			Prefix:    &c.prefix.Value,
			Action:    &c.action.Value,
			Operation: &c.operation.Value,
		},
	}

	return &input, nil
}
