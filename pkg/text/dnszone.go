package text

import (
	"fmt"
	"io"
	"strings"

	"github.com/segmentio/textio"

	"github.com/fastly/go-fastly/v15/fastly/dns/v1/dnszones"
)

// PrintDNSZone pretty prints a dnszones.Zone in verbose format to a given
// io.Writer. Consumers can provide a prefix string for indentation.
func PrintDNSZone(out io.Writer, prefix string, z *dnszones.Zone) {
	out = textio.NewPrefixWriter(out, prefix)

	fmt.Fprintf(out, "ID: %s\n", strOrEmpty(z.ID))
	fmt.Fprintf(out, "Name: %s\n", strOrEmpty(z.Name))
	fmt.Fprintf(out, "Type: %s\n", strOrEmpty(z.Type))
	fmt.Fprintf(out, "Description: %s\n", strOrEmpty(z.Description))
	fmt.Fprintf(out, "Serial: %s\n", strOrEmpty(z.Serial))
	fmt.Fprintf(out, "Nameservers: %s\n", strings.Join(z.Nameservers, ", "))
	if z.XfrConfigInbound != nil {
		fmt.Fprintf(out, "Inbound Transfer Config:\n")
		printXfrConfigInbound(out, "\t\t", z.XfrConfigInbound)
	}
	fmt.Fprintf(out, "Created at: %s\n", strOrEmpty(z.CreatedAt))
	fmt.Fprintf(out, "Updated at: %s\n", strOrEmpty(z.UpdatedAt))
}

// PrintDNSZoneTbl prints a slice of dnszones.Zone in table format to a given io.Writer.
func PrintDNSZoneTbl(out io.Writer, zones []dnszones.Zone) {
	tbl := NewTable(out)
	tbl.AddHeader("ID", "Name", "Type", "Description", "Created At", "Updated At")

	if zones == nil {
		tbl.Print()
		return
	}

	for _, z := range zones {
		tbl.AddLine(strOrEmpty(z.ID), strOrEmpty(z.Name), strOrEmpty(z.Type), strOrEmpty(z.Description), strOrEmpty(z.CreatedAt), strOrEmpty(z.UpdatedAt))
	}
	tbl.Print()
}

func printXfrConfigInbound(out io.Writer, indent string, x *dnszones.XfrConfigInbound) {
	out = textio.NewPrefixWriter(out, indent)

	fmt.Fprintf(out, "Inbound TSIG Key ID: %s\n", strOrEmpty(x.InboundTSIGKeyID))
	for i, p := range x.Primaries {
		fmt.Fprintf(out, "Primary[%d] Address: %s\n", i, strOrEmpty(p.Address))
		fmt.Fprintf(out, "Primary[%d] Description: %s\n", i, strOrEmpty(p.Description))
	}
	if x.NotifyIPAddresses != nil {
		fmt.Fprintf(out, "Notify IPv4 Addresses: %s\n", strings.Join(x.NotifyIPAddresses.IPv4, ", "))
	}
}

func strOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
