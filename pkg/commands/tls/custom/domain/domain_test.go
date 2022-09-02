package domain_test

import (
	"bytes"
	"testing"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v6/fastly"
)

const (
	mockResponseID     = "123"
	validateAPIError   = "validate API error"
	validateAPISuccess = "validate API success"
)

func TestList(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name: validateAPIError,
			API: mock.API{
				ListTLSDomainsFn: func(_ *fastly.ListTLSDomainsInput) ([]*fastly.TLSDomain, error) {
					return nil, testutil.Err
				},
			},
			Args:      args("tls-custom domain list"),
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
			Args:       args("tls-custom domain list --verbose"),
			WantOutput: "\nID: " + mockResponseID + "\nType: example\n\n",
		},
	}

	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(testcase.Name, func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.Args, &stdout)
			opts.APIClient = mock.APIClient(testcase.API)
			err := app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.WantOutput)
		})
	}
}
