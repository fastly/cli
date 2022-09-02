package service

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v6/fastly"
)

// CreateCommand calls the Fastly API to create services.
type CreateCommand struct {
	cmd.Base
	Input fastly.CreateServiceInput
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent cmd.Registerer, globals *config.Data) *CreateCommand {
	var c CreateCommand
	c.Globals = globals
	c.CmdClause = parent.Command("create", "Create a Fastly service").Alias("add")
	c.CmdClause.Flag("name", "Service name").Short('n').Required().StringVar(&c.Input.Name)
	c.CmdClause.Flag("type", `Service type. Can be one of "wasm" or "vcl", defaults to "vcl".`).Default("vcl").EnumVar(&c.Input.Type, "wasm", "vcl")
	c.CmdClause.Flag("comment", "Human-readable comment").StringVar(&c.Input.Comment)
	return &c
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(in io.Reader, out io.Writer) error {
	s, err := c.Globals.APIClient.CreateService(&c.Input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service Name": c.Input.Name,
			"Type":         c.Input.Type,
			"Comment":      c.Input.Comment,
		})
		return err
	}

	text.Success(out, "Created service %s", s.ID)
	return nil
}
