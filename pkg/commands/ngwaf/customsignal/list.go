package customsignal

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/go-fastly/v19/fastly/ngwaf/v1/scope"
	"github.com/fastly/go-fastly/v19/fastly/ngwaf/v1/signals"
)

// ListCommand calls the Fastly API to list all account-level custom signals for your API token.
type ListCommand struct {
	argparser.Base

	// Optional.
	argparser.JSONOutput
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	signals, err := signals.List(context.TODO(), fc, &signals.ListInput{
		Scope: &scope.Scope{
			Type:      scope.ScopeTypeAccount,
			AppliesTo: []string{"*"},
		},
	})
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, signals); ok {
		return err
	}

	text.PrintCustomSignalTbl(out, signals.Data)
	return nil
}
