package secretstore

import (
	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/kingpin"
)

// Secret Store flags.

func storeIDFlag(dst *string) cmd.StringFlagOpts {
	return cmd.StringFlagOpts{
		Name:        "store-id",
		Short:       's',
		Description: "Store ID",
		Dst:         dst,
		Required:    true,
	}
}

func storeNameFlag(dst *string) cmd.StringFlagOpts {
	return cmd.StringFlagOpts{
		Name:        "name",
		Short:       'n',
		Description: "Store name",
		Dst:         dst,
		Required:    true,
	}
}

func secretNameFlag(dst *string) cmd.StringFlagOpts {
	return cmd.StringFlagOpts{
		Name:        "name",
		Short:       'n',
		Description: "Secret name",
		Dst:         dst,
		Required:    true,
	}
}

func secretFileFlag(dst *string) cmd.StringFlagOpts {
	return cmd.StringFlagOpts{
		Name:        "file",
		Short:       'f',
		Description: "Read secret value from file instead of prompt",
		Dst:         dst,
		Required:    false,
	}
}

func secretStdinFlag(dst *bool) cmd.BoolFlagOpts {
	return cmd.BoolFlagOpts{
		Name:        "stdin",
		Description: "Read secret value from STDIN instead of prompt",
		Dst:         dst,
		Required:    false,
	}
}

func cursorFlag(dst *string) cmd.StringFlagOpts {
	return cmd.StringFlagOpts{
		Name:        "cursor",
		Short:       'c',
		Description: "Pagination cursor (Use 'next_cursor' value from list output)",
		Dst:         dst,
	}
}

func limitFlag(cmd *kingpin.CmdClause, dst *int) {
	limit := cmd.Flag("limit", "Maxiumum number of items to list")
	limit = limit.Default("50")
	limit = limit.Short('l')
	limit.IntVar(dst)
}
