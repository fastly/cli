package secretstore

import (
	"github.com/fastly/cli/v10/pkg/argparser"
)

func storeNameFlag(dst *string) argparser.StringFlagOpts {
	return argparser.StringFlagOpts{
		Name:        "name",
		Short:       'n',
		Description: "Store name",
		Dst:         dst,
		Required:    true,
	}
}
