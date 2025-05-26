package mcp

import (
	_ "embed"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/jedisct1/openapi-mcp/pkg/openapi2mcp"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
)

//go:embed fastly-openapi-mcp.yaml
var openapiSchema []byte

const mcpSchemaMountPath = "/mcp/api"

// APICommand handles the API MCP server subcommand.
type APICommand struct {
	argparser.Base
	httpAddr string
	servers  string // comma-separated list of MCP servers to enable (from positional arg)
}

// NewAPICommand returns a new API command registered under the parent.
func NewAPICommand(parent argparser.Registerer, g *global.Data) *APICommand {
	var c APICommand
	c.Globals = g
	c.CmdClause = parent.Command("api", `Start one or more MCP servers (comma-separated, e.g. 'api' or 'api,compute'). Only 'api' is supported currently.

Options:
  --http <address>   Serve MCP over HTTP on this address (e.g., :8080). Default is stdio.

Examples:
  # Run using stdio (default)
  fastly mcp api

  # Run using HTTP
  fastly mcp api --http :8080

  # Example JSON configuration for IDEs (e.g., VS Code, JetBrains)
  {
    "Fastly API": {
      "command": "fastly",
      "args": [
          "mcp", "api"
      ]
    }
  }
`)

	// Optional flags
	c.CmdClause.Flag("http", "Serve MCP over HTTP on this address (e.g., :8080). Default is stdio.").StringVar(&c.httpAddr)

	// Positional argument for servers (default: 'api')
	c.CmdClause.Arg("servers", "Comma-separated list of MCP servers to enable (default: 'api')").Default("api").StringVar(&c.servers)

	return &c
}

// Exec implements the command interface.
func (c *APICommand) Exec(_ io.Reader, out io.Writer) error {
	// Validate servers argument (only support 'api' for now)
	servers := strings.Split(c.servers, ",")
	for _, s := range servers {
		s = strings.TrimSpace(s)
		if s != "api" {
			return fmt.Errorf("unsupported MCP server: '%s' (only 'api' is supported at this time)", s)
		}
	}

	// Get the API token for authentication
	token, _ := c.Globals.Token()
	if token == "" {
		return fmt.Errorf(`no API token available

To use the Fastly API MCP server, you need to provide a Fastly API token.

You can set it in one of these ways:
1. Set the FASTLY_API_TOKEN environment variable:
   export FASTLY_API_TOKEN=your_token_here

2. Configure a profile with the fastly CLI:
   fastly profile create

3. Pass it when running the command:
   FASTLY_API_TOKEN=your_token_here fastly mcp api

You can get an API token from: https://manage.fastly.com/account/personal/tokens`)
	}

	// Get the API endpoint
	apiEndpoint, _ := c.Globals.APIEndpoint()

	// Configure authentication for the MCP server
	// The openapi-mcp library uses environment variables for authentication
	if err := os.Setenv("API_KEY", token); err != nil {
		return fmt.Errorf("failed to set API_KEY environment variable: %w", err)
	}

	// Set the base URL for API calls to match the Fastly API endpoint
	if err := os.Setenv("OPENAPI_BASE_URL", apiEndpoint); err != nil {
		return fmt.Errorf("failed to set OPENAPI_BASE_URL environment variable: %w", err)
	}

	// Load the OpenAPI specification from embedded data
	doc, err := openapi2mcp.LoadOpenAPISpecFromBytes(openapiSchema)
	if err != nil {
		return fmt.Errorf("failed to load OpenAPI spec: %w", err)
	}

	// Create a new MCP server instance with the correct arguments
	server := openapi2mcp.NewServer("fastly-api", doc.Info.Version, doc)

	// Mount the Fastly OpenAPI schema at /mcp/api (redundant, but for clarity)
	// If Mount is not needed, you can remove this block.
	// if err := server.Mount("/mcp/api", doc.Info.Version, doc); err != nil {
	// 	return fmt.Errorf("failed to mount OpenAPI schema: %w", err)
	// }

	// Start the server in the appropriate mode
	if c.httpAddr != "" {
		// HTTP mode
		fmt.Fprintf(out, "Starting Fastly API MCP server over HTTP...\n")
		fmt.Fprintf(out, "API endpoint: %s\n", apiEndpoint)
		fmt.Fprintf(out, "Authentication: API token configured\n")
		fmt.Fprintf(out, "Server listening on HTTP address: %s\n", c.httpAddr)

		mcpURL := openapi2mcp.GetSSEURL(c.httpAddr, mcpSchemaMountPath)
		fmt.Fprintf(out, "Use this URL in your MCP client: %s\n", mcpURL)

		if err := openapi2mcp.ServeHTTP(server, c.httpAddr, mcpSchemaMountPath); err != nil {
			return fmt.Errorf("MCP HTTP server error: %w", err)
		}
	} else {
		// Default stdio mode
		fmt.Fprint(os.Stderr, "Starting Fastly API MCP server...\n")
		fmt.Fprintf(os.Stderr, "API endpoint: %s\n", apiEndpoint)
		fmt.Fprint(os.Stderr, "Authentication: API token configured\n")
		fmt.Fprint(os.Stderr, "Server listening on stdio for MCP protocol messages...\n")

		if err := openapi2mcp.ServeStdio(server); err != nil {
			return fmt.Errorf("MCP server error: %w", err)
		}
	}

	return nil
}
