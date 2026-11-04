package domain

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/fastly/go-fastly/v19/fastly"
	"github.com/fastly/go-fastly/v19/fastly/domainmanagement/v1/tools/suggest"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// GetDomainSuggestionsCommand calls the Fastly API and results domain search results for a given query.
type GetDomainSuggestionsCommand struct {
	argparser.Base
	// Optional.
	Defaults argparser.OptionalString "help:\"Comma-separated list of default zones to include in the search results response, e.g. `--defaults=uk,co.uk`\""
	argparser.JSONOutput
	// Required.
	Query    []string                 `arg:"" required:"" sep:"none" help:"Search query, e.g. “acme coffee shop”"`
	Keywords argparser.OptionalString "help:\"Comma-separated list of keywords for seeding the search results, e.g. `--keywords=dance,party`\""
	Location argparser.OptionalString "help:\"Override IP geolocation with a two-character country code, e.g. `--location=in` to include Indian zones in the search results\""
	Vendor   argparser.OptionalString `help:"The domain name of a specific registrar or vendor, to filter the search results to the list of zones they support"`
}

// Exec invokes the application logic for the command.
func (g *GetDomainSuggestionsCommand) Exec(_ io.Reader, out io.Writer) error {
	if g.Globals.Verbose() && g.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input := &suggest.GetInput{
		Query: strings.Join(g.Query, " "),
	}

	if g.Defaults.WasSet {
		input.Defaults = new(g.Defaults.Value)
	}

	if g.Keywords.WasSet {
		input.Keywords = new(g.Keywords.Value)
	}

	if g.Location.WasSet {
		input.Location = new(g.Location.Value)
	}

	if g.Vendor.WasSet {
		input.Vendor = new(g.Vendor.Value)
	}

	fc, ok := g.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to acquire the Fastly API client")
	}

	suggestions, err := suggest.Get(context.TODO(), fc, input)
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
