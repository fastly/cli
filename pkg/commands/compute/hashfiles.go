package compute

import (
	"bytes"
	"crypto/sha512"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/kennygrant/sanitize"
	"github.com/mholt/archives"

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
	Dir                   argparser.OptionalString `short:"C" help:"Project directory to build (default: current directory)"`
	Env                   argparser.OptionalString `help:"The manifest environment config to use (e.g. 'stage' will attempt to read 'fastly.stage.toml')"`
	IncludeSrc            argparser.OptionalBool   `name:"include-source" help:"Include source code in built package"`
	Lang                  argparser.OptionalString `name:"language" help:"Language type"`
	MetadataDisable       argparser.OptionalBool   `help:"Disable Wasm binary metadata annotations"`
	MetadataFilterEnvVars argparser.OptionalString `name:"metadata-filter-envvars" help:"Redact specified environment variables from [scripts.env_vars] using comma-separated list"`
	MetadataShow          argparser.OptionalBool   `help:"Inspect the Wasm binary metadata"`
	Package               string                   `short:"p" help:"Path to a package tar.gz"`
	PackageName           argparser.OptionalString `help:"Package name"`
	SkipBuild             bool                     `help:"Skip the build step"`
	Timeout               argparser.OptionalInt    `help:"Timeout, in seconds, for the build compilation step"`

	buildCmd *BuildCommand
}

// NewHashFilesCommand returns a HashFilesCommand set up as it was before parsing.
// Its flags now come from the struct's tags.
func NewHashFilesCommand(g *global.Data, build *BuildCommand) *HashFilesCommand {
	var c HashFilesCommand
	c.buildCmd = build
	c.Globals = g

	return &c
}

// BeforeApply sets the command up with the values bound when the parser
// was built, before kong applies the command line.
func (c *HashFilesCommand) BeforeApply(g *global.Data) error {
	c.buildCmd = &BuildCommand{}
	c.buildCmd.Globals = g
	return nil
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
		manifestFilename := EnvironmentManifest(c.Env.Value)
		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current working directory: %w", err)
		}
		defer func() {
			_ = os.Chdir(wd)
		}()
		manifestPath := filepath.Join(wd, manifestFilename)

		projectDir, err := ChangeProjectDirectory(c.Dir.Value)
		if err != nil {
			return err
		}
		if projectDir != "" {
			if c.Globals.Verbose() {
				text.Info(out, ProjectDirMsg, projectDir)
			}
			manifestPath = filepath.Join(projectDir, manifestFilename)
		}

		if projectDir != "" || c.Env.WasSet {
			err = c.Globals.Manifest.File.Read(manifestPath)
		} else {
			err = c.Globals.Manifest.File.ReadError()
		}
		if err != nil {
			err = packageFlagManifestError(err)
			c.Globals.ErrLog.Add(err)
			return err
		}

		projectName, source := c.Globals.Manifest.Name()
		if source == manifest.SourceUndefined {
			return fsterr.ErrMissingManifestName
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
	if c.Dir.WasSet {
		c.buildCmd.Flags.Dir = c.Dir.Value
	}
	if c.Env.WasSet {
		c.buildCmd.Flags.Env = c.Env.Value
	}
	if c.IncludeSrc.WasSet {
		c.buildCmd.Flags.IncludeSrc = c.IncludeSrc.Value
	}
	if c.Lang.WasSet {
		c.buildCmd.Flags.Lang = c.Lang.Value
	}
	if c.PackageName.WasSet {
		c.buildCmd.Flags.PackageName = c.PackageName.Value
	}
	if c.Timeout.WasSet {
		c.buildCmd.Flags.Timeout = c.Timeout.Value
	}
	if c.MetadataDisable.WasSet {
		c.buildCmd.MetadataDisable = c.MetadataDisable.Value
	}
	if c.MetadataFilterEnvVars.WasSet {
		c.buildCmd.MetadataFilterEnvVars = c.MetadataFilterEnvVars.Value
	}
	if c.MetadataShow.WasSet {
		c.buildCmd.MetadataShow = c.MetadataShow.Value
	}
	return c.buildCmd.Exec(in, output)
}

// getFilesHash returns a hash of all the files in the package in sorted filename order.
func getFilesHash(pkgPath string) (string, error) {
	contents := make(map[string]*bytes.Buffer)

	if err := packageFiles(pkgPath, func(f archives.FileInfo) error {
		entry := f.NameInArchive
		contents[entry] = &bytes.Buffer{}
		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("error opening %s: %w", entry, err)
		}
		defer rc.Close()
		if _, err := io.Copy(contents[entry], rc); err != nil {
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
