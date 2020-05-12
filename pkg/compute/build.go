package compute

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/text"
	"github.com/kennygrant/sanitize"
	"github.com/mholt/archiver/v3"
)

// Toolchain abstracts a Compute@Edge source language toolchain.
type Toolchain interface {
	Verify(out io.Writer) error
	Build(out io.Writer, verbose bool) error
}

func createPackageArchive(files []string, destination string) error {
	tar := archiver.NewTarGz()
	tar.OverwriteExisting = true      //
	tar.ImplicitTopLevelFolder = true // prevent extracting to PWD
	tar.MkdirAll = true               // make destination directory if it doesn't exist

	err := tar.Archive(files, destination)
	if err != nil {
		return fmt.Errorf("error creating package archive: %w", err)
	}

	return nil
}

// BuildCommand produces a deployable artifact from files on the local disk.
type BuildCommand struct {
	common.Base
	client api.HTTPClient
	name   string
	lang   string
	force  bool
}

// NewBuildCommand returns a usable command registered under the parent.
func NewBuildCommand(parent common.Registerer, client api.HTTPClient, globals *config.Data) *BuildCommand {
	var c BuildCommand
	c.Globals = globals
	c.client = client
	c.CmdClause = parent.Command("build", "Build a Compute@Edge package locally")
	c.CmdClause.Flag("name", "Package name").StringVar(&c.name)
	c.CmdClause.Flag("language", "Language type").StringVar(&c.lang)
	c.CmdClause.Flag("force", "Skip verification steps and force build").BoolVar(&c.force)
	return &c
}

// Exec implements the command interface.
func (c *BuildCommand) Exec(in io.Reader, out io.Writer) (err error) {
	var progress text.Progress
	if c.Globals.Verbose() {
		progress = text.NewVerboseProgress(out)
	} else {
		progress = text.NewQuietProgress(out)
	}

	defer func() {
		if err != nil {
			progress.Fail() // progress.Done is handled inline
		}
	}()

	progress.Step("Verifying package manifest...")

	var m manifest.File
	if err := m.Read(ManifestFilename); err != nil {
		return fmt.Errorf("error reading package manifest: %w", err)
	}

	// Language from flag takes priority, otherwise infer from manifest and
	// error if neither are provided. Sanitize by trim and lowercase.
	var lang string
	if c.lang != "" {
		lang = c.lang
	} else if m.Language != "" {
		lang = m.Language
	} else {
		return fmt.Errorf("language cannot be empty, please provide a language")
	}
	lang = strings.ToLower(strings.TrimSpace(lang))

	// Name from flag takes priority, otherwise infer from manifest
	// error if neither are provided. Sanitize value to ensure it is a safe
	// filepath, replacing spaces with hyphens etc.
	var name string
	if c.name != "" {
		name = c.name
	} else if m.Name != "" {
		name = m.Name
	} else {
		return fmt.Errorf("name cannot be empty, please provide a name")
	}
	name = sanitize.BaseName(name)

	var toolchain Toolchain
	switch lang {
	case "rust":
		toolchain = &Rust{c.client}
	default:
		return fmt.Errorf("unsupported language %s", lang)
	}

	if !c.force {
		progress.Step(fmt.Sprintf("Verifying local %s toolchain...", lang))

		err = toolchain.Verify(progress)
		if err != nil {
			return err
		}
	}

	progress.Step(fmt.Sprintf("Building package using %s toolchain...", lang))

	if err := toolchain.Build(progress, c.Globals.Flag.Verbose); err != nil {
		return err
	}

	progress.Step("Creating package archive...")

	dest := filepath.Join("pkg", fmt.Sprintf("%s.tar.gz", name))
	err = createPackageArchive([]string{
		ManifestFilename,
		"Cargo.toml",
		"bin",
		"src",
	}, dest)
	if err != nil {
		return err
	}

	progress.Done()
	text.Success(out, "Built %s package %s (%s)", lang, name, dest)
	return nil
}
