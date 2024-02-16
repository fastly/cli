package snippet_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/v10/pkg/app"
	"github.com/fastly/cli/v10/pkg/global"
	"github.com/fastly/cli/v10/pkg/mock"
	"github.com/fastly/cli/v10/pkg/testutil"
)

func TestVCLSnippetCreate(t *testing.T) {
	var content string
	args := testutil.Args
	scenarios := []testutil.TestScenario{
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
					return nil, testutil.Err
				},
			},
			Args:      args("vcl snippet create --content ./testdata/snippet.vcl --name foo --type recv --service-id 123 --version 3"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate CreateSnippet API success for versioned Snippet",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CreateSnippetFn: func(i *fastly.CreateSnippetInput) (*fastly.Snippet, error) {
					// Track the contents parsed
					content = *i.Content
					if i.Content == nil {
						i.Content = fastly.ToPointer("")
					}
					if i.Dynamic == nil {
						i.Dynamic = fastly.ToPointer(0)
					}
					if i.Name == nil {
						i.Name = fastly.ToPointer("")
					}
					return &fastly.Snippet{
						Content:        i.Content,
						Dynamic:        i.Dynamic,
						Name:           i.Name,
						ServiceID:      fastly.ToPointer(i.ServiceID),
						ServiceVersion: fastly.ToPointer(i.ServiceVersion),
						SnippetID:      fastly.ToPointer("123"),
					}, nil
				},
			},
			Args:       args("vcl snippet create --content ./testdata/snippet.vcl --name foo --service-id 123 --type recv --version 3"),
			WantOutput: "Created VCL snippet 'foo' (service: 123, version: 3, dynamic: false, snippet id: 123, type: recv, priority: 0)",
		},
		{
			Name: "validate CreateSnippet API success for dynamic Snippet",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CreateSnippetFn: func(i *fastly.CreateSnippetInput) (*fastly.Snippet, error) {
					// Track the contents parsed
					content = *i.Content
					if i.Content == nil {
						i.Content = fastly.ToPointer("")
					}
					if i.Dynamic == nil {
						i.Dynamic = fastly.ToPointer(0)
					}
					if i.Name == nil {
						i.Name = fastly.ToPointer("")
					}
					return &fastly.Snippet{
						Content:        i.Content,
						Dynamic:        i.Dynamic,
						Name:           i.Name,
						ServiceID:      fastly.ToPointer(i.ServiceID),
						ServiceVersion: fastly.ToPointer(i.ServiceVersion),
						SnippetID:      fastly.ToPointer("123"),
					}, nil
				},
			},
			Args:       args("vcl snippet create --content ./testdata/snippet.vcl --dynamic --name foo --service-id 123 --type recv --version 3"),
			WantOutput: "Created VCL snippet 'foo' (service: 123, version: 3, dynamic: true, snippet id: 123, type: recv, priority: 0)",
		},
		{
			Name: "validate Priority set",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CreateSnippetFn: func(i *fastly.CreateSnippetInput) (*fastly.Snippet, error) {
					// Track the contents parsed
					content = *i.Content
					if i.Content == nil {
						i.Content = fastly.ToPointer("")
					}
					if i.Dynamic == nil {
						i.Dynamic = fastly.ToPointer(0)
					}
					if i.Name == nil {
						i.Name = fastly.ToPointer("")
					}
					return &fastly.Snippet{
						Content:        i.Content,
						Dynamic:        i.Dynamic,
						Name:           i.Name,
						Priority:       i.Priority,
						ServiceID:      fastly.ToPointer(i.ServiceID),
						ServiceVersion: fastly.ToPointer(i.ServiceVersion),
						SnippetID:      fastly.ToPointer("123"),
					}, nil
				},
			},
			Args:       args("vcl snippet create --content ./testdata/snippet.vcl --name foo --priority 1 --service-id 123 --type recv --version 3"),
			WantOutput: "Created VCL snippet 'foo' (service: 123, version: 3, dynamic: false, snippet id: 123, type: recv, priority: 1)",
		},
		{
			Name: "validate --autoclone results in cloned service version",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				CreateSnippetFn: func(i *fastly.CreateSnippetInput) (*fastly.Snippet, error) {
					// Track the contents parsed
					content = *i.Content
					if i.Content == nil {
						i.Content = fastly.ToPointer("")
					}
					if i.Dynamic == nil {
						i.Dynamic = fastly.ToPointer(0)
					}
					if i.Name == nil {
						i.Name = fastly.ToPointer("")
					}
					return &fastly.Snippet{
						Content:        i.Content,
						Dynamic:        i.Dynamic,
						Name:           i.Name,
						ServiceID:      fastly.ToPointer(i.ServiceID),
						ServiceVersion: fastly.ToPointer(i.ServiceVersion),
						SnippetID:      fastly.ToPointer("123"),
					}, nil
				},
			},
			Args:       args("vcl snippet create --autoclone --content ./testdata/snippet.vcl --name foo --service-id 123 --type recv --version 1"),
			WantOutput: "Created VCL snippet 'foo' (service: 123, version: 4, dynamic: false, snippet id: 123, type: recv, priority: 0)",
		},
		{
			Name: "validate CreateSnippet API success with inline Snippet content",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CreateSnippetFn: func(i *fastly.CreateSnippetInput) (*fastly.Snippet, error) {
					// Track the contents parsed
					content = *i.Content
					if i.Content == nil {
						i.Content = fastly.ToPointer("")
					}
					if i.Dynamic == nil {
						i.Dynamic = fastly.ToPointer(0)
					}
					if i.Name == nil {
						i.Name = fastly.ToPointer("")
					}
					return &fastly.Snippet{
						Content:        i.Content,
						Dynamic:        i.Dynamic,
						Name:           i.Name,
						ServiceID:      fastly.ToPointer(i.ServiceID),
						ServiceVersion: fastly.ToPointer(i.ServiceVersion),
						SnippetID:      fastly.ToPointer("123"),
					}, nil
				},
			},
			Args:       args("vcl snippet create --content inline_vcl --name foo --service-id 123 --type recv --version 3"),
			WantOutput: "Created VCL snippet 'foo' (service: 123, version: 3, dynamic: false, snippet id: 123, type: recv, priority: 0)",
		},
	}

	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(testcase.Name, func(t *testing.T) {
			var stdout bytes.Buffer
			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts := testutil.MockGlobalData(testcase.Args, &stdout)
				opts.APIClientFactory = mock.APIClient(testcase.API)
				return opts, nil
			}
			err := app.Run(testcase.Args, nil)
			t.Log(stdout.String())
			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertStringContains(t, testutil.StripNewLines(stdout.String()), testcase.WantOutput)
			testutil.AssertPathContentFlag("content", testcase.WantError, testcase.Args, "snippet.vcl", content, t)
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
					return testutil.Err
				},
			},
			Args:      args("vcl snippet delete --name foobar --service-id 123 --version 3"),
			WantError: testutil.Err.Error(),
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

	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(testcase.Name, func(t *testing.T) {
			var stdout bytes.Buffer
			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts := testutil.MockGlobalData(testcase.Args, &stdout)
				opts.APIClientFactory = mock.APIClient(testcase.API)
				return opts, nil
			}
			err := app.Run(testcase.Args, nil)
			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.WantOutput)
		})
	}
}

