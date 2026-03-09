package accesskeys

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v13/fastly"

	"github.com/fastly/go-fastly/v13/fastly/objectstorage/accesskeys"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// CreateCommand calls the Fastly API to create an access key.
type CreateCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	description string
	permission  string

	// Optional.
	buckets []string
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent argparser.Registerer, g *global.Data) *CreateCommand {
	c := CreateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}

	c.CmdClause = parent.Command("create", "Create an access key")

	// Required.
	c.CmdClause.Flag("description", "Description of the access key").Required().StringVar(&c.description)
	c.CmdClause.Flag("permission", "Permissions to be given to the access key (read-write-admin, read-only-admin, read-write-objects, read-only-objects)").Required().StringVar(&c.permission)

	// Optional.
	c.CmdClause.Flag("bucket", "Bucket to be associated with the access key. Set flag multiple times to include multiple buckets. If omitted, all buckets are associated").StringsVar(&c.buckets)
	c.RegisterFlagBool(c.JSONFlag())

	return &c
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
		Description: &c.description,
		Permission:  &c.permission,
		Buckets:     &c.buckets,
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
