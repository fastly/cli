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

	Manifest manifest.Data
	Package  string
}

// NewHashsumCommand returns a usable command registered under the parent.
func NewHashsumCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *HashsumCommand {
	var c HashsumCommand
	c.Globals = globals
	c.Manifest = data
	c.CmdClause = parent.Command("hashsum", "Generate a SHA512 digest from a Compute@Edge package")
	c.CmdClause.Flag("package", "Path to a package tar.gz").Short('p').StringVar(&c.Package)
	return &c
}

// Exec implements the command interface.
func (c *HashsumCommand) Exec(in io.Reader, out io.Writer) (err error) {
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
