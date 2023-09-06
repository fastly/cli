package acl_test

import (
	"bytes"
	"testing"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestACLCreate(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --name flag",
			Args:      args("acl create --version 3"),
			WantError: "error reading service: no service ID found",
		},
		{
			Name:      "validate missing --version flag",
			Args:      args("acl create --name foo"),
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Args:      args("acl create --name foo --version 3"),
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate missing --autoclone flag",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
			},
			Args:      args("acl create --name foo --service-id 123 --version 1"),
			WantError: "service version 1 is not editable",
		},
		{
			Name: "validate CreateACL API error",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CreateACLFn: func(i *fastly.CreateACLInput) (*fastly.ACL, error) {
					return nil, testutil.Err
				},
			},
			Args:      args("acl create --name foo --service-id 123 --version 3"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate CreateACL API success",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CreateACLFn: func(i *fastly.CreateACLInput) (*fastly.ACL, error) {
					return &fastly.ACL{
						ID:             "456",
						Name:           *i.Name,
						ServiceID:      i.ServiceID,
						ServiceVersion: i.ServiceVersion,
					}, nil
				},
			},
			Args:       args("acl create --name foo --service-id 123 --version 3"),
			WantOutput: "Created ACL 'foo' (id: 456, service: 123, version: 3)",
		},
		{
			Name: "validate --autoclone results in cloned service version",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				CreateACLFn: func(i *fastly.CreateACLInput) (*fastly.ACL, error) {
					return &fastly.ACL{
						ID:             "456",
						Name:           *i.Name,
						ServiceID:      i.ServiceID,
						ServiceVersion: i.ServiceVersion,
					}, nil
				},
			},
			Args:       args("acl create --autoclone --name foo --service-id 123 --version 1"),
			WantOutput: "Created ACL 'foo' (id: 456, service: 123, version: 4)",
		},
	}

	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(testcase.Name, func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.Args, &stdout)
			opts.APIClient = mock.APIClient(testcase.API)
			err := app.Run(opts)
			t.Log(stdout.String())
			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.WantOutput)
		})
	}
}

func TestACLDelete(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --name flag",
			Args:      args("acl delete --version 3"),
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Name:      "validate missing --version flag",
			Args:      args("acl delete --name foobar"),
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Args:      args("acl delete --name foobar --version 3"),
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate missing --autoclone flag",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
			},
			Args:      args("acl delete --name foobar --service-id 123 --version 1"),
			WantError: "service version 1 is not editable",
		},
		{
			Name: "validate DeleteACL API error",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				DeleteACLFn: func(i *fastly.DeleteACLInput) error {
					return testutil.Err
				},
			},
			Args:      args("acl delete --name foobar --service-id 123 --version 3"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate DeleteACL API success",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				DeleteACLFn: func(i *fastly.DeleteACLInput) error {
					return nil
				},
			},
			Args:       args("acl delete --name foobar --service-id 123 --version 3"),
			WantOutput: "Deleted ACL 'foobar' (service: 123, version: 3)",
		},
		{
			Name: "validate --autoclone results in cloned service version",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				DeleteACLFn: func(i *fastly.DeleteACLInput) error {
					return nil
				},
			},
			Args:       args("acl delete --autoclone --name foo --service-id 123 --version 1"),
			WantOutput: "Deleted ACL 'foo' (service: 123, version: 4)",
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

func TestACLDescribe(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --name flag",
			Args:      args("acl describe --version 3"),
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Name:      "validate missing --version flag",
			Args:      args("acl describe --name foobar"),
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Args:      args("acl describe --name foobar --version 3"),
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate GetACL API error",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetACLFn: func(i *fastly.GetACLInput) (*fastly.ACL, error) {
					return nil, testutil.Err
				},
			},
			Args:      args("acl describe --name foobar --service-id 123 --version 3"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate GetACL API success",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetACLFn:       getACL,
			},
			Args:       args("acl describe --name foobar --service-id 123 --version 3"),
			WantOutput: "\nService ID: 123\nService Version: 3\n\nName: foobar\nID: 456\n\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\nDeleted at: 2021-06-15 23:00:00 +0000 UTC\n",
		},
		{
			Name: "validate missing --autoclone flag is OK",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetACLFn:       getACL,
			},
			Args:       args("acl describe --name foobar --service-id 123 --version 1"),
			WantOutput: "\nService ID: 123\nService Version: 1\n\nName: foobar\nID: 456\n\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\nDeleted at: 2021-06-15 23:00:00 +0000 UTC\n",
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

