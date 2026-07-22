package starterkit

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/debug"
)

// DefaultEndpoint is the base URL of the starter-kit edge service.
const DefaultEndpoint = "https://compute-starter-kits.fastly.dev"

// Catalog represents the catalog-specific metadata of a starter kit, i.e. how
// it should be surfaced to end users (docs site, CLI, etc).
type Catalog struct {
	ShowOnDocs    bool     `json:"show_on_docs"`
	ShowOnCLI     bool     `json:"show_on_cli"`
	Tags          []string `json:"tags"`
	Topics        []string `json:"topics"`
	MinCLIVersion string   `json:"min_cli_version"`
	Slug          string   `json:"slug"`
}

// Kit represents a single starter kit entry in the manifest.
type Kit struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Path        string  `json:"path"`
	Language    string  `json:"language"`
	Description string  `json:"description"`
	Catalog     Catalog `json:"catalog"`
}

// KitName returns the kit-specific portion of the ID (the ID with the
// language prefix removed). This is safe to derive because the language is
// known explicitly from the Language field, unlike guessing a split point
// from the number of hyphens in the ID.
func (k Kit) KitName() string {
	return strings.TrimPrefix(k.ID, k.Language+"-")
}

// FromValue returns the canonical --from value that resolves to this kit,
// e.g. "starter-kit/javascript/typescript-default".
func (k Kit) FromValue() string {
	return "starter-kit/" + k.Language + "/" + k.KitName()
}

// Manifest represents the full /kits response.
type Manifest struct {
	GeneratedAt string `json:"generated_at"`
	Kits        []Kit  `json:"kits"`
}

// Client is a client for the starter-kit edge service.
type Client struct {
	endpoint   string
	httpClient api.HTTPClient
	debug      bool
}

// New returns a usable Client.
func New(endpoint string, httpClient api.HTTPClient, debugMode bool) *Client {
	return &Client{
		endpoint:   strings.TrimSuffix(endpoint, "/"),
		httpClient: httpClient,
		debug:      debugMode,
	}
}

// Kits fetches the starter-kit manifest, filtered server-side to kits with
// catalog.show_on_cli set. If lang is non-empty, results are additionally
// filtered server-side to that language.
func (c *Client) Kits(lang string) ([]Kit, error) {
	q := url.Values{}
	q.Set("cli", "true")
	if lang != "" {
		q.Set("lang", lang)
	}

	reqURL := c.endpoint + "/kits?" + q.Encode()
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to construct starter kit manifest request: %w", err)
	}

	if c.debug {
		debug.DumpHTTPRequest(req)
	}
	res, err := c.httpClient.Do(req)
	if c.debug {
		debug.DumpHTTPResponse(res)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch starter kit manifest: %w", err)
	}
	defer res.Body.Close() // #nosec G307

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch starter kit manifest '%s': %s", reqURL, res.Status)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read starter kit manifest response: %w", err)
	}

	var manifest Manifest
	if err := json.Unmarshal(body, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse starter kit manifest: %w", err)
	}

	return manifest.Kits, nil
}

// TarballURL builds the tarball download URL for the given language/kit-name
// pair. It performs no HTTP request.
func (c *Client) TarballURL(lang, name string) string {
	return fmt.Sprintf("%s/kits/%s/%s/tarball", c.endpoint, lang, name)
}

// ParseFrom parses a --from value of the form "starter-kit/<lang>/<name>"
// into its (lang, name) parts. Any other shape (missing prefix, wrong number
// of segments, empty segments) returns ok == false.
func ParseFrom(from string) (lang, name string, ok bool) {
	const prefix = "starter-kit/"
	if !strings.HasPrefix(from, prefix) {
		return "", "", false
	}

	rest := strings.TrimPrefix(from, prefix)
	parts := strings.Split(rest, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", false
	}

	return parts[0], parts[1], true
}
