package snippet_test

import (
	"bytes"
	"testing"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v3/fastly"
)

func TestVCLSnippetCreate(t *testing.T) {
	var content string
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --content flag",
			Args:      args("vcl snippet create --name foo --type recv --version 3"),
			WantError: "error parsing arguments: required flag --content not provided",
		},
		{
			Name:      "validate missing --name flag",
			Args:      args("vcl snippet create --content /path/to/snippet.vcl --type recv --version 3"),
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Name:      "validate missing --type flag",
			Args:      args("vcl snippet create --content /path/to/snippet.vcl --name foo --version 3"),
			WantError: "error parsing arguments: required flag --type not provided",
		},
		{
			Name:      "validate missing --version flag",
			Args:      args("vcl snippet create --content /path/to/snippet.vcl --name foo --type recv"),
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Args:      args("vcl snippet create --content /path/to/snippet.vcl --name foo --type recv --version 3"),
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate missing --autoclone flag",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
			},
			Args:      args("vcl snippet create --content ./testdata/snippet.vcl --name foo --type recv --service-id 123 --version 1"),
			WantError: "service version 1 is not editable",
		},
		{
			Name: "validate CreateSnippet API error",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CreateSnippetFn: func(i *fastly.CreateSnippetInput) (*fastly.Snippet, error) {
					return nil, testutil.ErrAPI
				},
			},
			Args:      args("vcl snippet create --content ./testdata/snippet.vcl --name foo --type recv --service-id 123 --version 3"),
			WantError: testutil.ErrAPI.Error(),
		},
		{
			Name: "validate CreateSnippet API success for versioned Snippet",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CreateSnippetFn: func(i *fastly.CreateSnippetInput) (*fastly.Snippet, error) {
					// Track the contents parsed
					content = i.Content

					return &fastly.Snippet{
						Content:        i.Content,
						Dynamic:        i.Dynamic,
						Name:           i.Name,
						Priority:       i.Priority,
						ServiceID:      i.ServiceID,
						ServiceVersion: i.ServiceVersion,
					}, nil
				},
			},
			Args:       args("vcl snippet create --content ./testdata/snippet.vcl --name foo --service-id 123 --type recv --version 3"),
			WantOutput: "Created VCL snippet 'foo' (service: 123, version: 3, dynamic: false, type: recv, priority: 0)",
		},
		{
			Name: "validate CreateSnippet API success for dynamic Snippet",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CreateSnippetFn: func(i *fastly.CreateSnippetInput) (*fastly.Snippet, error) {
					// Track the contents parsed
					content = i.Content

					return &fastly.Snippet{
						Content:        i.Content,
						Dynamic:        i.Dynamic,
						Name:           i.Name,
						Priority:       i.Priority,
						ServiceID:      i.ServiceID,
						ServiceVersion: i.ServiceVersion,
					}, nil
				},
			},
			Args:       args("vcl snippet create --content ./testdata/snippet.vcl --dynamic --name foo --service-id 123 --type recv --version 3"),
			WantOutput: "Created VCL snippet 'foo' (service: 123, version: 3, dynamic: true, type: recv, priority: 0)",
		},
		{
			Name: "validate Priority set",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CreateSnippetFn: func(i *fastly.CreateSnippetInput) (*fastly.Snippet, error) {
					// Track the contents parsed
					content = i.Content

					return &fastly.Snippet{
						Content:        i.Content,
						Dynamic:        i.Dynamic,
						Name:           i.Name,
						Priority:       i.Priority,
						ServiceID:      i.ServiceID,
						ServiceVersion: i.ServiceVersion,
					}, nil
				},
			},
			Args:       args("vcl snippet create --content ./testdata/snippet.vcl --name foo --priority 1 --service-id 123 --type recv --version 3"),
			WantOutput: "Created VCL snippet 'foo' (service: 123, version: 3, dynamic: false, type: recv, priority: 1)",
		},
		{
			Name: "validate --autoclone results in cloned service version",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				CreateSnippetFn: func(i *fastly.CreateSnippetInput) (*fastly.Snippet, error) {
					// Track the contents parsed
					content = i.Content

					return &fastly.Snippet{
						Content:        i.Content,
						Dynamic:        i.Dynamic,
						Name:           i.Name,
						Priority:       i.Priority,
						ServiceID:      i.ServiceID,
						ServiceVersion: i.ServiceVersion,
					}, nil
				},
			},
			Args:       args("vcl snippet create --autoclone --content ./testdata/snippet.vcl --name foo --service-id 123 --type recv --version 1"),
			WantOutput: "Created VCL snippet 'foo' (service: 123, version: 4, dynamic: false, type: recv, priority: 0)",
		},
		{
			Name: "validate CreateSnippet API success with inline Snippet content",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CreateSnippetFn: func(i *fastly.CreateSnippetInput) (*fastly.Snippet, error) {
					// Track the contents parsed
					content = i.Content

					return &fastly.Snippet{
						Content:        i.Content,
						Dynamic:        i.Dynamic,
						Name:           i.Name,
						Priority:       i.Priority,
						ServiceID:      i.ServiceID,
						ServiceVersion: i.ServiceVersion,
					}, nil
				},
			},
			Args:       args("vcl snippet create --content inline_vcl --name foo --service-id 123 --type recv --version 3"),
			WantOutput: "Created VCL snippet 'foo' (service: 123, version: 3, dynamic: false, type: recv, priority: 0)",
		},
	}

	for _, testcase := range scenarios {
		t.Run(testcase.Name, func(t *testing.T) {
			var buf bytes.Buffer
			ara := testutil.NewAppRunArgs(testcase.Args, testcase.API, &buf)
			err := app.Run(ara.Args, ara.Env, ara.File, ara.AppConfigFile, ara.ClientFactory, ara.HTTPClient, ara.CLIVersioner, ara.In, ara.Out)
			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertStringContains(t, buf.String(), testcase.WantOutput)
			testutil.AssertContentFlag(testcase.WantError, testcase.Args, "snippet.vcl", content, t)
		})
	}
}