func TestACLList(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --version flag",
			Args:      args("acl list"),
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Args:      args("acl list --version 3"),
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate ListACLs API error",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListACLsFn: func(i *fastly.ListACLsInput) ([]*fastly.ACL, error) {
					return nil, testutil.Err
				},
			},
			Args:      args("acl list --service-id 123 --version 3"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate ListACLs API success",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListACLsFn:     listACLs,
			},
			Args:       args("acl list --service-id 123 --version 3"),
			WantOutput: "SERVICE ID  VERSION  NAME  ID\n123         3        foo   456\n123         3        bar   789\n",
		},
		{
			Name: "validate missing --autoclone flag is OK",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListACLsFn:     listACLs,
			},
			Args:       args("acl list --service-id 123 --version 1"),
			WantOutput: "SERVICE ID  VERSION  NAME  ID\n123         1        foo   456\n123         1        bar   789\n",
		},
		{
			Name: "validate --verbose flag",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListACLsFn:     listACLs,
			},
			Args:       args("acl list --service-id 123 --verbose --version 1"),
			WantOutput: "Fastly API token provided via config file (profile: user)\nFastly API endpoint: https://api.fastly.com\n\nService ID (via --service-id): 123\n\nService Version: 1\n\nName: foo\nID: 456\n\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\nDeleted at: 2021-06-15 23:00:00 +0000 UTC\n\nName: bar\nID: 789\n\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\nDeleted at: 2021-06-15 23:00:00 +0000 UTC\n\n",
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

func TestACLUpdate(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --name flag",
			Args:      args("acl update --new-name beepboop --version 3"),
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Name:      "validate missing --new-name flag",
			Args:      args("acl update --name foobar --version 3"),
			WantError: "error parsing arguments: required flag --new-name not provided",
		},
		{
			Name:      "validate missing --version flag",
			Args:      args("acl update --name foobar --new-name beepboop"),
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Args:      args("acl update --name foobar --new-name beepboop --version 3"),
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate missing --autoclone flag",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
			},
			Args:      args("acl update --name foobar --new-name beepboop --service-id 123 --version 1"),
			WantError: "service version 1 is not editable",
		},
		{
			Name: "validate UpdateACL API error",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				UpdateACLFn: func(i *fastly.UpdateACLInput) (*fastly.ACL, error) {
					return nil, testutil.Err
				},
			},
			Args:      args("acl update --name foobar --new-name beepboop --service-id 123 --version 3"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate UpdateACL API success",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				UpdateACLFn: func(i *fastly.UpdateACLInput) (*fastly.ACL, error) {
					return &fastly.ACL{
						ID:             "456",
						Name:           *i.NewName,
						ServiceID:      i.ServiceID,
						ServiceVersion: i.ServiceVersion,
					}, nil
				},
			},
			Args:       args("acl update --name foobar --new-name beepboop --service-id 123 --version 3"),
			WantOutput: "Updated ACL 'beepboop' (previously: 'foobar', service: 123, version: 3)",
		},
		{
			Name: "validate --autoclone results in cloned service version",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				UpdateACLFn: func(i *fastly.UpdateACLInput) (*fastly.ACL, error) {
					return &fastly.ACL{
						ID:             "456",
						Name:           *i.NewName,
						ServiceID:      i.ServiceID,
						ServiceVersion: i.ServiceVersion,
					}, nil
				},
			},
			Args:       args("acl update --autoclone --name foobar --new-name beepboop --service-id 123 --version 1"),
			WantOutput: "Updated ACL 'beepboop' (previously: 'foobar', service: 123, version: 4)",
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

func getACL(i *fastly.GetACLInput) (*fastly.ACL, error) {
	t := testutil.Date

	return &fastly.ACL{
		ID:             "456",
		Name:           i.Name,
		ServiceID:      i.ServiceID,
		ServiceVersion: i.ServiceVersion,

		CreatedAt: &t,
		DeletedAt: &t,
		UpdatedAt: &t,
	}, nil
}

func listACLs(i *fastly.ListACLsInput) ([]*fastly.ACL, error) {
	t := testutil.Date
	vs := []*fastly.ACL{
		{
			ID:             "456",
			Name:           "foo",
			ServiceID:      i.ServiceID,
			ServiceVersion: i.ServiceVersion,

			CreatedAt: &t,
			DeletedAt: &t,
			UpdatedAt: &t,
		},
		{
			ID:             "789",
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