func TestVCLSnippetDescribe(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --version flag",
			Args:      args("vcl snippet describe"),
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Args:      args("vcl snippet describe --version 3"),
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate missing --name flag with versioned snippet",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
			},
			Args:      args("vcl snippet describe --service-id 123 --version 3"),
			WantError: "error parsing arguments: must provide --name with a versioned VCL snippet",
		},
		{
			Name: "validate missing --snippet-id flag with dynamic snippet",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
			},
			Args:      args("vcl snippet describe --dynamic --service-id 123 --version 3"),
			WantError: "error parsing arguments: must provide --snippet-id with a dynamic VCL snippet",
		},
		{
			Name: "validate GetSnippet API error",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetSnippetFn: func(i *fastly.GetSnippetInput) (*fastly.Snippet, error) {
					return nil, testutil.Err
				},
			},
			Args:      args("vcl snippet describe --name foobar --service-id 123 --version 3"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate GetSnippet API success",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetSnippetFn:   getSnippet,
			},
			Args:       args("vcl snippet describe --name foobar --service-id 123 --version 3"),
			WantOutput: "\nService ID: 123\nService Version: 3\n\nName: foobar\nID: 456\nPriority: 0\nDynamic: false\nType: recv\nContent: \n# some vcl content\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\nDeleted at: 2021-06-15 23:00:00 +0000 UTC\n",
		},
		{
			Name: "validate missing --autoclone flag is OK",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetSnippetFn:   getSnippet,
			},
			Args:       args("vcl snippet describe --name foobar --service-id 123 --version 1"),
			WantOutput: "\nService ID: 123\nService Version: 1\n\nName: foobar\nID: 456\nPriority: 0\nDynamic: false\nType: recv\nContent: \n# some vcl content\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\nDeleted at: 2021-06-15 23:00:00 +0000 UTC\n",
		},
		{
			Name: "validate dynamic GetSnippet API success",
			API: mock.API{
				ListVersionsFn:      testutil.ListVersions,
				GetDynamicSnippetFn: getDynamicSnippet,
			},
			Args:       args("vcl snippet describe --dynamic --service-id 123 --snippet-id 456 --version 3"),
			WantOutput: "\nService ID: 123\nID: 456\nContent: \n# some vcl content\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\n",
		},
	}

	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(testcase.Name, func(t *testing.T) {
			var stdout bytes.Buffer
			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts := testutil.MockGlobalData(testcase.Args, &stdout)
				opts.APIClientFactory = mock.APIClient(testcase.API)
				return opts, nil
			}
			err := app.Run(testcase.Args, nil)
			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.WantOutput)
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
					return nil, testutil.Err
				},
			},
			Args:      args("vcl snippet list --service-id 123 --version 3"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate ListSnippets API success",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListSnippetsFn: listSnippets,
			},
			Args:       args("vcl snippet list --service-id 123 --version 3"),
			WantOutput: "SERVICE ID  VERSION  NAME  DYNAMIC  SNIPPET ID\n123         3        foo   true     abc\n123         3        bar   false    abc\n",
		},
		{
			Name: "validate missing --autoclone flag is OK",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListSnippetsFn: listSnippets,
			},
			Args:       args("vcl snippet list --service-id 123 --version 1"),
			WantOutput: "SERVICE ID  VERSION  NAME  DYNAMIC  SNIPPET ID\n123         1        foo   true     abc\n123         1        bar   false    abc\n",
		},
		{
			Name: "validate missing --verbose flag",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListSnippetsFn: listSnippets,
			},
			Args:       args("vcl snippet list --service-id 123 --verbose --version 1"),
			WantOutput: "Fastly API endpoint: https://api.fastly.com\nFastly API token provided via config file (profile: user)\n\nService ID (via --service-id): 123\n\nService Version: 1\n\nName: foo\nID: abc\nPriority: 0\nDynamic: true\nType: recv\nContent: \n# some vcl content\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\nDeleted at: 2021-06-15 23:00:00 +0000 UTC\n\nName: bar\nID: abc\nPriority: 0\nDynamic: false\nType: recv\nContent: \n# some vcl content\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\nDeleted at: 2021-06-15 23:00:00 +0000 UTC\n",
		},
	}

	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(testcase.Name, func(t *testing.T) {
			var stdout bytes.Buffer
			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts := testutil.MockGlobalData(testcase.Args, &stdout)
				opts.APIClientFactory = mock.APIClient(testcase.API)
				return opts, nil
			}
			err := app.Run(testcase.Args, nil)
			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.WantOutput)
		})
	}
}

