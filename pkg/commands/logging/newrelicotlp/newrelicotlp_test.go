package newrelicotlp_test

import (
	"bytes"
	"testing"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestNewRelicOTLPCreate(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --service-id flag",
			Args:      args("logging newrelicotlp create --key abc --name foo --version 3"),
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate missing --autoclone flag",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
			},
			Args:      args("logging newrelicotlp create --key abc --name foo --service-id 123 --version 1"),
			WantError: "service version 1 is not editable",
		},
		{
			Name: "validate CreateNewRelicOTLP API error",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CreateNewRelicOTLPFn: func(i *fastly.CreateNewRelicOTLPInput) (*fastly.NewRelicOTLP, error) {
					return nil, testutil.Err
				},
			},
			Args:      args("logging newrelicotlp create --key abc --name foo --service-id 123 --version 3"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate CreateNewRelicOTLP API success",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CreateNewRelicOTLPFn: func(i *fastly.CreateNewRelicOTLPInput) (*fastly.NewRelicOTLP, error) {
					return &fastly.NewRelicOTLP{
						Name:           *i.Name,
						ServiceID:      i.ServiceID,
						ServiceVersion: i.ServiceVersion,
					}, nil
				},
			},
			Args:       args("logging newrelicotlp create --key abc --name foo --service-id 123 --version 3"),
			WantOutput: "Created New Relic OTLP logging endpoint 'foo' (service: 123, version: 3)",
		},
		{
			Name: "validate --autoclone results in cloned service version",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				CreateNewRelicOTLPFn: func(i *fastly.CreateNewRelicOTLPInput) (*fastly.NewRelicOTLP, error) {
					return &fastly.NewRelicOTLP{
						Name:           *i.Name,
						ServiceID:      i.ServiceID,
						ServiceVersion: i.ServiceVersion,
					}, nil
				},
			},
			Args:       args("logging newrelicotlp create --autoclone --key abc --name foo --service-id 123 --version 1"),
			WantOutput: "Created New Relic OTLP logging endpoint 'foo' (service: 123, version: 4)",
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

func TestNewRelicOTLPDelete(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --name flag",
			Args:      args("logging newrelicotlp delete --version 3"),
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Name:      "validate missing --version flag",
			Args:      args("logging newrelicotlp delete --name foobar"),
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Args:      args("logging newrelicotlp delete --name foobar --version 3"),
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate missing --autoclone flag",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
			},
			Args:      args("logging newrelicotlp delete --name foobar --service-id 123 --version 1"),
			WantError: "service version 1 is not editable",
		},
		{
			Name: "validate DeleteNewRelic API error",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				DeleteNewRelicOTLPFn: func(i *fastly.DeleteNewRelicOTLPInput) error {
					return testutil.Err
				},
			},
			Args:      args("logging newrelicotlp delete --name foobar --service-id 123 --version 3"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate DeleteNewRelic API success",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				DeleteNewRelicOTLPFn: func(i *fastly.DeleteNewRelicOTLPInput) error {
					return nil
				},
			},
			Args:       args("logging newrelicotlp delete --name foobar --service-id 123 --version 3"),
			WantOutput: "Deleted New Relic OTLP logging endpoint 'foobar' (service: 123, version: 3)",
		},
		{
			Name: "validate --autoclone results in cloned service version",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				DeleteNewRelicOTLPFn: func(i *fastly.DeleteNewRelicOTLPInput) error {
					return nil
				},
			},
			Args:       args("logging newrelicotlp delete --autoclone --name foo --service-id 123 --version 1"),
			WantOutput: "Deleted New Relic OTLP logging endpoint 'foo' (service: 123, version: 4)",
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

func TestNewRelicDescribe(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --name flag",
			Args:      args("logging newrelicotlp describe --version 3"),
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Name:      "validate missing --version flag",
			Args:      args("logging newrelicotlp describe --name foobar"),
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Args:      args("logging newrelicotlp describe --name foobar --version 3"),
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate GetNewRelic API error",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetNewRelicOTLPFn: func(i *fastly.GetNewRelicOTLPInput) (*fastly.NewRelicOTLP, error) {
					return nil, testutil.Err
				},
			},
			Args:      args("logging newrelicotlp describe --name foobar --service-id 123 --version 3"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate GetNewRelic API success",
			API: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				GetNewRelicOTLPFn: getNewRelic,
			},
			Args:       args("logging newrelicotlp describe --name foobar --service-id 123 --version 3"),
			WantOutput: "\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nDeleted at: 2021-06-15 23:00:00 +0000 UTC\nFormat: \nFormat Version: 0\nName: foobar\nPlacement: \nRegion: \nResponse Condition: \nService ID: 123\nService Version: 3\nToken: abc\nURL: \nUpdated at: 2021-06-15 23:00:00 +0000 UTC\n",
		},
		{
			Name: "validate missing --autoclone flag is OK",
			API: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				GetNewRelicOTLPFn: getNewRelic,
			},
			Args:       args("logging newrelicotlp describe --name foobar --service-id 123 --version 1"),
			WantOutput: "\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nDeleted at: 2021-06-15 23:00:00 +0000 UTC\nFormat: \nFormat Version: 0\nName: foobar\nPlacement: \nRegion: \nResponse Condition: \nService ID: 123\nService Version: 1\nToken: abc\nURL: \nUpdated at: 2021-06-15 23:00:00 +0000 UTC\n",
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

