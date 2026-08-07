package compute_test

import (
	"testing"

	toml "github.com/pelletier/go-toml"

	"github.com/fastly/cli/pkg/commands/compute"
	"github.com/fastly/cli/pkg/config"
)

// TestStaticConfigStarterKits validates that every language the CLI offers at
// the `compute init` prompt has at least one starter kit in the static config
// embedded into the binary.
//
// The starter kits are injected into the static config at build time by
// ./scripts/config.sh, which holds its own hardcoded list of starter kit
// repositories. Adding a language without also updating that list leaves users
// of the new language unable to init a project (see CDTOOL-1707), and this test
// is here to catch that drift.
//
// NOTE: The static config is generated, not committed, so run `make config`
// before running this test locally.
func TestStaticConfigStarterKits(t *testing.T) {
	var f config.File
	if err := toml.Unmarshal(config.Static, &f); err != nil {
		t.Fatalf("failed to unmarshal the static config: %s", err)
	}

	for _, language := range compute.NewLanguages(f.StarterKits) {
		// The 'other' language is for users bringing their own Wasm binary and so
		// has no starter kits by design.
		if language.Name == "other" {
			continue
		}
		if len(language.StarterKits) == 0 {
			t.Errorf("no starter kits found in the static config for language %q: add its starter kit repositories to the 'kits' list in ./scripts/config.sh", language.Name)
		}
	}
}
