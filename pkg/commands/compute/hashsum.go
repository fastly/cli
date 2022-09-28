package compute

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
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
func NewHashsumCommand(parent cmd.Registerer, globals *config.Data, build *BuildCommand, data manifest.Data) *HashsumCommand {
	var c HashsumCommand
	c.buildCmd = build
	c.Globals = globals
	c.Manifest = data
	c.CmdClause = parent.Command("hashsum", "Generate a SHA512 digest from a Compute@Edge package")
	c.CmdClause.Flag("package", "Path to a package tar.gz").Short('p').StringVar(&c.Package)
	c.CmdClause.Flag("skip-build", "Skip the build step (presumes a successful prior build)").BoolVar(&c.SkipBuild)
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

	_, _, hashSum, err := validatePackage(c.Manifest, c.Package, c.Globals.Verbose(), c.Globals.ErrLog, out)
	if err != nil {
		return fsterr.RemediationError{
			Inner:       fmt.Errorf("failed to validate package: %w", err),
			Remediation: "Run `fastly compute build` to produce a Compute@Edge package, alternatively use the --package flag to reference a package outside of the current project.",
		}
	}
	text.Output(out, hashSum)
	return nil
}

// Build constructs and executes the build logic.
func (c *HashsumCommand) Build(in io.Reader, out io.Writer) error {
	err := c.buildCmd.Exec(in, io.Discard)
	if err != nil {
		return err
	}
	return nil
}
