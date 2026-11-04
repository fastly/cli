package operations

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/fastly/go-fastly/v19/fastly"
	"github.com/fastly/go-fastly/v19/fastly/apisecurity/operations"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// AddTagsCommand calls the Fastly API to add tags to operations.
type AddTagsCommand struct {
	argparser.Base

	// Required.
	ServiceID   *string                         `short:"s" required:"" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`
	ServiceName argparser.OptionalServiceNameID `help:"The name of the service"`
	TagIDs      []string                        `name:"tag-ids" required:"" sep:"," help:"Comma-separated list of tag IDs to add"`

	// Optional.
	OperationIDs []string `name:"operation-ids" sep:"," help:"Comma-separated list of operation IDs to add tags to"`
	File         string   `help:"Add tags to operations in bulk from a JSON file"`

	argparser.JSONOutput
}

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *AddTagsCommand) AfterApply() error {
	if c.ServiceID != nil {
		c.Globals.Manifest.Flag.ServiceID = *c.ServiceID
	}
	return nil
}

// Exec invokes the application logic for the command.
func (c *AddTagsCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	if len(c.OperationIDs) == 0 && c.File == "" {
		return fmt.Errorf("error parsing arguments: must provide either --operation-ids or --file")
	}

	if len(c.OperationIDs) > 0 && c.File != "" {
		return fmt.Errorf("error parsing arguments: cannot use both --operation-ids and --file")
	}

	serviceID, source, flag, err := argparser.ServiceID(c.ServiceName, *c.Globals.Manifest, c.Globals.APIClient, c.Globals.ErrLog)
	if err != nil {
		return err
	}
	if c.Globals.Verbose() {
		argparser.DisplayServiceID(serviceID, flag, source, out)
	}

	// Get operation IDs and tag IDs from file or flags
	var operationIDs []string
	var tagIDs []string
	if c.File != "" {
		fileInput, err := c.readFromFile()
		if err != nil {
			c.Globals.ErrLog.Add(err)
			return err
		}
		operationIDs = fileInput.OperationIDs
		tagIDs = fileInput.TagIDs
	} else {
		operationIDs = c.OperationIDs
		tagIDs = c.TagIDs
	}

	if c.Globals.Verbose() {
		fmt.Fprintf(out, "Adding %d tag(s) to %d operation(s)\n", len(tagIDs), len(operationIDs))
	}

	input := &operations.BulkAddTagsInput{
		ServiceID:    &serviceID,
		OperationIDs: operationIDs,
		TagIDs:       tagIDs,
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	results, err := operations.BulkAddTags(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Operation Count": len(operationIDs),
			"Tag Count":       len(c.TagIDs),
		})
		return err
	}

	if ok, err := c.WriteJSON(out, results); ok {
		return err
	}

	return c.printResults(out, results)
}

// AddTagsFileInput represents the JSON file format for bulk add-tags operation.
type AddTagsFileInput struct {
	OperationIDs []string `json:"operation_ids"`
	TagIDs       []string `json:"tag_ids"`
}

// readFromFile reads operation IDs and tag IDs from a JSON file.
func (c *AddTagsCommand) readFromFile() (*AddTagsFileInput, error) {
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

	var input AddTagsFileInput
	if err := json.Unmarshal(byteValue, &input); err != nil {
		return nil, fmt.Errorf("invalid JSON format: %w", err)
	}

	if len(input.OperationIDs) == 0 {
		return nil, fmt.Errorf("no operation IDs found in file: %s", c.File)
	}

	if len(input.TagIDs) == 0 {
		return nil, fmt.Errorf("no tag IDs found in file: %s", c.File)
	}

	return &input, nil
}

// printResults displays the results of the bulk add tags operation.
func (c *AddTagsCommand) printResults(out io.Writer, results *operations.BulkOperationResultsResponse) error {
	var succeeded, failed int
	for _, result := range results.Data {
		if result.StatusCode >= 200 && result.StatusCode < 300 {
			succeeded++
		} else {
			failed++
		}
	}

	text.Success(out, "Added tags to %d operation(s)", succeeded)

	if failed > 0 {
		text.Warning(out, "%d operation(s) failed", failed)
	}

	if c.Globals.Verbose() {
		text.Break(out)
		tw := text.NewTable(out)
		tw.AddHeader("OPERATION ID", "STATUS CODE", "RESULT")
		for _, result := range results.Data {
			status := "Success"
			if result.StatusCode < 200 || result.StatusCode >= 300 {
				status = fmt.Sprintf("Failed: %s", result.Reason)
			}
			tw.AddLine(result.ID, fmt.Sprintf("%d", result.StatusCode), status)
		}
		tw.Print()
	}

	return nil
}
