package compute

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/filesystem"
	"github.com/fastly/cli/pkg/text"
	"github.com/kennygrant/sanitize"
	"github.com/mholt/archiver/v3"
)

// IgnoreFilePath is the filepath name of the Fastly ignore file.
const IgnoreFilePath = ".fastlyignore"

// Toolchain abstracts a Compute@Edge source language toolchain.
type Toolchain interface {
	Initialize(out io.Writer) error
	Verify(out io.Writer) error
	Build(out io.Writer, verbose bool) error
}

// Language models a Compute@Edge source language.
type Language struct {
	Name            string
	DisplayName     string
	StarterKits     []StarterKit
	SourceDirectory string
	IncludeFiles    []string

	Toolchain
}

// LanguageOptions models configuration options for a Language.
type LanguageOptions struct {
	Name            string
	DisplayName     string
	StarterKits     []StarterKit
	SourceDirectory string
	IncludeFiles    []string
	Toolchain       Toolchain
}

// NewLanguage constructs a new Language from a LangaugeOptions.
func NewLanguage(options *LanguageOptions) *Language {
	return &Language{
		options.Name,
		options.DisplayName,
		options.StarterKits,
		options.SourceDirectory,
		options.IncludeFiles,
		options.Toolchain,
	}
}

// BuildCommand produces a deployable artifact from files on the local disk.
type BuildCommand struct {
	common.Base
	client     api.HTTPClient
	name       string
	lang       string
	includeSrc bool
	force      bool
}

// NewBuildCommand returns a usable command registered under the parent.
func NewBuildCommand(parent common.Registerer, client api.HTTPClient, globals *config.Data) *BuildCommand {
	var c BuildCommand
	c.Globals = globals
	c.client = client
	c.CmdClause = parent.Command("build", "Build a Compute@Edge package locally")
	c.CmdClause.Flag("name", "Package name").StringVar(&c.name)
	c.CmdClause.Flag("language", "Language type").StringVar(&c.lang)
	c.CmdClause.Flag("include-source", "Include source code in built package").BoolVar(&c.includeSrc)
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

	var language *Language
	switch lang {
	case "assemblyscript":
		language = NewLanguage(&LanguageOptions{
			Name:            "assemblyscript",
			SourceDirectory: "assembly",
			IncludeFiles:    []string{"package.json"},
			Toolchain:       NewAssemblyScript(),
		})
	case "rust":
		language = NewLanguage(&LanguageOptions{
			Name:            "rust",
			SourceDirectory: "src",
			IncludeFiles:    []string{"Cargo.toml"},
			Toolchain:       NewRust(c.client, c.Globals),
		})
	default:
		return fmt.Errorf("unsupported language %s", lang)
	}

	if !c.force {
		progress.Step(fmt.Sprintf("Verifying local %s toolchain...", lang))

		err = language.Verify(progress)
		if err != nil {
			return err
		}
	}

	progress.Step(fmt.Sprintf("Building package using %s toolchain...", lang))

	if err := language.Build(progress, c.Globals.Flag.Verbose); err != nil {
		return err
	}

	progress.Step("Creating package archive...")

	dest := filepath.Join("pkg", fmt.Sprintf("%s.tar.gz", name))

	files := []string{
		ManifestFilename,
	}
	files = append(files, language.IncludeFiles...)

	ignoreFiles, err := getIgnoredFiles(IgnoreFilePath)
	if err != nil {
		return err
	}

	binFiles, err := getNonIgnoredFiles("bin", ignoreFiles)
	if err != nil {
		return err
	}
	files = append(files, binFiles...)

	if c.includeSrc {
		srcFiles, err := getNonIgnoredFiles(language.SourceDirectory, ignoreFiles)
		if err != nil {
			return err
		}
		files = append(files, srcFiles...)
	}

	err = createPackageArchive(files, dest)
	if err != nil {
		return fmt.Errorf("error creating package archive: %w", err)
	}

	progress.Done()

	text.Success(out, "Built %s package %s (%s)", lang, name, dest)
	return nil
}

// createPackageArchive packages build artifacts as a Fastly package, which
// must be a GZipped Tar archive such as: package-name.tar.gz.
//
// Due to a behavior of archiver.Archive() which recursively writes all files in
// a provided directory to the archive we first copy our input files to a
// temporary directory to ensure only the specified files are included and not
// any in the directory which may be ignored.
func createPackageArchive(files []string, destination string) error {
	// Create temporary directory to copy files into.
	p := make([]byte, 8)
	n, err := rand.Read(p)
	if err != nil {
		return fmt.Errorf("error creating temporary directory: %w", err)
	}

	tmpDir := filepath.Join(
		os.TempDir(),
		fmt.Sprintf("fastly-build-%x", p[:n]),
	)

	if err := os.MkdirAll(tmpDir, 0700); err != nil {
		return fmt.Errorf("error creating temporary directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create implicit top-level directory within temp which will become the
	// root of the archive. This replaces the `tar.ImplicitTopLevelFolder`
	// behavior.
	dir := filepath.Join(tmpDir, fileNameWithoutExtension(destination))
	if err := os.Mkdir(dir, 0700); err != nil {
		return fmt.Errorf("error creating temporary directory: %w", err)
	}

	for _, src := range files {
		dst := filepath.Join(dir, src)
		if err = filesystem.CopyFile(src, dst); err != nil {
			return fmt.Errorf("error copying file: %w", err)
		}
	}

	tar := archiver.NewTarGz()
	tar.OverwriteExisting = true //
	tar.MkdirAll = true          // make destination directory if it doesn't exist

	if err = tar.Archive([]string{dir}, destination); err != nil {
		return err
	}

	return nil
}

// fileNameWithoutExtension returns a filename with its extension stripped.
func fileNameWithoutExtension(filename string) string {
	base := filepath.Base(filename)
	firstDot := strings.Index(base, ".")
	if firstDot > -1 {
		return base[:firstDot]
	}
	return base
}

// getIgnoredFiles reads the .fastlyignore file line-by-line and expands the
// glob pattern into a map containing all files it matches. If no ignore file
// is present it returns an empty map.
func getIgnoredFiles(filePath string) (files map[string]bool, err error) {
	files = make(map[string]bool)

	if !filesystem.FileExists(filePath) {
		return files, nil
	}

	// gosec flagged this:
	// G304 (CWE-22): Potential file inclusion via variable
	// Disabling as we trust the source of the filepath variable as it comes
	// from the IgnoreFilePath constant.
	/* #nosec */
	file, err := os.Open(filePath)
	if err != nil {
		return files, err
	}
	defer func() {
		cerr := file.Close()
		if err == nil {
			err = cerr
		}
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		glob := strings.TrimSpace(scanner.Text())
		globFiles, err := filepath.Glob(glob)
		if err != nil {
			return files, fmt.Errorf("parsing glob %s: %w", glob, err)
		}
		for _, f := range globFiles {
			files[f] = true
		}
	}

	if err := scanner.Err(); err != nil {
		return files, fmt.Errorf("reading %s file: %w", filePath, err)
	}

	return files, nil
}

// getNonIgnoredFiles walks a filepath and returns all files don't exist in the
// provided ignore files map.
func getNonIgnoredFiles(base string, ignoredFiles map[string]bool) ([]string, error) {
	var files []string
	err := filepath.Walk(base, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if ignoredFiles[path] {
			return nil
		}
		files = append(files, path)
		return nil
	})

	return files, err
}
