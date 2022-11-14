package activation_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v7/fastly"
)

const (
	mockResponseID        = "123"
	mockResponseCertID    = "456"
	validateAPIError      = "validate API error"
	validateAPISuccess    = "validate API success"
	validateMissingIDFlag = "validate missing --id flag"
)

func TestCreate(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name:      validateMissingIDFlag,
			Args:      args("tls-custom activation enable --cert-id example"),
			WantError: "required flag --id not provided",
		},
		{
			Name:      validateMissingIDFlag,
			Args:      args("tls-custom activation enable --id example"),
			WantError: "required flag --cert-id not provided",
		},
		{
			Name: validateAPIError,
			API: mock.API{
				CreateTLSActivationFn: func(_ *fastly.CreateTLSActivationInput) (*fastly.TLSActivation, error) {
					return nil, testutil.Err
				},
			},
			Args:      args("tls-custom activation enable --cert-id example --id example"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: validateAPISuccess,
			API: mock.API{
				CreateTLSActivationFn: func(_ *fastly.CreateTLSActivationInput) (*fastly.TLSActivation, error) {
					return &fastly.TLSActivation{
						ID: mockResponseID,
						Certificate: &fastly.CustomTLSCertificate{
							ID: mockResponseCertID,
						},
					}, nil
				},
			},
			Args:       args("tls-custom activation enable --cert-id example --id example"),
			WantOutput: fmt.Sprintf("Enabled TLS Activation '%s' (Certificate '%s')", mockResponseID, mockResponseCertID),
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
			Args:      args("tls-custom activation disable"),
			WantError: "error parsing arguments: required flag --id not provided",
		},
		{
			Name: validateAPIError,
			API: mock.API{
				DeleteTLSActivationFn: func(_ *fastly.DeleteTLSActivationInput) error {
					return testutil.Err
				},
			},
			Args:      args("tls-custom activation disable --id example"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: validateAPISuccess,
			API: mock.API{
				DeleteTLSActivationFn: func(_ *fastly.DeleteTLSActivationInput) error {
					return nil
				},
			},
			Args:       args("tls-custom activation disable --id example"),
			WantOutput: "Disabled TLS Activation 'example'",
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
			Args:      args("tls-custom activation describe"),
			WantError: "error parsing arguments: required flag --id not provided",
		},
		{
			Name: validateAPIError,
			API: mock.API{
				GetTLSActivationFn: func(_ *fastly.GetTLSActivationInput) (*fastly.TLSActivation, error) {
					return nil, testutil.Err
				},
			},
			Args:      args("tls-custom activation describe --id example"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: validateAPISuccess,
			API: mock.API{
				GetTLSActivationFn: func(_ *fastly.GetTLSActivationInput) (*fastly.TLSActivation, error) {
					t := testutil.Date
					return &fastly.TLSActivation{
						ID:        mockResponseID,
						CreatedAt: &t,
					}, nil
				},
			},
			Args:       args("tls-custom activation describe --id example"),
			WantOutput: "\nID: " + mockResponseID + "\nCreated at: 2021-06-15 23:00:00 +0000 UTC\n",
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
				ListTLSActivationsFn: func(_ *fastly.ListTLSActivationsInput) ([]*fastly.TLSActivation, error) {
					return nil, testutil.Err
				},
			},
			Args:      args("tls-custom activation list"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: validateAPISuccess,
			API: mock.API{
				ListTLSActivationsFn: func(_ *fastly.ListTLSActivationsInput) ([]*fastly.TLSActivation, error) {
					t := testutil.Date
					return []*fastly.TLSActivation{
						{
							ID:        mockResponseID,
							CreatedAt: &t,
						},
					}, nil
				},
			},
			Args:       args("tls-custom activation list --verbose"),
			WantOutput: "\nID: " + mockResponseID + "\nCreated at: 2021-06-15 23:00:00 +0000 UTC\n",
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
			Name:      validateMissingIDFlag,
			Args:      args("tls-custom activation update --cert-id example"),
			WantError: "required flag --id not provided",
		},
		{
			Name:      validateMissingIDFlag,
			Args:      args("tls-custom activation update --id example"),
			WantError: "required flag --cert-id not provided",
		},
		{
			Name: validateAPIError,
			API: mock.API{
				UpdateTLSActivationFn: func(_ *fastly.UpdateTLSActivationInput) (*fastly.TLSActivation, error) {
					return nil, testutil.Err
				},
			},
			Args:      args("tls-custom activation update --cert-id example --id example"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: validateAPISuccess,
			API: mock.API{
				UpdateTLSActivationFn: func(_ *fastly.UpdateTLSActivationInput) (*fastly.TLSActivation, error) {
					return &fastly.TLSActivation{
						ID: mockResponseID,
						Certificate: &fastly.CustomTLSCertificate{
							ID: mockResponseCertID,
						},
					}, nil
				},
			},
			Args:       args("tls-custom activation update --cert-id example --id example"),
			WantOutput: fmt.Sprintf("Updated TLS Activation Certificate '%s' (previously: 'example')", mockResponseCertID),
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