func TestVCLSnippetUpdate(t *testing.T) {
	var content string
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --version flag",
			Args:      args("vcl snippet update"),
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Args:      args("vcl snippet update --version 3"),
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate missing --autoclone flag",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
			},
			Args:      args("vcl snippet update --service-id 123 --version 1"),
			WantError: "service version 1 is not editable",
		},
		{
			Name: "validate versioned snippet missing --name",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
			},
			Args:      args("vcl snippet update --content inline_vcl --new-name bar --service-id 123 --type recv --version 3"),
			WantError: "error parsing arguments: must provide --name to update a versioned VCL snippet",
		},
		{
			Name: "validate dynamic snippet missing --snippet-id",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
			},
			Args:      args("vcl snippet update --content inline_vcl --dynamic --service-id 123 --version 3"),
			WantError: "error parsing arguments: must provide --snippet-id to update a dynamic VCL snippet",
		},
		{
			Name: "validate versioned snippet with --snippet-id is not allowed",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
			},
			Args:      args("vcl snippet update --content inline_vcl --new-name foobar --service-id 123 --snippet-id 456 --version 3"),
			WantError: "error parsing arguments: --snippet-id is not supported when updating a versioned VCL snippet",
		},
		{
			Name: "validate dynamic snippet with --new-name is not allowed",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
			},
			Args:      args("vcl snippet update --content inline_vcl --dynamic --new-name foobar --service-id 123 --snippet-id 456 --version 3"),
			WantError: "error parsing arguments: --new-name is not supported when updating a dynamic VCL snippet",
		},
		{
			Name: "validate UpdateSnippet API error",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				UpdateSnippetFn: func(i *fastly.UpdateSnippetInput) (*fastly.Snippet, error) {
					return nil, testutil.Err
				},
			},
			Args:      args("vcl snippet update --content inline_vcl --name foo --new-name bar --service-id 123 --type recv --version 3"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate UpdateSnippet API success",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				UpdateSnippetFn: func(i *fastly.UpdateSnippetInput) (*fastly.Snippet, error) {
					// Track the contents parsed
					content = *i.Content

					return &fastly.Snippet{
						Content:        i.Content,
						Name:           i.NewName,
						Priority:       fastly.ToPointer(100),
						ServiceID:      fastly.ToPointer(i.ServiceID),
						ServiceVersion: fastly.ToPointer(i.ServiceVersion),
						Type:           i.Type,
					}, nil
				},
			},
			Args:       args("vcl snippet update --content inline_vcl --name foo --new-name bar --service-id 123 --type recv --version 3"),
			WantOutput: "Updated VCL snippet 'bar' (previously: 'foo', service: 123, version: 3, type: recv, priority: 100)",
		},
		{
			Name: "validate UpdateDynamicSnippet API success",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				UpdateDynamicSnippetFn: func(i *fastly.UpdateDynamicSnippetInput) (*fastly.DynamicSnippet, error) {
					// Track the contents parsed
					content = *i.Content

					return &fastly.DynamicSnippet{
						Content:   i.Content,
						SnippetID: fastly.ToPointer(i.SnippetID),
						ServiceID: fastly.ToPointer(i.ServiceID),
					}, nil
				},
			},
			Args:       args("vcl snippet update --content inline_vcl --dynamic --service-id 123 --snippet-id 456 --version 3"),
			WantOutput: "Updated dynamic VCL snippet '456' (service: 123)",
		},
		{
			Name: "validate --autoclone results in cloned service version",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				UpdateSnippetFn: func(i *fastly.UpdateSnippetInput) (*fastly.Snippet, error) {
					// Track the contents parsed
					content = *i.Content

					return &fastly.Snippet{
						Content:        i.Content,
						Name:           i.NewName,
						Priority:       i.Priority,
						ServiceID:      fastly.ToPointer(i.ServiceID),
						ServiceVersion: fastly.ToPointer(i.ServiceVersion),
						Type:           i.Type,
					}, nil
				},
			},
			Args:       args("vcl snippet update --autoclone --content inline_vcl --name foo --new-name bar --priority 1 --service-id 123 --type recv --version 1"),
			WantOutput: "Updated VCL snippet 'bar' (previously: 'foo', service: 123, version: 4, type: recv, priority: 1)",
		},
	}

	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(testcase.Name, func(t *testing.T) {
			var stdout bytes.Buffer
			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts := testutil.MockGlobalData(testcase.Args, &stdout)
				opts.APIClientFactory = mock.APIClient(testcase.API)
				return opts, nil
			}
			err := app.Run(testcase.Args, nil)
			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertStringContains(t, testutil.StripNewLines(stdout.String()), testcase.WantOutput)
			testutil.AssertPathContentFlag("content", testcase.WantError, testcase.Args, "snippet.vcl", content, t)
		})
	}
}

