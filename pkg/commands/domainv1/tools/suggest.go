package tools

import (
	"errors"
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v10/fastly"
	"github.com/fastly/go-fastly/v10/fastly/domains/v1/tools/suggest"
)

// GetDomainSuggestionsCommand calls the Fastly API to retrieve domain suggestions from the provided term(s).
type GetDomainSuggestionsCommand struct {
	argparser.Base
	argparser.JSONOutput
	// Required.
	query string
	// Optional.
	defaults argparser.OptionalString
	keywords argparser.OptionalString
	location argparser.OptionalString
	vendor   argparser.OptionalString
}

// NewDomainSuggestionsCommand returns a usable DomainSuggestionCommand registered under the parent.
func NewDomainSuggestionsCommand(parent argparser.Registerer, g *global.Data) *GetDomainSuggestionsCommand {
	cmd := GetDomainSuggestionsCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}

	cmd.CmdClause = parent.Command("suggest", "Performs real-time queries against the known zones database")
	// Required.
	cmd.CmdClause.Flag("query", "The term(s) to search against.").Required().StringVar(&cmd.query)
	// Optional.
	cmd.CmdClause.Flag("defaults", "Comma-separated list of default zones to include in the search results response").Action(cmd.defaults.Set).StringVar(&cmd.defaults.Value)
	cmd.RegisterFlagBool(cmd.JSONFlag())
	cmd.CmdClause.Flag("keywords", "Comma-separated list of keywords for seeding the results").Action(cmd.keywords.Set).StringVar(&cmd.keywords.Value)
	cmd.CmdClause.Flag("location", "Overrides the IP location detection for country-code zones, with a two-character country code").Action(cmd.location.Set).StringVar(&cmd.location.Value)
	cmd.CmdClause.Flag("vendor", "The domain name of a specific registrar or vendor ").Action(cmd.vendor.Set).StringVar(&cmd.vendor.Value)

	return &cmd
}

// Exec invokes the application logic for the command.
func (g *GetDomainSuggestionsCommand) Exec(_ io.Reader, out io.Writer) error {
	if g.Globals.Verbose() && g.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	fc, ok := g.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to acquire the Fastly API client")
	}

	input := &suggest.GetInput{
		Query: g.query,
	}

	if g.defaults.WasSet {
		input.Defaults = fastly.ToPointer(g.defaults.Value)
	}

	if g.keywords.WasSet {
		input.Keywords = fastly.ToPointer(g.keywords.Value)
	}

	if g.location.WasSet {
		input.Location = fastly.ToPointer(g.location.Value)
	}

	if g.vendor.WasSet {
		input.Vendor = fastly.ToPointer(g.vendor.Value)
	}

	suggestions, err := suggest.Get(fc, input)
	if err != nil {
		g.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := g.WriteJSON(out, suggestions); ok {
		return err
	}

	if g.Globals.Verbose() {
		printSuggestionsVerbose(out, suggestions)
	} else {
		printSuggestionsSummary(out, suggestions)
	}

	return nil
}

// printSuggestionsSummary displays the information returned from the API in a summarized
// format.
func printSuggestionsSummary(out io.Writer, suggestions *suggest.Suggestions) {
	t := text.NewTable(out)
	t.AddHeader("Domain", "Subdomain", "Zone", "Path")
	for _, suggestion := range suggestions.Results {
		var path string
		if suggestion.Path != nil {
			path = *suggestion.Path
		}
		t.AddLine(suggestion.Domain, suggestion.Subdomain, suggestion.Zone, path)
	}

	t.Print()
}

// printSuggestionsVerbose displays the information returned from the API in a verbose
// format.
func printSuggestionsVerbose(out io.Writer, suggestions *suggest.Suggestions) {
	for _, suggestion := range suggestions.Results {
		fmt.Fprintf(out, "Domain: %s\n", suggestion.Domain)
		fmt.Fprintf(out, "Subdomain: %s\n", suggestion.Subdomain)
		fmt.Fprintf(out, "Zone: %s\n", suggestion.Zone)

		if suggestion.Path != nil {
			fmt.Fprintf(out, "Path: %s\n", *suggestion.Path)
		}
		fmt.Fprintf(out, "\n")
	}
}
