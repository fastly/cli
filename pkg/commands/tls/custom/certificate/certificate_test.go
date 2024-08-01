package certificate_test

import (
	"fmt"
	"testing"

	"github.com/fastly/go-fastly/v9/fastly"

	root "github.com/fastly/cli/pkg/commands/tls/custom"
	sub "github.com/fastly/cli/pkg/commands/tls/custom/certificate"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

const (
	mockResponseID        = "123"
	mockFieldValue        = "example"
	validateAPIError      = "validate API error"
	validateAPISuccess    = "validate API success"
	validateMissingIDFlag = "validate missing --id flag"
)

func TestTLSCustomCertCreate(t *testing.T) {
	var content string
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --cert-blob and --cert-path flags",
			WantError: "neither --cert-path or --cert-blob provided, one must be provided",
		},
		{
			Name:      "validate specifying both --cert-blob and --cert-path flags",
			Arg:       "--cert-blob foo --cert-path bar",
			WantError: "cert-path and cert-blob provided, only one can be specified",
		},
		{
			Name:      "validate invalid --cert-path arg",
			Arg:       "--cert-path ............",
			WantError: "error reading cert-path",
		},
		{
			Name: "validate custom cert is submitted",
			API: mock.API{
				CreateCustomTLSCertificateFn: func(certInput *fastly.CreateCustomTLSCertificateInput) (*fastly.CustomTLSCertificate, error) {
					content = certInput.CertBlob
					return &fastly.CustomTLSCertificate{
						ID: mockResponseID,
					}, nil
				},
			},
			Arg:             "--cert-path ./testdata/certificate.crt",
			WantOutput:      fmt.Sprintf("Created TLS Certificate '%s'", mockResponseID),
			PathContentFlag: &testutil.PathContentFlag{Flag: "cert-path", Fixture: "certificate.crt", Content: func() string { return content }},
		},
		{
			Name: validateAPIError,
			API: mock.API{
				CreateCustomTLSCertificateFn: func(certInput *fastly.CreateCustomTLSCertificateInput) (*fastly.CustomTLSCertificate, error) {
					content = certInput.CertBlob
					return nil, testutil.Err
				},
			},
			Arg:             "--cert-blob example",
			WantError:       testutil.Err.Error(),
			PathContentFlag: &testutil.PathContentFlag{Flag: "cert-path", Fixture: "certificate.crt", Content: func() string { return content }},
		},
		{
			Name: validateAPISuccess,
			API: mock.API{
				CreateCustomTLSCertificateFn: func(certInput *fastly.CreateCustomTLSCertificateInput) (*fastly.CustomTLSCertificate, error) {
					content = certInput.CertBlob
					return &fastly.CustomTLSCertificate{
						ID: mockResponseID,
					}, nil
				},
			},
			Arg:             "--cert-blob example",
			WantOutput:      fmt.Sprintf("Created TLS Certificate '%s'", mockResponseID),
			PathContentFlag: &testutil.PathContentFlag{Flag: "cert-path", Fixture: "certificate.crt", Content: func() string { return content }},
		},
	}

	testutil.RunScenarios(t, []string{root.CommandName, sub.CommandName, "create"}, scenarios)
}

func TestTLSCustomCertDelete(t *testing.T) {
	scenarios := []testutil.TestScenario{
		{
			Name:      validateMissingIDFlag,
			WantError: "error parsing arguments: required flag --id not provided",
		},
		{
			Name: validateAPIError,
			API: mock.API{
				DeleteCustomTLSCertificateFn: func(_ *fastly.DeleteCustomTLSCertificateInput) error {
					return testutil.Err
				},
			},
			Arg:       "--id example",
			WantError: testutil.Err.Error(),
		},
		{
			Name: validateAPISuccess,
			API: mock.API{
				DeleteCustomTLSCertificateFn: func(_ *fastly.DeleteCustomTLSCertificateInput) error {
					return nil
				},
			},
			Arg:        "--id example",
			WantOutput: "Deleted TLS Certificate 'example'",
		},
	}

	testutil.RunScenarios(t, []string{root.CommandName, sub.CommandName, "delete"}, scenarios)
}

func TestTLSCustomCertDescribe(t *testing.T) {
	scenarios := []testutil.TestScenario{
		{
			Name:      validateMissingIDFlag,
			WantError: "error parsing arguments: required flag --id not provided",
		},
		{
			Name: validateAPIError,
			API: mock.API{
				GetCustomTLSCertificateFn: func(_ *fastly.GetCustomTLSCertificateInput) (*fastly.CustomTLSCertificate, error) {
					return nil, testutil.Err
				},
			},
			Arg:       "--id example",
			WantError: testutil.Err.Error(),
		},
		{
			Name: validateAPISuccess,
			API: mock.API{
				GetCustomTLSCertificateFn: func(_ *fastly.GetCustomTLSCertificateInput) (*fastly.CustomTLSCertificate, error) {
					t := testutil.Date
					return &fastly.CustomTLSCertificate{
						ID:                 mockResponseID,
						IssuedTo:           mockFieldValue,
						Issuer:             mockFieldValue,
						Name:               mockFieldValue,
						Replace:            true,
						SerialNumber:       mockFieldValue,
						SignatureAlgorithm: mockFieldValue,
						CreatedAt:          &t,
						UpdatedAt:          &t,
					}, nil
				},
			},
			Arg:        "--id example",
			WantOutput: "\nID: " + mockResponseID + "\nIssued to: " + mockFieldValue + "\nIssuer: " + mockFieldValue + "\nName: " + mockFieldValue + "\nReplace: true\nSerial number: " + mockFieldValue + "\nSignature algorithm: " + mockFieldValue + "\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\n",
		},
	}

	testutil.RunScenarios(t, []string{root.CommandName, sub.CommandName, "describe"}, scenarios)
}

