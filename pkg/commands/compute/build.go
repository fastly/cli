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

	"github.com/fastly/cli/pkg/cmd"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/filesystem"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/kennygrant/sanitize"
	"github.com/mholt/archiver/v3"
)

// IgnoreFilePath is the filepath name of the Fastly ignore file.
const IgnoreFilePath = ".fastlyignore"

// CustomPostBuildScriptMessage is the message displayed to a user when there is a
// custom post build script.
const CustomPostBuildScriptMessage = "This project has a custom post build script defined in the fastly.toml manifest"

// Flags represents the flags defined for the command.
type Flags struct {
	IncludeSrc  bool
	Lang        string
	PackageName string
	Timeout     int
}

// BuildCommand produces a deployable artifact from files on the local disk.
type BuildCommand struct {
	cmd.Base

	// NOTE: these are public so that the "serve" and "publish" composite
	// commands can set the values appropriately before calling Exec().
	Flags    Flags
	Manifest manifest.Data
}

// NewBuildCommand returns a usable command registered under the parent.
func NewBuildCommand(parent cmd.Registerer, g *global.Data, m manifest.Data) *BuildCommand {
	var c BuildCommand
	c.Globals = g
	c.Manifest = m
	c.CmdClause = parent.Command("build", "Build a Compute@Edge package locally")

	// NOTE: when updating these flags, be sure to update the composite commands:
	// `compute publish` and `compute serve`.
	c.CmdClause.Flag("include-source", "Include source code in built package").BoolVar(&c.Flags.IncludeSrc)
	c.CmdClause.Flag("language", "Language type").StringVar(&c.Flags.Lang)
	c.CmdClause.Flag("package-name", "Package name").StringVar(&c.Flags.PackageName)
	c.CmdClause.Flag("timeout", "Timeout, in seconds, for the build compilation step").IntVar(&c.Flags.Timeout)

	return &c
}

// Exec implements the command interface.
func (c *BuildCommand) Exec(in io.Reader, out io.Writer) (err error) {
	// We'll restore this at the end to print a final successful build output.
	originalOut := out

	if c.Globals.Flags.Quiet {
		out = io.Discard
	}
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

	packageName, err := packageName(c)
	if err != nil {
		return err
	}

	toolchain, err := toolchain(c)
	if err != nil {
		return err
	}

	language, err := language(toolchain, c, progress)
	if err != nil {
		return err
	}

	err = binDir(c)
	if err != nil {
		return err
	}

	// NOTE: We set the progress indicator to Done() so that any output we now
	// print doesn't get hidden by the progress status.
	progress.Done()

	if c.Globals.Verbose() {
		text.Break(out)
	}

	progress = text.ResetProgress(out, c.Globals.Verbose())

	postBuildCallback := func() error {
		if !c.Globals.Flags.AutoYes && !c.Globals.Flags.NonInteractive {
			err := promptForBuildContinue(CustomPostBuildScriptMessage, c.Manifest.File.Scripts.PostBuild, out, in, c.Globals.Verbose())
			if err != nil {
				return err
			}
		}
		return nil
	}

	if err := language.Build(out, progress, c.Globals.Flags.Verbose, postBuildCallback); err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Language": language.Name,
		})
		return err
	}

	if c.Globals.Verbose() {
		text.Break(out)
	}

	progress = text.ResetProgress(out, c.Globals.Verbose())
	progress.Step("Creating package archive...")

	dest := filepath.Join("pkg", fmt.Sprintf("%s.tar.gz", packageName))

	files := []string{
		manifest.Filename,
	}

	ignoreFiles, err := GetIgnoredFiles(IgnoreFilePath)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	binFiles, err := GetNonIgnoredFiles("bin", ignoreFiles)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Ignore files": ignoreFiles,
		})
		return err
	}
	files = append(files, binFiles...)

	if c.Flags.IncludeSrc {
		srcFiles, err := GetNonIgnoredFiles(language.SourceDirectory, ignoreFiles)
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
				"Source directory": language.SourceDirectory,
				"Ignore files":     ignoreFiles,
			})
			return err
		}
		files = append(files, srcFiles...)
	}

	err = CreatePackageArchive(files, dest)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Files":       files,
			"Destination": dest,
		})
		return fmt.Errorf("error creating package archive: %w", err)
	}

	progress.Done()

	out = originalOut
	text.Success(out, "Built package (%s)", dest)
	return nil
}

// packageName acquires the package name from either a flag or manifest.
// Additionally it will sanitize the name.
func packageName(c *BuildCommand) (string, error) {
	var name string

	switch {
	case c.Flags.PackageName != "":
		name = c.Flags.PackageName
	case c.Manifest.File.Name != "":
		name = c.Manifest.File.Name // use the project name as a fallback
	default:
		return "", fsterr.RemediationError{
			Inner:       fmt.Errorf("package name is missing"),
			Remediation: "Add a name to the fastly.toml 'name' field. Reference: https://developer.fastly.com/reference/compute/fastly-toml/",
		}
	}

	return sanitize.BaseName(name), nil
}

