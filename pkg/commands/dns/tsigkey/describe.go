package tsigkey

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"
	"github.com/fastly/go-fastly/v19/fastly/dns/v1/tsigkeys"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// DescribeCommand calls the Fastly API to describe a TSIG key.
type DescribeCommand struct {
	argparser.Base

	// Required.
	TSIGKeyID string `required:"" help:"The TSIG key ID to describe."`

	// Optional.
	argparser.JSONOutput
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	k, err := tsigkeys.Get(context.TODO(), fc, &tsigkeys.GetInput{
		TSIGKeyID: &c.TSIGKeyID,
	})
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"TSIG Key ID": c.TSIGKeyID,
		})
		return err
	}

	if ok, err := c.WriteJSON(out, k); ok {
		return err
	}

	text.PrintTSIGKey(out, "", k)
	return nil
}
