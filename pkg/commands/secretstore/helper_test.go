package secretstore_test

import (
	"bytes"

	"github.com/fastly/go-fastly/v10/fastly"

	"github.com/fastly/cli/pkg/text"
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
