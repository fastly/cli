package platform_test

import (
	"fmt"
	"testing"

	"github.com/fastly/go-fastly/v10/fastly"

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
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --cert-blob flag",
			Args:      "--intermediates-blob example",
			WantError: "required flag --cert-blob not provided",
		},
		{
			Name:      "validate missing --intermediates-blob flag",
			Args:      "--cert-blob example",
			WantError: "required flag --intermediates-blob not provided",
		},
		{
			Name: validateAPIError,
			API: mock.API{
				CreateBulkCertificateFn: func(_ *fastly.CreateBulkCertificateInput) (*fastly.BulkCertificate, error) {
					return nil, testutil.Err
				},
			},
			Args:      "--cert-blob example --intermediates-blob example",
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
			Args:       "--cert-blob example --intermediates-blob example",
			WantOutput: fmt.Sprintf("Uploaded TLS Bulk Certificate '%s'", mockResponseID),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, "upload"}, scenarios)
}

func TestTLSPlatformDelete(t *testing.T) {
	scenarios := []testutil.CLIScenario{
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
			Args:      "--id example",
			WantError: testutil.Err.Error(),
		},
		{
			Name: validateAPISuccess,
			API: mock.API{
				DeleteBulkCertificateFn: func(_ *fastly.DeleteBulkCertificateInput) error {
					return nil
				},
			},
			Args:       "--id example",
			WantOutput: "Deleted TLS Bulk Certificate 'example'",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, "delete"}, scenarios)
}

func TestTLSPlatformDescribe(t *testing.T) {
	scenarios := []testutil.CLIScenario{
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
			Args:      "--id example",
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
			Args:       "--id example",
			WantOutput: "\nID: 123\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\nReplace: true\n",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, "describe"}, scenarios)
}

func TestTLSPlatformList(t *testing.T) {
	scenarios := []testutil.CLIScenario{
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
			Args:       "--verbose",
			WantOutput: "\nID: " + mockResponseID + "\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\nReplace: true\n",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, "list"}, scenarios)
}

func TestTLSPlatformUpdate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      validateMissingIDFlag,
			Args:      "--cert-blob example --intermediates-blob example",
			WantError: "required flag --id not provided",
		},
		{
			Name:      "validate missing --cert-blob flag",
			Args:      "--id example --intermediates-blob example",
			WantError: "required flag --cert-blob not provided",
		},
		{
			Name:      "validate missing --intermediates-blob flag",
			Args:      "--id example --cert-blob example",
			WantError: "required flag --intermediates-blob not provided",
		},
		{
			Name: validateAPIError,
			API: mock.API{
				UpdateBulkCertificateFn: func(_ *fastly.UpdateBulkCertificateInput) (*fastly.BulkCertificate, error) {
					return nil, testutil.Err
				},
			},
			Args:      "--id example --cert-blob example --intermediates-blob example",
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
			Args:       "--id example --cert-blob example --intermediates-blob example",
			WantOutput: "Updated TLS Bulk Certificate '123'",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, "update"}, scenarios)
}
