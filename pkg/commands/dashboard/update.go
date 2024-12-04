package dashboard

import (
	"io"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
)

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent argparser.Registerer, globals *global.Data) *UpdateCommand {
	var c UpdateCommand
	c.CmdClause = parent.Command("update", "Update a custom dashboard")
	c.Globals = globals

	// Required flags

	// Optional flags
	c.RegisterFlagBool(c.JSONFlag()) // --json

	return &c
}

// UpdateCommand calls the Fastly API to update an appropriate resource.
type UpdateCommand struct {
	argparser.Base
	argparser.JSONOutput
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(in io.Reader, out io.Writer) error {
	// text.Success(out, "Updated <...> '%s' (service: %s, version: %d)", r.<...>, r.ServiceID, r.ServiceVersion)
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) constructInput() *fastly.UpdateObservabilityCustomDashboardInput {
	var input fastly.UpdateObservabilityCustomDashboardInput

	return &input
}
