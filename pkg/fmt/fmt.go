package fmt

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/fastly/cli/pkg/text"
)

func Success(format string, args ...any) string {
	var b bytes.Buffer
	text.Success(&b, format, args...)
	return b.String()
}

// JSON decodes then re-encodes back to JSON, with indentation matching
// that of jsonOutput.WriteJSON.
func JSON(format string, args ...any) string {
	var r json.RawMessage
	if err := json.Unmarshal([]byte(fmt.Sprintf(format, args...)), &r); err != nil {
		panic(err)
	}
	return EncodeJSON(r)
}

func EncodeJSON(value any) string {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetIndent("", "  ")
	_ = enc.Encode(value)
	return b.String()
}
