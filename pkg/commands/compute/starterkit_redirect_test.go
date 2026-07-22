package compute

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/fastly/cli/pkg/mock"
)

func TestGithubOrgRepo(t *testing.T) {
	for _, testcase := range []struct {
		name     string
		input    string
		wantOrg  string
		wantRepo string
		wantOK   bool
	}{
		{
			name:     "plain",
			input:    "https://github.com/fastly/compute-starter-kit-rust-default",
			wantOrg:  "fastly",
			wantRepo: "compute-starter-kit-rust-default",
			wantOK:   true,
		},
		{
			name:     "with .git suffix",
			input:    "https://github.com/fastly/compute-starter-kit-rust-default.git",
			wantOrg:  "fastly",
			wantRepo: "compute-starter-kit-rust-default",
			wantOK:   true,
		},
		{
			name:     "with trailing slash",
			input:    "https://github.com/fastly/compute-starter-kit-rust-default/",
			wantOrg:  "fastly",
			wantRepo: "compute-starter-kit-rust-default",
			wantOK:   true,
		},
		{
			name:   "too few path segments",
			input:  "https://github.com/fastly",
			wantOK: false,
		},
		{
			name:   "too many path segments",
			input:  "https://github.com/fastly/foo/archive/refs/heads/main.zip",
			wantOK: false,
		},
		{
			name:   "unparseable",
			input:  "://not a url",
			wantOK: false,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			org, repo, ok := githubOrgRepo(testcase.input)
			if ok != testcase.wantOK {
				t.Fatalf("ok: want %v, have %v", testcase.wantOK, ok)
			}
			if testcase.wantOK {
				if org != testcase.wantOrg {
					t.Errorf("org: want %q, have %q", testcase.wantOrg, org)
				}
				if repo != testcase.wantRepo {
					t.Errorf("repo: want %q, have %q", testcase.wantRepo, repo)
				}
			}
		})
	}
}

func TestStarterKitRedirect(t *testing.T) {
	for _, testcase := range []struct {
		name     string
		repoURL  string
		status   int
		body     string
		err      error
		wantLang string
		wantName string
		wantOK   bool
	}{
		{
			name:     "marker file present and valid",
			repoURL:  "https://github.com/fastly/compute-starter-kit-typescript-default",
			status:   http.StatusOK,
			body:     "starter-kit/javascript/typescript-default\n",
			wantLang: "javascript",
			wantName: "typescript-default",
			wantOK:   true,
		},
		{
			name:    "marker file absent (404)",
			repoURL: "https://github.com/fastly/compute-starter-kit-rust-empty",
			status:  http.StatusNotFound,
			body:    "",
			wantOK:  false,
		},
		{
			name:    "marker file present but malformed",
			repoURL: "https://github.com/fastly/compute-starter-kit-rust-empty",
			status:  http.StatusOK,
			body:    "not-a-valid-reference",
			wantOK:  false,
		},
		{
			name:    "network error",
			repoURL: "https://github.com/fastly/compute-starter-kit-rust-empty",
			err:     io.ErrUnexpectedEOF,
			wantOK:  false,
		},
		{
			name:    "repo URL doesn't parse into org/repo",
			repoURL: "https://github.com/fastly",
			wantOK:  false,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			var httpClient *mock.HTTPClient
			switch {
			case testcase.err != nil:
				httpClient = mock.NewHTTPClientWithErrors([]error{testcase.err})
			case testcase.status != 0:
				res := mock.NewHTTPResponse(testcase.status, nil, io.NopCloser(strings.NewReader(testcase.body)))
				httpClient = mock.NewHTTPClientWithResponses([]*http.Response{res})
			default:
				httpClient = mock.NewHTTPClientWithResponses(nil)
			}

			lang, name, ok := starterKitRedirect(httpClient, testcase.repoURL)
			if ok != testcase.wantOK {
				t.Fatalf("ok: want %v, have %v", testcase.wantOK, ok)
			}
			if testcase.wantOK {
				if lang != testcase.wantLang {
					t.Errorf("lang: want %q, have %q", testcase.wantLang, lang)
				}
				if name != testcase.wantName {
					t.Errorf("name: want %q, have %q", testcase.wantName, name)
				}
			}
		})
	}
}
