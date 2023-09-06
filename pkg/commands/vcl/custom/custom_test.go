package custom_test

import (
	"bytes"
	"testing"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestVCLCustomCreate(t *testing.T) {
	var content string
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name: "validate missing --autoclone flag",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
			},
			Args:      args("vcl custom create --content ./testdata/example.vcl --name foo --service-id 123 --version 1"),
			WantError: "service version 1 is not editable",
		},
		{
			Name: "validate CreateVCL API error",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CreateVCLFn: func(i *fastly.CreateVCLInput) (*fastly.VCL, error) {
					return nil, testutil.Err
				},
			},
			Args:      args("vcl custom create --content ./testdata/example.vcl --name foo --service-id 123 --version 3"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate CreateVCL API success for non-main VCL",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CreateVCLFn: func(i *fastly.CreateVCLInput) (*fastly.VCL, error) {
					// Track the contents parsed
					content = *i.Content
					if i.Content == nil {
						i.Content = fastly.String("")
					}
					if i.Main == nil {
						b := false
						i.Main = &b
					}
					if i.Name == nil {
						i.Name = fastly.String("")
					}
					return &fastly.VCL{
						Content:        *i.Content,
						Main:           *i.Main,
						Name:           *i.Name,
						ServiceID:      i.ServiceID,
						ServiceVersion: i.ServiceVersion,
					}, nil
				},
			},
			Args:       args("vcl custom create --content ./testdata/example.vcl --name foo --service-id 123 --version 3"),
			WantOutput: "Created custom VCL 'foo' (service: 123, version: 3, main: false)",
		},
		{
			Name: "validate CreateVCL API success for main VCL",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CreateVCLFn: func(i *fastly.CreateVCLInput) (*fastly.VCL, error) {
					// Track the contents parsed
					// Track the contents parsed
					content = *i.Content
					if i.Content == nil {
						i.Content = fastly.String("")
					}
					if i.Main == nil {
						b := false
						i.Main = &b
					}
					if i.Name == nil {
						i.Name = fastly.String("")
					}
					return &fastly.VCL{
						Content:        *i.Content,
						Main:           *i.Main,
						Name:           *i.Name,
						ServiceID:      i.ServiceID,
						ServiceVersion: i.ServiceVersion,
					}, nil
				},
			},
			Args:       args("vcl custom create --content ./testdata/example.vcl --main --name foo --service-id 123 --version 3"),
			WantOutput: "Created custom VCL 'foo' (service: 123, version: 3, main: true)",
		},
		{
			Name: "validate --autoclone results in cloned service version",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				CreateVCLFn: func(i *fastly.CreateVCLInput) (*fastly.VCL, error) {
					// Track the contents parsed
					content = *i.Content
					if i.Content == nil {
						i.Content = fastly.String("")
					}
					if i.Main == nil {
						b := false
						i.Main = &b
					}
					if i.Name == nil {
						i.Name = fastly.String("")
					}
					return &fastly.VCL{
						Content:        *i.Content,
						Main:           *i.Main,
						Name:           *i.Name,
						ServiceID:      i.ServiceID,
						ServiceVersion: i.ServiceVersion,
					}, nil
				},
			},
			Args:       args("vcl custom create --autoclone --content ./testdata/example.vcl --name foo --service-id 123 --version 1"),
			WantOutput: "Created custom VCL 'foo' (service: 123, version: 4, main: false)",
		},
		{
			Name: "validate CreateVCL API success with inline VCL content",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CreateVCLFn: func(i *fastly.CreateVCLInput) (*fastly.VCL, error) {
					// Track the contents parsed
					content = *i.Content
					if i.Content == nil {
						i.Content = fastly.String("")
					}
					if i.Main == nil {
						b := false
						i.Main = &b
					}
					if i.Name == nil {
						i.Name = fastly.String("")
					}
					return &fastly.VCL{
						Content:        *i.Content,
						Main:           *i.Main,
						Name:           *i.Name,
						ServiceID:      i.ServiceID,
						ServiceVersion: i.ServiceVersion,
					}, nil
				},
			},
			Args:       args("vcl custom create --content inline_vcl --name foo --service-id 123 --version 3"),
			WantOutput: "Created custom VCL 'foo' (service: 123, version: 3, main: false)",
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
			testutil.AssertPathContentFlag("content", testcase.WantError, testcase.Args, "example.vcl", content, t)
		})
	}
}

