package certificate_test

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/global"
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
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --cert-blob and --cert-path flags",
			Args:      args("tls-custom certificate create"),
			WantError: "neither --cert-path or --cert-blob provided, one must be provided",
		},
		{
			Name:      "validate specifying both --cert-blob and --cert-path flags",
			Args:      args("tls-custom certificate create --cert-blob foo --cert-path bar"),
			WantError: "cert-path and cert-blob provided, only one can be specified",
		},
		{
			Name:      "validate invalid --cert-path arg",
			Args:      args("tls-custom certificate create --cert-path ............"),
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
			Args:       args("tls-custom certificate create --cert-path ./testdata/certificate.crt"),
			WantOutput: fmt.Sprintf("Created TLS Certificate '%s'", mockResponseID),
		},
		{
			Name: validateAPIError,
			API: mock.API{
				CreateCustomTLSCertificateFn: func(certInput *fastly.CreateCustomTLSCertificateInput) (*fastly.CustomTLSCertificate, error) {
					content = certInput.CertBlob
					return nil, testutil.Err
				},
			},
			Args:      args("tls-custom certificate create --cert-blob example"),
			WantError: testutil.Err.Error(),
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
			Args:       args("tls-custom certificate create --cert-blob example"),
			WantOutput: fmt.Sprintf("Created TLS Certificate '%s'", mockResponseID),
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
			testutil.AssertPathContentFlag("cert-path", testcase.WantError, testcase.Args, "certificate.crt", content, t)
		})
	}
}

func TestTLSCustomCertDelete(t *testing.T) {
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

func TestTLSCustomCertDescribe(t *testing.T) {
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

func TestTLSCustomCertList(t *testing.T) {
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
			WantOutput: "Fastly API endpoint: https://api.fastly.com\nFastly API token provided via config file (profile: user)\n\nID: " + mockResponseID + "\nIssued to: " + mockFieldValue + "\nIssuer: " + mockFieldValue + "\nName: " + mockFieldValue + "\nReplace: true\nSerial number: " + mockFieldValue + "\nSignature algorithm: " + mockFieldValue + "\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\n",
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

func TestTLSCustomCertUpdate(t *testing.T) {
	var content string
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name:      validateMissingIDFlag,
			Args:      args("tls-custom certificate update --cert-blob example"),
			WantError: "required flag --id not provided",
		},
		{
			Name:      "validate missing --cert-blob and --cert-path flags",
			Args:      args("tls-custom certificate create"),
			WantError: "neither --cert-path or --cert-blob provided, one must be provided",
		},
		{
			Name:      "validate specifying both --cert-blob and --cert-path flags",
			Args:      args("tls-custom certificate create --cert-blob foo --cert-path bar"),
			WantError: "cert-path and cert-blob provided, only one can be specified",
		},
		{
			Name:      "validate invalid --cert-path arg",
			Args:      args("tls-custom certificate create --cert-path ............"),
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
			Args:      args("tls-custom certificate update --cert-blob example --id example"),
			WantError: testutil.Err.Error(),
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
			Args:       args("tls-custom certificate update --cert-blob example --id example"),
			WantOutput: fmt.Sprintf("Updated TLS Certificate '%s'", mockResponseID),
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
			Args:       args("tls-custom certificate update --cert-blob example --id example --name example"),
			WantOutput: "Updated TLS Certificate 'Updated' (previously: 'example')",
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
			Args:       args("tls-custom certificate update --id example --cert-path ./testdata/certificate.crt"),
			WantOutput: "SUCCESS: Updated TLS Certificate '123'",
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
			testutil.AssertPathContentFlag("cert-path", testcase.WantError, testcase.Args, "certificate.crt", content, t)
		})
	}
}
