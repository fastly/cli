package testutil

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

type ExpectedRequest struct {
	Method string
	Path   string

	// Body: nil means “don’t care”.
	// If non-nil:
	//   - empty string means “expect empty body”
	//   - otherwise compare as JSON or raw (see below)
	WantJSON *string

	// Header assertions:
	RequireHeaders http.Header // headers that must be present (value matching rules below)
	ForbidHeaders  []string    // header names that must NOT be present
}

// AssertRequest asserts on the request's properties.  If WantJSON has a value then
// the body of the Request will be consumed.
func AssertRequest(t *testing.T, got *http.Request, exp ExpectedRequest) {
	t.Helper()

	if got.Method != exp.Method {
		t.Fatalf("method got %q want %q", got.Method, exp.Method)
	}
	if got.URL.Path != exp.Path {
		t.Fatalf("path got %q want %q", got.URL.Path, exp.Path)
	}

	// Headers (require/forbid only)
	for _, k := range exp.ForbidHeaders {
		ck := http.CanonicalHeaderKey(k)
		if _, ok := got.Header[ck]; ok {
			t.Fatalf("header %q must not be present", k)
		}
	}
	for k, wantVals := range exp.RequireHeaders {
		ck := http.CanonicalHeaderKey(k)
		gotVals, ok := got.Header[ck]
		if !ok {
			t.Fatalf("missing required header %q", k)
		}
		if len(wantVals) == 0 {
			continue // presence-only
		}
		if !reflect.DeepEqual(gotVals, wantVals) {
			t.Fatalf("header %q got %v want %v", k, gotVals, wantVals)
		}
	}

	// Body (JSON semantic compare)
	if exp.WantJSON != nil {
		gotBody, err := io.ReadAll(got.Body)
		if err != nil {
			t.Fatalf("can't read body")
		}
		gotTrim := bytes.TrimSpace(gotBody)
		if len(strings.TrimSpace(*exp.WantJSON)) == 0 {
			if len(gotTrim) != 0 {
				t.Fatalf("expected empty body, got %q", string(gotBody))
			}
			return
		}

		var gv any
		var ev any
		if err := json.Unmarshal(gotTrim, &gv); err != nil {
			t.Fatalf("got body not valid JSON: %v; body=%q", err, string(gotBody))
		}
		if err := json.Unmarshal([]byte(*exp.WantJSON), &ev); err != nil {
			t.Fatalf("expected JSON not valid: %v; json=%q", err, *exp.WantJSON)
		}
		if !reflect.DeepEqual(gv, ev) {
			t.Fatalf("JSON body mismatch\n got: %s\nwant: %s", string(gotBody), *exp.WantJSON)
		}
	}
}
