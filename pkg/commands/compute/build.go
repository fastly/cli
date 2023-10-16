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
	"time"

	"github.com/kennygrant/sanitize"
	"github.com/mholt/archiver/v3"

	"github.com/fastly/cli/pkg/cmd"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/filesystem"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
)

// IgnoreFilePath is the filepath name of the Fastly ignore file.
const IgnoreFilePath = ".fastlyignore"

// CustomPostScriptMessage is the message displayed to a user when there is
// either a post_init or post_build script defined.
const CustomPostScriptMessage = "This project has a custom post_%s script defined in the %s manifest"

// Flags represents the flags defined for the command.
type Flags struct {
	Dir         string
	Env         string
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
	Flags Flags
}

// NewBuildCommand returns a usable command registered under the parent.
func NewBuildCommand(parent cmd.Registerer, g *global.Data) *BuildCommand {
	var c BuildCommand
	c.Globals = g
	c.CmdClause = parent.Command("build", "Build a Compute package locally")

	// NOTE: when updating these flags, be sure to update the composite commands:
	// `compute publish` and `compute serve`.
	c.CmdClause.Flag("dir", "Project directory to build (default: current directory)").Short('C').StringVar(&c.Flags.Dir)
	c.CmdClause.Flag("env", "The manifest environment config to use (e.g. 'stage' will attempt to read 'fastly.stage.toml')").StringVar(&c.Flags.Env)
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

	manifestFilename := EnvironmentManifest(c.Flags.Env)
	if c.Flags.Env != "" {
		if c.Globals.Verbose() {
			text.Info(out, EnvManifestMsg, manifestFilename, manifest.Filename)
		}
	}
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}
	defer os.Chdir(wd)
	manifestPath := filepath.Join(wd, manifestFilename)

	projectDir, err := ChangeProjectDirectory(c.Flags.Dir)
	if err != nil {
		return err
	}
	if projectDir != "" {
		if c.Globals.Verbose() {
			text.Info(out, ProjectDirMsg, projectDir)
		}
		manifestPath = filepath.Join(projectDir, manifestFilename)
	}

	spinner, err := text.NewSpinner(out)
	if err != nil {
		return err
	}

	defer func(errLog fsterr.LogInterface) {
		if err != nil {
			errLog.Add(err)
		}
	}(c.Globals.ErrLog)

	err = spinner.Process(fmt.Sprintf("Verifying %s", manifestFilename), func(_ *text.SpinnerWrapper) error {
		if projectDir != "" || c.Flags.Env != "" {
			err = c.Globals.Manifest.File.Read(manifestPath)
		} else {
			err = c.Globals.Manifest.File.ReadError()
		}
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				err = fsterr.ErrReadingManifest
			}
			c.Globals.ErrLog.Add(err)
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	var pkgName string
	err = spinner.Process("Identifying package name", func(_ *text.SpinnerWrapper) error {
		pkgName, err = c.PackageName(manifestFilename)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	var toolchain string
	err = spinner.Process("Identifying toolchain", func(_ *text.SpinnerWrapper) error {
		toolchain, err = identifyToolchain(c)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	language, err := language(toolchain, manifestFilename, c, in, out, spinner)
	if err != nil {
		return err
	}

	err = binDir(c)
	if err != nil {
		return err
	}

	if err := language.Build(); err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Language": language.Name,
		})
		return err
	}

	dest := filepath.Join("pkg", fmt.Sprintf("%s.tar.gz", pkgName))

	err = spinner.Process("Creating package archive", func(_ *text.SpinnerWrapper) error {
		// IMPORTANT: The minimum package requirement is `fastly.toml` and `main.wasm`.
		//
		// The Fastly platform will reject a package that doesn't have a manifest
		// named exactly fastly.toml which means if the user is building and
		// deploying a package with an environment manifest (e.g. fastly.stage.toml)
		// then we need to:
		//
		// 1. Rename any existing fastly.toml to fastly.toml.backup.<TIMESTAMP>
		// 2. Make a temp copy of the environment manifest and name it fastly.toml
		// 3. Remove the newly created fastly.toml once the packaging is done
		// 4. Rename the fastly.toml.backup back to fastly.toml
		if c.Flags.Env != "" {
			// 1. Rename any existing fastly.toml to fastly.toml.backup.<TIMESTAMP>
			//
			// For example, the user is trying to deploy a fastly.stage.toml rather
			// than the standard fastly.toml manifest.
			if _, err := os.Stat(manifest.Filename); err == nil {
				backup := fmt.Sprintf("%s.backup.%d", manifest.Filename, time.Now().Unix())
				if err := os.Rename(manifest.Filename, backup); err != nil {
					return fmt.Errorf("failed to backup primary manifest file: %w", err)
				}
				defer func() {
					// 4. Rename the fastly.toml.backup back to fastly.toml
					if err = os.Rename(backup, manifest.Filename); err != nil {
						text.Error(out, err.Error())
					}
				}()
			} else {
				// 3. Remove the newly created fastly.toml once the packaging is done
				//
				// If there wasn't an existing fastly.toml because the user only wants
				// to work with environment manifests (e.g. fastly.stage.toml and
				// fastly.production.toml) then we should remove the fastly.toml that we
				// created just for the packaging process (see step 2. below).
				defer func() {
					if err = os.Remove(manifest.Filename); err != nil {
						text.Error(out, err.Error())
					}
				}()
			}
			// 2. Make a temp copy of the environment manifest and name it fastly.toml
			//
			// If there was no existing fastly.toml then this step will create one, so
			// we need to make sure we remove it after packaging has finished so as to
			// not confuse the user with a fastly.toml that has suddenly appeared (see
			// step 3. above).
			if err := filesystem.CopyFile(manifestFilename, manifest.Filename); err != nil {
				return fmt.Errorf("failed to copy environment manifest file: %w", err)
			}
		}

		files := []string{
			manifest.Filename,
			"bin/main.wasm",
		}
		files, err = c.includeSourceCode(files, language.SourceDirectory)
		if err != nil {
			return err
		}
		err = CreatePackageArchive(files, dest)
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
				"Files":       files,
				"Destination": dest,
			})
			return fmt.Errorf("error creating package archive: %w", err)
		}
		return nil
	})
	if err != nil {
		return err
	}

	out = originalOut
	text.Success(out, "\nBuilt package (%s)", dest)
	return nil
}

