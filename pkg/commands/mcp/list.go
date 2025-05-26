package mcp

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
)

// ListCommand handles listing available MCP server types.
type ListCommand struct {
	argparser.Base
	argparser.JSONOutput
}

// NewListCommand returns a new list command registered under the parent.
func NewListCommand(parent argparser.Registerer, g *global.Data) *ListCommand {
	var c ListCommand
	c.Globals = g
	c.CmdClause = parent.Command("list", "List available MCP server types")

	// Optional flags
	c.RegisterFlagBool(c.JSONFlag()) // --json

	return &c
}

// ServerType represents an available MCP server type.
type ServerType struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Available   bool   `json:"available"`
}

// Exec implements the command interface.
func (c *ListCommand) Exec(_ io.Reader, out io.Writer) error {
	serverTypes := []ServerType{
		{
			Name:        "api",
			Description: "MCP server for Fastly API interaction using OpenAPI specification",
			Available:   true,
		},
	}

	if ok, err := c.WriteJSON(out, serverTypes); ok {
		return err
	}

	fmt.Fprintln(out, "Available MCP server types:")
	fmt.Fprintln(out)

	for _, serverType := range serverTypes {
		status := "✓"
		if !serverType.Available {
			status = "✗"
		}
		fmt.Fprintf(out, "  %s %s - %s\n", status, serverType.Name, serverType.Description)
	}

	fmt.Fprintln(out)
	fmt.Fprintln(out, "Usage:")
	fmt.Fprintln(out, "  fastly mcp <server-type>")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Example:")
	fmt.Fprintln(out, "  fastly mcp api")

	return nil
}
