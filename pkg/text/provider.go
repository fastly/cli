package text

import (
	"io"
	"strconv"

	"github.com/fastly/go-fastly/v17/fastly/airuntimecontrol/v1/provider"
)

// PrintProvidersTbl displays ARC providers in a table format. Model lists can
// run into the dozens per provider, so only a count is shown here; use
// `list-models --provider-id <id>` to see the full list for a given provider.
func PrintProvidersTbl(out io.Writer, providersToPrint []provider.Provider) {
	tbl := NewTable(out)
	tbl.AddHeader("ID", "Display Name", "Default Base URL", "Model Count")

	if providersToPrint == nil {
		tbl.Print()
		return
	}

	for _, p := range providersToPrint {
		tbl.AddLine(p.ID, p.DisplayName, p.DefaultBaseURL, strconv.Itoa(len(p.Models)))
	}
	tbl.Print()
}

// PrintModelsTbl displays ARC models in a table format.
func PrintModelsTbl(out io.Writer, modelsToPrint []provider.Model) {
	tbl := NewTable(out)
	tbl.AddHeader("ID", "Display Name", "Provider ID")

	if modelsToPrint == nil {
		tbl.Print()
		return
	}

	for _, m := range modelsToPrint {
		tbl.AddLine(m.ID, m.DisplayName, m.ProviderID)
	}
	tbl.Print()
}