func TestVCLSnippetDelete(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --name flag",
			Args:      args("vcl snippet delete --version 3"),
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Name:      "validate missing --version flag",
			Args:      args("vcl snippet delete --name foobar"),
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Args:      args("vcl snippet delete --name foobar --version 3"),
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate missing --autoclone flag",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
			},
			Args:      args("vcl snippet delete --name foobar --service-id 123 --version 1"),
			WantError: "service version 1 is not editable",
		},
		{
			Name: "validate DeleteSnippet API error",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				DeleteSnippetFn: func(i *fastly.DeleteSnippetInput) error {
					return testutil.ErrAPI
				},
			},
			Args:      args("vcl snippet delete --name foobar --service-id 123 --version 3"),
			WantError: testutil.ErrAPI.Error(),
		},
		{
			Name: "validate DeleteSnippet API success",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				DeleteSnippetFn: func(i *fastly.DeleteSnippetInput) error {
					return nil
				},
			},
			Args:       args("vcl snippet delete --name foobar --service-id 123 --version 3"),
			WantOutput: "Deleted VCL snippet 'foobar' (service: 123, version: 3)",
		},
		{
			Name: "validate --autoclone results in cloned service version",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				DeleteSnippetFn: func(i *fastly.DeleteSnippetInput) error {
					return nil
				},
			},
			Args:       args("vcl snippet delete --autoclone --name foo --service-id 123 --version 1"),
			WantOutput: "Deleted VCL snippet 'foo' (service: 123, version: 4)",
		},
	}

	for _, testcase := range scenarios {
		t.Run(testcase.Name, func(t *testing.T) {
			var buf bytes.Buffer
			ara := testutil.NewAppRunArgs(testcase.Args, testcase.API, &buf)
			err := app.Run(ara.Args, ara.Env, ara.File, ara.AppConfigFile, ara.ClientFactory, ara.HTTPClient, ara.CLIVersioner, ara.In, ara.Out)
			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertStringContains(t, buf.String(), testcase.WantOutput)
		})
	}
}

