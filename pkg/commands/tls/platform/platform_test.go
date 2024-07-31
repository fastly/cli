package platform_test

import (
	"fmt"
	"testing"

	"github.com/fastly/go-fastly/v9/fastly"

	root "github.com/fastly/cli/pkg/commands/tls/platform"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

const (
	validateAPIError      = "validate API error"
	validateAPISuccess    = "validate API success"
	validateMissingIDFlag = "validate missing --id flag"
	mockResponseID        = "123"
)

func TestTLSPlatformUpload(t *testing.T) {
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --cert-blob flag",
			Arg:       "--intermediates-blob example",
			WantError: "required flag --cert-blob not provided",
		},
		{
			Name:      "validate missing --intermediates-blob flag",
			Arg:       "--cert-blob example",
			WantError: "required flag --intermediates-blob not provided",
		},
		{
			Name: validateAPIError,
			API: mock.API{
				CreateBulkCertificateFn: func(_ *fastly.CreateBulkCertificateInput) (*fastly.BulkCertificate, error) {
					return nil, testutil.Err
				},
			},
			Arg:       "--cert-blob example --intermediates-blob example",
			WantError: testutil.Err.Error(),
		},
		{
			Name: validateAPISuccess,
			API: mock.API{
				CreateBulkCertificateFn: func(_ *fastly.CreateBulkCertificateInput) (*fastly.BulkCertificate, error) {
					return &fastly.BulkCertificate{
						ID: mockResponseID,
					}, nil
				},
			},
			Arg:        "--cert-blob example --intermediates-blob example",
			WantOutput: fmt.Sprintf("Uploaded TLS Bulk Certificate '%s'", mockResponseID),
		},
	}

	testutil.RunScenarios(t, []string{root.CommandName, "upload"}, scenarios)
}

func TestTLSPlatformDelete(t *testing.T) {
	scenarios := []testutil.TestScenario{
		{
			Name:      validateMissingIDFlag,
			WantError: "error parsing arguments: required flag --id not provided",
		},
		{
			Name: validateAPIError,
			API: mock.API{
				DeleteBulkCertificateFn: func(_ *fastly.DeleteBulkCertificateInput) error {
					return testutil.Err
				},
			},
			Arg:       "--id example",
			WantError: testutil.Err.Error(),
		},
		{
			Name: validateAPISuccess,
			API: mock.API{
				DeleteBulkCertificateFn: func(_ *fastly.DeleteBulkCertificateInput) error {
					return nil
				},
			},
			Arg:        "--id example",
			WantOutput: "Deleted TLS Bulk Certificate 'example'",
		},
	}

	testutil.RunScenarios(t, []string{root.CommandName, "delete"}, scenarios)
}

func TestTLSPlatformDescribe(t *testing.T) {
	scenarios := []testutil.TestScenario{
		{
			Name:      validateMissingIDFlag,
			WantError: "error parsing arguments: required flag --id not provided",
		},
		{
			Name: validateAPIError,
			API: mock.API{
				GetBulkCertificateFn: func(_ *fastly.GetBulkCertificateInput) (*fastly.BulkCertificate, error) {
					return nil, testutil.Err
				},
			},
			Arg:       "--id example",
			WantError: testutil.Err.Error(),
		},
		{
			Name: validateAPISuccess,
			API: mock.API{
				GetBulkCertificateFn: func(_ *fastly.GetBulkCertificateInput) (*fastly.BulkCertificate, error) {
					t := testutil.Date
					return &fastly.BulkCertificate{
						ID:        "123",
						CreatedAt: &t,
						UpdatedAt: &t,
						Replace:   true,
					}, nil
				},
			},
			Arg:        "--id example",
			WantOutput: "\nID: 123\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\nReplace: true\n",
		},
	}

	testutil.RunScenarios(t, []string{root.CommandName, "describe"}, scenarios)
}

func TestTLSPlatformList(t *testing.T) {
	scenarios := []testutil.TestScenario{
		{
			Name: validateAPIError,
			API: mock.API{
				ListBulkCertificatesFn: func(_ *fastly.ListBulkCertificatesInput) ([]*fastly.BulkCertificate, error) {
					return nil, testutil.Err
				},
			},
			WantError: testutil.Err.Error(),
		},
		{
			Name: validateAPISuccess,
			API: mock.API{
				ListBulkCertificatesFn: func(_ *fastly.ListBulkCertificatesInput) ([]*fastly.BulkCertificate, error) {
					t := testutil.Date
					return []*fastly.BulkCertificate{
						{
							ID:        mockResponseID,
							CreatedAt: &t,
							UpdatedAt: &t,
							Replace:   true,
						},
					}, nil
				},
			},
			Arg:        "--verbose",
			WantOutput: "\nID: " + mockResponseID + "\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\nReplace: true\n",
		},
	}

	testutil.RunScenarios(t, []string{root.CommandName, "list"}, scenarios)
}

func TestTLSPlatformUpdate(t *testing.T) {
	scenarios := []testutil.TestScenario{
		{
			Name:      validateMissingIDFlag,
			Arg:       "--cert-blob example --intermediates-blob example",
			WantError: "required flag --id not provided",
		},
		{
			Name:      "validate missing --cert-blob flag",
			Arg:       "--id example --intermediates-blob example",
			WantError: "required flag --cert-blob not provided",
		},
		{
			Name:      "validate missing --intermediates-blob flag",
			Arg:       "--id example --cert-blob example",
			WantError: "required flag --intermediates-blob not provided",
		},
		{
			Name: validateAPIError,
			API: mock.API{
				UpdateBulkCertificateFn: func(_ *fastly.UpdateBulkCertificateInput) (*fastly.BulkCertificate, error) {
					return nil, testutil.Err
				},
			},
			Arg:       "--id example --cert-blob example --intermediates-blob example",
			WantError: testutil.Err.Error(),
		},
		{
			Name: validateAPISuccess,
			API: mock.API{
				UpdateBulkCertificateFn: func(_ *fastly.UpdateBulkCertificateInput) (*fastly.BulkCertificate, error) {
					return &fastly.BulkCertificate{
						ID: mockResponseID,
					}, nil
				},
			},
			Arg:        "--id example --cert-blob example --intermediates-blob example",
			WantOutput: "Updated TLS Bulk Certificate '123'",
		},
	}

	testutil.RunScenarios(t, []string{root.CommandName, "update"}, scenarios)
}
