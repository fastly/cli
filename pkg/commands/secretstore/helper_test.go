package secretstore_test

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v7/fastly"
)

func fmtSuccess(format string, args ...any) string {
	var b bytes.Buffer
	text.Success(&b, format, args...)
	return b.String()
}

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

// fmtJSON decodes then re-encodes back to JSON, with indentation matching
// that of jsonOutput.WriteJSON.
func fmtJSON(format string, args ...any) string {
	var r json.RawMessage
	if err := json.Unmarshal([]byte(fmt.Sprintf(format, args...)), &r); err != nil {
		panic(err)
	}
	return encodeJSON(r)
}

func encodeJSON(value any) string {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetIndent("", "  ")
	enc.Encode(value)
	return b.String()
}
