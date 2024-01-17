package certificate_test

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/fastly/go-fastly/v8/fastly"

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
	//validateFilePathSuccess = "validate passing file path to --cert-blob works"
	//validateFilePathFailure = "validate passing invalid file path fails"
	validateCertBlobSuccess = "validate passing in full cert blob works"
)

func TestTLSCustomCertCreate(t *testing.T) {
	args := testutil.Args

	scenarios := []testutil.TestScenario{
		//{
		//	Name:      "validate missing --cert-blob flag",
		//	Args:      args("tls-custom certificate create"),
		//	WantError: "required flag --cert-blob not provided",
		//},
		//{
		//	Name: validateAPIError,
		//	API: mock.API{
		//		CreateCustomTLSCertificateFn: func(_ *fastly.CreateCustomTLSCertificateInput) (*fastly.CustomTLSCertificate, error) {
		//			return nil, testutil.Err
		//		},
		//	},
		//	Args:      args("tls-custom certificate create --cert-blob example"),
		//	WantError: testutil.Err.Error(),
		//},
		//{
		//	Name: validateAPISuccess,
		//	API: mock.API{
		//		CreateCustomTLSCertificateFn: func(_ *fastly.CreateCustomTLSCertificateInput) (*fastly.CustomTLSCertificate, error) {
		//			return &fastly.CustomTLSCertificate{
		//				ID: mockResponseID,
		//			}, nil
		//		},
		//	},
		//	Args:       args("tls-custom certificate create --cert-blob example"),
		//	WantOutput: fmt.Sprintf("Created TLS Certificate '%s'", mockResponseID),
		//},
		//{
		//	Name: validateFilePathFailure,
		//	API: mock.API{
		//		CreateCustomTLSCertificateFn: func(_ *fastly.CreateCustomTLSCertificateInput) (*fastly.CustomTLSCertificate, error) {
		//			return nil, testutil.Err
		//		},
		//	},
		//	Args:      args(fmt.Sprintf("tls-custom certificate create --cert-blob %s", filepath.Join("invalid", "path", "missing.crt"))),
		//	WantError: testutil.Err.Error(),
		//},
		//{
		//	Name: validateFilePathSuccess,
		//	API: mock.API{
		//		CreateCustomTLSCertificateFn: func(_ *fastly.CreateCustomTLSCertificateInput) (*fastly.CustomTLSCertificate, error) {
		//			return &fastly.CustomTLSCertificate{
		//				ID: mockResponseID,
		//			}, nil
		//		},
		//	},
		//	Args:       args(fmt.Sprintf("tls-custom certificate create --cert-blob %s", filepath.Join("./", "testdata", "certificate.crt"))),
		//	WantOutput: fmt.Sprintf("Created TLS Certificate '%s'", mockResponseID),
		//},
		{
			Name: validateCertBlobSuccess,
			API: mock.API{
				CreateCustomTLSCertificateFn: func(_ *fastly.CreateCustomTLSCertificateInput) (*fastly.CustomTLSCertificate, error) {
					return &fastly.CustomTLSCertificate{
						ID: mockResponseID,
					}, nil
				},
			},
			Args:       args("tls-custom certificate create --cert-blob `\"-----BEGINCERTIFICATE-----\nMIIDnDCCAoSgAwIBAgIJAIaudhxZj1xDMA0GCSqGSIb3DQEBCwUAMD8xGDAWBgNV\nBAMMD3J5YW5kb2hlcnR5Lm5ldDELMAkGA1UEBhMCVVMxFjAUBgNVBAcMDVNhbiBG\ncmFuc2lzY28wHhcNMjQwMTAyMjIxNzI3WhcNMjUwMTAxMjIxNzI3WjCBgjELMAkG\nA1UEBhMCVVMxEzARBgNVBAgMCkNhbGlmb3JuaWExFjAUBgNVBAcMDVNhbiBGcmFu\nc2lzY28xFTATBgNVBAoMDFJ5YW4gRG9oZXJ0eTEVMBMGA1UECwwMUnlhbiBEb2hl\ncnR5MRgwFgYDVQQDDA9yeWFuZG9oZXJ0eS5uZXQwggEiMA0GCSqGSIb3DQEBAQUA\nA4IBDwAwggEKAoIBAQCwje6gREvHarGjvKlX5KOahPuErSwX5HbsfueWWXG8grkq\nGuAKSwrXV3jIja7rHsmWZrMooJdg5pDl5hHqiVzRUrK8E1qJd83HyCbEK62rVaDQ\nxUf9QT0MlvIANGBncUBKPETPH53TYeLxWzvXRtYIxhWpPrpTK8p8i4cmB4Io2abc\npirBejL3cgmg/PTzou+LjKv1jerEkGPvp5GqlgpRqWHeq5KCUcsZA0eSTPmPzZBX\nCENyXcXX+P7CrrFLvbZ+Q2BZO3tEOp+kdfR0Tit4ZeR6cx0ZxUhcWV2J7B0tNiQ3\nE0IEIpYt7p66jsPNX5rW0p5qiJejomCnmywTnVIRAgMBAAGjVzBVMB8GA1UdIwQY\nMBaAFGyYO+Vpyy+RqvjbQGdZg84q6IBxMAkGA1UdEwQCMAAwCwYDVR0PBAQDAgTw\nMBoGA1UdEQQTMBGCD3J5YW5kb2hlcnR5Lm5ldDANBgkqhkiG9w0BAQsFAAOCAQEA\nAKpwxBQ5raWRgja4YLWQDL1WgkE7O/hxvpeBKgNn0xr1ZjYZ06Kxb1HzOHbUW9D8\n8Sc37ClTrWXAdLilkVbTEOO5yzVwu0iQeeQp8KIFJPlVAQJU1SLea5832Fqhxu/7\nfae7OK59bfKaPGUf5ZFGXeatkMbw0bMcZ4p4zQmCNfG1ABEKiZ6LNq9ujw15PMPr\nutJ6pqiZNDejxESCZhoR7hSvahIpbiQuIHOTzXz02DcqnEzUpz1w77QTwyDGhxgl\nCeU7HTlQvIJx18Z/C5p60Yafzl2lThDRk199MvITBHwxGJeHS7oW3GayI9b8V+lb\nEhaGRxHY3wavIR7FxZNF3w==\n-----ENDCERTIFICATE-----`\""),
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