func TestVCLSnippetDescribe(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --name flag",
			Args:      args("vcl snippet describe --snippet-id 123 --version 3"),
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Name:      "validate missing --snippet-id flag",
			Args:      args("vcl snippet describe --name foobar --version 3"),
			WantError: "error parsing arguments: required flag --snippet-id not provided",
		},
		{
			Name:      "validate missing --version flag",
			Args:      args("vcl snippet describe --name foobar --snippet-id 123"),
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Args:      args("vcl snippet describe --name foobar --snippet-id 123 --version 3"),
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate GetSnippet API error",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetSnippetFn: func(i *fastly.GetSnippetInput) (*fastly.Snippet, error) {
					return nil, testutil.ErrAPI
				},
			},
			Args:      args("vcl snippet describe --name foobar --service-id 123 --snippet-id 456 --version 3"),
			WantError: testutil.ErrAPI.Error(),
		},
		{
			Name: "validate GetSnippet API success",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetSnippetFn:   getSnippet,
			},
			Args:       args("vcl snippet describe --name foobar --service-id 123 --snippet-id 456 --version 3"),
			WantOutput: "Service ID: 123\nService Version: 3\nName: foobar\nID: 456\nPriority: 0\nDynamic: false\nType: recv\nContent: # some vcl content\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\nDeleted at: 2021-06-15 23:00:00 +0000 UTC\n",
		},
		{
			Name: "validate missing --autoclone flag is OK",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetSnippetFn:   getSnippet,
			},
			Args:       args("vcl snippet describe --name foobar --service-id 123 --snippet-id 456 --version 1"),
			WantOutput: "Service ID: 123\nService Version: 1\nName: foobar\nID: 456\nPriority: 0\nDynamic: false\nType: recv\nContent: # some vcl content\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\nDeleted at: 2021-06-15 23:00:00 +0000 UTC\n",
		},
		{
			Name: "validate dynamic GetSnippet API success",
			API: mock.API{
				ListVersionsFn:      testutil.ListVersions,
				GetDynamicSnippetFn: getDynamicSnippet,
			},
			Args:       args("vcl snippet describe --dynamic --name foobar --service-id 123 --snippet-id 456 --version 3"),
			WantOutput: "Service ID: 123\nID: 456\nContent: # some vcl content\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\n",
		},
	}

	for _, testcase := range scenarios {
		t.Run(testcase.Name, func(t *testing.T) {
			var buf bytes.Buffer
			ara := testutil.NewAppRunArgs(testcase.Args, testcase.API, &buf)
			err := app.Run(ara.Args, ara.Env, ara.File, ara.AppConfigFile, ara.ClientFactory, ara.HTTPClient, ara.CLIVersioner, ara.In, ara.Out)
			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertStringContains(t, buf.String(), testcase.WantOutput)
		})
	}
}

