package fmt

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/fastly/cli/pkg/text"
)

// Success is a test helper used to generate output for asserting against.
func Success(format string, args ...any) string {
	var b bytes.Buffer
	text.Success(&b, format, args...)
	return b.String()
}

// Info is a test helper used to generate output for asserting against.
func Info(format string, args ...any) string {
	var b bytes.Buffer
	text.Info(&b, format, args...)
	return b.String()
}

// JSON decodes then re-encodes back to JSON, with indentation matching
// that of ../cmd/argparser.go's argparser.WriteJSON.
func JSON(format string, args ...any) string {
	var r json.RawMessage
	if err := json.Unmarshal([]byte(fmt.Sprintf(format, args...)), &r); err != nil {
		panic(err)
	}
	return EncodeJSON(r)
}

// EncodeJSON is a test helper that encodes any Go type into JSON.
func EncodeJSON(value any) string {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetIndent("", "  ")
	_ = enc.Encode(value)
	return b.String()
}