func TestVCLCustomDelete(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --name flag",
			Args:      args("vcl custom delete --version 3"),
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Name:      "validate missing --version flag",
			Args:      args("vcl custom delete --name foobar"),
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Args:      args("vcl custom delete --name foobar --version 3"),
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate missing --autoclone flag",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
			},
			Args:      args("vcl custom delete --name foobar --service-id 123 --version 1"),
			WantError: "service version 1 is not editable",
		},
		{
			Name: "validate DeleteVCL API error",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				DeleteVCLFn: func(i *fastly.DeleteVCLInput) error {
					return testutil.Err
				},
			},
			Args:      args("vcl custom delete --name foobar --service-id 123 --version 3"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate DeleteVCL API success",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				DeleteVCLFn: func(i *fastly.DeleteVCLInput) error {
					return nil
				},
			},
			Args:       args("vcl custom delete --name foobar --service-id 123 --version 3"),
			WantOutput: "Deleted custom VCL 'foobar' (service: 123, version: 3)",
		},
		{
			Name: "validate --autoclone results in cloned service version",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				DeleteVCLFn: func(i *fastly.DeleteVCLInput) error {
					return nil
				},
			},
			Args:       args("vcl custom delete --autoclone --name foo --service-id 123 --version 1"),
			WantOutput: "Deleted custom VCL 'foo' (service: 123, version: 4)",
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

func TestVCLCustomDescribe(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --name flag",
			Args:      args("vcl custom describe --version 3"),
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Name:      "validate missing --version flag",
			Args:      args("vcl custom describe --name foobar"),
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Args:      args("vcl custom describe --name foobar --version 3"),
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate GetVCL API error",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetVCLFn: func(i *fastly.GetVCLInput) (*fastly.VCL, error) {
					return nil, testutil.Err
				},
			},
			Args:      args("vcl custom describe --name foobar --service-id 123 --version 3"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate GetVCL API success",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetVCLFn:       getVCL,
			},
			Args:       args("vcl custom describe --name foobar --service-id 123 --version 3"),
			WantOutput: "\nService ID: 123\nService Version: 3\n\nName: foobar\nMain: true\nContent: \n# some vcl content\n\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\nDeleted at: 2021-06-15 23:00:00 +0000 UTC\n",
		},
		{
			Name: "validate missing --autoclone flag is OK",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetVCLFn:       getVCL,
			},
			Args:       args("vcl custom describe --name foobar --service-id 123 --version 1"),
			WantOutput: "\nService ID: 123\nService Version: 1\n\nName: foobar\nMain: true\nContent: \n# some vcl content\n\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\nDeleted at: 2021-06-15 23:00:00 +0000 UTC\n",
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

func TestVCLCustomList(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --version flag",
			Args:      args("vcl custom list"),
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Args:      args("vcl custom list --version 3"),
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate ListVCLs API error",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListVCLsFn: func(i *fastly.ListVCLsInput) ([]*fastly.VCL, error) {
					return nil, testutil.Err
				},
			},
			Args:      args("vcl custom list --service-id 123 --version 3"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate ListVCLs API success",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListVCLsFn:     listVCLs,
			},
			Args:       args("vcl custom list --service-id 123 --version 3"),
			WantOutput: "SERVICE ID  VERSION  NAME  MAIN\n123         3        foo   true\n123         3        bar   false\n",
		},
		{
			Name: "validate missing --autoclone flag is OK",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListVCLsFn:     listVCLs,
			},
			Args:       args("vcl custom list --service-id 123 --version 1"),
			WantOutput: "SERVICE ID  VERSION  NAME  MAIN\n123         1        foo   true\n123         1        bar   false\n",
		},
		{
			Name: "validate missing --verbose flag",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListVCLsFn:     listVCLs,
			},
			Args:       args("vcl custom list --service-id 123 --verbose --version 1"),
			WantOutput: "Fastly API token provided via config file (profile: user)\nFastly API endpoint: https://api.fastly.com\n\nService ID (via --service-id): 123\n\nService Version: 1\n\nName: foo\nMain: true\nContent: \n# some vcl content\n\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\nDeleted at: 2021-06-15 23:00:00 +0000 UTC\n\nName: bar\nMain: false\nContent: \n# some vcl content\n\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\nDeleted at: 2021-06-15 23:00:00 +0000 UTC\n",
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

