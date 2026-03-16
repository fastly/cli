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

	"github.com/fastly/go-fastly/v13/fastly"
	"github.com/fastly/go-fastly/v13/fastly/apisecurity/operations"

	"github.com/fastly/kingpin"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// CreateCommand calls the Fastly API to create an operation.
type CreateCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	serviceName argparser.OptionalServiceNameID
	method      string
	domain      string
	path        string

	// Optional.
	description string
	tagIDs      []string
	file        string
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

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent argparser.Registerer, g *global.Data) *CreateCommand {
	c := CreateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("create", "Create an operation").Alias("add")

	// Required.
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagServiceIDName,
		Description: argparser.FlagServiceIDDesc,
		Dst:         &g.Manifest.Flag.ServiceID,
		Short:       's',
		Required:    true,
	})
	c.RegisterFlag(argparser.StringFlagOpts{
		Action:      c.serviceName.Set,
		Name:        argparser.FlagServiceName,
		Description: argparser.FlagServiceNameDesc,
		Dst:         &c.serviceName.Value,
	})
	c.CmdClause.Flag("method", "The HTTP method for the operation (e.g., GET, POST, PUT)").StringVar(&c.method)
	c.CmdClause.Flag("domain", "Domain for the operation").StringVar(&c.domain)
	c.CmdClause.Flag("path", "The path for the operation, which may include path parameters.(e.g., /api/users)").StringVar(&c.path)

	// Optional.
	c.CmdClause.Flag("description", "Description of what the operation does").StringVar(&c.description)
	c.CmdClause.Flag("tag-ids", "A comma-separated array of operation tag IDs associated with this operation").StringsVar(&c.tagIDs, kingpin.Separator(","))
	c.CmdClause.Flag("file", "Create operations in bulk from a JSON file").StringVar(&c.file)
	c.RegisterFlagBool(c.JSONFlag()) // --json

	return &c
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	// Validate flags
	if c.file != "" && (c.method != "" || c.domain != "" || c.path != "") {
		return fmt.Errorf("error parsing arguments: cannot use both --file and individual operation flags (--method, --domain, --path)")
	}

	if c.file == "" && (c.method == "" || c.domain == "" || c.path == "") {
		return fmt.Errorf("error parsing arguments: must provide either --file or all of --method, --domain, and --path")
	}

	serviceID, source, flag, err := argparser.ServiceID(c.serviceName, *c.Globals.Manifest, c.Globals.APIClient, c.Globals.ErrLog)
	if err != nil {
		return err
	}
	if c.Globals.Verbose() {
		argparser.DisplayServiceID(serviceID, flag, source, out)
	}

	// Handle bulk mode from file
	if c.file != "" {
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
			"Method":     c.method,
			"Domain":     c.domain,
			"Path":       c.path,
		})
		return err
	}

	if ok, err := c.WriteJSON(out, o); ok {
		return err
	}

	text.Success(out, "Created operation %s %s%s (ID: %s)", strings.ToUpper(o.Method), o.Domain, o.Path, o.ID)
	if c.description != "" {
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
		Method:    &c.method,
		Domain:    &c.domain,
		Path:      &c.path,
	}

	if c.description != "" {
		input.Description = &c.description
	}

	if len(c.tagIDs) > 0 {
		input.TagIDs = c.tagIDs
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
	path, err := filepath.Abs(c.file)
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
		return nil, fmt.Errorf("no operations found in file: %s", c.file)
	}

	// Validate required fields
	for i, op := range input.Operations {
		if op.Method == "" || op.Domain == "" || op.Path == "" {
			return nil, fmt.Errorf("operation %d: missing required fields (method, domain, path)", i+1)
		}
	}

	return input.Operations, nil
}
