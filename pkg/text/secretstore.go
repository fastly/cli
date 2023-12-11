package text

import (
	"encoding/hex"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v8/fastly"
	"github.com/segmentio/textio"
)

// PrintSecretStoresTbl displays store data in a table format.
func PrintSecretStoresTbl(out io.Writer, stores []fastly.SecretStore) {
	tbl := NewTable(out)
	tbl.AddHeader("Name", "ID")

	for _, store := range stores {
		tbl.AddLine(store.Name, store.StoreID)
	}
	tbl.Print()
}

// PrintSecretsTbl displays secrets data in a table format.
func PrintSecretsTbl(out io.Writer, secrets *fastly.Secrets) {
	tbl := NewTable(out)
	tbl.AddHeader("Name", "Digest")

	if secrets == nil {
		tbl.Print()
		return
	}

	for _, s := range secrets.Data {
		// avoid gosec loop aliasing check :/
		s := s
		tbl.AddLine(s.Name, hex.EncodeToString(s.Digest))
	}
	tbl.Print()

	if secrets.Meta.NextCursor != "" {
		fmt.Fprintf(out, "\nNext cursor: %s\n", secrets.Meta.NextCursor)
	}
}

// PrintSecretStore displays store data.
func PrintSecretStore(out io.Writer, prefix string, s *fastly.SecretStore) {
	out = textio.NewPrefixWriter(out, prefix)

	fmt.Fprintf(out, "Name: %s\n", s.Name)
	fmt.Fprintf(out, "ID: %s\n", s.StoreID)
}

// PrintSecret displays store data.
func PrintSecret(out io.Writer, prefix string, s *fastly.Secret) {
	out = textio.NewPrefixWriter(out, prefix)

	fmt.Fprintf(out, "Name: %s\n", s.Name)
	fmt.Fprintf(out, "Digest: %s\n", hex.EncodeToString(s.Digest))
}
