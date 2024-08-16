package domain_test

import (
	"testing"

	"github.com/fastly/go-fastly/v9/fastly"

	root "github.com/fastly/cli/pkg/commands/tls/custom"
	sub "github.com/fastly/cli/pkg/commands/tls/custom/domain"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

const (
	mockResponseID     = "123"
	validateAPIError   = "validate API error"
	validateAPISuccess = "validate API success"
)

func TestList(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name: validateAPIError,
			API: mock.API{
				ListTLSDomainsFn: func(_ *fastly.ListTLSDomainsInput) ([]*fastly.TLSDomain, error) {
					return nil, testutil.Err
				},
			},
			WantError: testutil.Err.Error(),
		},
		{
			Name: validateAPISuccess,
			API: mock.API{
				ListTLSDomainsFn: func(_ *fastly.ListTLSDomainsInput) ([]*fastly.TLSDomain, error) {
					return []*fastly.TLSDomain{
						{
							ID:   mockResponseID,
							Type: "example",
						},
					}, nil
				},
			},
			Args:       "--verbose",
			WantOutput: "\nID: " + mockResponseID + "\nType: example\n\n",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "list"}, scenarios)
}
