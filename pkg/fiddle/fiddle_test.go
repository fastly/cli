package fiddle_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/fastly/cli/pkg/fiddle"
)

// fakeFiddle is a minimal in-memory Fiddle API.
func fakeFiddle(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("POST /fiddle", func(w http.ResponseWriter, r *http.Request) {
		var spec map[string]any
		if err := json.NewDecoder(r.Body).Decode(&spec); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if _, hasVCL := spec["vcl"]; !hasVCL {
			http.Error(w, "missing vcl key", http.StatusInternalServerError)
			return
		}
		fmt.Fprint(w, `{"fiddle":{"id":"cafe0123","srcVersion":0},"valid":true,"lintStatus":{"recv":[]}}`)
	})
	mux.HandleFunc("PUT /fiddle/{id}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"fiddle":{"id":%q,"srcVersion":1},"valid":false,"lintStatus":{"recv":[{"level":"error","line":0,"message":"boom"}]}}`, r.PathValue("id"))
	})
	mux.HandleFunc("POST /fiddle/{id}/execute", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `{"sessionID":"feedbeef00","streamHost":""}`)
	})
	mux.HandleFunc("GET /results/{sid}/stream", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, "event: waitingForSync\ndata: {}\n\n")
		fmt.Fprint(w, "event: updateResult\ndata: {\"execHost\":\"s-r-cafe0123v0-1-100.exec9.fiddle.fastly.dev\",\n")
		fmt.Fprint(w, "data: \"clientFetches\":{\"a\":{\"status\":200,\"complete\":true,\"resp\":\"HTTP/2 200 OK\\ncontent-type: text/plain\",\"bodyPreview\":\"hi\",\"bodyBytesReceived\":2,\"isText\":true}}}\n\n")
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return server
}

func testClient(server *httptest.Server) *fiddle.Client {
	return &fiddle.Client{
		Endpoint:   server.URL,
		HTTPClient: server.Client(),
		UserAgent:  "test",
	}
}

func TestCreateAndUpdate(t *testing.T) {
	client := testClient(fakeFiddle(t))
	saved, err := client.Create(context.Background(), fiddle.Spec{VCL: map[string]string{"recv": "error 601;"}})
	if err != nil {
		t.Fatal(err)
	}
	if saved.ID != "cafe0123" || !saved.Valid {
		t.Fatalf("unexpected create result: %+v", saved)
	}

	saved, err = client.Update(context.Background(), saved.ID, fiddle.Spec{})
	if err != nil {
		t.Fatal(err)
	}
	if saved.Valid {
		t.Fatal("expected invalid VCL on update")
	}
	if len(saved.LintStatus["recv"]) != 1 || saved.LintStatus["recv"][0].Message != "boom" {
		t.Fatalf("unexpected lint status: %+v", saved.LintStatus)
	}
}

func TestRunParsesStream(t *testing.T) {
	client := testClient(fakeFiddle(t))
	result, err := client.Run(context.Background(), "cafe0123", 1, fiddle.StreamOptions{
		WantFetches: 1,
		MaxWait:     5 * time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.CompleteFetches() != 1 {
		t.Fatalf("expected one complete fetch, got %+v", result)
	}
	fetch := result.ClientFetch["a"]
	if fetch.Status != 200 || fetch.BodyPreview != "hi" {
		t.Fatalf("unexpected fetch: %+v", fetch)
	}
	if !strings.Contains(result.ExecHost, "cafe0123v0") {
		t.Fatalf("unexpected exec host: %q", result.ExecHost)
	}
}

func TestErrorBodiesAreTruncatedAndLegible(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "Fiddle validation error in key 'origins': 42\nsecond line", http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)
	client := testClient(server)
	_, err := client.Create(context.Background(), fiddle.Spec{})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "Fiddle validation error") || strings.Contains(err.Error(), "second line") {
		t.Fatalf("unexpected error: %v", err)
	}
}
