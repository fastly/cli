package vcl

import (
	"testing"

	"github.com/fastly/cli/pkg/fiddle"
)

func TestPickFetchPrefersRedirectedRequest(t *testing.T) {
	result := &fiddle.Result{ClientFetch: map[string]fiddle.ClientFetch{
		"a": {Status: 301, Complete: true},
		"b": {Status: 302, Complete: true, IsRedirected: true},
	}}
	if f := pickFetch(result); f == nil || f.Status != 302 {
		t.Fatalf("expected the response the redirect led to, got %+v", f)
	}
}
