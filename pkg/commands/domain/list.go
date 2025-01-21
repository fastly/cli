package domain

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/fastly/go-fastly/v9/fastly"
	v1 "github.com/fastly/go-fastly/v9/fastly/domains/v1"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
)

// ListCommand calls the Fastly API to list domains.
type ListCommand struct {
	argparser.Base
	argparser.JSONOutput

	Input          fastly.ListDomainsInput
	apiVersion     argparser.OptionalString
	cursor         argparser.OptionalString
	fqdn           argparser.OptionalString
	limit          argparser.OptionalInt
	serviceName    argparser.OptionalServiceNameID
	serviceVersion argparser.OptionalServiceVersion
	sort           argparser.OptionalString
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent argparser.Registerer, g *global.Data) *ListCommand {
	c := ListCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("list", "List domains")

	// Optional.
	c.CmdClause.Flag("api-version", fmt.Sprintf("The Fastly API version (%s)", strings.Join(APIVersions, ","))).HintOptions(APIVersions...).Action(c.apiVersion.Set).EnumVar(&c.apiVersion.Value, APIVersions...)
	c.CmdClause.Flag("cursor", "Cursor value from the next_cursor field of a previous response, used to retrieve the next page (version support: v1)").Action(c.cursor.Set).StringVar(&c.cursor.Value)
	c.CmdClause.Flag("fqdn", "Filters results by the FQDN using a fuzzy/partial match (version support: v1)").Action(c.fqdn.Set).StringVar(&c.fqdn.Value)
	c.RegisterFlagBool(c.JSONFlag()) // --json
	c.CmdClause.Flag("limit", "Limit how many results are returned (version support: v1)").Action(c.limit.Set).IntVar(&c.limit.Value)
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagServiceIDName,
		Description: argparser.FlagServiceIDDesc,
		Dst:         &g.Manifest.Flag.ServiceID,
		Short:       's',
	})
	c.RegisterFlag(argparser.StringFlagOpts{
		Action:      c.serviceName.Set,
		Name:        argparser.FlagServiceName,
		Description: argparser.FlagServiceNameDesc,
		Dst:         &c.serviceName.Value,
	})
	c.CmdClause.Flag("sort", "The order in which to list the results (version support: v1)").Action(c.sort.Set).StringVar(&c.sort.Value)
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagVersionName,
		Description: argparser.FlagVersionDesc,
		Dst:         &c.serviceVersion.Value,
	})
	return &c
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(in io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}
	if c.apiVersion.Value == "v1" {
		return c.v1(in, out)
	}
	return c.v0(out)
}

func (c *ListCommand) v1(in io.Reader, out io.Writer) error {
	input := &v1.ListInput{}

	serviceID, source, _, _ := argparser.ServiceID(c.serviceName, *c.Globals.Manifest, c.Globals.APIClient, c.Globals.ErrLog)
	if source == manifest.SourceFlag {
		input.ServiceID = &serviceID
	}
	if c.cursor.WasSet {
		input.Cursor = &c.cursor.Value
	}
	if c.fqdn.WasSet {
		input.FQDN = &c.fqdn.Value
	}
	if c.limit.WasSet {
		input.Limit = &c.limit.Value
	}
	if c.sort.WasSet {
		input.Sort = &c.sort.Value
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	for {
		cl, err := v1.List(fc, input)
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
				"Cursor":     c.cursor.Value,
				"FQDN":       c.fqdn.Value,
				"Limit":      c.limit.Value,
				"Service ID": serviceID,
				"Sort":       c.sort.Value,
			})
			return err
		}

		if ok, err := c.WriteJSON(out, cl); ok {
			// No pagination prompt w/ JSON output.
			return err
		}

		// clPtr := make([]*v1.Data, len(cl.Data))
		// for i := range cl.Data {
		// 	clPtr[i] = &cl.Data[i]
		// }

		if c.Globals.Verbose() {
			printVerbose(out, cl.Data)
		} else {
			printSummary(out, cl.Data)
		}

		if cl != nil && cl.Meta.NextCursor != "" {
			// Check if 'out' is interactive before prompting.
			if !c.Globals.Flags.NonInteractive && !c.Globals.Flags.AutoYes && text.IsTTY(out) {
				printNext, err := text.AskYesNo(out, "Print next page [y/N]: ", in)
				if err != nil {
					return err
				}
				if printNext {
					input.Cursor = &cl.Meta.NextCursor
					continue
				}
			}
		}

		return nil
	}
}

func (c *ListCommand) v0(out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	serviceID, serviceVersion, err := argparser.ServiceDetails(argparser.ServiceDetailsOpts{
		APIClient:          c.Globals.APIClient,
		Manifest:           *c.Globals.Manifest,
		Out:                out,
		ServiceNameFlag:    c.serviceName,
		ServiceVersionFlag: c.serviceVersion,
		VerboseMode:        c.Globals.Flags.Verbose,
	})
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": fsterr.ServiceVersion(serviceVersion),
		})
		return err
	}

	c.Input.ServiceID = serviceID
	c.Input.ServiceVersion = fastly.ToValue(serviceVersion.Number)

	o, err := c.Globals.APIClient.ListDomains(&c.Input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": fastly.ToValue(serviceVersion.Number),
		})
		return err
	}

	if ok, err := c.WriteJSON(out, o); ok {
		return err
	}

	if !c.Globals.Verbose() {
		tw := text.NewTable(out)
		tw.AddHeader("SERVICE", "VERSION", "NAME", "COMMENT")
		for _, domain := range o {
			tw.AddLine(
				fastly.ToValue(domain.ServiceID),
				fastly.ToValue(domain.ServiceVersion),
				fastly.ToValue(domain.Name),
				fastly.ToValue(domain.Comment),
			)
		}
		tw.Print()
		return nil
	}

	fmt.Fprintf(out, "Version: %d\n", c.Input.ServiceVersion)
	for i, domain := range o {
		fmt.Fprintf(out, "\tDomain %d/%d\n", i+1, len(o))
		fmt.Fprintf(out, "\t\tName: %s\n", fastly.ToValue(domain.Name))
		fmt.Fprintf(out, "\t\tComment: %v\n", fastly.ToValue(domain.Comment))
	}
	fmt.Fprintln(out)

	return nil
}
