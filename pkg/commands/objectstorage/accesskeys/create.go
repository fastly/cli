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

// CreateCommand calls the Fastly API to create an access key.
type CreateCommand struct {
	argparser.Base

	// Required.
	Description string `required:"" help:"Description of the access key"`
	Permission  string `required:"" help:"Permissions to be given to the access key (read-write-admin, read-only-admin, read-write-objects, read-only-objects)"`

	// Optional.
	Buckets []string `name:"bucket" sep:"none" help:"Bucket to be associated with the access key. Set flag multiple times to include multiple buckets. If omitted, all buckets are associated"`

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

	accessKey, err := accesskeys.Create(context.TODO(), fc, &accesskeys.CreateInput{
		Description: &c.Description,
		Permission:  &c.Permission,
		Buckets:     &c.Buckets,
	})
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, accessKey); ok {
		return err
	}

	text.Success(out, "Created access key (id: %s, secret: %s)", accessKey.AccessKeyID, accessKey.SecretKey)
	return nil
}
