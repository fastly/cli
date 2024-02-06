package service

import (
	"io"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// CreateCommand calls the Fastly API to create services.
type CreateCommand struct {
	argparser.Base

	// Optional.
	comment argparser.OptionalString
	name    argparser.OptionalString
	stype   argparser.OptionalString
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent argparser.Registerer, g *global.Data) *CreateCommand {
	c := CreateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("create", "Create a Fastly service").Alias("add")

	// Optional.
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

	text.Success(out, "Created service %s", fastly.ToValue(s.ServiceID))
	return nil
}
