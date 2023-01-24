package secretstore

import (
	"github.com/fastly/cli/pkg/cmd"
)

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

func limitFlag(dst *int) cmd.IntFlagOpts {
	return cmd.IntFlagOpts{
		Name:        "limit",
		Short:       'l',
		Description: "Maxiumum number of items to list",
		Default:     50,
		Dst:         dst,
	}
}
