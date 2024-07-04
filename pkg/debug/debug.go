package debug

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
)

// PrintStruct pretty prints the given struct.
func PrintStruct(v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err == nil {
		fmt.Println(string(b))
	}
	return err
}

// DumpHTTPRequest dumps the HTTP network request if --debug-mode is set.
func DumpHTTPRequest(r *http.Request) {
	req := r.Clone(context.Background())
	if req.Header.Get("Fastly-Key") != "" {
		req.Header.Set("Fastly-Key", "REDACTED")
	}
	dump, _ := httputil.DumpRequest(r, true)
	fmt.Printf("\n\nhttp.Request (dump): %q\n\n", dump)
}

// DumpHTTPResponse dumps the HTTP network response if --debug-mode is set.
func DumpHTTPResponse(r *http.Response) {
	if r != nil {
		dump, _ := httputil.DumpResponse(r, true)
		fmt.Printf("\n\nhttp.Response (dump): %q\n\n", dump)
	}
}
