package text

import (
	"fmt"
	"io"
	"strings"

	"github.com/fastly/cli/pkg/time"
	"github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/lists"
)

// PrintList displays an NGWAF list.
func PrintList(out io.Writer, listToPrint *lists.List) {
	fmt.Fprintf(out, "ID: %s\n", listToPrint.ListID)
	fmt.Fprintf(out, "Name: %s\n", listToPrint.Name)
	fmt.Fprintf(out, "Description: %s\n", listToPrint.Description)
	fmt.Fprintf(out, "Type: %s\n", listToPrint.Type)
	fmt.Fprintf(out, "Entries: %s\n", strings.Join(listToPrint.Entries, ", "))
	fmt.Fprintf(out, "Scope: %s\n", listToPrint.Scope.Type)
	fmt.Fprintf(out, "Updated (UTC): %s\n", listToPrint.UpdatedAt.UTC().Format(time.Format))
	fmt.Fprintf(out, "Created (UTC): %s\n", listToPrint.CreatedAt.UTC().Format(time.Format))
}

// PrintWorkspaceTbl displays workspaces in a table format.
func PrintListTbl(out io.Writer, listsToPrint []lists.List) {
	tbl := NewTable(out)
	tbl.AddHeader("ID", "Name", "Description", "Type", "Scope", "Entries", "Updated At", "Created At")

	if listsToPrint == nil {
		tbl.Print()
		return
	}

	for _, listToPrint := range listsToPrint {
		tbl.AddLine(
			listToPrint.ListID,
			listToPrint.Name,
			listToPrint.Description,
			listToPrint.Type,
			listToPrint.Scope.Type,
			strings.Join(listToPrint.Entries, ", "),
			listToPrint.UpdatedAt,
			listToPrint.CreatedAt,
		)
	}
	tbl.Print()
}