func TestVCLSnippetList(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --version flag",
			Args:      args("vcl snippet list"),
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Args:      args("vcl snippet list --version 3"),
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate ListSnippets API error",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListSnippetsFn: func(i *fastly.ListSnippetsInput) ([]*fastly.Snippet, error) {
					return nil, testutil.ErrAPI
				},
			},
			Args:      args("vcl snippet list --service-id 123 --version 3"),
			WantError: testutil.ErrAPI.Error(),
		},
		{
			Name: "validate ListSnippets API success",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListSnippetsFn: listSnippets,
			},
			Args:       args("vcl snippet list --service-id 123 --version 3"),
			WantOutput: "SERVICE ID  VERSION  NAME  DYNAMIC\n123         3        foo   true\n123         3        bar   false\n",
		},
		{
			Name: "validate missing --autoclone flag is OK",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListSnippetsFn: listSnippets,
			},
			Args:       args("vcl snippet list --service-id 123 --version 1"),
			WantOutput: "SERVICE ID  VERSION  NAME  DYNAMIC\n123         1        foo   true\n123         1        bar   false\n",
		},
		{
			Name: "validate missing --verbose flag",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListSnippetsFn: listSnippets,
			},
			Args:       args("vcl snippet list --service-id 123 --verbose --version 1"),
			WantOutput: "Fastly API token not provided\nFastly API endpoint: https://api.fastly.com\nService ID: 123\nService Version: 1\nName: foo\nID: abc\nPriority: 0\nDynamic: true\nType: recv\nContent: # some vcl content\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\nDeleted at: 2021-06-15 23:00:00 +0000 UTC\nName: bar\nID: abc\nPriority: 0\nDynamic: false\nType: recv\nContent: # some vcl content\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\nDeleted at: 2021-06-15 23:00:00 +0000 UTC\n",
		},
	}

	for _, testcase := range scenarios {
		t.Run(testcase.Name, func(t *testing.T) {
			var buf bytes.Buffer
			ara := testutil.NewAppRunArgs(testcase.Args, testcase.API, &buf)
			err := app.Run(ara.Args, ara.Env, ara.File, ara.AppConfigFile, ara.ClientFactory, ara.HTTPClient, ara.CLIVersioner, ara.In, ara.Out)
			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertStringContains(t, buf.String(), testcase.WantOutput)
		})
	}
}

func TestVCLSnippetUpdate(t *testing.T) {
	var content string
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --name flag",
			Args:      args("vcl snippet update --version 3"),
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Name:      "validate missing --version flag",
			Args:      args("vcl snippet update --name foobar"),
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Args:      args("vcl snippet update --name foobar --version 3"),
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate missing --autoclone flag",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
			},
			Args:      args("vcl snippet update --name foobar --service-id 123 --version 1"),
			WantError: "service version 1 is not editable",
		},
		{
			Name: "validate missing either --new-name or --content",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
			},
			Args:      args("vcl snippet update --name foobar --service-id 123 --version 3"),
			WantError: "error parsing arguments: must provide either --new-name or --content to update the VCL snippet",
		},
		{
			Name: "validate missing --snippet-id when updating dynamic snippet",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
			},
			Args:      args("vcl snippet update --dynamic --name foobar --service-id 123 --version 3"),
			WantError: "error parsing arguments: must provide --snippet-id to update a dynamic VCL snippet",
		},
		{
			Name: "validate UpdateSnippet API error",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				UpdateSnippetFn: func(i *fastly.UpdateSnippetInput) (*fastly.Snippet, error) {
					return nil, testutil.ErrAPI
				},
			},
			Args:      args("vcl snippet update --name foobar --new-name beepboop --service-id 123 --version 3"),
			WantError: testutil.ErrAPI.Error(),
		},
		{
			Name: "validate UpdateSnippet API success with --new-name",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				UpdateSnippetFn: func(i *fastly.UpdateSnippetInput) (*fastly.Snippet, error) {
					return &fastly.Snippet{
						Content:        "# untouched",
						Dynamic:        i.Dynamic,
						Name:           i.NewName,
						ServiceID:      i.ServiceID,
						ServiceVersion: i.ServiceVersion,
					}, nil
				},
			},
			Args:       args("vcl snippet update --name foobar --new-name beepboop --service-id 123 --version 3"),
			WantOutput: "Updated VCL snippet 'beepboop' (previously: 'foobar', service: 123, version: 3)",
		},
		{
			Name: "validate UpdateSnippet API success with --content",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				UpdateSnippetFn: func(i *fastly.UpdateSnippetInput) (*fastly.Snippet, error) {
					// Track the contents parsed
					content = i.Content

					return &fastly.Snippet{
						Content:        i.Content,
						Dynamic:        i.Dynamic,
						Name:           i.Name,
						ServiceID:      i.ServiceID,
						ServiceVersion: i.ServiceVersion,
					}, nil
				},
			},
			Args:       args("vcl snippet update --content updated --name foobar --service-id 123 --version 3"),
			WantOutput: "Updated VCL snippet 'foobar' (service: 123, version: 3)",
		},
		{
			Name: "validate UpdateDynamicSnippet API success with --content",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				UpdateDynamicSnippetFn: func(i *fastly.UpdateDynamicSnippetInput) (*fastly.DynamicSnippet, error) {
					// Track the contents parsed
					content = i.Content

					return &fastly.DynamicSnippet{
						Content:   i.Content,
						ID:        i.ID,
						ServiceID: i.ServiceID,
					}, nil
				},
			},
			Args:       args("vcl snippet update --content updated --dynamic --name foobar --service-id 123 --snippet-id foobar --version 3"),
			WantOutput: "Updated dynamic VCL snippet 'foobar' (service: 123)",
		},
		{
			Name: "validate --autoclone results in cloned service version",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				UpdateSnippetFn: func(i *fastly.UpdateSnippetInput) (*fastly.Snippet, error) {
					// Track the contents parsed
					content = i.Content

					return &fastly.Snippet{
						Content:        i.Content,
						Dynamic:        i.Dynamic,
						Name:           i.Name,
						ServiceID:      i.ServiceID,
						ServiceVersion: i.ServiceVersion,
					}, nil
				},
			},
			Args:       args("vcl snippet update --autoclone --content ./testdata/snippet.vcl --name foo --service-id 123 --version 1"),
			WantOutput: "Updated VCL snippet 'foo' (service: 123, version: 4)",
		},
	}

	for _, testcase := range scenarios {
		t.Run(testcase.Name, func(t *testing.T) {
			var buf bytes.Buffer
			ara := testutil.NewAppRunArgs(testcase.Args, testcase.API, &buf)
			err := app.Run(ara.Args, ara.Env, ara.File, ara.AppConfigFile, ara.ClientFactory, ara.HTTPClient, ara.CLIVersioner, ara.In, ara.Out)
			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertStringContains(t, buf.String(), testcase.WantOutput)
			testutil.AssertContentFlag(testcase.WantError, testcase.Args, "snippet.vcl", content, t)
		})
	}
}

