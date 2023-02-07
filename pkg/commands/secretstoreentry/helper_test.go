package secretstoreentry_test

import (
	"bytes"

	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v7/fastly"
)

func fmtSecret(s *fastly.Secret) string {
	var b bytes.Buffer
	text.PrintSecret(&b, "", s)
	return b.String()
}

func fmtSecrets(s *fastly.Secrets) string {
	var b bytes.Buffer
	text.PrintSecretsTbl(&b, s)
	return b.String()
}
