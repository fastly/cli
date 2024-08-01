package activation_test

import (
	"fmt"
	"testing"

	"github.com/fastly/go-fastly/v9/fastly"

	root "github.com/fastly/cli/pkg/commands/tls/custom"
	sub "github.com/fastly/cli/pkg/commands/tls/custom/activation"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

const (
	mockResponseID        = "123"
	mockResponseCertID    = "456"
	validateAPIError      = "validate API error"
	validateAPISuccess    = "validate API success"
	validateMissingIDFlag = "validate missing --id flag"
)

func TestTLSCustomActivationEnable(t *testing.T) {
	scenarios := []testutil.TestScenario{
		{
			Name:      validateMissingIDFlag,
			Arg:       "--cert-id example",
			WantError: "required flag --id not provided",
		},
		{
			Name:      validateMissingIDFlag,
			Arg:       "--id example",
			WantError: "required flag --cert-id not provided",
		},
		{
			Name: validateAPIError,
			API: mock.API{
				CreateTLSActivationFn: func(_ *fastly.CreateTLSActivationInput) (*fastly.TLSActivation, error) {
					return nil, testutil.Err
				},
			},
			Arg:       "--cert-id example --id example",
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
			Arg:        "--cert-id example --id example",
			WantOutput: fmt.Sprintf("Enabled TLS Activation '%s' (Certificate '%s')", mockResponseID, mockResponseCertID),
		},
	}

	testutil.RunScenarios(t, []string{root.CommandName, sub.CommandName, "enable"}, scenarios)
}

func TestTLSCustomActivationDisable(t *testing.T) {
	scenarios := []testutil.TestScenario{
		{
			Name:      validateMissingIDFlag,
			WantError: "error parsing arguments: required flag --id not provided",
		},
		{
			Name: validateAPIError,
			API: mock.API{
				DeleteTLSActivationFn: func(_ *fastly.DeleteTLSActivationInput) error {
					return testutil.Err
				},
			},
			Arg:       "--id example",
			WantError: testutil.Err.Error(),
		},
		{
			Name: validateAPISuccess,
			API: mock.API{
				DeleteTLSActivationFn: func(_ *fastly.DeleteTLSActivationInput) error {
					return nil
				},
			},
			Arg:        "--id example",
			WantOutput: "Disabled TLS Activation 'example'",
		},
	}

	testutil.RunScenarios(t, []string{root.CommandName, sub.CommandName, "disable"}, scenarios)
}

func TestTLSCustomActivationDescribe(t *testing.T) {
	scenarios := []testutil.TestScenario{
		{
			Name:      validateMissingIDFlag,
			WantError: "error parsing arguments: required flag --id not provided",
		},
		{
			Name: validateAPIError,
			API: mock.API{
				GetTLSActivationFn: func(_ *fastly.GetTLSActivationInput) (*fastly.TLSActivation, error) {
					return nil, testutil.Err
				},
			},
			Arg:       "--id example",
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
			Arg:        "--id example",
			WantOutput: "\nID: " + mockResponseID + "\nCreated at: 2021-06-15 23:00:00 +0000 UTC\n",
		},
	}

	testutil.RunScenarios(t, []string{root.CommandName, sub.CommandName, "describe"}, scenarios)
}

func TestTLSCustomActivationList(t *testing.T) {
	scenarios := []testutil.TestScenario{
		{
			Name: validateAPIError,
			API: mock.API{
				ListTLSActivationsFn: func(_ *fastly.ListTLSActivationsInput) ([]*fastly.TLSActivation, error) {
					return nil, testutil.Err
				},
			},
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
			Arg:        "--verbose",
			WantOutput: "\nID: " + mockResponseID + "\nCreated at: 2021-06-15 23:00:00 +0000 UTC\n",
		},
	}

	testutil.RunScenarios(t, []string{root.CommandName, sub.CommandName, "list"}, scenarios)
}

func TestTLSCustomActivationUpdate(t *testing.T) {
	scenarios := []testutil.TestScenario{
		{
			Name:      validateMissingIDFlag,
			Arg:       "--cert-id example",
			WantError: "required flag --id not provided",
		},
		{
			Name:      validateMissingIDFlag,
			Arg:       "--id example",
			WantError: "required flag --cert-id not provided",
		},
		{
			Name: validateAPIError,
			API: mock.API{
				UpdateTLSActivationFn: func(_ *fastly.UpdateTLSActivationInput) (*fastly.TLSActivation, error) {
					return nil, testutil.Err
				},
			},
			Arg:       "--cert-id example --id example",
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
			Arg:        "--cert-id example --id example",
			WantOutput: fmt.Sprintf("Updated TLS Activation Certificate '%s' (previously: 'example')", mockResponseCertID),
		},
	}

	testutil.RunScenarios(t, []string{root.CommandName, sub.CommandName, "update"}, scenarios)
}
