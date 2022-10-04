package common

import (
	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/kingpin"
)

func Format(cmd *kingpin.CmdClause, c cmd.OptionalString) {
	cmd.Flag("format", "Apache style log formatting").Action(c.Set).StringVar(&c.Value)
}
