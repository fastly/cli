package mcp_test

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/testutil"
)

func TestMCPListCommand_Works(t *testing.T) {
	var stdout bytes.Buffer

	app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
		return testutil.MockGlobalData([]string{"mcp", "list"}, &stdout), nil
	}
	_ = app.Run([]string{"mcp", "list"}, nil) // ignore error, just check output

	helpOutput := stdout.String()
	if want := "api"; !bytes.Contains([]byte(helpOutput), []byte(want)) {
		t.Errorf("expected output to contain %q, got: %s", want, helpOutput)
	}
}

func TestMCPAPICommand_Works(t *testing.T) {
	// Set up a pipe to capture os.Stderr
	r, w, pipeErr := os.Pipe()
	if pipeErr != nil {
		t.Fatalf("failed to create pipe: %v", pipeErr)
	}
	origStderr := os.Stderr
	os.Stderr = w
	defer func() {
		os.Stderr = origStderr
		w.Close()
		r.Close()
	}()

	// Provide a dummy token and API endpoint via the mock global data
	data := testutil.MockGlobalData([]string{"mcp", "api"}, io.Discard)
	data.Config.Profiles["user"].Token = "dummy-token"
	data.Config.Profiles["user"].Email = "dummy@example.com"
	data.Config.Profiles["user"].Default = true
	data.ConfigPath = "/dev/null"
	data.Env.APIEndpoint = "https://api.fastly.com"

	app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
		return data, nil
	}

	done := make(chan struct {
		output string
		err    error
	})
	go func() {
		var buf bytes.Buffer
		n, copyErr := io.Copy(&buf, r)
		_ = n // ignore n
		done <- struct {
			output string
			err    error
		}{buf.String(), copyErr}
	}()

	_ = app.Run([]string{"mcp", "api"}, nil)
	w.Close() // close writer to signal end of output
	result := <-done

	if result.err != nil {
		t.Fatalf("failed to copy stderr: %v", result.err)
	}
	if !bytes.Contains([]byte(result.output), []byte("Starting Fastly API MCP server")) {
		t.Errorf("expected output to contain startup message, got: %s", result.output)
	}
}
