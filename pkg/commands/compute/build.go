package compute

import (
	"bufio"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/filesystem"
	"github.com/fastly/cli/pkg/manifest"
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

// Flags represents the flags defined for the command.
type Flags struct {
	AcceptCustomBuild bool
	IncludeSrc        bool
	Lang              string
	PackageName       string
	SkipVerification  bool
	Timeout           int
}

// BuildCommand produces a deployable artifact from files on the local disk.
type BuildCommand struct {
	cmd.Base
	client api.HTTPClient

	// NOTE: these are public so that the "serve" and "publish" composite
	// commands can set the values appropriately before calling Exec().
	Flags    Flags
	Manifest manifest.Data
}

// NewBuildCommand returns a usable command registered under the parent.
func NewBuildCommand(parent cmd.Registerer, client api.HTTPClient, globals *config.Data, data manifest.Data) *BuildCommand {
	var c BuildCommand
	c.client = client
	c.Globals = globals
	c.Manifest = data
	c.CmdClause = parent.Command("build", "Build a Compute@Edge package locally")

	// NOTE: when updating these flags, be sure to update the composite commands:
	// `compute publish` and `compute serve`.
	c.CmdClause.Flag("accept-custom-build", "Do not prompt when project manifest defines [scripts.build]").BoolVar(&c.Flags.AcceptCustomBuild)
	c.CmdClause.Flag("include-source", "Include source code in built package").BoolVar(&c.Flags.IncludeSrc)
	c.CmdClause.Flag("language", "Language type").StringVar(&c.Flags.Lang)
	c.CmdClause.Flag("name", "Package name").StringVar(&c.Flags.PackageName)
	c.CmdClause.Flag("skip-verification", "Skip verification steps and force build").BoolVar(&c.Flags.SkipVerification)
	c.CmdClause.Flag("timeout", "Timeout, in seconds, for the build compilation step").IntVar(&c.Flags.Timeout)

	return &c
}

// Exec implements the command interface.
func (c *BuildCommand) Exec(in io.Reader, out io.Writer) (err error) {
	progress := text.NewProgress(out, c.Globals.Verbose())

	defer func(errLog fsterr.LogInterface) {
		if err != nil {
			errLog.Add(err)
			progress.Fail() // progress.Done is handled inline
		}
	}(c.Globals.ErrLog)

	progress.Step("Verifying package manifest...")

	err = c.Manifest.File.ReadError()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = fsterr.ErrReadingManifest
		}
		c.Globals.ErrLog.Add(err)
		return err
	}

	// Language from flag takes priority, otherwise infer from manifest and
	// error if neither are provided. Sanitize by trim and lowercase.
	var identLang string
	if c.Flags.Lang != "" {
		identLang = c.Flags.Lang
	} else if c.Manifest.File.Language != "" {
		identLang = c.Manifest.File.Language
	} else {
		return fmt.Errorf("language cannot be empty, please provide a language")
	}
	identLang = strings.ToLower(strings.TrimSpace(identLang))

	// Name from flag takes priority, otherwise infer from manifest
	// error if neither are provided. Sanitize value to ensure it is a safe
	// filepath, replacing spaces with hyphens etc.
	var name string
	if c.Flags.PackageName != "" {
		name = c.Flags.PackageName
	} else if c.Manifest.File.Name != "" {
		name = c.Manifest.File.Name
	} else {
		return fmt.Errorf("name cannot be empty, please provide a name")
	}
	name = sanitize.BaseName(name)

	var language *Language
	switch identLang {
	case "assemblyscript":
		language = NewLanguage(&LanguageOptions{
			Name:            "assemblyscript",
			SourceDirectory: "assembly",
			IncludeFiles:    []string{"package.json"},
			Toolchain:       NewAssemblyScript(c.Flags.Timeout, c.Manifest.File.Scripts.Build, c.Globals.ErrLog),
		})
	case "javascript":
		language = NewLanguage(&LanguageOptions{
			Name:            "javascript",
			SourceDirectory: "src",
			IncludeFiles:    []string{"package.json"},
			Toolchain:       NewJavaScript(c.Flags.Timeout, c.Manifest.File.Scripts.Build, c.Globals.ErrLog),
		})
	case "rust":
		language = NewLanguage(&LanguageOptions{
			Name:            "rust",
			SourceDirectory: "src",
			IncludeFiles:    []string{"Cargo.toml"},
			Toolchain:       NewRust(c.client, c.Globals.File.Language.Rust, c.Globals.ErrLog, c.Flags.Timeout, c.Manifest.File.Scripts.Build),
		})
	case "other":
		language = NewLanguage(&LanguageOptions{
			Name:      "other",
			Toolchain: NewOther(c.Flags.Timeout, c.Manifest.File.Scripts.Build, c.Globals.ErrLog),
		})
	default:
		return fmt.Errorf("unsupported language %s", identLang)
	}

	toolchain := identLang
	if c.Manifest.File.Scripts.Build != "" {
		toolchain = "custom"
	}

	// NOTE: We don't verify custom build scripts.
	if c.Manifest.File.Scripts.Build == "" && !c.Flags.SkipVerification {
		progress.Step(fmt.Sprintf("Verifying local %s toolchain...", toolchain))

		err = language.Verify(progress)
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
				"Language": language.Name,
			})
			return err
		}
	}

	// NOTE: We set the progress indicator to Done() so that any output we now
	// print doesn't get hidden by the progress status.
	progress.Done()

	customBuildMsg := "This project has a custom build script defined in the fastly.toml manifest.\n"

	if toolchain == "custom" {
		if c.Globals.Flag.Verbose {
			text.Info(out, customBuildMsg)
			text.Break(out)
			text.Indent(out, 4, "%s", c.Manifest.File.Scripts.Build)
			customBuildMsg = ""
		}
		if !c.Flags.AcceptCustomBuild {
			// NOTE: A third-party could share a project with a build command for a
			// language that wouldn't normally require one (e.g. Rust), and do evil
			// things. So we should notify the user and confirm they would like to
			// continue with the build.
			label := fmt.Sprintf("\n%sAre you sure you want to continue with the build step? [y/N] ", customBuildMsg)
			cont, err := text.Input(out, label, in)
			if err != nil {
				return fmt.Errorf("error reading input %w", err)
			}
			contl := strings.ToLower(cont)
			if contl != "y" && contl != "yes" {
				text.Info(out, "Stopping the build process.")
				return nil
			}
			text.Break(out)
		}
	}

	progress = text.ResetProgress(out, c.Globals.Verbose())
	progress.Step(fmt.Sprintf("Building package using %s toolchain...", toolchain))

	if err := language.Build(progress, c.Globals.Flag.Verbose); err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
			"Language": language.Name,
		})
		return err
	}

	progress.Step("Creating package archive...")

	dest := filepath.Join("pkg", fmt.Sprintf("%s.tar.gz", name))

	files := []string{
		manifest.Filename,
	}
	files = append(files, language.IncludeFiles...)

	ignoreFiles, err := GetIgnoredFiles(IgnoreFilePath)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	binFiles, err := GetNonIgnoredFiles("bin", ignoreFiles)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
			"Ignore files": ignoreFiles,
		})
		return err
	}
	files = append(files, binFiles...)

	if c.Flags.IncludeSrc {
		srcFiles, err := GetNonIgnoredFiles(language.SourceDirectory, ignoreFiles)
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
				"Source directory": language.SourceDirectory,
				"Ignore files":     ignoreFiles,
			})
			return err
		}
		files = append(files, srcFiles...)
	}

	err = CreatePackageArchive(files, dest)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
			"Files":       files,
			"Destination": dest,
		})
		return fmt.Errorf("error creating package archive: %w", err)
	}

	progress.Done()

	text.Success(out, "Built package '%s' (%s)", name, dest)
	return nil
}

// CreatePackageArchive packages build artifacts as a Fastly package, which
// must be a GZipped Tar archive such as: package-name.tar.gz.
//
// Due to a behavior of archiver.Archive() which recursively writes all files in
// a provided directory to the archive we first copy our input files to a
// temporary directory to ensure only the specified files are included and not
// any in the directory which may be ignored.
func CreatePackageArchive(files []string, destination string) error {
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
	dir := filepath.Join(tmpDir, FileNameWithoutExtension(destination))
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

// FileNameWithoutExtension returns a filename with its extension stripped.
func FileNameWithoutExtension(filename string) string {
	base := filepath.Base(filename)
	firstDot := strings.Index(base, ".")
	if firstDot > -1 {
		return base[:firstDot]
	}
	return base
}

// GetIgnoredFiles reads the .fastlyignore file line-by-line and expands the
// glob pattern into a map containing all files it matches. If no ignore file
// is present it returns an empty map.
func GetIgnoredFiles(filePath string) (files map[string]bool, err error) {
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

// GetNonIgnoredFiles walks a filepath and returns all files don't exist in the
// provided ignore files map.
func GetNonIgnoredFiles(base string, ignoredFiles map[string]bool) ([]string, error) {
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