// toolchain determines the programming language.
//
// It prioritises the --language flag over the manifest field.
// Will error if neither are provided.
// Lastly, it will normalise with a trim and lowercase.
func toolchain(c *BuildCommand) (string, error) {
	var toolchain string

	switch {
	case c.Flags.Lang != "":
		toolchain = c.Flags.Lang
	case c.Manifest.File.Language != "":
		toolchain = c.Manifest.File.Language
	default:
		return "", fmt.Errorf("language cannot be empty, please provide a language")
	}

	return strings.ToLower(strings.TrimSpace(toolchain)), nil
}

// language returns a pointer to a supported language.
func language(toolchain string, c *BuildCommand, progress text.Progress) (*Language, error) {
	var language *Language
	switch toolchain {
	case "assemblyscript":
		language = NewLanguage(&LanguageOptions{
			Name:            "assemblyscript",
			SourceDirectory: AsSourceDirectory,
			Toolchain: NewAssemblyScript(
				&c.Manifest.File,
				c.Globals.ErrLog,
				c.Flags.Timeout,
				progress,
				c.Globals.Verbose(),
			),
		})
	case "go":
		language = NewLanguage(&LanguageOptions{
			Name:            "go",
			SourceDirectory: GoSourceDirectory,
			Toolchain: NewGo(
				&c.Manifest.File,
				c.Globals.ErrLog,
				c.Flags.Timeout,
				c.Globals.Config.Language.Go,
				progress,
				c.Globals.Verbose(),
			),
		})
	case "javascript":
		language = NewLanguage(&LanguageOptions{
			Name:            "javascript",
			SourceDirectory: JsSourceDirectory,
			Toolchain: NewJavaScript(
				&c.Manifest.File,
				c.Globals.ErrLog,
				c.Flags.Timeout,
				progress,
				c.Globals.Verbose(),
			),
		})
	case "rust":
		language = NewLanguage(&LanguageOptions{
			Name:            "rust",
			SourceDirectory: RustSourceDirectory,
			Toolchain: NewRust(
				&c.Manifest.File,
				c.Globals.ErrLog,
				c.Flags.Timeout,
				c.Globals.Config.Language.Rust,
				progress,
				c.Globals.Verbose(),
			),
		})
	case "other":
		language = NewLanguage(&LanguageOptions{
			Name: "other",
			Toolchain: NewOther(
				c.Manifest.File.Scripts,
				c.Globals.ErrLog,
				c.Flags.Timeout,
			),
		})
	default:
		return nil, fmt.Errorf("unsupported language %s", toolchain)
	}

	return language, nil
}

// binDir ensures a ./bin directory exists.
// The directory is required so a main.wasm can be placed inside it.
func binDir(c *BuildCommand) error {
	if c.Globals.Verbose() {
		text.Info(c.Globals.Output, "Creating ./bin directory (for Wasm binary)")
	}
	dir, err := os.Getwd()
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return fmt.Errorf("failed to identify the current working directory: %w", err)
	}
	binDir := filepath.Join(dir, "bin")
	if err := filesystem.MakeDirectoryIfNotExists(binDir); err != nil {
		c.Globals.ErrLog.Add(err)
		return fmt.Errorf("failed to create bin directory: %w", err)
	}
	return nil
}

// promptForBuildContinue ensures the user is happy to continue with the build
// when there is either a custom build or post build in the fastly.toml
// manifest file.
func promptForBuildContinue(msg, script string, out io.Writer, in io.Reader, verbose bool) error {
	text.Info(out, "%s:\n", msg)
	text.Break(out)
	text.Indent(out, 4, "%s", script)

	var post string
	if msg == CustomPostBuildScriptMessage {
		post = "post "
	}

	label := fmt.Sprintf("\nAre you sure you want to continue with the %sbuild step? [y/N] ", post)
	answer, err := text.AskYesNo(out, label, in)
	if err != nil {
		return err
	}
	if !answer {
		text.Info(out, "Stopping the %sbuild process.", post)
		if !verbose {
			text.Break(out)
		}
		return fsterr.ErrBuildStopped
	}
	text.Break(out)
	return nil
}

// CreatePackageArchive packages build artifacts as a Fastly package.
// The package must be a GZipped Tar archive.
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

	if err := os.MkdirAll(tmpDir, 0o700); err != nil {
		return fmt.Errorf("error creating temporary directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create implicit top-level directory within temp which will become the
	// root of the archive. This replaces the `tar.ImplicitTopLevelFolder`
	// behavior.
	dir := filepath.Join(tmpDir, FileNameWithoutExtension(destination))
	if err := os.Mkdir(dir, 0o700); err != nil {
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

	return tar.Archive([]string{dir}, destination)
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

// GetNonIgnoredFiles walks a filepath and returns all files that don't exist in
// the provided ignore files map.
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
