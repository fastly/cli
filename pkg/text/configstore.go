package text

import (
	"fmt"
	"io"
	"strconv"
	"time"

	fsttime "github.com/fastly/cli/pkg/time"
	"github.com/fastly/go-fastly/v7/fastly"
	"github.com/segmentio/textio"
)

// PrintConfigStoresTbl displays store data in a table format.
func PrintConfigStoresTbl(out io.Writer, stores []*fastly.ConfigStore) {
	tbl := NewTable(out)
	tbl.AddHeader("Name", "ID", "Created (UTC)", "Updated (UTC)")

	if stores == nil {
		tbl.Print()
		return
	}

	for _, cs := range stores {
		// avoid gosec loop aliasing check :/
		cs := cs
		tbl.AddLine(cs.Name, cs.ID, fmtConfigStoreTime(cs.CreatedAt), fmtConfigStoreTime(cs.UpdatedAt))
	}
	tbl.Print()
}

type configStoreWithMetdata interface {
	GetConfigStore() *fastly.ConfigStore
	GetConfigStoreMetadata() *fastly.ConfigStoreMetadata // May be nil.
}

// PrintConfigStore displays store data.
func PrintConfigStore(out io.Writer, prefix string, s configStoreWithMetdata) {
	out = textio.NewPrefixWriter(out, prefix)

	cs := s.GetConfigStore()

	fmt.Fprintf(out, "Name: %s\n", cs.Name)
	fmt.Fprintf(out, "ID: %s\n", cs.ID)
	fmt.Fprintf(out, "Created (UTC): %s\n", fmtConfigStoreTime(cs.CreatedAt))
	fmt.Fprintf(out, "Updated (UTC): %s\n", fmtConfigStoreTime(cs.UpdatedAt))
	if meta := s.GetConfigStoreMetadata(); meta != nil {
		fmt.Fprintf(out, "Item Count: %d\n", meta.ItemCount)
	}
}

// PrintConfigStoreServicesTbl displays table of a config store's services.
func PrintConfigStoreServicesTbl(out io.Writer, s []*fastly.Service) {
	tw := NewTable(out)
	tw.AddHeader("NAME", "ID", "TYPE")
	for _, service := range s {
		tw.AddLine(service.Name, service.ID, service.Type)
	}
	tw.Print()
}

func fmtConfigStoreTime(t *time.Time) string {
	if t == nil {
		return "n/a"
	}
	return t.UTC().Format(fsttime.Format)
}

// PrintConfigStoreItemsTbl displays store item data in a table format.
func PrintConfigStoreItemsTbl(out io.Writer, items []*fastly.ConfigStoreItem) {
	tbl := NewTable(out)
	tbl.AddHeader("Key", "Value", "Created (UTC)", "Updated (UTC)")

	if items == nil {
		tbl.Print()
		return
	}

	for _, csi := range items {
		// avoid gosec loop aliasing check :/
		csi := csi

		// Quote and truncate 'value' to an arbitrary length.
		// Note that this operates on the number of bytes, and not
		// character or grapheme clusters.
		value := csi.Value
		var truncated bool
		if len(csi.Value) > 64 {
			value = value[:64]
			truncated = true
		}
		value = strconv.Quote(value)
		if truncated {
			value += " (truncated)"
		}

		tbl.AddLine(csi.Key, value, fmtConfigStoreTime(csi.CreatedAt), fmtConfigStoreTime(csi.UpdatedAt))
	}
	tbl.Print()
}

// PrintConfigStoreItem displays store item data.
func PrintConfigStoreItem(out io.Writer, prefix string, csi *fastly.ConfigStoreItem) {
	out = textio.NewPrefixWriter(out, prefix)

	fmt.Fprintf(out, "StoreID: %s\n", csi.StoreID)
	fmt.Fprintf(out, "Key: %s\n", csi.Key)
	fmt.Fprintf(out, "Value: %s\n", csi.Value)
	fmt.Fprintf(out, "Created (UTC): %s\n", fmtConfigStoreTime(csi.CreatedAt))
	fmt.Fprintf(out, "Updated (UTC): %s\n", fmtConfigStoreTime(csi.UpdatedAt))
	if csi.DeletedAt != nil {
		fmt.Fprintf(out, "Deleted (UTC): %s\n", fmtConfigStoreTime(csi.DeletedAt))
	}
}
