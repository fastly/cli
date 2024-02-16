package secretstore_test

import (
	"bytes"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/v10/pkg/text"
)

func fmtStore(s *fastly.SecretStore) string {
	var b bytes.Buffer
	text.PrintSecretStore(&b, "", s)
	return b.String()
}

func fmtStores(s []fastly.SecretStore) string {
	var b bytes.Buffer
	text.PrintSecretStoresTbl(&b, s)
	return b.String()
}