func getSnippet(i *fastly.GetSnippetInput) (*fastly.Snippet, error) {
	t := testutil.Date

	return &fastly.Snippet{
		Content:        fastly.ToPointer("# some vcl content"),
		Dynamic:        fastly.ToPointer(0),
		SnippetID:      fastly.ToPointer("456"),
		Name:           fastly.ToPointer(i.Name),
		Priority:       fastly.ToPointer(0),
		ServiceID:      fastly.ToPointer(i.ServiceID),
		ServiceVersion: fastly.ToPointer(i.ServiceVersion),
		Type:           fastly.ToPointer(fastly.SnippetTypeRecv),

		CreatedAt: &t,
		DeletedAt: &t,
		UpdatedAt: &t,
	}, nil
}

func getDynamicSnippet(i *fastly.GetDynamicSnippetInput) (*fastly.DynamicSnippet, error) {
	t := testutil.Date

	return &fastly.DynamicSnippet{
		Content:   fastly.ToPointer("# some vcl content"),
		SnippetID: fastly.ToPointer(i.SnippetID),
		ServiceID: fastly.ToPointer(i.ServiceID),

		CreatedAt: &t,
		UpdatedAt: &t,
	}, nil
}

func listSnippets(i *fastly.ListSnippetsInput) ([]*fastly.Snippet, error) {
	t := testutil.Date
	vs := []*fastly.Snippet{
		{
			Content:        fastly.ToPointer("# some vcl content"),
			Dynamic:        fastly.ToPointer(1),
			SnippetID:      fastly.ToPointer("abc"),
			Name:           fastly.ToPointer("foo"),
			Priority:       fastly.ToPointer(0),
			ServiceID:      fastly.ToPointer(i.ServiceID),
			ServiceVersion: fastly.ToPointer(i.ServiceVersion),
			Type:           fastly.ToPointer(fastly.SnippetTypeRecv),

			CreatedAt: &t,
			DeletedAt: &t,
			UpdatedAt: &t,
		},
		{
			Content:        fastly.ToPointer("# some vcl content"),
			Dynamic:        fastly.ToPointer(0),
			SnippetID:      fastly.ToPointer("abc"),
			Name:           fastly.ToPointer("bar"),
			Priority:       fastly.ToPointer(0),
			ServiceID:      fastly.ToPointer(i.ServiceID),
			ServiceVersion: fastly.ToPointer(i.ServiceVersion),
			Type:           fastly.ToPointer(fastly.SnippetTypeRecv),

			CreatedAt: &t,
			DeletedAt: &t,
			UpdatedAt: &t,
		},
	}
	return vs, nil
}
