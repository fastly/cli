package compute

import (
	"crypto/sha512"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/kennygrant/sanitize"

	"github.com/fastly/cli/v10/pkg/argparser"
	fsterr "github.com/fastly/cli/v10/pkg/errors"
	"github.com/fastly/cli/v10/pkg/global"
	"github.com/fastly/cli/v10/pkg/manifest"
	"github.com/fastly/cli/v10/pkg/text"
)

// HashsumCommand produces a deployable artifact from files on the local disk.
type HashsumCommand struct {
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

	buildCmd    *BuildCommand
	PackagePath string
	SkipBuild   bool
}

// NewHashsumCommand returns a usable command registered under the parent.
// Deprecated: Use NewHashFilesCommand instead.
func NewHashsumCommand(parent argparser.Registerer, g *global.Data, build *BuildCommand) *HashsumCommand {
	var c HashsumCommand
	c.buildCmd = build
	c.Globals = g
	c.CmdClause = parent.Command("hashsum", "Generate a SHA512 digest from a Compute package").Hidden()
	c.CmdClause.Flag("dir", "Project directory to build (default: current directory)").Short('C').Action(c.dir.Set).StringVar(&c.dir.Value)
	c.CmdClause.Flag("env", "The manifest environment config to use (e.g. 'stage' will attempt to read 'fastly.stage.toml')").Action(c.env.Set).StringVar(&c.env.Value)
	c.CmdClause.Flag("include-source", "Include source code in built package").Action(c.includeSrc.Set).BoolVar(&c.includeSrc.Value)
	c.CmdClause.Flag("language", "Language type").Action(c.lang.Set).StringVar(&c.lang.Value)
	c.CmdClause.Flag("metadata-disable", "Disable Wasm binary metadata annotations").Action(c.metadataDisable.Set).BoolVar(&c.metadataDisable.Value)
	c.CmdClause.Flag("metadata-filter-envvars", "Redact specified environment variables from [scripts.env_vars] using comma-separated list").Action(c.metadataFilterEnvVars.Set).StringVar(&c.metadataFilterEnvVars.Value)
	c.CmdClause.Flag("metadata-show", "Inspect the Wasm binary metadata").Action(c.metadataShow.Set).BoolVar(&c.metadataShow.Value)
	c.CmdClause.Flag("package", "Path to a package tar.gz").Short('p').StringVar(&c.PackagePath)
	c.CmdClause.Flag("package-name", "Package name").Action(c.packageName.Set).StringVar(&c.packageName.Value)
	c.CmdClause.Flag("skip-build", "Skip the build step").BoolVar(&c.SkipBuild)
	c.CmdClause.Flag("timeout", "Timeout, in seconds, for the build compilation step").Action(c.timeout.Set).IntVar(&c.timeout.Value)

	return &c
}

// Exec implements the command interface.
func (c *HashsumCommand) Exec(in io.Reader, out io.Writer) (err error) {
	if !c.Globals.Flags.Quiet {
		// FIXME: Remove `hashsum` subcommand before v11.0.0 is released.
		text.Warning(out, "This command is deprecated. Use `fastly compute hash-files` instead.")
	}

	// No point in building a package if the user provides a package path.
	if !c.SkipBuild && c.PackagePath == "" {
		err = c.Build(in, out)
		if err != nil {
			return err
		}
		if !c.Globals.Flags.Quiet {
			text.Break(out)
		}
	}

	pkgPath := c.PackagePath
	if pkgPath == "" {
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
	}

	err = validatePackage(pkgPath)
	if err != nil {
		var skipBuildMsg string
		if c.SkipBuild {
			skipBuildMsg = " avoid using --skip-build, or"
		}
		return fsterr.RemediationError{
			Inner:       fmt.Errorf("failed to validate package: %w", err),
			Remediation: fmt.Sprintf("Run `fastly compute build` to produce a Compute package, alternatively%s use the --package flag to reference a package outside of the current project.", skipBuildMsg),
		}
	}

	hashSum, err := getHashSum(pkgPath)
	if err != nil {
		return err
	}

	text.Output(out, hashSum)
	return nil
}

// Build constructs and executes the build logic.
func (c *HashsumCommand) Build(in io.Reader, out io.Writer) error {
	output := out
	if !c.Globals.Verbose() && !c.metadataShow.WasSet {
		output = io.Discard
	} else {
		text.Break(out)
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

// getHashSum returns a hash of the package.
func getHashSum(pkg string) (string, error) {
	// gosec flagged this:
	// G304 (CWE-22): Potential file inclusion via variable
	// Disabling as we trust the source of the filepath variable.
	/* #nosec */
	f, err := os.Open(pkg)
	if err != nil {
		return "", err
	}

	h := sha512.New()
	if _, err := io.Copy(h, f); err != nil {
		_ = f.Close()
		return "", err
	}

	if err = f.Close(); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
