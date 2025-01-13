package compute

import (
	"archive/tar"
	"bytes"
	"crypto/sha512"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/kennygrant/sanitize"
	"github.com/mholt/archiver/v3"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
)

// MaxPackageSize represents the max package size that can be uploaded to the
// Fastly Package API endpoint.
//
// NOTE: This is variable not a constant for the sake of test manipulations.
// https://www.fastly.com/documentation/guides/compute#limitations-and-constraints
var MaxPackageSize int64 = 100000000 // 100MB in bytes

// HashFilesCommand produces a deployable artifact from files on the local disk.
type HashFilesCommand struct {
	argparser.Base

	// Build fields
	dir                   argparser.OptionalString
	env                   argparser.OptionalString
	includeSrc            argparser.OptionalBool
	lang                  argparser.OptionalString
	metadataDisable       argparser.OptionalBool
	metadataFilterEnvVars argparser.OptionalString
	metadataShow          argparser.OptionalBool
	packageName           argparser.OptionalString
	timeout               argparser.OptionalInt

	buildCmd  *BuildCommand
	Package   string
	SkipBuild bool
}

// NewHashFilesCommand returns a usable command registered under the parent.
func NewHashFilesCommand(parent argparser.Registerer, g *global.Data, build *BuildCommand) *HashFilesCommand {
	var c HashFilesCommand
	c.buildCmd = build
	c.Globals = g
	c.CmdClause = parent.Command("hash-files", "Generate a SHA512 digest from the contents of the Compute package")
	c.CmdClause.Flag("dir", "Project directory to build (default: current directory)").Short('C').Action(c.dir.Set).StringVar(&c.dir.Value)
	c.CmdClause.Flag("env", "The manifest environment config to use (e.g. 'stage' will attempt to read 'fastly.stage.toml')").Action(c.env.Set).StringVar(&c.env.Value)
	c.CmdClause.Flag("include-source", "Include source code in built package").Action(c.includeSrc.Set).BoolVar(&c.includeSrc.Value)
	c.CmdClause.Flag("language", "Language type").Action(c.lang.Set).StringVar(&c.lang.Value)
	c.CmdClause.Flag("metadata-disable", "Disable Wasm binary metadata annotations").Action(c.metadataDisable.Set).BoolVar(&c.metadataDisable.Value)
	c.CmdClause.Flag("metadata-filter-envvars", "Redact specified environment variables from [scripts.env_vars] using comma-separated list").Action(c.metadataFilterEnvVars.Set).StringVar(&c.metadataFilterEnvVars.Value)
	c.CmdClause.Flag("metadata-show", "Inspect the Wasm binary metadata").Action(c.metadataShow.Set).BoolVar(&c.metadataShow.Value)
	c.CmdClause.Flag("package", "Path to a package tar.gz").Short('p').StringVar(&c.Package)
	c.CmdClause.Flag("package-name", "Package name").Action(c.packageName.Set).StringVar(&c.packageName.Value)
	c.CmdClause.Flag("skip-build", "Skip the build step").BoolVar(&c.SkipBuild)
	c.CmdClause.Flag("timeout", "Timeout, in seconds, for the build compilation step").Action(c.timeout.Set).IntVar(&c.timeout.Value)

	return &c
}

// Exec implements the command interface.
func (c *HashFilesCommand) Exec(in io.Reader, out io.Writer) (err error) {
	if !c.SkipBuild && c.Package == "" {
		err = c.Build(in, out)
		if err != nil {
			return err
		}
		if c.Globals.Verbose() {
			text.Break(out)
		}
	}

	var pkgPath string

	if c.Package == "" {
		manifestFilename := EnvironmentManifest(c.env.Value)
		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current working directory: %w", err)
		}
		defer func() {
			_ = os.Chdir(wd)
		}()
		manifestPath := filepath.Join(wd, manifestFilename)

		projectDir, err := ChangeProjectDirectory(c.dir.Value)
		if err != nil {
			return err
		}
		if projectDir != "" {
			if c.Globals.Verbose() {
				text.Info(out, ProjectDirMsg, projectDir)
			}
			manifestPath = filepath.Join(projectDir, manifestFilename)
		}

		if projectDir != "" || c.env.WasSet {
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

		projectName, source := c.Globals.Manifest.Name()
		if source == manifest.SourceUndefined {
			return fsterr.ErrReadingManifest
		}
		pkgPath = filepath.Join(projectDir, "pkg", fmt.Sprintf("%s.tar.gz", sanitize.BaseName(projectName)))
	} else {
		pkgPath, err = filepath.Abs(c.Package)
		if err != nil {
			return fmt.Errorf("failed to locate package path '%s': %w", c.Package, err)
		}
	}

	hash, err := getFilesHash(pkgPath)
	if err != nil {
		return err
	}

	text.Output(out, hash)
	return nil
}

// Build constructs and executes the build logic.
func (c *HashFilesCommand) Build(in io.Reader, out io.Writer) error {
	output := out
	if !c.Globals.Verbose() {
		output = io.Discard
	}
	if c.dir.WasSet {
		c.buildCmd.Flags.Dir = c.dir.Value
	}
	if c.env.WasSet {
		c.buildCmd.Flags.Env = c.env.Value
	}
	if c.includeSrc.WasSet {
		c.buildCmd.Flags.IncludeSrc = c.includeSrc.Value
	}
	if c.lang.WasSet {
		c.buildCmd.Flags.Lang = c.lang.Value
	}
	if c.packageName.WasSet {
		c.buildCmd.Flags.PackageName = c.packageName.Value
	}
	if c.timeout.WasSet {
		c.buildCmd.Flags.Timeout = c.timeout.Value
	}
	if c.metadataDisable.WasSet {
		c.buildCmd.MetadataDisable = c.metadataDisable.Value
	}
	if c.metadataFilterEnvVars.WasSet {
		c.buildCmd.MetadataFilterEnvVars = c.metadataFilterEnvVars.Value
	}
	if c.metadataShow.WasSet {
		c.buildCmd.MetadataShow = c.metadataShow.Value
	}
	return c.buildCmd.Exec(in, output)
}

// getFilesHash returns a hash of all the files in the package in sorted filename order.
func getFilesHash(pkgPath string) (string, error) {
	contents := make(map[string]*bytes.Buffer)

	if err := packageFiles(pkgPath, func(f archiver.File) error {
		// We want the full path here and not f.Name(), which is only the
		// filename.
		//
		// This is safe to do - we already verified it in packageFiles().
		header, ok := f.Header.(*tar.Header)
		if !ok {
			return errors.New("failed to convert file type into *tar.Header")
		}
		entry := header.Name
		contents[entry] = &bytes.Buffer{}
		if _, err := io.Copy(contents[entry], f); err != nil {
			return fmt.Errorf("error reading %s: %w", entry, err)
		}
		return nil
	}); err != nil {
		return "", err
	}

	keys := make([]string, 0, len(contents))
	for k := range contents {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	h := sha512.New()
	for _, entry := range keys {
		if _, err := io.Copy(h, contents[entry]); err != nil {
			return "", fmt.Errorf("failed to generate hash from package files: %w", err)
		}
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
