package testutil

import "encoding/json"

// GenJSON returns JSON encoding of data, or empty object in case of an error.
func GenJSON(data any) []byte {
	b, err := json.Marshal(data)
	if err != nil {
		return []byte("{}")
	}
	return b
}