func TestNewRelicList(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --version flag",
			Args:      args("logging newrelicotlp list"),
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Args:      args("logging newrelicotlp list --version 3"),
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate ListNewRelics API error",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListNewRelicOTLPFn: func(i *fastly.ListNewRelicOTLPInput) ([]*fastly.NewRelicOTLP, error) {
					return nil, testutil.Err
				},
			},
			Args:      args("logging newrelicotlp list --service-id 123 --version 3"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate ListNewRelics API success",
			API: mock.API{
				ListVersionsFn:     testutil.ListVersions,
				ListNewRelicOTLPFn: listNewRelic,
			},
			Args:       args("logging newrelicotlp list --service-id 123 --version 3"),
			WantOutput: "SERVICE ID  VERSION  NAME\n123         3        foo\n123         3        bar\n",
		},
		{
			Name: "validate missing --autoclone flag is OK",
			API: mock.API{
				ListVersionsFn:     testutil.ListVersions,
				ListNewRelicOTLPFn: listNewRelic,
			},
			Args:       args("logging newrelicotlp list --service-id 123 --version 1"),
			WantOutput: "SERVICE ID  VERSION  NAME\n123         1        foo\n123         1        bar\n",
		},
		{
			Name: "validate missing --verbose flag",
			API: mock.API{
				ListVersionsFn:     testutil.ListVersions,
				ListNewRelicOTLPFn: listNewRelic,
			},
			Args:       args("logging newrelicotlp list --service-id 123 --verbose --version 1"),
			WantOutput: "Fastly API token provided via config file (profile: user)\nFastly API endpoint: https://api.fastly.com\n\nService ID (via --service-id): 123\n\nService Version: 1\n\nName: foo\n\nToken: \n\nFormat: \n\nFormat Version: 0\n\nPlacement: \n\nRegion: \n\nResponse Condition: \n\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\nDeleted at: 2021-06-15 23:00:00 +0000 UTC\n\nName: bar\n\nToken: \n\nFormat: \n\nFormat Version: 0\n\nPlacement: \n\nRegion: \n\nResponse Condition: \n\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\nDeleted at: 2021-06-15 23:00:00 +0000 UTC\n",
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

func TestNewRelicUpdate(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --name flag",
			Args:      args("logging newrelicotlp update --service-id 123 --version 3"),
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Name:      "validate missing --version flag",
			Args:      args("logging newrelicotlp update --name foobar --service-id 123"),
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Args:      args("logging newrelicotlp update --name foobar --version 3"),
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate missing --autoclone flag",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
			},
			Args:      args("logging newrelicotlp update --name foobar --service-id 123 --version 1"),
			WantError: "service version 1 is not editable",
		},
		{
			Name: "validate UpdateNewRelic API error",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				UpdateNewRelicOTLPFn: func(i *fastly.UpdateNewRelicOTLPInput) (*fastly.NewRelicOTLP, error) {
					return nil, testutil.Err
				},
			},
			Args:      args("logging newrelicotlp update --name foobar --new-name beepboop --service-id 123 --version 3"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate UpdateNewRelic API success",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				UpdateNewRelicOTLPFn: func(i *fastly.UpdateNewRelicOTLPInput) (*fastly.NewRelicOTLP, error) {
					return &fastly.NewRelicOTLP{
						Name:           *i.NewName,
						ServiceID:      i.ServiceID,
						ServiceVersion: i.ServiceVersion,
					}, nil
				},
			},
			Args:       args("logging newrelicotlp update --name foobar --new-name beepboop --service-id 123 --version 3"),
			WantOutput: "Updated New Relic OTLP logging endpoint 'beepboop' (previously: foobar, service: 123, version: 3)",
		},
		{
			Name: "validate --autoclone results in cloned service version",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				UpdateNewRelicOTLPFn: func(i *fastly.UpdateNewRelicOTLPInput) (*fastly.NewRelicOTLP, error) {
					return &fastly.NewRelicOTLP{
						Name:           *i.NewName,
						ServiceID:      i.ServiceID,
						ServiceVersion: i.ServiceVersion,
					}, nil
				},
			},
			Args:       args("logging newrelicotlp update --autoclone --name foobar --new-name beepboop --service-id 123 --version 1"),
			WantOutput: "Updated New Relic OTLP logging endpoint 'beepboop' (previously: foobar, service: 123, version: 4)",
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

func getNewRelic(i *fastly.GetNewRelicOTLPInput) (*fastly.NewRelicOTLP, error) {
	t := testutil.Date

	return &fastly.NewRelicOTLP{
		Name:           i.Name,
		Token:          "abc",
		ServiceID:      i.ServiceID,
		ServiceVersion: i.ServiceVersion,

		CreatedAt: &t,
		DeletedAt: &t,
		UpdatedAt: &t,
	}, nil
}

func listNewRelic(i *fastly.ListNewRelicOTLPInput) ([]*fastly.NewRelicOTLP, error) {
	t := testutil.Date
	vs := []*fastly.NewRelicOTLP{
		{
			Name:           "foo",
			ServiceID:      i.ServiceID,
			ServiceVersion: i.ServiceVersion,

			CreatedAt: &t,
			DeletedAt: &t,
			UpdatedAt: &t,
		},
		{
			Name:           "bar",
			ServiceID:      i.ServiceID,
			ServiceVersion: i.ServiceVersion,

			CreatedAt: &t,
			DeletedAt: &t,
			UpdatedAt: &t,
		},
	}
	return vs, nil
}
