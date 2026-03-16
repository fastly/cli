package text

import (
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v13/fastly/apisecurity/operations"
)

// PrintOperationTag displays an operation tag.
func PrintOperationTag(out io.Writer, tag *operations.OperationTag) {
	fmt.Fprintf(out, "ID: %s\n", tag.ID)
	fmt.Fprintf(out, "Name: %s\n", tag.Name)
	if tag.Description != "" {
		fmt.Fprintf(out, "Description: %s\n", tag.Description)
	}
	if tag.Count > 0 {
		fmt.Fprintf(out, "Operation Count: %d\n", tag.Count)
	}
	if tag.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", tag.CreatedAt)
	}
	if tag.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", tag.UpdatedAt)
	}
}

// PrintOperationTagsTbl displays operation tags in a table format.
func PrintOperationTagsTbl(out io.Writer, tags []operations.OperationTag) {
	tbl := NewTable(out)
	tbl.AddHeader("ID", "Name", "Description", "Operations", "Created At", "Updated At")

	if tags == nil {
		tbl.Print()
		return
	}

	for _, tag := range tags {
		description := tag.Description
		if description == "" {
			description = "-"
		}
		tbl.AddLine(tag.ID, tag.Name, description, fmt.Sprintf("%d", tag.Count), tag.CreatedAt, tag.UpdatedAt)
	}
	tbl.Print()
}