func TestTLSCustomCertList(t *testing.T) {
	scenarios := []testutil.TestScenario{
		{
			Name: validateAPIError,
			API: mock.API{
				ListCustomTLSCertificatesFn: func(_ *fastly.ListCustomTLSCertificatesInput) ([]*fastly.CustomTLSCertificate, error) {
					return nil, testutil.Err
				},
			},
			WantError: testutil.Err.Error(),
		},
		{
			Name: validateAPISuccess,
			API: mock.API{
				ListCustomTLSCertificatesFn: func(_ *fastly.ListCustomTLSCertificatesInput) ([]*fastly.CustomTLSCertificate, error) {
					t := testutil.Date
					return []*fastly.CustomTLSCertificate{
						{
							ID:                 mockResponseID,
							IssuedTo:           mockFieldValue,
							Issuer:             mockFieldValue,
							Name:               mockFieldValue,
							Replace:            true,
							SerialNumber:       mockFieldValue,
							SignatureAlgorithm: mockFieldValue,
							CreatedAt:          &t,
							UpdatedAt:          &t,
						},
					}, nil
				},
			},
			Arg:        "--verbose",
			WantOutput: "Fastly API endpoint: https://api.fastly.com\nFastly API token provided via config file (profile: user)\n\nID: " + mockResponseID + "\nIssued to: " + mockFieldValue + "\nIssuer: " + mockFieldValue + "\nName: " + mockFieldValue + "\nReplace: true\nSerial number: " + mockFieldValue + "\nSignature algorithm: " + mockFieldValue + "\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\n",
		},
	}

	testutil.RunScenarios(t, []string{root.CommandName, sub.CommandName, "list"}, scenarios)
}

func TestTLSCustomCertUpdate(t *testing.T) {
	var content string
	scenarios := []testutil.TestScenario{
		{
			Name:      validateMissingIDFlag,
			Arg:       "--cert-blob example",
			WantError: "required flag --id not provided",
		},
		{
			Name:      "validate missing --cert-blob and --cert-path flags",
			Arg:       "--id example",
			WantError: "neither --cert-path or --cert-blob provided, one must be provided",
		},
		{
			Name:      "validate specifying both --cert-blob and --cert-path flags",
			Arg:       "--id example --cert-blob foo --cert-path bar",
			WantError: "cert-path and cert-blob provided, only one can be specified",
		},
		{
			Name:      "validate invalid --cert-path arg",
			Arg:       "--id example --cert-path ............",
			WantError: "error reading cert-path",
		},
		{
			Name: validateAPIError,
			API: mock.API{
				UpdateCustomTLSCertificateFn: func(certInput *fastly.UpdateCustomTLSCertificateInput) (*fastly.CustomTLSCertificate, error) {
					content = certInput.CertBlob
					return nil, testutil.Err
				},
			},
			Arg:             "--cert-blob example --id example",
			WantError:       testutil.Err.Error(),
			PathContentFlag: &testutil.PathContentFlag{Flag: "cert-path", Fixture: "certificate.crt", Content: func() string { return content }},
		},
		{
			Name: validateAPISuccess,
			API: mock.API{
				UpdateCustomTLSCertificateFn: func(certInput *fastly.UpdateCustomTLSCertificateInput) (*fastly.CustomTLSCertificate, error) {
					content = certInput.CertBlob
					return &fastly.CustomTLSCertificate{
						ID: mockResponseID,
					}, nil
				},
			},
			Arg:             "--cert-blob example --id example",
			WantOutput:      fmt.Sprintf("Updated TLS Certificate '%s'", mockResponseID),
			PathContentFlag: &testutil.PathContentFlag{Flag: "cert-path", Fixture: "certificate.crt", Content: func() string { return content }},
		},
		{
			Name: validateAPISuccess + " with --name for different output",
			API: mock.API{
				UpdateCustomTLSCertificateFn: func(certInput *fastly.UpdateCustomTLSCertificateInput) (*fastly.CustomTLSCertificate, error) {
					content = certInput.CertBlob
					return &fastly.CustomTLSCertificate{
						ID:   mockResponseID,
						Name: "Updated",
					}, nil
				},
			},
			Arg:             "--cert-blob example --id example --name example",
			WantOutput:      "Updated TLS Certificate 'Updated' (previously: 'example')",
			PathContentFlag: &testutil.PathContentFlag{Flag: "cert-path", Fixture: "certificate.crt", Content: func() string { return content }},
		},
		{
			Name: "validate custom cert is submitted",
			API: mock.API{
				UpdateCustomTLSCertificateFn: func(certInput *fastly.UpdateCustomTLSCertificateInput) (*fastly.CustomTLSCertificate, error) {
					content = certInput.CertBlob
					return &fastly.CustomTLSCertificate{
						ID: mockResponseID,
					}, nil
				},
			},
			Arg:             "--id example --cert-path ./testdata/certificate.crt",
			WantOutput:      "SUCCESS: Updated TLS Certificate '123'",
			PathContentFlag: &testutil.PathContentFlag{Flag: "cert-path", Fixture: "certificate.crt", Content: func() string { return content }},
		},
	}

	testutil.RunScenarios(t, []string{root.CommandName, sub.CommandName, "update"}, scenarios)
}
