package secretstore_test

import (
	"bytes"

	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v8/fastly"
)

func fmtStore(s *fastly.SecretStore) string {
	var b bytes.Buffer
	text.PrintSecretStore(&b, "", s)
	return b.String()
}

func fmtStores(s *fastly.SecretStores) string {
	var b bytes.Buffer
	text.PrintSecretStoresTbl(&b, s)
	return b.String()
}
