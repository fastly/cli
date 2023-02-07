package compute

import (
	"fmt"
	"io"

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
func NewHashsumCommand(parent cmd.Registerer, g *global.Data, build *BuildCommand, m manifest.Data) *HashsumCommand {
	var c HashsumCommand
	c.buildCmd = build
	c.Globals = g
	c.Manifest = m
	c.CmdClause = parent.Command("hashsum", "Generate a SHA512 digest from a Compute@Edge package")
	c.CmdClause.Flag("package", "Path to a package tar.gz").Short('p').StringVar(&c.Package)
	c.CmdClause.Flag("skip-build", "Skip the build step").BoolVar(&c.SkipBuild)
	return &c
}

// Exec implements the command interface.
func (c *HashsumCommand) Exec(in io.Reader, out io.Writer) (err error) {
	if !c.SkipBuild {
		err = c.Build(in, out)
		if err != nil {
			return err
		}
	}

	_, hashSum, err := validatePackage(c.Manifest, c.Package, c.Globals.Verbose(), c.Globals.ErrLog, out)
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

	text.Output(out, hashSum)
	return nil
}

// Build constructs and executes the build logic.
func (c *HashsumCommand) Build(in io.Reader, out io.Writer) error {
	output := out
	if !c.Globals.Verbose() {
		output = io.Discard
	}

	err := c.buildCmd.Exec(in, output)
	if err != nil {
		return err
	}
	return nil
}
