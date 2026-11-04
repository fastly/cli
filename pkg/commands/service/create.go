package service

import (
	"context"
	"fmt"
	"io"
	"slices"

	"github.com/alecthomas/kong"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/internal/kongcompat"
	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/text"
)

// CreateCommand calls the Fastly API to create services.
type CreateCommand struct {
	argparser.Base

	// Optional.
	Comment  argparser.OptionalString `help:"Human-readable comment"`
	NameFlag argparser.OptionalString `name:"name" short:"n" help:"Service name"`
	Stype    argparser.OptionalString `name:"type" default:"vcl" hints:"wasm,vcl" help:"Service type. Can be one of \"wasm\" or \"vcl\", defaults to \"vcl\"."`
}

// AfterApply checks the values kong cannot check by itself.
func (c *CreateCommand) AfterApply(ctx *kong.Context) error {
	if (c.Stype.Value != "" || kongcompat.Given(ctx, "type")) && !slices.Contains([]string{"wasm", "vcl"}, c.Stype.Value) {
		return fmt.Errorf("enum value must be one of wasm,vcl, got '%s'", c.Stype.Value)
	}
	return nil
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	input := fastly.CreateServiceInput{}

	if c.NameFlag.WasSet {
		input.Name = &c.NameFlag.Value
	}
	if c.Comment.WasSet {
		input.Comment = &c.Comment.Value
	}
	if c.Stype.WasSet {
		input.Type = &c.Stype.Value
	}
	s, err := c.Globals.APIClient.CreateService(context.TODO(), &input)
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
