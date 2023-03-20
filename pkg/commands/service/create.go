package service

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v7/fastly"
)

// CreateCommand calls the Fastly API to create services.
type CreateCommand struct {
	cmd.Base

	// optional
	comment cmd.OptionalString
	name    cmd.OptionalString
	stype   cmd.OptionalString
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent cmd.Registerer, g *global.Data) *CreateCommand {
	c := CreateCommand{
		Base: cmd.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("create", "Create a Fastly service").Alias("add")

	// optional
	c.CmdClause.Flag("comment", "Human-readable comment").Action(c.comment.Set).StringVar(&c.comment.Value)
	c.CmdClause.Flag("name", "Service name").Short('n').Action(c.name.Set).StringVar(&c.name.Value)
	c.CmdClause.Flag("type", `Service type. Can be one of "wasm" or "vcl", defaults to "vcl".`).Default("vcl").Action(c.stype.Set).EnumVar(&c.stype.Value, "wasm", "vcl")
	return &c
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	input := fastly.CreateServiceInput{}

	if c.name.WasSet {
		input.Name = &c.name.Value
	}
	if c.comment.WasSet {
		input.Comment = &c.comment.Value
	}
	if c.stype.WasSet {
		input.Type = &c.stype.Value
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
