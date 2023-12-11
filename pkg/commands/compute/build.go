package compute

import (
	"bufio"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/kennygrant/sanitize"
	"github.com/mholt/archiver/v3"
	"golang.org/x/text/cases"
	textlang "golang.org/x/text/language"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/check"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/filesystem"
	"github.com/fastly/cli/pkg/github"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/revision"
	"github.com/fastly/cli/pkg/text"
)

// IgnoreFilePath is the filepath name of the Fastly ignore file.
const IgnoreFilePath = ".fastlyignore"

// CustomPostScriptMessage is the message displayed to a user when there is
// either a post_init or post_build script defined.
const CustomPostScriptMessage = "This project has a custom post_%s script defined in the %s manifest"

// ErrWasmtoolsNotFound represents an error finding the binary installed.
var ErrWasmtoolsNotFound = fsterr.RemediationError{
	Inner:       fmt.Errorf("wasm-tools not found"),
	Remediation: fsterr.BugRemediation,
}

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
	argparser.Base

	// NOTE: Composite commands require these build flags to be public.
	// e.g. serve, publish, hashsum, hash-files
	// This is so they can set values appropriately before calling Build.Exec().
	Flags                 Flags
	MetadataDisable       bool
	MetadataFilterEnvVars string
	MetadataShow          bool
}

