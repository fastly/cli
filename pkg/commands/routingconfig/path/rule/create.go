package rule

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/fastly/go-fastly/v19/fastly"
	"github.com/fastly/go-fastly/v19/fastly/domainmanagement/v1/routingconfigs/paths/rules"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/text"
)

// CreateCommand calls the Fastly API to create a rule within a path.
type CreateCommand struct {
	argparser.Base

	// Required.
	File            string `required:"" help:"Path to a JSON file describing the rule, e.g. {\"action\": {\"type\": \"service\", \"value\": \"...\"}, \"conditions\": [...]}. Omit \"conditions\" to create the default (catch-all) rule for the path"`
	PathID          string `required:"" help:"The Path Identifier"`
	RoutingConfigID string `required:"" help:"The Routing Config Identifier"`
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	input := &rules.CreateInput{
		PathID:          &c.PathID,
		RoutingConfigID: &c.RoutingConfigID,
	}

	if err := readRuleFile(c.File, input); err != nil {
		return err
	}
	if input.Action == nil {
		return fmt.Errorf("the --file JSON at '%s' is missing the required 'action' field", c.File)
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	d, err := rules.Create(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Path ID":           c.PathID,
			"Routing Config ID": c.RoutingConfigID,
		})
		return err
	}

	text.Success(out, "Created rule (rule-id: %s) in path (path-id: %s)", d.RuleID, c.PathID)
	return nil
}

// readRuleFile reads a JSON file describing a rule and unmarshals it
// directly into i (a *rules.CreateInput or *rules.UpdateInput). PathID,
// RoutingConfigID, and RuleID are tagged to be ignored by JSON (un)marshaling,
// so they are left untouched by this call.
func readRuleFile(path string, i any) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("error parsing path '%s': %w", path, err)
	}

	f, err := os.Open(abs) // #nosec G304
	if err != nil {
		return fmt.Errorf("error reading path '%s': %w", path, err)
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("failed to read file '%s': %w", path, err)
	}

	if err := json.Unmarshal(b, i); err != nil {
		return fmt.Errorf("failed to unmarshal JSON in '%s': %w", path, err)
	}
	return nil
}
