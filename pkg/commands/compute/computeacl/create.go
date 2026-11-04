package computeacl

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/go-fastly/v19/fastly/computeacls"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// CreateCommand calls the Fastly API to create a compute ACL.
type CreateCommand struct {
	argparser.Base

	// Required.
	NameFlag string `name:"name" required:"" help:"Name of the compute ACL"`

	// Optional.
	argparser.JSONOutput
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	acl, err := computeacls.Create(context.TODO(), fc, &computeacls.CreateInput{
		Name: &c.NameFlag,
	})
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, acl); ok {
		return err
	}

	text.Success(out, "Created compute ACL '%s' (id: %s)", acl.Name, acl.ComputeACLID)
	return nil
}
