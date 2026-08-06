package text

import (
	"fmt"
	"io"
	"strings"

	"github.com/fastly/cli/pkg/time"
	"github.com/fastly/go-fastly/v17/fastly/notifications/v1/eventmappings"
	"github.com/fastly/go-fastly/v17/fastly/notifications/v1/eventmappings/eventtypes"
	"github.com/fastly/go-fastly/v17/fastly/notifications/v1/eventmappings/scopetypes"
)

// PrintEventMapping displays an audit log event mapping.
func PrintEventMapping(out io.Writer, em *eventmappings.EventMapping) {
	fmt.Fprintf(out, "ID: %s\n", em.ID)
	fmt.Fprintf(out, "Name: %s\n", em.Name)
	fmt.Fprintf(out, "Description: %s\n", em.Description)
	fmt.Fprintf(out, "Scope Type: %s\n", em.ScopeType)
	fmt.Fprintf(out, "Scope IDs: %s\n", strings.Join(em.ScopeIDs, ", "))
	fmt.Fprintf(out, "Event Types: %s\n", strings.Join(em.EventTypes, ", "))
	fmt.Fprintf(out, "Integration IDs: %s\n", strings.Join(em.IntegrationIDs, ", "))
	fmt.Fprintf(out, "Mapping Status: %s\n", em.MappingStatus)
	fmt.Fprintf(out, "Created (UTC): %s\n", em.CreatedAt.UTC().Format(time.Format))
	fmt.Fprintf(out, "Updated (UTC): %s\n", em.UpdatedAt.UTC().Format(time.Format))
}

// PrintEventMappingsTbl displays audit log event mappings in a table format.
func PrintEventMappingsTbl(out io.Writer, ems []eventmappings.EventMapping) {
	tbl := NewTable(out)
	tbl.AddHeader("ID", "NAME", "SCOPE TYPE", "EVENT TYPES", "INTEGRATION IDS", "STATUS")

	for _, em := range ems {
		tbl.AddLine(
			em.ID,
			em.Name,
			em.ScopeType,
			strings.Join(em.EventTypes, ", "),
			strings.Join(em.IntegrationIDs, ", "),
			em.MappingStatus,
		)
	}
	tbl.Print()
}

// PrintEventTypesTbl displays supported audit event types in a table format.
func PrintEventTypesTbl(out io.Writer, types []eventtypes.EventType) {
	tbl := NewTable(out)
	tbl.AddHeader("EVENT TYPE", "DISPLAY NAME", "SCOPE TYPES")

	for _, et := range types {
		tbl.AddLine(et.EventType, et.DisplayName, strings.Join(et.ScopeTypes, ", "))
	}
	tbl.Print()
}

// PrintScopeTypesTbl displays supported scope types in a table format.
func PrintScopeTypesTbl(out io.Writer, types []scopetypes.ScopeType) {
	tbl := NewTable(out)
	tbl.AddHeader("SCOPE TYPE")

	for _, st := range types {
		tbl.AddLine(st.ScopeType)
	}
	tbl.Print()
}
