package products_test

import (
	"testing"

	root "github.com/fastly/cli/pkg/commands/products"
	"github.com/fastly/cli/pkg/testutil"
)

func TestProductEnablement(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing Service ID",
			WantError: "failed to identify Service ID: error reading service: no service ID found",
		},
		{
			Name:      "validate invalid enable/disable flag combo",
			Args:      "--enable fanout --disable fanout",
			WantError: "invalid flag combination: --enable and --disable",
		},
		{
			Name:      "validate flag parsing error for enabling product",
			Args:      "--service-id 123 --enable foo",
			WantError: "error parsing arguments: enum value must be one of bot_management,brotli_compression,domain_inspector,fanout,image_optimizer,log_explorer_insights,origin_inspector,websockets, got 'foo'",
		},
		{
			Name:      "validate flag parsing error for disabling product",
			Args:      "--service-id 123 --disable foo",
			WantError: "error parsing arguments: enum value must be one of bot_management,brotli_compression,domain_inspector,fanout,image_optimizer,log_explorer_insights,origin_inspector,websockets, got 'foo'",
		},
		{
			Name:      "validate invalid json/verbose flag combo",
			Args:      "--service-id 123 --json --verbose",
			WantError: "invalid flag combination, --verbose and --json",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName}, scenarios)
}
