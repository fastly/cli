package discoveredoperations

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

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// UpdateCommand calls the Fastly API to update a discovered API operation's status.
type UpdateCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required .
	serviceName argparser.OptionalServiceNameID
	file        string
	operationID argparser.OptionalString
	status      argparser.OptionalString
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent argparser.Registerer, g *global.Data) *UpdateCommand {
	c := UpdateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("update", "Update the status of discovered operation(s)")

	// Required.
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagServiceIDName,
		Description: argparser.FlagServiceIDDesc,
		Dst:         &g.Manifest.Flag.ServiceID,
		Short:       's',
		Required:    true,
	})
	c.CmdClause.Flag("operation-id", "The ID of the discovered operation (comma-separated for multiple)").Action(c.operationID.Set).StringVar(&c.operationID.Value)
	c.CmdClause.Flag("file", "Update operations in bulk from a JSON file").StringVar(&c.file)

	// Optional.
	c.CmdClause.Flag("status", "The new status to apply. Valid values are: 'discovered', 'ignored'").Action(c.status.Set).StringVar(&c.status.Value)
	c.RegisterFlagBool(c.JSONFlag()) // --json

	return &c
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	if !c.operationID.WasSet && c.file == "" {
		return fmt.Errorf("error parsing arguments: must provide either --operation-id or --file")
	}

	if c.operationID.WasSet && c.file != "" {
		return fmt.Errorf("error parsing arguments: cannot use both --operation-id and --file")
	}

	// When using --file, status should not be provided via flag.
	if c.file != "" && c.status.WasSet {
		return fmt.Errorf("error parsing arguments: cannot use both --file and --status (status should be specified in the JSON file)")
	}

	// When using --operation-id, status is required.
	if c.operationID.WasSet && !c.status.WasSet {
		return fmt.Errorf("error parsing arguments: --status is required when using --operation-id")
	}

	serviceID, source, flag, err := argparser.ServiceID(c.serviceName, *c.Globals.Manifest, c.Globals.APIClient, c.Globals.ErrLog)
	if err != nil {
		return err
	}
	if c.Globals.Verbose() {
		argparser.DisplayServiceID(serviceID, flag, source, out)
	}

	// Handle bulk mode from file.
	if c.file != "" {
		fileInput, err := c.readFromFile()
		if err != nil {
			c.Globals.ErrLog.Add(err)
			return err
		}

		// Convert status from file to uppercase to map to API.
		status, err := c.validateStatus(fileInput.Status)
		if err != nil {
			c.Globals.ErrLog.Add(err)
			return err
		}

		return c.executeBulkUpdate(out, serviceID, status, fileInput.OperationIDs)
	}

	// Convert status to uppercase for API.
	status, err := c.validateStatus(c.status.Value)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	// Handle comma-separated operation IDs.
	operationIDs := strings.Split(c.operationID.Value, ",")
	// Trim whitespace from each ID
	for i, id := range operationIDs {
		operationIDs[i] = strings.TrimSpace(id)
	}

	// If multiple operation IDs, use bulk update.
	if len(operationIDs) > 1 {
		return c.executeBulkUpdate(out, serviceID, status, operationIDs)
	}

	// Handle single operation mode.
	input := operations.UpdateDiscoveredStatusInput{
		ServiceID:   &serviceID,
		OperationID: &c.operationID.Value,
		Status:      &status,
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	o, err := operations.UpdateDiscoveredStatus(context.TODO(), fc, &input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":   serviceID,
			"Operation ID": c.operationID.Value,
			"Status":       c.status.Value,
		})
		return err
	}

	if ok, err := c.WriteJSON(out, o); ok {
		return err
	}

	if !c.Globals.Verbose() {
		return c.printSummary(out, o)
	}

	return c.printVerbose(out, o)
}

// validateStatus converts and validates the status value.
func (c *UpdateCommand) validateStatus(statusValue string) (string, error) {
	switch statusValue {
	case "discovered", "DISCOVERED":
		return "DISCOVERED", nil
	case "ignored", "IGNORED":
		return "IGNORED", nil
	default:
		return "", fmt.Errorf("invalid status: %s. Valid options: 'discovered', 'ignored'", statusValue)
	}
}

