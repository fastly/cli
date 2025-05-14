package tools

import (
	"errors"
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/go-fastly/v10/fastly"
	"github.com/fastly/go-fastly/v10/fastly/domains/v1/tools/status"
)

// GetDomainStatusCommand calls the Fastly API to check the availability status of a domain.
type GetDomainStatusCommand struct {
	argparser.Base
	argparser.JSONOutput
	// Required.
	domain string
	// Optional.
	scope argparser.OptionalString
}

// NewDomainStatusCommand returns a usable DomainStatusCommand registered under the parent.
func NewDomainStatusCommand(parent argparser.Registerer, g *global.Data) *GetDomainStatusCommand {
	cmd := GetDomainStatusCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}

	cmd.CmdClause = parent.Command("status", "Check the availability status of a single domain name.")

	// Required.
	cmd.CmdClause.Flag("domain", "Domain being checked for availability").Required().StringVar(&cmd.domain)
	// Optional.
	cmd.CmdClause.Flag("scope", "Scope determines the availability check to perform, specify `estimate` for an estimated check").Action(cmd.scope.Set).StringVar(&cmd.scope.Value)
	cmd.RegisterFlagBool(cmd.JSONFlag())

	return &cmd
}

// Exec invokes the application logic for the command.
func (g *GetDomainStatusCommand) Exec(_ io.Reader, out io.Writer) error {
	if g.Globals.Verbose() && g.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	fc, ok := g.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	input := &status.GetInput{
		Domain: g.domain,
	}

	if g.scope.WasSet {
		input.Scope = fastly.ToPointer(status.Scope(g.scope.Value))
	}

	st, err := status.Get(fc, input)
	if err != nil {
		g.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := g.WriteJSON(out, st); ok {
		return err
	}

	writeStatusSummary(out, st)
	return nil
}

func writeStatusSummary(w io.Writer, st *status.Status) {
	fmt.Fprintf(w, "Domain: %s\n", st.Domain)
	fmt.Fprintf(w, "Zone: %s\n", st.Zone)
	fmt.Fprintf(w, "Status: %s\n", st.Status)
	fmt.Fprintf(w, "Tags: %s\n", st.Tags)

	if st.Scope != nil {
		fmt.Fprintf(w, "Scope: %s\n", *st.Scope)
	}

	if len(st.Offers) > 0 {
		fmt.Fprintf(w, "Offers:\n")
		for _, o := range st.Offers {
			fmt.Fprintf(w, "  - Vendor: %s\n", o.Vendor)
			fmt.Fprintf(w, "    Currency: %s\n", o.Currency)
			fmt.Fprintf(w, "    Price: %s\n", o.Price)
		}
	}
}
