package platform_test

import (
	"bytes"
	"testing"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v6/fastly"
)

func TestCreate(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --cert-blob flag",
			Args:      args("tls-platform upload --intermediates-blob example"),
			WantError: "error parsing arguments: required flag --cert-blob not provided",
		},
		{
			Name:      "validate missing --intermediates-blob flag",
			Args:      args("tls-platform upload --cert-blob example"),
			WantError: "error parsing arguments: required flag --intermediates-blob not provided",
		},
		{
			Name: "validate API error",
			API: mock.API{
				CreateBulkCertificateFn: func(i *fastly.CreateBulkCertificateInput) (*fastly.BulkCertificate, error) {
					return nil, testutil.Err
				},
			},
			Args:      args("tls-platform upload --cert-blob example --intermediates-blob example"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate API success",
			API: mock.API{
				CreateBulkCertificateFn: func(i *fastly.CreateBulkCertificateInput) (*fastly.BulkCertificate, error) {
					return &fastly.BulkCertificate{
						ID: "123",
					}, nil
				},
			},
			Args:       args("tls-platform upload --cert-blob example --intermediates-blob example"),
			WantOutput: "Uploaded TLS Bulk Certificate '123'",
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
