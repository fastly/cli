package domain

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v11/fastly"
	"github.com/fastly/go-fastly/v11/fastly/domainmanagement/v1/tools/status"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
)

// GetDomainStatusCommand calls the Fastly API to check the availability of a domain name.
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

	cmd.CmdClause = parent.Command("status", "Check domain name availability")
	// Required.
	cmd.CmdClause.Arg("domain", "Domain name to check").Required().StringVar(&cmd.domain)
	// Optional.
	cmd.RegisterFlagBool(cmd.JSONFlag())
	cmd.CmdClause.Flag("scope", "Specify `--scope=estimate` to perform an “estimated” availability check, which checks the DNS and domain aftermarkets, not domain registries").Action(cmd.scope.Set).StringVar(&cmd.scope.Value)

	return &cmd
}

// Exec invokes the application logic for the command.
func (g *GetDomainStatusCommand) Exec(_ io.Reader, out io.Writer) error {
	if g.Globals.Verbose() && g.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input := &status.GetInput{
		Domain: g.domain,
	}

	if g.scope.WasSet {
		scope := status.Scope(g.scope.Value)
		if scope != status.ScopeEstimate {
			return fsterr.RemediationError{
				Inner:       errors.New("invalid scope provided"),
				Remediation: "Use `--scope=estimate` for an estimated status check",
			}
		}
		input.Scope = fastly.ToPointer(scope)
	}

	fc, ok := g.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to acquire the Fastly API client")
	}

	st, err := status.Get(context.TODO(), fc, input)
	if err != nil {
		g.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := g.WriteJSON(out, st); ok {
		return err
	}

	printStatusSummary(out, st)
	return nil
}

// printStatusSummary displays the information returned from the API in a summarized format.
func printStatusSummary(w io.Writer, st *status.Status) {
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
