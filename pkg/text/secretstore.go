package text

import (
	"encoding/hex"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v7/fastly"
	"github.com/segmentio/textio"
)

func PrintSecretStoresTbl(out io.Writer, stores *fastly.SecretStores) {
	tbl := NewTable(out)
	tbl.AddHeader("Name", "ID")

	if stores == nil {
		tbl.Print()
		return
	}

	for _, s := range stores.Data {
		// avoid gosec loop aliasing check :/
		s := s
		tbl.AddLine(s.Name, s.ID)
	}
	tbl.Print()

	if stores.Meta.NextCursor != "" {
		fmt.Fprintf(out, "\nNext cursor: %s\n", stores.Meta.NextCursor)
	}
}

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

func PrintSecretStore(out io.Writer, prefix string, s *fastly.SecretStore) {
	out = textio.NewPrefixWriter(out, prefix)

	fmt.Fprintf(out, "Name: %s\n", s.Name)
	fmt.Fprintf(out, "ID: %s\n", s.ID)
}

func PrintSecret(out io.Writer, prefix string, s *fastly.Secret) {
	out = textio.NewPrefixWriter(out, prefix)

	fmt.Fprintf(out, "Name: %s\n", s.Name)
	fmt.Fprintf(out, "Digest: %s\n", hex.EncodeToString(s.Digest))
}
