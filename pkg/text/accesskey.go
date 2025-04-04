package text

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/time"
	"github.com/fastly/go-fastly/v10/fastly/objectstorage/accesskeys"
)

// PrintAccessKey displays an access key.
func PrintAccessKey(out io.Writer, accessKey *accesskeys.AccessKey) {

	fmt.Fprintf(out, "ID: %s\n", accessKey.AccessKeyID)
	fmt.Fprintf(out, "Secret: %s\n", accessKey.SecretKey)
	fmt.Fprintf(out, "Description: %s\n", accessKey.Description)
	fmt.Fprintf(out, "Permission: %s\n", accessKey.Permission)
	fmt.Fprintf(out, "Buckets: %s\n", accessKey.Buckets)
	fmt.Fprintf(out, "Created (UTC): %s\n", accessKey.CreatedAt.UTC().Format(time.Format))
}

// PrintAccessKeyTbl displays access keys in a table format.
func PrintAccessKeyTbl(out io.Writer, accessKeys []accesskeys.AccessKey) {
	tbl := NewTable(out)
	tbl.AddHeader("ID", "Secret", "Description", "Permssion", "Buckets", "Created At")

	if accessKeys == nil {
		tbl.Print()
		return
	}

	for _, accessKey := range accessKeys {
		// avoid gosec loop aliasing check :/
		accessKey := accessKey
		tbl.AddLine(accessKey.AccessKeyID, accessKey.SecretKey, accessKey.Description, accessKey.Permission, accessKey.Buckets, accessKey.CreatedAt)
	}
	tbl.Print()
}
