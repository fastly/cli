package domain

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"
	"github.com/fastly/go-fastly/v19/fastly/domainmanagement/v1/tools/status"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
)

// GetDomainStatusCommand calls the Fastly API to check the availability of a domain name.
type GetDomainStatusCommand struct {
	argparser.Base
	// Optional.
	argparser.JSONOutput
	// Required.
	Domain string `arg:"" required:"" help:"Domain name to check"`
	// Optional.
	Scope argparser.OptionalString "help:\"Specify `--scope=estimate` to perform an “estimated” availability check, which checks the DNS and domain aftermarkets, not domain registries\""
}

// Exec invokes the application logic for the command.
func (g *GetDomainStatusCommand) Exec(_ io.Reader, out io.Writer) error {
	if g.Globals.Verbose() && g.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input := &status.GetInput{
		Domain: g.Domain,
	}

	if g.Scope.WasSet {
		scope := status.Scope(g.Scope.Value)
		if scope != status.ScopeEstimate {
			return fsterr.RemediationError{
				Inner:       errors.New("invalid scope provided"),
				Remediation: "Use `--scope=estimate` for an estimated status check",
			}
		}
		input.Scope = new(scope)
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
