package activation_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/fastly/go-fastly/v13/fastly"

	root "github.com/fastly/cli/pkg/commands/tls/custom"
	sub "github.com/fastly/cli/pkg/commands/tls/custom/activation"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

const (
	mockResponseID        = "123"
	mockResponseCertID    = "456"
	mockResponseConfigID  = "789"
	mockResponseDomain    = "tls.example.com"
	validateAPIError      = "validate API error"
	validateAPISuccess    = "validate API success"
	validateMissingIDFlag = "validate missing --id flag"
)

func TestTLSCustomActivationEnable(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing CertID flag",
			Args:      fmt.Sprintf("--tls-config-id %s --tls-domain %s", mockResponseConfigID, mockResponseDomain),
			WantError: "required flag --cert-id not provided",
		},
		{
			Name:      "validate missing ConfigID flag",
			Args:      fmt.Sprintf("--cert-id %s --tls-domain %s", mockResponseCertID, mockResponseDomain),
			WantError: "required flag --tls-config-id not provided",
		},
		{
			Name:      "validate missing Domain Flag",
			Args:      fmt.Sprintf("--cert-id %s --tls-config-id %s", mockResponseCertID, mockResponseConfigID),
			WantError: "required flag --tls-domain not provided",
		},
		{
			Name: validateAPIError,
			API: mock.API{
				CreateTLSActivationFn: func(_ context.Context, _ *fastly.CreateTLSActivationInput) (*fastly.TLSActivation, error) {
					return nil, testutil.Err
				},
			},
			Args:      fmt.Sprintf("--cert-id %s --tls-config-id %s --tls-domain %s", mockResponseCertID, mockResponseConfigID, mockResponseDomain),
			WantError: testutil.Err.Error(),
		},
		{
			Name: validateAPISuccess,
			API: mock.API{
				CreateTLSActivationFn: func(_ context.Context, _ *fastly.CreateTLSActivationInput) (*fastly.TLSActivation, error) {
					return &fastly.TLSActivation{
						ID: mockResponseID,
					}, nil
				},
			},
			Args:       fmt.Sprintf("--cert-id %s --tls-config-id %s --tls-domain %s", mockResponseCertID, mockResponseConfigID, mockResponseDomain),
			WantOutput: fmt.Sprintf("SUCCESS: Enabled TLS Activation '%s' (Certificate '%s', Configuration '%s')", mockResponseID, mockResponseCertID, mockResponseConfigID),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "enable"}, scenarios)
}

func TestTLSCustomActivationDisable(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      validateMissingIDFlag,
			WantError: "error parsing arguments: required flag --id not provided",
		},
		{
			Name: validateAPIError,
			API: mock.API{
				DeleteTLSActivationFn: func(_ context.Context, _ *fastly.DeleteTLSActivationInput) error {
					return testutil.Err
				},
			},
			Args:      "--id example",
			WantError: testutil.Err.Error(),
		},
		{
			Name: validateAPISuccess,
			API: mock.API{
				DeleteTLSActivationFn: func(_ context.Context, _ *fastly.DeleteTLSActivationInput) error {
					return nil
				},
			},
			Args:       "--id example",
			WantOutput: "Disabled TLS Activation 'example'",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "disable"}, scenarios)
}

func TestTLSCustomActivationDescribe(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      validateMissingIDFlag,
			WantError: "error parsing arguments: required flag --id not provided",
		},
		{
			Name: validateAPIError,
			API: mock.API{
				GetTLSActivationFn: func(_ context.Context, _ *fastly.GetTLSActivationInput) (*fastly.TLSActivation, error) {
					return nil, testutil.Err
				},
			},
			Args:      "--id example",
			WantError: testutil.Err.Error(),
		},
		{
			Name: validateAPISuccess,
			API: mock.API{
				GetTLSActivationFn: func(_ context.Context, _ *fastly.GetTLSActivationInput) (*fastly.TLSActivation, error) {
					t := testutil.Date
					return &fastly.TLSActivation{
						ID:        mockResponseID,
						CreatedAt: &t,
					}, nil
				},
			},
			Args:       "--id example",
			WantOutput: "\nID: " + mockResponseID + "\nCreated at: 2021-06-15 23:00:00 +0000 UTC\n",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "describe"}, scenarios)
}

func TestTLSCustomActivationList(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name: validateAPIError,
			API: mock.API{
				ListTLSActivationsFn: func(_ context.Context, _ *fastly.ListTLSActivationsInput) ([]*fastly.TLSActivation, error) {
					return nil, testutil.Err
				},
			},
			WantError: testutil.Err.Error(),
		},
		{
			Name: validateAPISuccess,
			API: mock.API{
				ListTLSActivationsFn: func(_ context.Context, _ *fastly.ListTLSActivationsInput) ([]*fastly.TLSActivation, error) {
					t := testutil.Date
					return []*fastly.TLSActivation{
						{
							ID:        mockResponseID,
							CreatedAt: &t,
						},
					}, nil
				},
			},
			Args:       "--verbose",
			WantOutput: "\nID: " + mockResponseID + "\nCreated at: 2021-06-15 23:00:00 +0000 UTC\n",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "list"}, scenarios)
}

func TestTLSCustomActivationUpdate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      validateMissingIDFlag,
			Args:      "--cert-id example",
			WantError: "required flag --id not provided",
		},
		{
			Name:      validateMissingIDFlag,
			Args:      "--id example",
			WantError: "required flag --cert-id not provided",
		},
		{
			Name: validateAPIError,
			API: mock.API{
				UpdateTLSActivationFn: func(_ context.Context, _ *fastly.UpdateTLSActivationInput) (*fastly.TLSActivation, error) {
					return nil, testutil.Err
				},
			},
			Args:      "--cert-id example --id example",
			WantError: testutil.Err.Error(),
		},
		{
			Name: validateAPISuccess,
			API: mock.API{
				UpdateTLSActivationFn: func(_ context.Context, _ *fastly.UpdateTLSActivationInput) (*fastly.TLSActivation, error) {
					return &fastly.TLSActivation{
						ID: mockResponseID,
						Certificate: &fastly.CustomTLSCertificate{
							ID: mockResponseCertID,
						},
					}, nil
				},
			},
			Args:       "--cert-id example --id example",
			WantOutput: fmt.Sprintf("Updated TLS Activation Certificate '%s' (previously: 'example')", mockResponseCertID),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "update"}, scenarios)
}
