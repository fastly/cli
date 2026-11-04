package operations

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/fastly/go-fastly/v19/fastly"
	"github.com/fastly/go-fastly/v19/fastly/apisecurity/operations"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// CreateCommand calls the Fastly API to create an operation.
type CreateCommand struct {
	argparser.Base

	// Required.
	ServiceID   *string                         `short:"s" required:"" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`
	ServiceName argparser.OptionalServiceNameID `help:"The name of the service"`
	Method      string                          `help:"The HTTP method for the operation (e.g., GET, POST, PUT)"`
	Domain      string                          `help:"Domain for the operation"`
	Path        string                          `help:"The path for the operation, which may include path parameters.(e.g., /api/users)"`

	// Optional.
	Description string   `help:"Description of what the operation does"`
	TagIDs      []string `name:"tag-ids" sep:"," help:"A comma-separated array of operation tag IDs associated with this operation"`
	File        string   `help:"Create operations in bulk from a JSON file"`

	argparser.JSONOutput
}

// OperationInput represents a single operation to be created from JSON.
type OperationInput struct {
	Method      string   `json:"method"`
	Domain      string   `json:"domain"`
	Path        string   `json:"path"`
	Description string   `json:"description,omitempty"`
	TagIDs      []string `json:"tag_ids,omitempty"`
}

// CreateFileInput represents the JSON file format for bulk create operations.
type CreateFileInput struct {
	Operations []OperationInput `json:"operations"`
}

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *CreateCommand) AfterApply() error {
	if c.ServiceID != nil {
		c.Globals.Manifest.Flag.ServiceID = *c.ServiceID
	}
	return nil
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	// Validate flags
	if c.File != "" && (c.Method != "" || c.Domain != "" || c.Path != "") {
		return fmt.Errorf("error parsing arguments: cannot use both --file and individual operation flags (--method, --domain, --path)")
	}

	if c.File == "" && (c.Method == "" || c.Domain == "" || c.Path == "") {
		return fmt.Errorf("error parsing arguments: must provide either --file or all of --method, --domain, and --path")
	}

	serviceID, source, flag, err := argparser.ServiceID(c.ServiceName, *c.Globals.Manifest, c.Globals.APIClient, c.Globals.ErrLog)
	if err != nil {
		return err
	}
	if c.Globals.Verbose() {
		argparser.DisplayServiceID(serviceID, flag, source, out)
	}

	// Handle bulk mode from file
	if c.File != "" {
		return c.createFromFile(serviceID, out)
	}

	// Handle single operation mode
	input := c.constructInput(serviceID)

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	o, err := operations.Create(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID": serviceID,
			"Method":     c.Method,
			"Domain":     c.Domain,
			"Path":       c.Path,
		})
		return err
	}

	if ok, err := c.WriteJSON(out, o); ok {
		return err
	}

	text.Success(out, "Created operation %s %s%s (ID: %s)", strings.ToUpper(o.Method), o.Domain, o.Path, o.ID)
	if c.Description != "" {
		fmt.Fprintf(out, "\nDescription: %s\n", o.Description)
	}
	if len(o.TagIDs) > 0 {
		fmt.Fprintf(out, "Tags: %d associated\n", len(o.TagIDs))
	}

	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *CreateCommand) constructInput(serviceID string) *operations.CreateInput {
	input := &operations.CreateInput{
		ServiceID: &serviceID,
		Method:    &c.Method,
		Domain:    &c.Domain,
		Path:      &c.Path,
	}

	if c.Description != "" {
		input.Description = &c.Description
	}

	if len(c.TagIDs) > 0 {
		input.TagIDs = c.TagIDs
	}

	return input
}

// createFromFile creates operations in bulk from a newline-delimited JSON file.
func (c *CreateCommand) createFromFile(serviceID string, out io.Writer) error {
	ops, err := c.readOperationsFromFile()
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if c.Globals.Verbose() {
		fmt.Fprintf(out, "Creating %d operation(s) from file\n", len(ops))
	}

	type result struct {
		Operation *operations.Operation
		Error     error
	}

	results := make([]result, 0, len(ops))

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	for _, op := range ops {
		input := &operations.CreateInput{
			ServiceID:   &serviceID,
			Method:      &op.Method,
			Domain:      &op.Domain,
			Path:        &op.Path,
			Description: &op.Description,
			TagIDs:      op.TagIDs,
		}

		o, err := operations.Create(context.TODO(), fc, input)
		results = append(results, result{
			Operation: o,
			Error:     err,
		})
	}

	// Count successes and failures
	var succeeded, failed int
	for _, r := range results {
		if r.Error == nil {
			succeeded++
		} else {
			failed++
		}
	}

	if c.JSONOutput.Enabled {
		type jsonResult struct {
			Success    int                     `json:"success"`
			Failed     int                     `json:"failed"`
			Operations []*operations.Operation `json:"operations,omitempty"`
			Errors     []string                `json:"errors,omitempty"`
		}

		jr := jsonResult{
			Success: succeeded,
			Failed:  failed,
		}

		for _, r := range results {
			if r.Error == nil {
				jr.Operations = append(jr.Operations, r.Operation)
			} else {
				jr.Errors = append(jr.Errors, r.Error.Error())
			}
		}

		_, err := c.WriteJSON(out, jr)
		return err
	}

	text.Success(out, "Created %d operation(s)", succeeded)

	if failed > 0 {
		text.Warning(out, "%d operation(s) failed to create", failed)
	}

	if c.Globals.Verbose() {
		text.Break(out)
		tw := text.NewTable(out)
		tw.AddHeader("METHOD", "DOMAIN", "PATH", "RESULT")
		for i, r := range results {
			status := "Success"
			if r.Error != nil {
				status = fmt.Sprintf("Failed: %s", r.Error.Error())
			}
			op := ops[i]
			tw.AddLine(strings.ToUpper(op.Method), op.Domain, op.Path, status)
		}
		tw.Print()
	}

	if failed > 0 {
		return fmt.Errorf("%d operation(s) failed to create", failed)
	}

	return nil
}

// readOperationsFromFile reads operations from a JSON file.
func (c *CreateCommand) readOperationsFromFile() ([]OperationInput, error) {
	path, err := filepath.Abs(c.File)
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(path); err != nil {
		return nil, err
	}

	file, err := os.Open(path) /* #nosec */
	if err != nil {
		return nil, err
	}
	defer file.Close()

	byteValue, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var input CreateFileInput
	if err := json.Unmarshal(byteValue, &input); err != nil {
		return nil, fmt.Errorf("invalid JSON format: %w", err)
	}

	if len(input.Operations) == 0 {
		return nil, fmt.Errorf("no operations found in file: %s", c.File)
	}

	// Validate required fields
	for i, op := range input.Operations {
		if op.Method == "" || op.Domain == "" || op.Path == "" {
			return nil, fmt.Errorf("operation %d: missing required fields (method, domain, path)", i+1)
		}
	}

	return input.Operations, nil
}
