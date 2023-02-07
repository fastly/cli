package secretstoreentry

import (
	"github.com/fastly/cli/pkg/cmd"
)

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
