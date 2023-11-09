package whoami

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"

	"github.com/fastly/cli/pkg/api/undocumented"
	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/useragent"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	cmd.Base
}

// NewRootCommand returns a new command registered in the parent.
func NewRootCommand(parent cmd.Registerer, g *global.Data) *RootCommand {
	var c RootCommand
	c.Globals = g
	c.CmdClause = parent.Command("whoami", "Get information about the currently authenticated account")
	return &c
}

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, out io.Writer) error {
	debugMode, _ := strconv.ParseBool(c.Globals.Env.DebugMode)
	token, _ := c.Globals.Token()
	apiEndpoint, _ := c.Globals.APIEndpoint()
	data, err := undocumented.Call(undocumented.CallOptions{
		APIEndpoint: apiEndpoint,
		HTTPClient:  c.Globals.HTTPClient,
		HTTPHeaders: []undocumented.HTTPHeader{
			{
				Key:   "Accept",
				Value: "application/json",
			},
			{
				Key:   "User-Agent",
				Value: useragent.Name,
			},
		},
		Method: http.MethodGet,
		Path:   "/verify",
		Token:  token,
		Debug:  debugMode,
	})
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return fmt.Errorf("error executing API request: %w", err)
	}

	var response VerifyResponse
	if err := json.Unmarshal(data, &response); err != nil {
		c.Globals.ErrLog.Add(err)
		return fmt.Errorf("error decoding API response: %w", err)
	}

	if !c.Globals.Verbose() {
		fmt.Fprintf(out, "%s <%s>\n", response.User.Name, response.User.Login)
		return nil
	}

	keys := make([]string, 0, len(response.Services))
	for k := range response.Services {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	fmt.Fprintf(out, "Customer ID: %s\n", response.Customer.ID)
	fmt.Fprintf(out, "Customer name: %s\n", response.Customer.Name)
	fmt.Fprintf(out, "User ID: %s\n", response.User.ID)
	fmt.Fprintf(out, "User name: %s\n", response.User.Name)
	fmt.Fprintf(out, "User login: %s\n", response.User.Login)
	fmt.Fprintf(out, "Token ID: %s\n", response.Token.ID)
	fmt.Fprintf(out, "Token name: %s\n", response.Token.Name)
	fmt.Fprintf(out, "Token created at: %s\n", response.Token.CreatedAt)
	if response.Token.ExpiresAt != "" {
		fmt.Fprintf(out, "Token expires at: %s\n", response.Token.ExpiresAt)
	}
	fmt.Fprintf(out, "Token scope: %s\n", response.Token.Scope)
	fmt.Fprintf(out, "Service count: %d\n", len(response.Services))
	for _, k := range keys {
		fmt.Fprintf(out, "\t%s (%s)\n", response.Services[k], k)
	}

	return nil
}

// VerifyResponse models the Fastly API response for the whoami command.
type VerifyResponse struct {
	Customer Customer          `json:"customer"`
	User     User              `json:"user"`
	Services map[string]string `json:"services"`
	Token    Token             `json:"token"`
}

// Customer is part of the Fastly API response for the whoami command.
type Customer struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// User is part of the Fastly API response for the whoami command.
type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Login string `json:"login"`
}

// Token is part of the Fastly API response for the whoami command.
type Token struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	ExpiresAt string `json:"expires_at"`
	Scope     string `json:"scope"`
}
