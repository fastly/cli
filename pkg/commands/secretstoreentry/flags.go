package secretstoreentry

import (
	"github.com/fastly/cli/v10/pkg/argparser"
)

func secretNameFlag(dst *string) argparser.StringFlagOpts {
	return argparser.StringFlagOpts{
		Name:        "name",
		Short:       'n',
		Description: "Secret name",
		Dst:         dst,
		Required:    true,
	}
}

func secretFileFlag(dst *string) argparser.StringFlagOpts {
	return argparser.StringFlagOpts{
		Name:        "file",
		Short:       'f',
		Description: "Read secret value from file instead of prompt",
		Dst:         dst,
		Required:    false,
	}
}

func secretStdinFlag(dst *bool) argparser.BoolFlagOpts {
	return argparser.BoolFlagOpts{
		Name:        "stdin",
		Description: "Read secret value from STDIN instead of prompt",
		Dst:         dst,
		Required:    false,
	}
}
