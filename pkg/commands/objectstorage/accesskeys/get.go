package accesskeys

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/go-fastly/v19/fastly/objectstorage/accesskeys"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// GetCommand calls the Fastly API to get an access key.
type GetCommand struct {
	argparser.Base

	// Required.
	AccessKeyID string `name:"ak-id" required:"" help:"Access key ID"`

	// Optional.
	argparser.JSONOutput
}

// Exec invokes the application logic for the command.
func (c *GetCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	accessKey, err := accesskeys.Get(context.TODO(), fc, &accesskeys.GetInput{
		AccessKeyID: &c.AccessKeyID,
	})
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, accessKey); ok {
		return err
	}

	text.PrintAccessKey(out, accessKey)
	return nil
}
