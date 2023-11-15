package domain_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
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
			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts := testutil.MockGlobalData(testcase.Args, &stdout)
				opts.APIClientFactory = mock.APIClient(testcase.API)
				return opts, nil
			}
			err := app.Run(testcase.Args, nil)
			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.WantOutput)
		})
	}
}
