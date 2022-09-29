package compute

import (
	"fmt"
	"io"
	"runtime"
	"sort"
	"strings"

	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/manifest"
)

// NewLanguages returns a list of supported programming languages.
//
// NOTE: The 'timeout' value zero is passed into each New<Language> call as it's
// only useful during the `compute build` phase and is expected to be
// provided by the user via a flag on the build command.
func NewLanguages(kits config.StarterKitLanguages, d *config.Data, fastlyManifest manifest.File, out io.Writer) []*Language {
	// WARNING: Do not reorder these options as they affect the rendered output.
	// They are placed in order of language maturity/importance.
	//
	// A change to this order will also break the tests, as the logic defaults to
	// the first language in the list if nothing entered at the relevant language
	// prompt.
	return []*Language{
		NewLanguage(&LanguageOptions{
			Name:        "rust",
			DisplayName: "Rust (limited availability)",
			StarterKits: kits.Rust,
			Toolchain: NewRust(
				&fastlyManifest,
				d.ErrLog,
				0,
				d.File.Language.Rust,
				out,
				nil,
			),
		}),
		NewLanguage(&LanguageOptions{
			Name:        "javascript",
			DisplayName: "JavaScript (limited availability)",
			StarterKits: kits.JavaScript,
			Toolchain: NewJavaScript(
				&fastlyManifest,
				d.ErrLog,
				0,
				out,
				nil,
			),
		}),
		NewLanguage(&LanguageOptions{
			Name:        "go",
			DisplayName: "Go (beta)",
			StarterKits: kits.Go,
			Toolchain: NewGo(
				&fastlyManifest,
				d.ErrLog,
				0,
				d.File.Language.Go,
				out,
				nil,
			),
		}),
		NewLanguage(&LanguageOptions{
			Name:        "assemblyscript",
			DisplayName: "AssemblyScript (beta)",
			StarterKits: kits.AssemblyScript,
			Toolchain: NewAssemblyScript(
				&fastlyManifest,
				d.ErrLog,
				0,
				out,
				nil,
			),
		}),
		NewLanguage(&LanguageOptions{
			Name:        "other",
			DisplayName: "Other ('bring your own' Wasm binary)",
		}),
	}
}

// Language models a Compute@Edge source language.
type Language struct {
	Name            string
	DisplayName     string
	StarterKits     []config.StarterKit
	SourceDirectory string
	IncludeFiles    []string

	Toolchain
}

// LanguageOptions models configuration options for a Language.
type LanguageOptions struct {
	Name            string
	DisplayName     string
	StarterKits     []config.StarterKit
	SourceDirectory string
	IncludeFiles    []string
	Toolchain       Toolchain
}

// NewLanguage constructs a new Language from a LangaugeOptions.
func NewLanguage(options *LanguageOptions) *Language {
	// Ensure the 'default' starter kit is always first.
	sort.Slice(options.StarterKits, func(i, j int) bool {
		suffix := fmt.Sprintf("%s-default", options.Name)
		a := strings.HasSuffix(options.StarterKits[i].Path, suffix)
		b := strings.HasSuffix(options.StarterKits[j].Path, suffix)
		var (
			bitSetA int8
			bitSetB int8
		)
		if a {
			bitSetA = 1
		}
		if b {
			bitSetB = 1
		}
		return bitSetA > bitSetB
	})

	return &Language{
		options.Name,
		options.DisplayName,
		options.StarterKits,
		options.SourceDirectory,
		options.IncludeFiles,
		options.Toolchain,
	}
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
// sh -c "yarn install && yarn build"
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