// includeSourceCode calculates what source code files to include in the final
// package.tar.gz that is uploaded to the Fastly API.
//
// TODO: Investigate possible change to --include-source flag.
// The following implementation presumes source code is stored in a constant
// location, which might not be true for all users. We should look at whether
// we should change the --include-source flag to not be a boolean but to
// accept a 'source code' path instead.
func (c *BuildCommand) includeSourceCode(files []string, srcDir string) ([]string, error) {
	empty := make([]string, 0)

	if c.Flags.IncludeSrc {
		ignoreFiles, err := GetIgnoredFiles(IgnoreFilePath)
		if err != nil {
			c.Globals.ErrLog.Add(err)
			return empty, err
		}

		binFiles, err := GetNonIgnoredFiles("bin", ignoreFiles)
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
				"Ignore files": ignoreFiles,
			})
			return empty, err
		}
		files = append(files, binFiles...)

		srcFiles, err := GetNonIgnoredFiles(srcDir, ignoreFiles)
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
				"Source directory": srcDir,
				"Ignore files":     ignoreFiles,
			})
			return empty, err
		}
		files = append(files, srcFiles...)
	}

	return files, nil
}

// packageName acquires the package name from either a flag or manifest.
// Additionally it will sanitize the name.
func (c *BuildCommand) PackageName(manifestFilename string) (string, error) {
	var name string

	switch {
	case c.Flags.PackageName != "":
		name = c.Flags.PackageName
	case c.Globals.Manifest.File.Name != "":
		name = c.Globals.Manifest.File.Name // use the project name as a fallback
	default:
		return "", fsterr.RemediationError{
			Inner:       fmt.Errorf("package name is missing"),
			Remediation: fmt.Sprintf("Add a name to the %s 'name' field. Reference: https://developer.fastly.com/reference/compute/fastly-toml/", manifestFilename),
		}
	}

	return sanitize.BaseName(name), nil
}

// identifyToolchain determines the programming language.
//
// It prioritises the --language flag over the manifest field.
// Will error if neither are provided.
// Lastly, it will normalise with a trim and lowercase.
func identifyToolchain(c *BuildCommand) (string, error) {
	var toolchain string

	switch {
	case c.Flags.Lang != "":
		toolchain = c.Flags.Lang
	case c.Globals.Manifest.File.Language != "":
		toolchain = c.Globals.Manifest.File.Language
	default:
		return "", fmt.Errorf("language cannot be empty, please provide a language")
	}

	return strings.ToLower(strings.TrimSpace(toolchain)), nil
}

// language returns a pointer to a supported language.
func language(toolchain, manifestFilename string, c *BuildCommand, in io.Reader, out io.Writer, spinner text.Spinner) (*Language, error) {
	var language *Language
	switch toolchain {
	case "assemblyscript":
		language = NewLanguage(&LanguageOptions{
			Name:            "assemblyscript",
			SourceDirectory: AsSourceDirectory,
			Toolchain: NewAssemblyScript(
				&c.Globals.Manifest.File,
				c.Globals,
				c.Flags,
				in,
				manifestFilename,
				out,
				spinner,
			),
		})
	case "go":
		language = NewLanguage(&LanguageOptions{
			Name:            "go",
			SourceDirectory: GoSourceDirectory,
			Toolchain: NewGo(
				&c.Globals.Manifest.File,
				c.Globals,
				c.Flags,
				in,
				manifestFilename,
				out,
				spinner,
			),
		})
	case "javascript":
		language = NewLanguage(&LanguageOptions{
			Name:            "javascript",
			SourceDirectory: JsSourceDirectory,
			Toolchain: NewJavaScript(
				&c.Globals.Manifest.File,
				c.Globals,
				c.Flags,
				in,
				manifestFilename,
				out,
				spinner,
			),
		})
	case "rust":
		language = NewLanguage(&LanguageOptions{
			Name:            "rust",
			SourceDirectory: RustSourceDirectory,
			Toolchain: NewRust(
				&c.Globals.Manifest.File,
				c.Globals,
				c.Flags,
				in,
				manifestFilename,
				out,
				spinner,
			),
		})
	case "other":
		language = NewLanguage(&LanguageOptions{
			Name: "other",
			Toolchain: NewOther(
				&c.Globals.Manifest.File,
				c.Globals,
				c.Flags,
				in,
				manifestFilename,
				out,
				spinner,
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
		text.Info(c.Globals.Output, "\nCreating ./bin directory (for Wasm binary)\n\n")
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