// executeBulkUpdate performs a bulk update operation for multiple operation IDs.
func (c *UpdateCommand) executeBulkUpdate(out io.Writer, serviceID string, status string, operationIDs []string) error {
	if c.Globals.Verbose() {
		fmt.Fprintf(out, "Updating %d operation(s) with status: %s\n", len(operationIDs), status)
		fmt.Fprintf(out, "Operation IDs: %v\n", operationIDs)
	}

	input := operations.BulkUpdateDiscoveredStatusInput{
		ServiceID:    &serviceID,
		OperationIDs: operationIDs,
		Status:       &status,
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	results, err := operations.BulkUpdateDiscoveredStatus(context.TODO(), fc, &input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID": serviceID,
			"Status":     c.status.Value,
			"Count":      len(operationIDs),
		})
		return err
	}

	if ok, err := c.WriteJSON(out, results); ok {
		return err
	}

	return c.printBulkResults(out, results)
}

// UpdateFileInput represents the JSON file format for bulk update operation.
type UpdateFileInput struct {
	OperationIDs []string `json:"operation_ids"`
	Status       string   `json:"status"`
}

// readFromFile reads operation IDs and status from a JSON file.
func (c *UpdateCommand) readFromFile() (*UpdateFileInput, error) {
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

	var input UpdateFileInput
	if err := json.Unmarshal(byteValue, &input); err != nil {
		return nil, fmt.Errorf("invalid JSON format: %w", err)
	}

	if len(input.OperationIDs) == 0 {
		return nil, fmt.Errorf("no operation IDs found in file: %s", c.file)
	}

	if input.Status == "" {
		return nil, fmt.Errorf("status not specified in file: %s", c.file)
	}

	return &input, nil
}

// printBulkResults displays the results of a bulk update operation.
func (c *UpdateCommand) printBulkResults(out io.Writer, results *operations.BulkOperationResultsResponse) error {
	var succeeded, failed int
	for _, result := range results.Data {
		if result.StatusCode >= 200 && result.StatusCode < 300 {
			succeeded++
		} else {
			failed++
		}
	}

	text.Success(out, "Updated %d discovered operation(s)", succeeded)

	if failed > 0 {
		text.Warning(out, "%d operation(s) failed to update", failed)
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

// printSummary displays the discovered operation in a simple format.
func (c *UpdateCommand) printSummary(out io.Writer, op *operations.DiscoveredOperation) error {
	fmt.Fprintf(out, "Updated discovered operation:\n")
	fmt.Fprintf(out, "  ID: %s\n", op.ID)
	fmt.Fprintf(out, "  Method: %s\n", op.Method)
	fmt.Fprintf(out, "  Domain: %s\n", op.Domain)
	fmt.Fprintf(out, "  Path: %s\n", op.Path)
	fmt.Fprintf(out, "  Status: %s\n", op.Status)

	return nil
}

// printVerbose displays detailed information for the discovered operation.
func (c *UpdateCommand) printVerbose(out io.Writer, op *operations.DiscoveredOperation) error {
	fmt.Fprintf(out, "\nUpdated Discovered Operation\n")
	fmt.Fprintf(out, "\tID: %s\n", op.ID)
	fmt.Fprintf(out, "\tMethod: %s\n", op.Method)
	fmt.Fprintf(out, "\tDomain: %s\n", op.Domain)
	fmt.Fprintf(out, "\tPath: %s\n", op.Path)
	fmt.Fprintf(out, "\tStatus: %s\n", op.Status)
	fmt.Fprintf(out, "\tRPS: %.2f\n", op.RPS)
	if op.LastSeenAt != "" {
		fmt.Fprintf(out, "\tLast Seen: %s\n", op.LastSeenAt)
	}
	if op.UpdatedAt != "" {
		fmt.Fprintf(out, "\tUpdated At: %s\n", op.UpdatedAt)
	}
	fmt.Fprintln(out)

	return nil
}
