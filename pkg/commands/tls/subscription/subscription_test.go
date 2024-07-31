package subscription_test

import (
	"fmt"
	"testing"

	"github.com/fastly/go-fastly/v9/fastly"

	root "github.com/fastly/cli/pkg/commands/tls/subscription"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

const (
	certificateAuthority  = "lets-encrypt"
	mockResponseID        = "123"
	validateAPIError      = "validate API error"
	validateAPISuccess    = "validate API success"
	validateMissingIDFlag = "validate missing --id flag"
)

func TestTLSSubscriptionCreate(t *testing.T) {
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --domain flag",
			WantError: "required flag --domain not provided",
		},
		{
			Name: validateAPIError,
			API: mock.API{
				CreateTLSSubscriptionFn: func(_ *fastly.CreateTLSSubscriptionInput) (*fastly.TLSSubscription, error) {
					return nil, testutil.Err
				},
			},
			Arg:       "--domain example.com",
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
			Arg:        "--domain example.com",
			WantOutput: fmt.Sprintf("Created TLS Subscription '%s' (Authority: %s, Common Name: example.com)", mockResponseID, certificateAuthority),
		},
	}

	testutil.RunScenarios(t, []string{root.CommandName, "create"}, scenarios)
}

func TestTLSSubscriptionDelete(t *testing.T) {
	scenarios := []testutil.TestScenario{
		{
			Name:      validateMissingIDFlag,
			WantError: "error parsing arguments: required flag --id not provided",
		},
		{
			Name: validateAPIError,
			API: mock.API{
				DeleteTLSSubscriptionFn: func(_ *fastly.DeleteTLSSubscriptionInput) error {
					return testutil.Err
				},
			},
			Arg:       "--id example",
			WantError: testutil.Err.Error(),
		},
		{
			Name: validateAPISuccess,
			API: mock.API{
				DeleteTLSSubscriptionFn: func(_ *fastly.DeleteTLSSubscriptionInput) error {
					return nil
				},
			},
			Arg:        "--id example",
			WantOutput: "Deleted TLS Subscription 'example' (force: false)",
		},
	}

	testutil.RunScenarios(t, []string{root.CommandName, "delete"}, scenarios)
}

func TestTLSSubscriptionDescribe(t *testing.T) {
	scenarios := []testutil.TestScenario{
		{
			Name:      validateMissingIDFlag,
			WantError: "error parsing arguments: required flag --id not provided",
		},
		{
			Name: validateAPIError,
			API: mock.API{
				GetTLSSubscriptionFn: func(_ *fastly.GetTLSSubscriptionInput) (*fastly.TLSSubscription, error) {
					return nil, testutil.Err
				},
			},
			Arg:       "--id example",
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
			Arg:        "--id example",
			WantOutput: "\nID: " + mockResponseID + "\nCertificate Authority: " + certificateAuthority + "\nState: pending\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\n",
		},
	}

	testutil.RunScenarios(t, []string{root.CommandName, "describe"}, scenarios)
}

func TestTLSSubscriptionList(t *testing.T) {
	scenarios := []testutil.TestScenario{
		{
			Name: validateAPIError,
			API: mock.API{
				ListTLSSubscriptionsFn: func(_ *fastly.ListTLSSubscriptionsInput) ([]*fastly.TLSSubscription, error) {
					return nil, testutil.Err
				},
			},
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
			Arg:        "--verbose",
			WantOutput: "\nID: " + mockResponseID + "\nCertificate Authority: " + certificateAuthority + "\nState: pending\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\n",
		},
	}

	testutil.RunScenarios(t, []string{root.CommandName, "list"}, scenarios)
}

func TestTLSSubscriptionUpdate(t *testing.T) {
	scenarios := []testutil.TestScenario{
		{
			Name:      validateMissingIDFlag,
			WantError: "required flag --id not provided",
		},
		{
			Name: validateAPIError,
			API: mock.API{
				UpdateTLSSubscriptionFn: func(_ *fastly.UpdateTLSSubscriptionInput) (*fastly.TLSSubscription, error) {
					return nil, testutil.Err
				},
			},
			Arg:       "--id example",
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
			Arg:        "--id example",
			WantOutput: fmt.Sprintf("Updated TLS Subscription '%s' (Authority: %s, Common Name: example.com)", mockResponseID, certificateAuthority),
		},
	}

	testutil.RunScenarios(t, []string{root.CommandName, "update"}, scenarios)
}
