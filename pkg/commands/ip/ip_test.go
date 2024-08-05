package ip_test

import (
	"testing"

	"github.com/fastly/go-fastly/v9/fastly"

	root "github.com/fastly/cli/pkg/commands/ip"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestAllIPs(t *testing.T) {
	scenarios := []testutil.TestScenario{
		{
			Name: "validate listing IP addresses",
			API: mock.API{
				AllIPsFn: func() (v4, v6 fastly.IPAddrs, err error) {
					return []string{
							"00.123.45.6/78",
						}, []string{
							"0a12:3b45::/67",
						}, nil
				},
			},
			WantOutput: "\nIPv4\n\t00.123.45.6/78\n\nIPv6\n\t0a12:3b45::/67\n",
		},
	}

	testutil.RunScenarios(t, []string{root.CommandName}, scenarios)
}
