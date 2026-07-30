package compute

import (
	"runtime"
	"sort"
	"strings"

	"github.com/blang/semver"

	"github.com/fastly/cli/pkg/revision"
	"github.com/fastly/cli/pkg/starterkit"
)

// NewLanguages returns a list of supported programming languages.
//
// NOTE: The 'timeout' value zero is passed into each New<Language> call as it's
// only useful during the `compute build` phase and is expected to be
// provided by the user via a flag on the build command.
//
// StarterKits are not populated here -- they depend on a live fetch from the
// starter-kit edge service, and are populated lazily via FetchStarterKits
// only when the interactive starter-kit prompt is actually about to run.
func NewLanguages() []*Language {
	// WARNING: Do not reorder these options as they affect the rendered output.
	// They are placed in order of language maturity/importance.
	//
	// A change to this order will also break the tests, as the logic defaults to
	// the first language in the list if nothing entered at the relevant language
	// prompt.
	return []*Language{
		NewLanguage(&LanguageOptions{
			Name:        "rust",
			DisplayName: "Rust",
		}),
		NewLanguage(&LanguageOptions{
			Name:        "javascript",
			DisplayName: "JavaScript",
		}),
		NewLanguage(&LanguageOptions{
			Name:        "go",
			DisplayName: "Go",
		}),
		NewLanguage(&LanguageOptions{
			Name:        "cpp",
			DisplayName: "C++",
		}),
		NewLanguage(&LanguageOptions{
			Name:        "other",
			DisplayName: "Other ('bring your own' Wasm binary)",
		}),
	}
}

// NewLanguage constructs a new Language from a LangaugeOptions.
func NewLanguage(options *LanguageOptions) *Language {
	return &Language{
		options.Name,
		options.DisplayName,
		nil,
		options.SourceDirectory,
		options.Toolchain,
	}
}

// Language models a Compute source language.
type Language struct {
	Name            string
	DisplayName     string
	StarterKits     []starterkit.Kit
	SourceDirectory string

	Toolchain
}

// LanguageOptions models configuration options for a Language.
type LanguageOptions struct {
	Name            string
	DisplayName     string
	SourceDirectory string
	Toolchain       Toolchain
}

// FetchStarterKits populates l.StarterKits from the starter-kit edge
// service, filtered server-side to this language, further filtered to kits
// marked for CLI display and supported by the running CLI version, with the
// "default" kit (if present) sorted first.
func (l *Language) FetchStarterKits(client *starterkit.Client) error {
	kits, err := client.Kits(l.Name)
	if err != nil {
		return err
	}

	kits = filterByShowOnCLI(kits)
	kits = filterByMinCLIVersion(kits)

	sort.Slice(kits, func(i, j int) bool {
		a := kits[i].KitName() == "default"
		b := kits[j].KitName() == "default"
		return a && !b
	})

	l.StarterKits = kits
	return nil
}

// filterByShowOnCLI drops kits not marked catalog.show_on_cli. The edge
// service is also asked to filter server-side (Client.Kits passes ?cli=true),
// but we don't rely on that alone since Kit already carries this field.
func filterByShowOnCLI(kits []starterkit.Kit) []starterkit.Kit {
	filtered := make([]starterkit.Kit, 0, len(kits))
	for _, kit := range kits {
		if kit.Catalog.ShowOnCLI {
			filtered = append(filtered, kit)
		}
	}
	return filtered
}

// filterByMinCLIVersion drops kits whose catalog.min_cli_version exceeds the
// running CLI's version. Kits with an unparseable or missing
// min_cli_version, or a running CLI version that itself can't be parsed
// (e.g. local dev builds without version info baked in via LDFLAGS), are
// kept rather than hidden.
func filterByMinCLIVersion(kits []starterkit.Kit) []starterkit.Kit {
	// revision.None is the AppVersion for local/dev builds without LDFLAGS
	// version info baked in -- treat it the same as "unknown", not as a real
	// (very old) version, otherwise every kit with a min_cli_version would be
	// hidden for anyone running from source.
	if revision.AppVersion == revision.None {
		return kits
	}

	current, err := semver.Parse(strings.TrimPrefix(revision.AppVersion, "v"))
	if err != nil {
		return kits
	}

	filtered := make([]starterkit.Kit, 0, len(kits))
	for _, kit := range kits {
		minVersion, err := semver.Parse(strings.TrimPrefix(kit.Catalog.MinCLIVersion, "v"))
		if kit.Catalog.MinCLIVersion == "" || err != nil || !current.LT(minVersion) {
			filtered = append(filtered, kit)
		}
	}
	return filtered
}

// Shell represents a subprocess shell used by `compute` environment where
// `[scripts.build]` has been defined within fastly.toml manifest.
type Shell struct{}

// Build expects a command that can be prefixed with an appropriate subprocess
// shell.
//
// Example:
// build = "yarn install && yarn build"
//
// Should be converted into a command such as (on unix):
// sh -c "yarn install && yarn build".
func (s Shell) Build(command string) (cmd string, args []string) {
	cmd = "sh"
	args = []string{"-c"}

	if runtime.GOOS == "windows" {
		cmd = "cmd.exe"
		args = []string{"/C"}
	}

	args = append(args, command)

	return cmd, args
}
