package vcl

import (
	"context"
	"io"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// CheckCommand compiles VCL with the real Fastly VCL compiler, without
// executing anything, and reports diagnostics.
type CheckCommand struct {
	argparser.Base
	argparser.JSONOutput

	spec  specFlags
	files []string
}

// NewCheckCommand returns a usable command registered under the parent.
func NewCheckCommand(parent argparser.Registerer, g *global.Data) *CheckCommand {
	c := CheckCommand{
		Base: argparser.Base{
			Globals:         g,
			SuppressVerbose: true,
		},
	}
	c.CmdClause = parent.Command("check", "Check that VCL compiles, using the real Fastly VCL compiler (no service or API token needed)")
	c.CmdClause.Arg("file", "VCL files holding subroutine bodies, named after their subroutine (recv.vcl, fetch.vcl, ...)").StringsVar(&c.files)
	c.spec.register(c.CmdClause)
	c.RegisterFlagBool(c.JSONFlag()) // --json
	return &c
}

// Exec invokes the application logic for the command.
func (c *CheckCommand) Exec(_ io.Reader, out io.Writer) error {
	vcl, sources, err := c.spec.buildVCL(c.files)
	if err != nil {
		return err
	}
	origins, err := c.spec.buildOrigins()
	if err != nil {
		return err
	}

	state := loadSandboxState()
	saved, err := publish(context.Background(), c.spec.client(c.Globals), newSpec(origins, vcl), &state)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	diags := flattenDiagnostics(saved.LintStatus, sources)
	if ok, err := c.WriteJSON(out, struct {
		Valid       bool         `json:"valid"`
		Diagnostics []diagnostic `json:"diagnostics"`
	}{saved.Valid, diags}); ok {
		if err == nil && !saved.Valid {
			return errCompileFailed
		}
		return err
	}

	printDiagnostics(out, diags)
	if !saved.Valid {
		return errCompileFailed
	}
	text.Success(out, "VCL compiled cleanly.")
	return nil
}
