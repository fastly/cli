package compute

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/starterkit"
)

// starterKitIDFile is the well-known marker file a legacy
// fastly/compute-starter-kit-* repo may contain at its root, redirecting
// `compute init` to the new starter-kit edge service instead of cloning the
// legacy repo directly.
const starterKitIDFile = ".starter-kit-id"

// starterKitRedirect checks whether the given (Fastly-owned) GitHub repo URL
// has a root-level .starter-kit-id marker file, and if so, parses its
// content as a "starter-kit/<lang>/<name>" reference.
//
// Any failure along the way (the repo URL doesn't parse into an org/repo, the
// marker file is missing or unreadable, the request errors, or the content
// doesn't parse) is treated identically: ok is false, and the caller should
// fall back to cloning the legacy repo directly.
func starterKitRedirect(httpClient api.HTTPClient, repoURL string) (lang, name string, ok bool) {
	org, repo, ok := githubOrgRepo(repoURL)
	if !ok {
		return "", "", false
	}

	rawURL := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/HEAD/%s", org, repo, starterKitIDFile)
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return "", "", false
	}

	res, err := httpClient.Do(req)
	if err != nil {
		return "", "", false
	}
	defer res.Body.Close() // #nosec G307

	if res.StatusCode != http.StatusOK {
		return "", "", false
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", "", false
	}

	return starterkit.ParseFrom(strings.TrimSpace(string(body)))
}

// githubOrgRepo extracts the "org" and "repo" path segments from a
// github.com repository URL, stripping any ".git" suffix from the repo name.
func githubOrgRepo(repoURL string) (org, repo string, ok bool) {
	u, err := url.Parse(repoURL)
	if err != nil {
		return "", "", false
	}

	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", false
	}

	return parts[0], strings.TrimSuffix(parts[1], ".git"), true
}
