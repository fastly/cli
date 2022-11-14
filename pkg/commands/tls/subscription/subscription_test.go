package subscription_test

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
	certificateAuthority  = "lets-encrypt"
	mockResponseID        = "123"
	validateAPIError      = "validate API error"
	validateAPISuccess    = "validate API success"
	validateMissingIDFlag = "validate missing --id flag"
)

func TestCreate(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --domain flag",
			Args:      args("tls-subscription create"),
			WantError: "required flag --domain not provided",
		},
		{
			Name: validateAPIError,
			API: mock.API{
				CreateTLSSubscriptionFn: func(_ *fastly.CreateTLSSubscriptionInput) (*fastly.TLSSubscription, error) {
					return nil, testutil.Err
				},
			},
			Args:      args("tls-subscription create --domain example.com"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: validateAPISuccess,
			API: mock.API{
				CreateTLSSubscriptionFn: func(_ *fastly.CreateTLSSubscriptionInput) (*fastly.TLSSubscription, error) {
					return &fastly.TLSSubscription{
						ID:                   mockResponseID,
						CertificateAuthority: certificateAuthority,
						CommonName: &fastly.TLSDomain{
							ID: "example.com",
						},
					}, nil
				},
			},
			Args:       args("tls-subscription create --domain example.com"),
			WantOutput: fmt.Sprintf("Created TLS Subscription '%s' (Authority: %s, Common Name: example.com)", mockResponseID, certificateAuthority),
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
			Args:      args("tls-subscription delete"),
			WantError: "error parsing arguments: required flag --id not provided",
		},
		{
			Name: validateAPIError,
			API: mock.API{
				DeleteTLSSubscriptionFn: func(_ *fastly.DeleteTLSSubscriptionInput) error {
					return testutil.Err
				},
			},
			Args:      args("tls-subscription delete --id example"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: validateAPISuccess,
			API: mock.API{
				DeleteTLSSubscriptionFn: func(_ *fastly.DeleteTLSSubscriptionInput) error {
					return nil
				},
			},
			Args:       args("tls-subscription delete --id example"),
			WantOutput: "Deleted TLS Subscription 'example' (force: false)",
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
			Args:      args("tls-subscription describe"),
			WantError: "error parsing arguments: required flag --id not provided",
		},
		{
			Name: validateAPIError,
			API: mock.API{
				GetTLSSubscriptionFn: func(_ *fastly.GetTLSSubscriptionInput) (*fastly.TLSSubscription, error) {
					return nil, testutil.Err
				},
			},
			Args:      args("tls-subscription describe --id example"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: validateAPISuccess,
			API: mock.API{
				GetTLSSubscriptionFn: func(_ *fastly.GetTLSSubscriptionInput) (*fastly.TLSSubscription, error) {
					t := testutil.Date
					return &fastly.TLSSubscription{
						ID:                   mockResponseID,
						CertificateAuthority: certificateAuthority,
						State:                "pending",
						CreatedAt:            &t,
						UpdatedAt:            &t,
					}, nil
				},
			},
			Args:       args("tls-subscription describe --id example"),
			WantOutput: "\nID: " + mockResponseID + "\nCertificate Authority: " + certificateAuthority + "\nState: pending\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\n",
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
				ListTLSSubscriptionsFn: func(_ *fastly.ListTLSSubscriptionsInput) ([]*fastly.TLSSubscription, error) {
					return nil, testutil.Err
				},
			},
			Args:      args("tls-subscription list"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: validateAPISuccess,
			API: mock.API{
				ListTLSSubscriptionsFn: func(_ *fastly.ListTLSSubscriptionsInput) ([]*fastly.TLSSubscription, error) {
					t := testutil.Date
					return []*fastly.TLSSubscription{
						{
							ID:                   mockResponseID,
							CertificateAuthority: certificateAuthority,
							State:                "pending",
							CreatedAt:            &t,
							UpdatedAt:            &t,
						},
					}, nil
				},
			},
			Args:       args("tls-subscription list --verbose"),
			WantOutput: "\nID: " + mockResponseID + "\nCertificate Authority: " + certificateAuthority + "\nState: pending\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\n",
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
			Args:      args("tls-subscription update"),
			WantError: "required flag --id not provided",
		},
		{
			Name: validateAPIError,
			API: mock.API{
				UpdateTLSSubscriptionFn: func(_ *fastly.UpdateTLSSubscriptionInput) (*fastly.TLSSubscription, error) {
					return nil, testutil.Err
				},
			},
			Args:      args("tls-subscription update --id example"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: validateAPISuccess,
			API: mock.API{
				UpdateTLSSubscriptionFn: func(_ *fastly.UpdateTLSSubscriptionInput) (*fastly.TLSSubscription, error) {
					return &fastly.TLSSubscription{
						ID:                   mockResponseID,
						CertificateAuthority: certificateAuthority,
						CommonName: &fastly.TLSDomain{
							ID: "example.com",
						},
					}, nil
				},
			},
			Args:       args("tls-subscription update --id example"),
			WantOutput: fmt.Sprintf("Updated TLS Subscription '%s' (Authority: %s, Common Name: example.com)", mockResponseID, certificateAuthority),
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
