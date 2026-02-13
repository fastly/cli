package text

import (
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v13/fastly/ngwaf/v1/workspaces/virtualpatches"
)

// PrintVirtualPatch displays a single virtual patch.
func PrintVirtualPatch(out io.Writer, virtualPatchToPrint *virtualpatches.VirtualPatch) {
	fmt.Fprintf(out, "ID: %s\n", virtualPatchToPrint.ID)
	fmt.Fprintf(out, "Description: %s\n", virtualPatchToPrint.Description)
	fmt.Fprintf(out, "Enabled: %t\n", virtualPatchToPrint.Enabled)
	fmt.Fprintf(out, "Mode: %s\n", virtualPatchToPrint.Mode)
}

// PrintVirtualPatchTbl prints a table of virtual patches.
func PrintVirtualPatchTbl(out io.Writer, virtualPatchesToPrint []virtualpatches.VirtualPatch) {
	tbl := NewTable(out)
	tbl.AddHeader("ID", "Description", "Enabled", "Mode")

	if virtualPatchesToPrint == nil {
		tbl.Print()
		return
	}

	for _, vp := range virtualPatchesToPrint {
		tbl.AddLine(vp.ID, vp.Description, vp.Enabled, vp.Mode)
	}
	tbl.Print()
}