// NewBuildCommand returns a usable command registered under the parent.
func NewBuildCommand(parent argparser.Registerer, g *global.Data) *BuildCommand {
	var c BuildCommand
	c.Globals = g
	c.CmdClause = parent.Command("build", "Build a Compute package locally")

	// NOTE: when updating these flags, be sure to update the composite commands:
	// `compute publish` and `compute serve`.
	c.CmdClause.Flag("dir", "Project directory to build (default: current directory)").Short('C').StringVar(&c.Flags.Dir)
	c.CmdClause.Flag("env", "The manifest environment config to use (e.g. 'stage' will attempt to read 'fastly.stage.toml')").StringVar(&c.Flags.Env)
	c.CmdClause.Flag("include-source", "Include source code in built package").BoolVar(&c.Flags.IncludeSrc)
	c.CmdClause.Flag("language", "Language type").StringVar(&c.Flags.Lang)
	c.CmdClause.Flag("metadata-disable", "Disable Wasm binary metadata annotations").BoolVar(&c.MetadataDisable)
	c.CmdClause.Flag("metadata-filter-envvars", "Redact specified environment variables from [scripts.env_vars] using comma-separated list").StringVar(&c.MetadataFilterEnvVars)
	c.CmdClause.Flag("metadata-show", "Inspect the Wasm binary metadata").BoolVar(&c.MetadataShow)
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
	defer func() {
		_ = os.Chdir(wd)
	}()
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

	wasmtools, wasmtoolsErr := GetWasmTools(spinner, out, c.Globals.Versioners.WasmTools, c.Globals)

	var pkgName string
	err = spinner.Process("Identifying package name", func(_ *text.SpinnerWrapper) error {
		pkgName, err = c.PackageName(manifestFilename)
		// nosemgrep: Users.integralist.Code.go.semgrep-go.err-nil-check (false positive)
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
		// nosemgrep: Users.integralist.Code.go.semgrep-go.err-nil-check (false positive)
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

	// IMPORTANT: We ignore errors downloading wasm-tools.
	// This is because we don't want to block a user from building their project.
	// Annotating the compiled binary with metadata isn't that important.
	if wasmtoolsErr == nil {
		metadataProcessedBy := fmt.Sprintf(
			"--processed-by=fastly=%s (%s)",
			revision.AppVersion, cases.Title(textlang.English).String(language.Name),
		)
		metadataArgs := []string{
			"metadata", "add", "bin/main.wasm", metadataProcessedBy,
		}

		metadataDisable, _ := strconv.ParseBool(c.Globals.Env.WasmMetadataDisable)
		if !c.MetadataDisable && !metadataDisable {
			if err := c.AnnotateWasmBinaryLong(wasmtools, metadataArgs, language); err != nil {
				return err
			}
		} else {
			if err := c.AnnotateWasmBinaryShort(wasmtools, metadataArgs); err != nil {
				return err
			}
		}
		if c.MetadataShow {
			if err := c.ShowMetadata(wasmtools, out); err != nil {
				return err
			}
		}
	} else {
		if !c.Globals.Verbose() {
			text.Break(out)
		}
		text.Info(out, "There was an error downloading the wasm-tools (used for binary annotations) but we don't let that block you building your project. For reference here is the error (in case you want to let us know about it): %s\n\n", wasmtoolsErr.Error())
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

// AnnotateWasmBinaryShort annotates the Wasm binary with only the CLI version.
func (c *BuildCommand) AnnotateWasmBinaryShort(wasmtools string, args []string) error {
	return c.Globals.ExecuteWasmTools(wasmtools, args)
}

// AnnotateWasmBinaryLong annotates the Wasm binary will all available data.
func (c *BuildCommand) AnnotateWasmBinaryLong(wasmtools string, args []string, language *Language) error {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	// Allow customer to specify their own env variables to be filtered.
	ExtendStaticSecretEnvVars(c.MetadataFilterEnvVars)

	dc := DataCollection{}

	metadata := c.Globals.Config.WasmMetadata

	// Only record basic data if user has disabled all other metadata collection.
	if metadata.BuildInfo == "disable" && metadata.MachineInfo == "disable" && metadata.PackageInfo == "disable" && metadata.ScriptInfo == "disable" {
		return c.AnnotateWasmBinaryShort(wasmtools, args)
	}

	if metadata.BuildInfo == "enable" {
		dc.BuildInfo = DataCollectionBuildInfo{
			MemoryHeapAlloc: ms.HeapAlloc,
		}
	}
	if metadata.MachineInfo == "enable" {
		dc.MachineInfo = DataCollectionMachineInfo{
			Arch:      runtime.GOARCH,
			CPUs:      runtime.NumCPU(),
			Compiler:  runtime.Compiler,
			GoVersion: runtime.Version(),
			OS:        runtime.GOOS,
		}
	}
	if metadata.PackageInfo == "enable" {
		dc.PackageInfo = DataCollectionPackageInfo{
			ClonedFrom: c.Globals.Manifest.File.ClonedFrom,
			Packages:   language.Dependencies(),
		}
	}
	if metadata.ScriptInfo == "enable" {
		dc.ScriptInfo = DataCollectionScriptInfo{
			DefaultBuildUsed: language.DefaultBuildScript(),
			BuildScript:      FilterSecretsFromString(c.Globals.Manifest.File.Scripts.Build),
			EnvVars:          FilterSecretsFromSlice(c.Globals.Manifest.File.Scripts.EnvVars),
			PostInitScript:   FilterSecretsFromString(c.Globals.Manifest.File.Scripts.PostInit),
			PostBuildScript:  FilterSecretsFromString(c.Globals.Manifest.File.Scripts.PostBuild),
		}
	}

	data, err := json.Marshal(dc)
	if err != nil {
		return err
	}

	args = append(args, fmt.Sprintf("--processed-by=fastly_data=%s", data))

	return c.Globals.ExecuteWasmTools(wasmtools, args)
}

// ShowMetadata displays the metadata attached to the Wasm binary.
func (c *BuildCommand) ShowMetadata(wasmtools string, out io.Writer) error {
	// gosec flagged this:
	// G204 (CWE-78): Subprocess launched with variable
	// Disabling as the variables come from trusted sources.
	// #nosec
	// nosemgrep
	command := exec.Command(wasmtools, "metadata", "show", "bin/main.wasm")
	wasmtoolsOutput, err := command.Output()
	if err != nil {
		return fmt.Errorf("failed to execute wasm-tools metadata command: %w", err)
	}
	text.Info(out, "\nBelow is the metadata attached to the Wasm binary\n\n")
	fmt.Fprintln(out, string(wasmtoolsOutput))
	text.Break(out)
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

// PackageName acquires the package name from either a flag or manifest.
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

// ExecuteWasmTools calls the wasm-tools binary.
func ExecuteWasmTools(wasmtools string, args []string) error {
	// gosec flagged this:
	// G204 (CWE-78): Subprocess launched with function call as argument or command arguments
	// Disabling as we trust the source of the variable.
	// #nosec
	// nosemgrep: go.lang.security.audit.dangerous-exec-command.dangerous-exec-command
	command := exec.Command(wasmtools, args...)
	wasmtoolsOutput, err := command.Output()
	if err != nil {
		return fmt.Errorf("failed to annotate binary with metadata: %w", err)
	}
	// Ensure the Wasm binary can be executed.
	//
	// G302 (CWE-276): Expect file permissions to be 0600 or less
	// gosec flagged this:
	// Disabling as we want all users to be able to execute this binary.
	// #nosec
	err = os.WriteFile("bin/main.wasm", wasmtoolsOutput, 0o777)
	if err != nil {
		return fmt.Errorf("failed to annotate binary with metadata: %w", err)
	}
	return nil
}

// GetWasmTools returns the path to the wasm-tools binary.
// If there is no version installed, install the latest version.
// If there is a version installed, update to the latest version if not already.
func GetWasmTools(spinner text.Spinner, out io.Writer, wasmtoolsVersioner github.AssetVersioner, g *global.Data) (binPath string, err error) {
	binPath = wasmtoolsVersioner.InstallPath()

	// NOTE: When checking if wasm-tools is installed we don't use $PATH.
	//
	// $PATH is unreliable across OS platforms, but also we actually install
	// wasm-tools in the same location as the CLI's app config, which means it
	// wouldn't be found in the $PATH any way. We could pass the path for the app
	// config into exec.LookPath() but it's simpler to attempt executing the binary.
	//
	// gosec flagged this:
	// G204 (CWE-78): Subprocess launched with variable
	// Disabling as the variables come from trusted sources.
	// #nosec
	// nosemgrep
	c := exec.Command(binPath, "--version")

	var installedVersion string

	stdoutStderr, err := c.CombinedOutput()
	if err != nil {
		g.ErrLog.Add(err)
	} else {
		// Check the version output has the expected format: `wasm-tools 1.0.40`
		installedVersion = strings.TrimSpace(string(stdoutStderr))
		segs := strings.Split(installedVersion, " ")
		if len(segs) < 2 {
			return binPath, ErrWasmtoolsNotFound
		}
		installedVersion = segs[1]
	}

	if installedVersion == "" {
		if g.Verbose() {
			text.Info(out, "\nwasm-tools is not already installed, so we will install the latest version.\n\n")
		}
		err = installLatestWasmtools(binPath, spinner, wasmtoolsVersioner)
		if err != nil {
			g.ErrLog.Add(err)
			return binPath, err
		}

		latestVersion, err := wasmtoolsVersioner.LatestVersion()
		if err != nil {
			return binPath, fmt.Errorf("failed to retrieve wasm-tools latest version: %w", err)
		}

		g.Config.WasmTools.LatestVersion = latestVersion
		g.Config.WasmTools.LastChecked = time.Now().Format(time.RFC3339)

		err = g.Config.Write(g.ConfigPath)
		if err != nil {
			return binPath, err
		}
	}

	if installedVersion != "" {
		err = updateWasmtools(binPath, spinner, out, g, wasmtoolsVersioner, installedVersion)
		if err != nil {
			g.ErrLog.Add(err)
			return binPath, err
		}
	}

	err = github.SetBinPerms(binPath)
	if err != nil {
		g.ErrLog.Add(err)
		return binPath, err
	}

	return binPath, nil
}

func installLatestWasmtools(binPath string, spinner text.Spinner, wasmtoolsVersioner github.AssetVersioner) error {
	return spinner.Process("Fetching latest wasm-tools release", func(_ *text.SpinnerWrapper) error {
		tmpBin, err := wasmtoolsVersioner.DownloadLatest()
		if err != nil {
			return fmt.Errorf("failed to download latest wasm-tools release: %w", err)
		}
		defer os.RemoveAll(tmpBin)
		if err := os.Rename(tmpBin, binPath); err != nil {
			if err := filesystem.CopyFile(tmpBin, binPath); err != nil {
				return fmt.Errorf("failed to move wasm-tools binary to accessible location: %w", err)
			}
		}
		return nil
	})
}

func updateWasmtools(
	binPath string,
	spinner text.Spinner,
	out io.Writer,
	g *global.Data,
	wasmtoolsVersioner github.AssetVersioner,
	installedVersion string,
) error {
	cfg := g.Config
	cfgPath := g.ConfigPath

	// NOTE: We shouldn't see LastChecked with no value if wasm-tools installed.
	if cfg.WasmTools.LastChecked == "" {
		cfg.WasmTools.LastChecked = time.Now().Format(time.RFC3339)
		if err := cfg.Write(cfgPath); err != nil {
			return err
		}
	}
	if !check.Stale(cfg.WasmTools.LastChecked, cfg.WasmTools.TTL) {
		if g.Verbose() {
			text.Info(out, "\nwasm-tools is installed but the CLI config (`fastly config`) shows the TTL, checking for a newer version, hasn't expired.\n\n")
		}
		return nil
	}

	var latestVersion string
	err := spinner.Process("Checking latest wasm-tools release", func(_ *text.SpinnerWrapper) error {
		var err error
		latestVersion, err = wasmtoolsVersioner.LatestVersion()
		if err != nil {
			return fsterr.RemediationError{
				Inner:       fmt.Errorf("error fetching latest version: %w", err),
				Remediation: fsterr.NetworkRemediation,
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	cfg.WasmTools.LatestVersion = latestVersion
	cfg.WasmTools.LastChecked = time.Now().Format(time.RFC3339)

	err = cfg.Write(cfgPath)
	if err != nil {
		return err
	}
	if g.Verbose() {
		text.Info(out, "\nThe CLI config (`fastly config`) has been updated with the latest wasm-tools version: %s\n\n", latestVersion)
	}
	if installedVersion == latestVersion {
		return nil
	}

	return installLatestWasmtools(binPath, spinner, wasmtoolsVersioner)
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
//
// TODO: Fix the mess that is New<language>()'s argument list.
func language(toolchain, manifestFilename string, c *BuildCommand, in io.Reader, out io.Writer, spinner text.Spinner) (*Language, error) {
	var language *Language
	switch toolchain {
	case "assemblyscript":
		language = NewLanguage(&LanguageOptions{
			Name:            "assemblyscript",
			SourceDirectory: AsSourceDirectory,
			Toolchain:       NewAssemblyScript(c, in, manifestFilename, out, spinner),
		})
	case "go":
		language = NewLanguage(&LanguageOptions{
			Name:            "go",
			SourceDirectory: GoSourceDirectory,
			Toolchain:       NewGo(c, in, manifestFilename, out, spinner),
		})
	case "javascript":
		language = NewLanguage(&LanguageOptions{
			Name:            "javascript",
			SourceDirectory: JsSourceDirectory,
			Toolchain:       NewJavaScript(c, in, manifestFilename, out, spinner),
		})
	case "rust":
		language = NewLanguage(&LanguageOptions{
			Name:            "rust",
			SourceDirectory: RustSourceDirectory,
			Toolchain:       NewRust(c, in, manifestFilename, out, spinner),
		})
	case "other":
		language = NewLanguage(&LanguageOptions{
			Name:      "other",
			Toolchain: NewOther(c, in, manifestFilename, out, spinner),
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

// DataCollection represents data annotated onto the Wasm binary.
type DataCollection struct {
	BuildInfo   DataCollectionBuildInfo   `json:"build_info,omitempty"`
	MachineInfo DataCollectionMachineInfo `json:"machine_info,omitempty"`
	PackageInfo DataCollectionPackageInfo `json:"package_info,omitempty"`
	ScriptInfo  DataCollectionScriptInfo  `json:"script_info,omitempty"`
}

// DataCollectionBuildInfo represents build data annotated onto the Wasm binary.
type DataCollectionBuildInfo struct {
	MemoryHeapAlloc uint64 `json:"mem_heap_alloc,omitempty"`
}

// DataCollectionMachineInfo represents machine data annotated onto the Wasm binary.
type DataCollectionMachineInfo struct {
	Arch      string `json:"arch,omitempty"`
	CPUs      int    `json:"cpus,omitempty"`
	Compiler  string `json:"compiler,omitempty"`
	GoVersion string `json:"go_version,omitempty"`
	OS        string `json:"os,omitempty"`
}

// DataCollectionPackageInfo represents package data annotated onto the Wasm binary.
type DataCollectionPackageInfo struct {
	// ClonedFrom indicates if the Starter Kit used was cloned from a specific
	// repository (e.g. using the `compute init` --from flag).
	ClonedFrom string `json:"cloned_from,omitempty"`
	// Packages is a map where the key is the name of the package and the value is
	// the package version.
	Packages map[string]string `json:"packages,omitempty"`
}

// DataCollectionScriptInfo represents script data annotated onto the Wasm binary.
type DataCollectionScriptInfo struct {
	DefaultBuildUsed bool     `json:"default_build_used,omitempty"`
	BuildScript      string   `json:"build_script,omitempty"`
	EnvVars          []string `json:"env_vars,omitempty"`
	PostInitScript   string   `json:"post_init_script,omitempty"`
	PostBuildScript  string   `json:"post_build_script,omitempty"`
}
