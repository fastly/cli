package text

import (
	"io"
	"strconv"

	"github.com/fastly/go-fastly/v17/fastly/airuntimecontrol/v1/session"
)

// PrintSessionsTbl displays ARC session logs in a table format.
func PrintSessionsTbl(out io.Writer, sessionsToPrint []session.Session) {
	tbl := NewTable(out)
	tbl.AddHeader("ID", "Virtual Key", "Model", "Provider", "Requests", "Input Tokens", "Output Tokens", "Created At")

	if sessionsToPrint == nil {
		tbl.Print()
		return
	}

	for _, s := range sessionsToPrint {
		tbl.AddLine(
			s.ID,
			s.VirtualKeyName,
			s.Model,
			s.Provider,
			strconv.Itoa(s.Requests),
			strconv.Itoa(s.InputTokens),
			strconv.Itoa(s.OutputTokens),
			s.CreatedAt,
		)
	}
	tbl.Print()
}
