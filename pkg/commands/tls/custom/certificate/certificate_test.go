package certificate_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/app"
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

func TestCreate(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --cert-blob flag",
			Args:      args("tls-custom certificate create"),
			WantError: "required flag --cert-blob not provided",
		},
		{
			Name: validateAPIError,
			API: mock.API{
				CreateCustomTLSCertificateFn: func(_ *fastly.CreateCustomTLSCertificateInput) (*fastly.CustomTLSCertificate, error) {
					return nil, testutil.Err
				},
			},
			Args:      args("tls-custom certificate create --cert-blob example"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: validateAPISuccess,
			API: mock.API{
				CreateCustomTLSCertificateFn: func(_ *fastly.CreateCustomTLSCertificateInput) (*fastly.CustomTLSCertificate, error) {
					return &fastly.CustomTLSCertificate{
						ID: mockResponseID,
					}, nil
				},
			},
			Args:       args("tls-custom certificate create --cert-blob example"),
			WantOutput: fmt.Sprintf("Created TLS Certificate '%s'", mockResponseID),
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

func TestDelete(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name:      validateMissingIDFlag,
			Args:      args("tls-custom certificate delete"),
			WantError: "error parsing arguments: required flag --id not provided",
		},
		{
			Name: validateAPIError,
			API: mock.API{
				DeleteCustomTLSCertificateFn: func(_ *fastly.DeleteCustomTLSCertificateInput) error {
					return testutil.Err
				},
			},
			Args:      args("tls-custom certificate delete --id example"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: validateAPISuccess,
			API: mock.API{
				DeleteCustomTLSCertificateFn: func(_ *fastly.DeleteCustomTLSCertificateInput) error {
					return nil
				},
			},
			Args:       args("tls-custom certificate delete --id example"),
			WantOutput: "Deleted TLS Certificate 'example'",
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

func TestDescribe(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name:      validateMissingIDFlag,
			Args:      args("tls-custom certificate describe"),
			WantError: "error parsing arguments: required flag --id not provided",
		},
		{
			Name: validateAPIError,
			API: mock.API{
				GetCustomTLSCertificateFn: func(_ *fastly.GetCustomTLSCertificateInput) (*fastly.CustomTLSCertificate, error) {
					return nil, testutil.Err
				},
			},
			Args:      args("tls-custom certificate describe --id example"),
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
			Args:       args("tls-custom certificate describe --id example"),
			WantOutput: "\nID: " + mockResponseID + "\nIssued to: " + mockFieldValue + "\nIssuer: " + mockFieldValue + "\nName: " + mockFieldValue + "\nReplace: true\nSerial number: " + mockFieldValue + "\nSignature algorithm: " + mockFieldValue + "\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\n",
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

func TestList(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name: validateAPIError,
			API: mock.API{
				ListCustomTLSCertificatesFn: func(_ *fastly.ListCustomTLSCertificatesInput) ([]*fastly.CustomTLSCertificate, error) {
					return nil, testutil.Err
				},
			},
			Args:      args("tls-custom certificate list"),
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
			Args:       args("tls-custom certificate list --verbose"),
			WantOutput: "Fastly API token provided via config file (profile: user)\nFastly API endpoint: https://api.fastly.com\n\nID: " + mockResponseID + "\nIssued to: " + mockFieldValue + "\nIssuer: " + mockFieldValue + "\nName: " + mockFieldValue + "\nReplace: true\nSerial number: " + mockFieldValue + "\nSignature algorithm: " + mockFieldValue + "\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\n",
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

func TestUpdate(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --cert-blob flag",
			Args:      args("tls-custom certificate update --id example"),
			WantError: "required flag --cert-blob not provided",
		},
		{
			Name:      validateMissingIDFlag,
			Args:      args("tls-custom certificate update --cert-blob example"),
			WantError: "required flag --id not provided",
		},
		{
			Name: validateAPIError,
			API: mock.API{
				UpdateCustomTLSCertificateFn: func(_ *fastly.UpdateCustomTLSCertificateInput) (*fastly.CustomTLSCertificate, error) {
					return nil, testutil.Err
				},
			},
			Args:      args("tls-custom certificate update --cert-blob example --id example"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: validateAPISuccess,
			API: mock.API{
				UpdateCustomTLSCertificateFn: func(_ *fastly.UpdateCustomTLSCertificateInput) (*fastly.CustomTLSCertificate, error) {
					return &fastly.CustomTLSCertificate{
						ID: mockResponseID,
					}, nil
				},
			},
			Args:       args("tls-custom certificate update --cert-blob example --id example"),
			WantOutput: fmt.Sprintf("Updated TLS Certificate '%s'", mockResponseID),
		},
		{
			Name: validateAPISuccess + " with --name for different output",
			API: mock.API{
				UpdateCustomTLSCertificateFn: func(_ *fastly.UpdateCustomTLSCertificateInput) (*fastly.CustomTLSCertificate, error) {
					return &fastly.CustomTLSCertificate{
						ID:   mockResponseID,
						Name: "Updated",
					}, nil
				},
			},
			Args:       args("tls-custom certificate update --cert-blob example --id example --name example"),
			WantOutput: "Updated TLS Certificate 'Updated' (previously: 'example')",
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
