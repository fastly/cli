package dashboard

import (
	"errors"
	"io"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent argparser.Registerer, globals *global.Data) *ListCommand {
	var c ListCommand
	c.CmdClause = parent.Command("list", "List custom dashboards")
	c.Globals = globals

	// Optional Flags
	c.RegisterFlagBool(c.JSONFlag()) // --json

	return &c
}

// ListCommand calls the Fastly API to list appropriate resources.
type ListCommand struct {
	argparser.Base
	argparser.JSONOutput
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(in io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}
	for {
		return nil
	}
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *ListCommand) constructInput() (*fastly.ListObservabilityCustomDashboardsInput, error) {
	var input fastly.ListObservabilityCustomDashboardsInput

	return &input, nil
}
