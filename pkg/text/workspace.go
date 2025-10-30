package text

import (
	"fmt"
	"io"
	"strings"

	"github.com/fastly/cli/pkg/time"
	"github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/workspaces"
)

// PrintWorkspace displays a workspace.
func PrintWorkspace(out io.Writer, workspaceToPrint *workspaces.Workspace) {
	fmt.Fprintf(out, "ID: %s\n", workspaceToPrint.WorkspaceID)
	fmt.Fprintf(out, "Name: %s\n", workspaceToPrint.Name)
	fmt.Fprintf(out, "Description: %s\n", workspaceToPrint.Description)
	fmt.Fprintf(out, "Mode: %s\n", workspaceToPrint.Mode)
	fmt.Fprintf(out, "Attack Signal Thresholds: Immediate %t, One Minute %d, Ten Minutes %d, One Hour %d\n", workspaceToPrint.AttackSignalThresholds.Immediate, workspaceToPrint.AttackSignalThresholds.OneMinute, workspaceToPrint.AttackSignalThresholds.TenMinutes, workspaceToPrint.AttackSignalThresholds.OneHour)
	if len(workspaceToPrint.ClientIPHeaders) != 0 {
		fmt.Fprintf(out, "Client IP Headers: %s\n", strings.Join(workspaceToPrint.ClientIPHeaders, ", "))
	}
	if workspaceToPrint.DefaultBlockingResponseCode > 0 {
		fmt.Fprintf(out, "Default Blocking Response Code: %d\n", workspaceToPrint.DefaultBlockingResponseCode)
	}
	if workspaceToPrint.DefaultRedirectURL != "" {
		fmt.Fprintf(out, "Default Redirect URL: %s\n", workspaceToPrint.DefaultRedirectURL)
	}
	if workspaceToPrint.IPAnonymization != "" {
		fmt.Fprintf(out, "IP Anonymization: %s\n", workspaceToPrint.IPAnonymization)
	}
	fmt.Fprintf(out, "Created (UTC): %s\n", workspaceToPrint.CreatedAt.UTC().Format(time.Format))
}

// PrintWorkspaceTbl displays workspaces in a table format.
func PrintWorkspaceTbl(out io.Writer, workspacesToPrint []workspaces.Workspace) {
	tbl := NewTable(out)
	tbl.AddHeader("ID", "Name", "Description", "Mode", "Created At")

	if workspacesToPrint == nil {
		tbl.Print()
		return
	}

	for _, workspaceToPrint := range workspacesToPrint {
		tbl.AddLine(workspaceToPrint.WorkspaceID, workspaceToPrint.Name, workspaceToPrint.Description, workspaceToPrint.Mode, workspaceToPrint.CreatedAt)
	}
	tbl.Print()
}
