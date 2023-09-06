package newrelic_test

import (
	"bytes"
	"testing"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestNewRelicCreate(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --service-id flag",
			Args:      args("logging newrelic create --key abc --name foo --version 3"),
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate missing --autoclone flag",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
			},
			Args:      args("logging newrelic create --key abc --name foo --service-id 123 --version 1"),
			WantError: "service version 1 is not editable",
		},
		{
			Name: "validate CreateNewRelic API error",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CreateNewRelicFn: func(i *fastly.CreateNewRelicInput) (*fastly.NewRelic, error) {
					return nil, testutil.Err
				},
			},
			Args:      args("logging newrelic create --key abc --name foo --service-id 123 --version 3"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate CreateNewRelic API success",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CreateNewRelicFn: func(i *fastly.CreateNewRelicInput) (*fastly.NewRelic, error) {
					return &fastly.NewRelic{
						Name:           *i.Name,
						ServiceID:      i.ServiceID,
						ServiceVersion: i.ServiceVersion,
					}, nil
				},
			},
			Args:       args("logging newrelic create --key abc --name foo --service-id 123 --version 3"),
			WantOutput: "Created New Relic logging endpoint 'foo' (service: 123, version: 3)",
		},
		{
			Name: "validate --autoclone results in cloned service version",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				CreateNewRelicFn: func(i *fastly.CreateNewRelicInput) (*fastly.NewRelic, error) {
					return &fastly.NewRelic{
						Name:           *i.Name,
						ServiceID:      i.ServiceID,
						ServiceVersion: i.ServiceVersion,
					}, nil
				},
			},
			Args:       args("logging newrelic create --autoclone --key abc --name foo --service-id 123 --version 1"),
			WantOutput: "Created New Relic logging endpoint 'foo' (service: 123, version: 4)",
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

