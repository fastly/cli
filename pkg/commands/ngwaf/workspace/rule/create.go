package rule

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/fastly/go-fastly/v12/fastly"
	"github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/rules"
	"github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/scope"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// CreateCommand calls the Fastly API to create workspace-level rules.
type CreateCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	path        string
	workspaceID argparser.OptionalWorkspaceID
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent argparser.Registerer, g *global.Data) *CreateCommand {
	c := CreateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("create", "Create a workspace-level rule").Alias("add")

	// Required.
	c.CmdClause.Flag("path", "Path to a json file that contains the rule schema.").Required().StringVar(&c.path)
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagNGWAFWorkspaceID,
		Description: argparser.FlagNGWAFWorkspaceIDDesc,
		Dst:         &c.workspaceID.Value,
		Action:      c.workspaceID.Set,
	})

	// Optional.
	c.RegisterFlagBool(c.JSONFlag())

	return &c
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}
	if err := c.workspaceID.Parse(); err != nil {
		return err
	}
	input := &rules.CreateInput{}
	if c.path != "" {
		path, err := filepath.Abs(c.path)
		if err != nil {
			return fmt.Errorf("error parsing path '%s': %q", c.path, err)
		}

		jsonFile, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("error reading cert-path '%s': %q", c.path, err)
		}
		defer jsonFile.Close()

		byteValue, err := io.ReadAll(jsonFile)
		if err != nil {
			return fmt.Errorf("failed to read json file: %v", err)
		}

		if err := json.Unmarshal(byteValue, input); err != nil {
			return fmt.Errorf("failed to unmarshal json data: %v", err)
		}
	}
	input.Scope = &scope.Scope{
		Type:      scope.ScopeTypeWorkspace,
		AppliesTo: []string{c.workspaceID.Value},
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	data, err := rules.Create(context.TODO(), fc, input)
	if err != nil {
		return err
	}

	if ok, err := c.WriteJSON(out, data); ok {
		return err
	}

	text.Success(out, "Created workspace-level rule with ID %s", data.RuleID)
	return nil
}
