package stats

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
)

// RootCommand dispatches all "stats" commands.
type RootCommand struct {
	argparser.Base

	Aggregate       AggregateCommand       `cmd:"" help:"View aggregated stats across all services"`
	DomainInspector DomainInspectorCommand `cmd:"" help:"View domain metrics for a Fastly service"`
	Historical      HistoricalCommand      `cmd:"" help:"View historical stats for a Fastly service"`
	OriginInspector OriginInspectorCommand `cmd:"" help:"View origin metrics for a Fastly service"`
	Realtime        RealtimeCommand        `cmd:"" help:"View realtime stats for a Fastly service"`
	Regions         RegionsCommand         `cmd:"" help:"List stats regions"`
	Usage           UsageCommand           `cmd:"" help:"View usage stats (bandwidth, requests)"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "stats"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}

func resolveJSONFormat(formatFlag *string, jsonFlag bool, g *global.Data) error {
	if jsonFlag {
		*formatFlag = "json"
	}
	if *formatFlag == "json" {
		g.Flags.JSON = true
		if g.Verbose() {
			return fsterr.ErrInvalidVerboseJSONCombo
		}
	}
	return nil
}