func getSnippet(i *fastly.GetSnippetInput) (*fastly.Snippet, error) {
	t := testutil.Date

	return &fastly.Snippet{
		Content:        "# some vcl content",
		Dynamic:        0,
		ID:             "456",
		Name:           i.Name,
		Priority:       0,
		ServiceID:      i.ServiceID,
		ServiceVersion: i.ServiceVersion,
		Type:           "recv",

		CreatedAt: &t,
		DeletedAt: &t,
		UpdatedAt: &t,
	}, nil
}

func getDynamicSnippet(i *fastly.GetDynamicSnippetInput) (*fastly.DynamicSnippet, error) {
	t := testutil.Date

	return &fastly.DynamicSnippet{
		Content:   "# some vcl content",
		ID:        i.ID,
		ServiceID: i.ServiceID,

		CreatedAt: &t,
		UpdatedAt: &t,
	}, nil
}

func listSnippets(i *fastly.ListSnippetsInput) ([]*fastly.Snippet, error) {
	t := testutil.Date
	vs := []*fastly.Snippet{
		{
			Content:        "# some vcl content",
			Dynamic:        1,
			ID:             "abc",
			Name:           "foo",
			Priority:       0,
			ServiceID:      i.ServiceID,
			ServiceVersion: i.ServiceVersion,
			Type:           "recv",

			CreatedAt: &t,
			DeletedAt: &t,
			UpdatedAt: &t,
		},
		{
			Content:        "# some vcl content",
			Dynamic:        0,
			ID:             "abc",
			Name:           "bar",
			Priority:       0,
			ServiceID:      i.ServiceID,
			ServiceVersion: i.ServiceVersion,
			Type:           "recv",

			CreatedAt: &t,
			DeletedAt: &t,
			UpdatedAt: &t,
		},
	}
	return vs, nil
}
