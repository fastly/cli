package secretstore

import (
	"github.com/fastly/cli/pkg/cmd"
)

func storeNameFlag(dst *string) cmd.StringFlagOpts {
	return cmd.StringFlagOpts{
		Name:        "name",
		Short:       'n',
		Description: "Store name",
		Dst:         dst,
		Required:    true,
	}
}
