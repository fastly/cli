package starterkit_test

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/starterkit"
	"github.com/fastly/cli/pkg/testutil"
)

func TestParseFrom(t *testing.T) {
	for _, testcase := range []struct {
		name     string
		input    string
		wantLang string
		wantName string
		wantOK   bool
	}{
		{
			name:     "valid",
			input:    "starter-kit/javascript/typescript-default",
			wantLang: "javascript",
			wantName: "typescript-default",
			wantOK:   true,
		},
		{
			name:     "valid with hyphenated kit name",
			input:    "starter-kit/rust/connect-google-bigquery",
			wantLang: "rust",
			wantName: "connect-google-bigquery",
			wantOK:   true,
		},
		{
			name:   "missing prefix",
			input:  "javascript/typescript-default",
			wantOK: false,
		},
		{
			name:   "just the prefix",
			input:  "starter-kit/",
			wantOK: false,
		},
		{
			name:   "missing name segment",
			input:  "starter-kit/javascript",
			wantOK: false,
		},
		{
			name:   "too many segments",
			input:  "starter-kit/javascript/typescript-default/extra",
			wantOK: false,
		},
		{
			name:   "empty lang segment",
			input:  "starter-kit//typescript-default",
			wantOK: false,
		},
		{
			name:   "empty name segment",
			input:  "starter-kit/javascript/",
			wantOK: false,
		},
		{
			name:   "plain github url",
			input:  "https://github.com/fastly/compute-starter-kit-rust-default",
			wantOK: false,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			lang, name, ok := starterkit.ParseFrom(testcase.input)
			testutil.AssertBool(t, testcase.wantOK, ok)
			if testcase.wantOK {
				testutil.AssertString(t, testcase.wantLang, lang)
				testutil.AssertString(t, testcase.wantName, name)
			}
		})
	}
}

func TestKitFromValue(t *testing.T) {
	kit := starterkit.Kit{
		ID:       "javascript-typescript-kv-store",
		Language: "javascript",
	}
	testutil.AssertString(t, "typescript-kv-store", kit.KitName())
	testutil.AssertString(t, "starter-kit/javascript/typescript-kv-store", kit.FromValue())
}

func TestClientTarballURL(t *testing.T) {
	c := starterkit.New("https://example.com/", nil, false)
	testutil.AssertString(t, "https://example.com/kits/javascript/typescript-default/tarball", c.TarballURL("javascript", "typescript-default"))
}

func TestClientKits(t *testing.T) {
	for _, testcase := range []struct {
		name      string
		status    int
		body      string
		wantErr   string
		wantCount int
	}{
		{
			name:      "success",
			status:    200,
			body:      `{"generated_at":"2026-07-13T00:00:00Z","kits":[{"id":"javascript-typescript-default","name":"TypeScript","language":"javascript","description":"desc","catalog":{"show_on_cli":true,"min_cli_version":"16.0.0"}}]}`,
			wantCount: 1,
		},
		{
			name:    "non-200",
			status:  500,
			body:    "",
			wantErr: "failed to fetch starter kit manifest",
		},
		{
			name:    "malformed json",
			status:  200,
			body:    "not json",
			wantErr: "failed to parse starter kit manifest",
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			res := mock.NewHTTPResponse(testcase.status, nil, io.NopCloser(strings.NewReader(testcase.body)))
			httpClient := mock.HTMLClient([]*http.Response{res}, []error{nil})

			c := starterkit.New("https://example.com", httpClient, false)
			kits, err := c.Kits("")

			if testcase.wantErr != "" {
				testutil.AssertErrorContains(t, err, testcase.wantErr)
				return
			}
			testutil.AssertNoError(t, err)
			testutil.AssertLength(t, testcase.wantCount, kits)
		})
	}
}
