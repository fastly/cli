package text

import (
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v9/fastly/computeacls"
	"github.com/segmentio/textio"
)

// PrintComputeACL displays a compute ACL.
func PrintComputeACL(out io.Writer, prefix string, acl *computeacls.ComputeACL) {
	out = textio.NewPrefixWriter(out, prefix)

	fmt.Fprintf(out, "\nID: %s\n", acl.ComputeACLID)
	fmt.Fprintf(out, "Name: %s\n", acl.Name)
}

// PrintComputeACLsTbl displays compute ACLs in a table format.
func PrintComputeACLsTbl(out io.Writer, acls []computeacls.ComputeACL) {
	tbl := NewTable(out)
	tbl.AddHeader("Name", "ID")

	if acls == nil {
		tbl.Print()
		return
	}

	for _, acl := range acls {
		// avoid gosec loop aliasing check :/
		acl := acl
		tbl.AddLine(acl.Name, acl.ComputeACLID)
	}
	tbl.Print()
}

// PrintComputeACLEntry displays a compute ACL entry.
func PrintComputeACLEntry(out io.Writer, prefix string, entry *computeacls.ComputeACLEntry) {
	out = textio.NewPrefixWriter(out, prefix)

	fmt.Fprintf(out, "\nPrefix: %s\n", entry.Prefix)
	fmt.Fprintf(out, "Action: %s\n", entry.Action)
}

// PrintComputeACLEntriesTbl displays compute ACL entries in a table format.
func PrintComputeACLEntriesTbl(out io.Writer, entries []computeacls.ComputeACLEntry) {
	tbl := NewTable(out)
	tbl.AddHeader("Prefix", "Action")

	if entries == nil {
		tbl.Print()
		return
	}

	for _, entry := range entries {
		// avoid gosec loop aliasing check :/
		entry := entry
		tbl.AddLine(entry.Prefix, entry.Action)
	}
	tbl.Print()
}
