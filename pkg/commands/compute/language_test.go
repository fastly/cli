package compute

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/revision"
	"github.com/fastly/cli/pkg/starterkit"
)

func TestFilterByShowOnCLI(t *testing.T) {
	kits := []starterkit.Kit{
		{ID: "rust-default", Catalog: starterkit.Catalog{ShowOnCLI: true}},
		{ID: "rust-internal-only", Catalog: starterkit.Catalog{ShowOnCLI: false}},
		{ID: "rust-unset"},
	}

	got := filterByShowOnCLI(kits)
	if len(got) != 1 {
		t.Fatalf("want 1 kit, have %d", len(got))
	}
	if got[0].ID != "rust-default" {
		t.Errorf("want rust-default to survive filtering, have %q", got[0].ID)
	}
}

func TestFilterByMinCLIVersion(t *testing.T) {
	kits := []starterkit.Kit{
		{ID: "rust-default", Language: "rust", Catalog: starterkit.Catalog{MinCLIVersion: "16.0.0"}},
		{ID: "rust-old", Language: "rust", Catalog: starterkit.Catalog{MinCLIVersion: "1.0.0"}},
		{ID: "rust-no-min", Language: "rust"},
		{ID: "rust-bad-min", Language: "rust", Catalog: starterkit.Catalog{MinCLIVersion: "not-a-version"}},
	}

	t.Run("dev build (revision.None) keeps everything", func(t *testing.T) {
		orig := revision.AppVersion
		revision.AppVersion = revision.None
		defer func() { revision.AppVersion = orig }()

		got := filterByMinCLIVersion(kits)
		if len(got) != len(kits) {
			t.Fatalf("want %d kits, have %d", len(kits), len(got))
		}
	})

	t.Run("real version filters out kits requiring a newer CLI", func(t *testing.T) {
		orig := revision.AppVersion
		revision.AppVersion = "v15.4.0"
		defer func() { revision.AppVersion = orig }()

		got := filterByMinCLIVersion(kits)
		if len(got) != 3 {
			t.Fatalf("want 3 kits, have %d", len(got))
		}
		for _, id := range []string{"rust-old", "rust-no-min", "rust-bad-min"} {
			found := false
			for _, k := range got {
				if k.ID == id {
					found = true
				}
			}
			if !found {
				t.Errorf("expected kit %q to survive filtering", id)
			}
		}
		for _, k := range got {
			if k.ID == "rust-default" {
				t.Errorf("expected rust-default (min_cli_version 16.0.0) to be filtered out")
			}
		}
	})

	t.Run("unparseable running version keeps everything", func(t *testing.T) {
		orig := revision.AppVersion
		revision.AppVersion = "not-a-version"
		defer func() { revision.AppVersion = orig }()

		got := filterByMinCLIVersion(kits)
		if len(got) != len(kits) {
			t.Fatalf("want %d kits, have %d", len(kits), len(got))
		}
	})
}

func TestLanguageFetchStarterKits(t *testing.T) {
	orig := revision.AppVersion
	revision.AppVersion = revision.None // avoid min_cli_version filtering noise in this test
	defer func() { revision.AppVersion = orig }()

	body := `{"generated_at":"2026-07-13T00:00:00Z","kits":[
		{"id":"rust-websockets","name":"WebSockets","language":"rust","description":"desc","catalog":{"show_on_cli":true}},
		{"id":"rust-default","name":"Default","language":"rust","description":"desc","catalog":{"show_on_cli":true}},
		{"id":"rust-internal-only","name":"Internal","language":"rust","description":"desc","catalog":{"show_on_cli":false}}
	]}`
	res := mock.NewHTTPResponse(http.StatusOK, nil, io.NopCloser(strings.NewReader(body)))
	httpClient := mock.NewHTTPClientWithResponses([]*http.Response{res})
	client := starterkit.New("https://example.com", httpClient, false)

	lang := NewLanguage(&LanguageOptions{Name: "rust", DisplayName: "Rust"})
	if err := lang.FetchStarterKits(client); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(lang.StarterKits) != 2 {
		t.Fatalf("want 2 starter kits, have %d", len(lang.StarterKits))
	}
	// The "default" kit should be sorted first regardless of manifest order.
	if got := lang.StarterKits[0].KitName(); got != "default" {
		t.Errorf("want first starter kit to be %q, have %q", "default", got)
	}
}
