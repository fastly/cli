package user

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v7/fastly"
)

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *CreateCommand {
	var c CreateCommand
	c.CmdClause = parent.Command("create", "Create a user of the Fastly API and web interface").Alias("add")
	c.Globals = globals
	c.manifest = data

	// Required flags
	c.CmdClause.Flag("login", "The login associated with the user (typically, an email address)").Required().StringVar(&c.login)
	c.CmdClause.Flag("name", "The real life name of the user").Required().StringVar(&c.name)

	// Optional flags
	c.CmdClause.Flag("role", "The permissions role assigned to the user. Can be user, billing, engineer, or superuser").EnumVar(&c.role, "user", "billing", "engineer", "superuser")

	return &c
}

// CreateCommand calls the Fastly API to create an appropriate resource.
type CreateCommand struct {
	cmd.Base

	login    string
	manifest manifest.Data
	name     string
	role     string
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	_, s := c.Globals.Token()
	if s == config.SourceUndefined {
		return errors.ErrNoToken
	}

	input := c.constructInput()

	r, err := c.Globals.APIClient.CreateUser(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"User Login": c.login,
			"User Name":  c.name,
		})
		return err
	}

	text.Success(out, "Created user '%s' (role: %s)", r.Name, r.Role)
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *CreateCommand) constructInput() *fastly.CreateUserInput {
	var input fastly.CreateUserInput

	input.Login = c.login
	input.Name = c.name
	if c.role == "" {
		input.Role = c.role
	}

	return &input
}
