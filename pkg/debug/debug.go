package debug

import (
	"encoding/json"
	"fmt"
)

// PrintStruct pretty prints the given struct.
func PrintStruct(v any) (err error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err == nil {
		fmt.Println(string(b))
	}
	return
}
