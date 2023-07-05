package compute

import (
	"crypto/sha512"
	"fmt"
	"io"
	"os"

	"github.com/fastly/cli/pkg/cmd"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
)

// HashsumCommand produces a deployable artifact from files on the local disk.
type HashsumCommand struct {
	cmd.Base

	buildCmd  *BuildCommand
	Manifest  manifest.Data
	Package   string
	SkipBuild bool
}

// NewHashsumCommand returns a usable command registered under the parent.
// Deprecated: Use NewHashFilesCommand instead.
func NewHashsumCommand(parent cmd.Registerer, g *global.Data, build *BuildCommand, m manifest.Data) *HashsumCommand {
	var c HashsumCommand
	c.buildCmd = build
	c.Globals = g
	c.Manifest = m
	c.CmdClause = parent.Command("hashsum", "Generate a SHA512 digest from a Compute@Edge package").Hidden()
	c.CmdClause.Flag("package", "Path to a package tar.gz").Short('p').StringVar(&c.Package)
	c.CmdClause.Flag("skip-build", "Skip the build step").BoolVar(&c.SkipBuild)
	return &c
}

// Exec implements the command interface.
func (c *HashsumCommand) Exec(in io.Reader, out io.Writer) (err error) {
	if !c.Globals.Flags.Quiet {
		text.Warning(out, "This command is deprecated. Use `fastly compute hash-files` instead.")
		text.Break(out)
	}

	if !c.SkipBuild {
		err = c.Build(in, out)
		if err != nil {
			return err
		}
	}

	pkgPath, err := validatePackage(c.Manifest, c.Package, c.Globals.Verbose(), c.Globals.ErrLog, out)
	if err != nil {
		var skipBuildMsg string
		if c.SkipBuild {
			skipBuildMsg = " avoid using --skip-build, or"
		}
		return fsterr.RemediationError{
			Inner:       fmt.Errorf("failed to validate package: %w", err),
			Remediation: fmt.Sprintf("Run `fastly compute build` to produce a Compute@Edge package, alternatively%s use the --package flag to reference a package outside of the current project.", skipBuildMsg),
		}
	}

	if c.Globals.Verbose() {
		text.Break(out)
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
	if !c.Globals.Verbose() {
		output = io.Discard
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
		f.Close()
		return "", err
	}

	if err = f.Close(); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
