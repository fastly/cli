package vcl

import (
	"testing"

	"github.com/fastly/cli/pkg/fiddle"
)

func TestPickFetchRedirectChain(t *testing.T) {
	fetches := map[string]fiddle.ClientFetch{
		"a": {Status: 301, Complete: true},
		"b": {Status: 302, Complete: true, IsRedirected: true},
	}
	if f := pickFetch(&fiddle.Result{ClientFetch: fetches}); f == nil || f.Status != 302 {
		t.Fatalf("expected the redirect Fiddle stopped on, got %+v", f)
	}

	fetches["c"] = fiddle.ClientFetch{Status: 200, Complete: true, IsRedirected: true}
	if f := pickFetch(&fiddle.Result{ClientFetch: fetches}); f == nil || f.Status != 200 {
		t.Fatalf("expected the end of the redirect chain, got %+v", f)
	}
}
