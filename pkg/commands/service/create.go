package service

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v7/fastly"
)

// CreateCommand calls the Fastly API to create services.
type CreateCommand struct {
	cmd.Base
	name string

	// optional
	stype   cmd.OptionalString
	comment cmd.OptionalString
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent cmd.Registerer, globals *config.Data) *CreateCommand {
	var c CreateCommand
	c.Globals = globals
	c.CmdClause = parent.Command("create", "Create a Fastly service").Alias("add")
	c.CmdClause.Flag("name", "Service name").Short('n').Required().StringVar(&c.name)
	c.CmdClause.Flag("type", `Service type. Can be one of "wasm" or "vcl", defaults to "vcl".`).Default("vcl").Action(c.stype.Set).EnumVar(&c.stype.Value, "wasm", "vcl")
	c.CmdClause.Flag("comment", "Human-readable comment").Action(c.comment.Set).StringVar(&c.comment.Value)
	return &c
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	input := fastly.CreateServiceInput{
		Name: &c.name,
	}
	if c.stype.WasSet {
		input.Type = &c.stype.Value
	}
	if c.comment.WasSet {
		input.Type = &c.comment.Value
	}
	s, err := c.Globals.APIClient.CreateService(&input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service Name": input.Name,
			"Type":         input.Type,
			"Comment":      input.Comment,
		})
		return err
	}

	text.Success(out, "Created service %s", s.ID)
	return nil
}
