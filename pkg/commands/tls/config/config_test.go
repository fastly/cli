package config_test

import (
	"bytes"
	"testing"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v6/fastly"
)

func TestDescribe(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --id flag",
			Args:      args("tls-config describe"),
			WantError: "error parsing arguments: required flag --id not provided",
		},
		{
			Name: "validate API error",
			API: mock.API{
				GetCustomTLSConfigurationFn: func(i *fastly.GetCustomTLSConfigurationInput) (*fastly.CustomTLSConfiguration, error) {
					return nil, testutil.Err
				},
			},
			Args:      args("tls-config describe --id example"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate API success",
			API: mock.API{
				GetCustomTLSConfigurationFn: func(_ *fastly.GetCustomTLSConfigurationInput) (*fastly.CustomTLSConfiguration, error) {
					t := testutil.Date
					return &fastly.CustomTLSConfiguration{
						ID:   "123",
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
			Args:       args("tls-config describe --id example"),
			WantOutput: "\nID: 123\nName: Foo\nDNS Record ID: 456\nDNS Record Type: Bar\nDNS Record Region: Baz\nBulk: true\nDefault: true\nHTTP Protocol: 1.1\nTLS Protocol: 1.3\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\n",
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
			Name: "validate API error",
			API: mock.API{
				ListCustomTLSConfigurationsFn: func(_ *fastly.ListCustomTLSConfigurationsInput) ([]*fastly.CustomTLSConfiguration, error) {
					return nil, testutil.Err
				},
			},
			Args:      args("tls-config list"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate API success",
			API: mock.API{
				ListCustomTLSConfigurationsFn: func(i *fastly.ListCustomTLSConfigurationsInput) ([]*fastly.CustomTLSConfiguration, error) {
					t := testutil.Date
					return []*fastly.CustomTLSConfiguration{
						{
							ID:   "123",
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
			Args:       args("tls-config list --verbose"),
			WantOutput: "\nID: 123\nName: Foo\nDNS Record ID: 456\nDNS Record Type: Bar\nDNS Record Region: Baz\nBulk: true\nDefault: true\nHTTP Protocol: 1.1\nTLS Protocol: 1.3\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\n",
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
			Name:      "validate missing --id flag",
			Args:      args("tls-config update --name example"),
			WantError: "error parsing arguments: required flag --id not provided",
		},
		{
			Name:      "validate missing --name flag",
			Args:      args("tls-config update --id 123"),
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Name: "validate API error",
			API: mock.API{
				UpdateCustomTLSConfigurationFn: func(i *fastly.UpdateCustomTLSConfigurationInput) (*fastly.CustomTLSConfiguration, error) {
					return nil, testutil.Err
				},
			},
			Args:      args("tls-config update --id example --name example"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate API success",
			API: mock.API{
				UpdateCustomTLSConfigurationFn: func(_ *fastly.UpdateCustomTLSConfigurationInput) (*fastly.CustomTLSConfiguration, error) {
					return &fastly.CustomTLSConfiguration{
						ID: "123",
					}, nil
				},
			},
			Args:       args("tls-config update --id example --name example"),
			WantOutput: "Updated TLS Configuration '123'",
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
