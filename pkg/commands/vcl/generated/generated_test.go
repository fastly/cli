package generated_test

import (
	"bytes"
	"testing"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v3/fastly"
)

func TestDescribe(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --version flag",
			Args:      args("vcl generated describe"),
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Args:      args("vcl generated describe --version 3"),
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate GetGeneratedVCL API error",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetGeneratedVCLFn: func(i *fastly.GetGeneratedVCLInput) (*fastly.VCL, error) {
					return nil, testutil.Err
				},
			},
			Args:      args("vcl generated describe --service-id 123 --version 3"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate GetGeneratedVCL API success",
			API: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				GetGeneratedVCLFn: getGeneratedVCL,
			},
			Args:       args("vcl generated describe --service-id 123 --version 3"),
			WantOutput: "\nService ID: 123\nService Version: 3\n\nContent: \n# some vcl content\n",
		},
	}

	for _, testcase := range scenarios {
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

func getGeneratedVCL(i *fastly.GetGeneratedVCLInput) (*fastly.VCL, error) {
	return &fastly.VCL{
		ServiceID:      i.ServiceID,
		ServiceVersion: i.ServiceVersion,
		Content:        "# some vcl content",
	}, nil
}
