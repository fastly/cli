package dashboard

import (
	"io"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent argparser.Registerer, globals *global.Data) *CreateCommand {
	var c CreateCommand
	c.CmdClause = parent.Command("create", "Create a custom dashboard").Alias("add")
	c.Globals = globals

	// Required flags
	c.CmdClause.Flag("name", "A human-readable name for the dashboard").Short('n').Required().StringVar(&c.name) // --name

	// Optional flags
	c.RegisterFlagBool(c.JSONFlag())                                                                                                  // --json
	c.CmdClause.Flag("description", "A short description of the dashboard").Action(c.description.Set).StringVar(&c.description.Value) // --description

	return &c
}

// CreateCommand calls the Fastly API to create an appropriate resource.
type CreateCommand struct {
	argparser.Base
	argparser.JSONOutput

	name        string
	description argparser.OptionalString
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input := c.constructInput()
	dashboard, err := c.Globals.APIClient.CreateObservabilityCustomDashboard(input)
	if err != nil {
		return err
	}

	if ok, err := c.WriteJSON(out, dashboard); ok {
		return err
	}

	text.Success(out, `Created Custom Dashboard "%s" (id: %s)`, dashboard.Name, dashboard.ID)
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *CreateCommand) constructInput() *fastly.CreateObservabilityCustomDashboardInput {
	input := fastly.CreateObservabilityCustomDashboardInput{
		Name:  c.name,
		Items: []fastly.DashboardItem{},
	}

	if c.description.WasSet {
		input.Description = &c.description.Value
	}

	return &input
}
