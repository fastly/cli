package config_test

import (
	"fmt"
	"testing"

	"github.com/fastly/go-fastly/v9/fastly"

	root "github.com/fastly/cli/pkg/commands/tls/config"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

const (
	validateAPIError   = "validate API error"
	validateAPISuccess = "validate API success"
	mockResponseID     = "123"
)

func TestDescribe(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --id flag",
			WantError: "error parsing arguments: required flag --id not provided",
		},
		{
			Name: validateAPIError,
			API: mock.API{
				GetCustomTLSConfigurationFn: func(i *fastly.GetCustomTLSConfigurationInput) (*fastly.CustomTLSConfiguration, error) {
					return nil, testutil.Err
				},
			},
			Args:      "--id example",
			WantError: testutil.Err.Error(),
		},
		{
			Name: validateAPISuccess,
			API: mock.API{
				GetCustomTLSConfigurationFn: func(_ *fastly.GetCustomTLSConfigurationInput) (*fastly.CustomTLSConfiguration, error) {
					t := testutil.Date
					return &fastly.CustomTLSConfiguration{
						ID:   mockResponseID,
						Name: "Foo",
						DNSRecords: []*fastly.DNSRecord{
							{
								ID:         "456",
								RecordType: "Bar",
								Region:     "Baz",
							},
						},
						Bulk:          true,
						Default:       true,
						HTTPProtocols: []string{"1.1"},
						TLSProtocols:  []string{"1.3"},
						CreatedAt:     &t,
						UpdatedAt:     &t,
					}, nil
				},
			},
			Args:       "--id example",
			WantOutput: "\nID: " + mockResponseID + "\nName: Foo\nDNS Record ID: 456\nDNS Record Type: Bar\nDNS Record Region: Baz\nBulk: true\nDefault: true\nHTTP Protocol: 1.1\nTLS Protocol: 1.3\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\n",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, "describe"}, scenarios)
}

func TestList(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name: validateAPIError,
			API: mock.API{
				ListCustomTLSConfigurationsFn: func(_ *fastly.ListCustomTLSConfigurationsInput) ([]*fastly.CustomTLSConfiguration, error) {
					return nil, testutil.Err
				},
			},
			WantError: testutil.Err.Error(),
		},
		{
			Name: validateAPISuccess,
			API: mock.API{
				ListCustomTLSConfigurationsFn: func(_ *fastly.ListCustomTLSConfigurationsInput) ([]*fastly.CustomTLSConfiguration, error) {
					t := testutil.Date
					return []*fastly.CustomTLSConfiguration{
						{
							ID:   mockResponseID,
							Name: "Foo",
							DNSRecords: []*fastly.DNSRecord{
								{
									ID:         "456",
									RecordType: "Bar",
									Region:     "Baz",
								},
							},
							Bulk:          true,
							Default:       true,
							HTTPProtocols: []string{"1.1"},
							TLSProtocols:  []string{"1.3"},
							CreatedAt:     &t,
							UpdatedAt:     &t,
						},
					}, nil
				},
			},
			Args:       "--verbose",
			WantOutput: "\nID: " + mockResponseID + "\nName: Foo\nDNS Record ID: 456\nDNS Record Type: Bar\nDNS Record Region: Baz\nBulk: true\nDefault: true\nHTTP Protocol: 1.1\nTLS Protocol: 1.3\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\n",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, "list"}, scenarios)
}

func TestUpdate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --id flag",
			Args:      "--name example",
			WantError: "error parsing arguments: required flag --id not provided",
		},
		{
			Name:      "validate missing --name flag",
			Args:      "--id 123",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Name: validateAPIError,
			API: mock.API{
				UpdateCustomTLSConfigurationFn: func(i *fastly.UpdateCustomTLSConfigurationInput) (*fastly.CustomTLSConfiguration, error) {
					return nil, testutil.Err
				},
			},
			Args:      "--id example --name example",
			WantError: testutil.Err.Error(),
		},
		{
			Name: validateAPISuccess,
			API: mock.API{
				UpdateCustomTLSConfigurationFn: func(_ *fastly.UpdateCustomTLSConfigurationInput) (*fastly.CustomTLSConfiguration, error) {
					return &fastly.CustomTLSConfiguration{
						ID: mockResponseID,
					}, nil
				},
			},
			Args:       "--id example --name example",
			WantOutput: fmt.Sprintf("Updated TLS Configuration '%s'", mockResponseID),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, "update"}, scenarios)
}