func TestNewRelicDelete(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --name flag",
			Args:      args("logging newrelic delete --version 3"),
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Name:      "validate missing --version flag",
			Args:      args("logging newrelic delete --name foobar"),
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Args:      args("logging newrelic delete --name foobar --version 3"),
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate missing --autoclone flag",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
			},
			Args:      args("logging newrelic delete --name foobar --service-id 123 --version 1"),
			WantError: "service version 1 is not editable",
		},
		{
			Name: "validate DeleteNewRelic API error",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				DeleteNewRelicFn: func(i *fastly.DeleteNewRelicInput) error {
					return testutil.Err
				},
			},
			Args:      args("logging newrelic delete --name foobar --service-id 123 --version 3"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate DeleteNewRelic API success",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				DeleteNewRelicFn: func(i *fastly.DeleteNewRelicInput) error {
					return nil
				},
			},
			Args:       args("logging newrelic delete --name foobar --service-id 123 --version 3"),
			WantOutput: "Deleted New Relic logging endpoint 'foobar' (service: 123, version: 3)",
		},
		{
			Name: "validate --autoclone results in cloned service version",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				DeleteNewRelicFn: func(i *fastly.DeleteNewRelicInput) error {
					return nil
				},
			},
			Args:       args("logging newrelic delete --autoclone --name foo --service-id 123 --version 1"),
			WantOutput: "Deleted New Relic logging endpoint 'foo' (service: 123, version: 4)",
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
			Args:      args("logging newrelic describe --version 3"),
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Name:      "validate missing --version flag",
			Args:      args("logging newrelic describe --name foobar"),
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Args:      args("logging newrelic describe --name foobar --version 3"),
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate GetNewRelic API error",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetNewRelicFn: func(i *fastly.GetNewRelicInput) (*fastly.NewRelic, error) {
					return nil, testutil.Err
				},
			},
			Args:      args("logging newrelic describe --name foobar --service-id 123 --version 3"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate GetNewRelic API success",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetNewRelicFn:  getNewRelic,
			},
			Args:       args("logging newrelic describe --name foobar --service-id 123 --version 3"),
			WantOutput: "\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nDeleted at: 2021-06-15 23:00:00 +0000 UTC\nFormat: \nFormat Version: 0\nName: foobar\nPlacement: \nRegion: \nResponse Condition: \nService ID: 123\nService Version: 3\nToken: abc\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\n",
		},
		{
			Name: "validate missing --autoclone flag is OK",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetNewRelicFn:  getNewRelic,
			},
			Args:       args("logging newrelic describe --name foobar --service-id 123 --version 1"),
			WantOutput: "\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nDeleted at: 2021-06-15 23:00:00 +0000 UTC\nFormat: \nFormat Version: 0\nName: foobar\nPlacement: \nRegion: \nResponse Condition: \nService ID: 123\nService Version: 1\nToken: abc\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\n",
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
			Args:      args("logging newrelic list"),
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Args:      args("logging newrelic list --version 3"),
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate ListNewRelics API error",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListNewRelicFn: func(i *fastly.ListNewRelicInput) ([]*fastly.NewRelic, error) {
					return nil, testutil.Err
				},
			},
			Args:      args("logging newrelic list --service-id 123 --version 3"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate ListNewRelics API success",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListNewRelicFn: listNewRelic,
			},
			Args:       args("logging newrelic list --service-id 123 --version 3"),
			WantOutput: "SERVICE ID  VERSION  NAME\n123         3        foo\n123         3        bar\n",
		},
		{
			Name: "validate missing --autoclone flag is OK",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListNewRelicFn: listNewRelic,
			},
			Args:       args("logging newrelic list --service-id 123 --version 1"),
			WantOutput: "SERVICE ID  VERSION  NAME\n123         1        foo\n123         1        bar\n",
		},
		{
			Name: "validate missing --verbose flag",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListNewRelicFn: listNewRelic,
			},
			Args:       args("logging newrelic list --service-id 123 --verbose --version 1"),
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
			Args:      args("logging newrelic update --service-id 123 --version 3"),
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Name:      "validate missing --version flag",
			Args:      args("logging newrelic update --name foobar --service-id 123"),
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Args:      args("logging newrelic update --name foobar --version 3"),
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate missing --autoclone flag",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
			},
			Args:      args("logging newrelic update --name foobar --service-id 123 --version 1"),
			WantError: "service version 1 is not editable",
		},
		{
			Name: "validate UpdateNewRelic API error",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				UpdateNewRelicFn: func(i *fastly.UpdateNewRelicInput) (*fastly.NewRelic, error) {
					return nil, testutil.Err
				},
			},
			Args:      args("logging newrelic update --name foobar --new-name beepboop --service-id 123 --version 3"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate UpdateNewRelic API success",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				UpdateNewRelicFn: func(i *fastly.UpdateNewRelicInput) (*fastly.NewRelic, error) {
					return &fastly.NewRelic{
						Name:           *i.NewName,
						ServiceID:      i.ServiceID,
						ServiceVersion: i.ServiceVersion,
					}, nil
				},
			},
			Args:       args("logging newrelic update --name foobar --new-name beepboop --service-id 123 --version 3"),
			WantOutput: "Updated New Relic logging endpoint 'beepboop' (previously: foobar, service: 123, version: 3)",
		},
		{
			Name: "validate --autoclone results in cloned service version",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				UpdateNewRelicFn: func(i *fastly.UpdateNewRelicInput) (*fastly.NewRelic, error) {
					return &fastly.NewRelic{
						Name:           *i.NewName,
						ServiceID:      i.ServiceID,
						ServiceVersion: i.ServiceVersion,
					}, nil
				},
			},
			Args:       args("logging newrelic update --autoclone --name foobar --new-name beepboop --service-id 123 --version 1"),
			WantOutput: "Updated New Relic logging endpoint 'beepboop' (previously: foobar, service: 123, version: 4)",
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

func getNewRelic(i *fastly.GetNewRelicInput) (*fastly.NewRelic, error) {
	t := testutil.Date

	return &fastly.NewRelic{
		Name:           i.Name,
		Token:          "abc",
		ServiceID:      i.ServiceID,
		ServiceVersion: i.ServiceVersion,

		CreatedAt: &t,
		DeletedAt: &t,
		UpdatedAt: &t,
	}, nil
}

func listNewRelic(i *fastly.ListNewRelicInput) ([]*fastly.NewRelic, error) {
	t := testutil.Date
	vs := []*fastly.NewRelic{
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