func TestVCLCustomUpdate(t *testing.T) {
	var content string
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --name flag",
			Args:      args("vcl custom update --version 3"),
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Name:      "validate missing --version flag",
			Args:      args("vcl custom update --name foobar"),
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Args:      args("vcl custom update --name foobar --version 3"),
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate missing --autoclone flag",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
			},
			Args:      args("vcl custom update --name foobar --service-id 123 --version 1"),
			WantError: "service version 1 is not editable",
		},
		{
			Name: "validate UpdateVCL API error",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				UpdateVCLFn: func(i *fastly.UpdateVCLInput) (*fastly.VCL, error) {
					return nil, testutil.Err
				},
			},
			Args:      args("vcl custom update --name foobar --new-name beepboop --service-id 123 --version 3"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate UpdateVCL API success with --new-name",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				UpdateVCLFn: func(i *fastly.UpdateVCLInput) (*fastly.VCL, error) {
					return &fastly.VCL{
						Content:        "# untouched",
						Main:           true,
						Name:           *i.NewName,
						ServiceID:      i.ServiceID,
						ServiceVersion: i.ServiceVersion,
					}, nil
				},
			},
			Args:       args("vcl custom update --name foobar --new-name beepboop --service-id 123 --version 3"),
			WantOutput: "Updated custom VCL 'beepboop' (previously: 'foobar', service: 123, version: 3)",
		},
		{
			Name: "validate UpdateVCL API success with --content",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				UpdateVCLFn: func(i *fastly.UpdateVCLInput) (*fastly.VCL, error) {
					// Track the contents parsed
					content = *i.Content

					return &fastly.VCL{
						Content:        *i.Content,
						Main:           true,
						Name:           i.Name,
						ServiceID:      i.ServiceID,
						ServiceVersion: i.ServiceVersion,
					}, nil
				},
			},
			Args:       args("vcl custom update --content updated --name foobar --service-id 123 --version 3"),
			WantOutput: "Updated custom VCL 'foobar' (service: 123, version: 3)",
		},
		{
			Name: "validate --autoclone results in cloned service version",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				UpdateVCLFn: func(i *fastly.UpdateVCLInput) (*fastly.VCL, error) {
					// Track the contents parsed
					content = *i.Content

					return &fastly.VCL{
						Content:        *i.Content,
						Main:           true,
						Name:           i.Name,
						ServiceID:      i.ServiceID,
						ServiceVersion: i.ServiceVersion,
					}, nil
				},
			},
			Args:       args("vcl custom update --autoclone --content ./testdata/example.vcl --name foo --service-id 123 --version 1"),
			WantOutput: "Updated custom VCL 'foo' (service: 123, version: 4)",
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
			testutil.AssertPathContentFlag("content", testcase.WantError, testcase.Args, "example.vcl", content, t)
		})
	}
}

func getVCL(i *fastly.GetVCLInput) (*fastly.VCL, error) {
	t := testutil.Date

	return &fastly.VCL{
		Content:        "# some vcl content",
		Main:           true,
		Name:           i.Name,
		ServiceID:      i.ServiceID,
		ServiceVersion: i.ServiceVersion,

		CreatedAt: &t,
		DeletedAt: &t,
		UpdatedAt: &t,
	}, nil
}

func listVCLs(i *fastly.ListVCLsInput) ([]*fastly.VCL, error) {
	t := testutil.Date
	vs := []*fastly.VCL{
		{
			Content:        "# some vcl content",
			Main:           true,
			Name:           "foo",
			ServiceID:      i.ServiceID,
			ServiceVersion: i.ServiceVersion,

			CreatedAt: &t,
			DeletedAt: &t,
			UpdatedAt: &t,
		},
		{
			Content:        "# some vcl content",
			Main:           false,
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
