package text_test

import (
	"bytes"
	"testing"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/cli/pkg/text"
)

func TestPrintService(t *testing.T) {
	for _, testcase := range []struct {
		name       string
		prefix     string
		service    *fastly.Service
		wantOutput string
	}{
		{
			name:   "without prefix",
			prefix: "",
			service: &fastly.Service{
				ServiceID:     fastly.ToPointer("1"),
				Name:          fastly.ToPointer("2"),
				Type:          fastly.ToPointer("3"),
				CustomerID:    fastly.ToPointer("4"),
				ActiveVersion: fastly.ToPointer(5),
			},
			wantOutput: "ID: 1\nName: 2\nType: 3\nCustomer ID: 4\nActive version: 5\nVersions: 0\n",
		},
		{
			name:   "with prefix",
			prefix: "\t",
			service: &fastly.Service{
				ServiceID:     fastly.ToPointer("1"),
				Name:          fastly.ToPointer("2"),
				Type:          fastly.ToPointer("3"),
				CustomerID:    fastly.ToPointer("4"),
				ActiveVersion: fastly.ToPointer(5),
			},
			wantOutput: "\tID: 1\n\tName: 2\n\tType: 3\n\tCustomer ID: 4\n\tActive version: 5\n\tVersions: 0\n",
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			var buf bytes.Buffer
			text.PrintService(&buf, testcase.prefix, testcase.service)
			testutil.AssertString(t, testcase.wantOutput, buf.String())
		})
	}
}

func TestPrintVersion(t *testing.T) {
	for _, testcase := range []struct {
		name       string
		prefix     string
		version    *fastly.Version
		wantOutput string
	}{
		{
			name:   "without prefix",
			prefix: "",
			version: &fastly.Version{
				Number:    fastly.ToPointer(1),
				ServiceID: fastly.ToPointer("example"),
				Active:    fastly.ToPointer(true),
				Locked:    fastly.ToPointer(true),
				Deployed:  fastly.ToPointer(true),
				Staging:   fastly.ToPointer(true),
				Testing:   fastly.ToPointer(false),
			},
			wantOutput: "Number: 1\nService ID: example\nActive: true\nLocked: true\nDeployed: true\nStaging: true\nTesting: false\n",
		},
		{
			name:   "with",
			prefix: "\t",
			version: &fastly.Version{
				Number:    fastly.ToPointer(1),
				ServiceID: fastly.ToPointer("example"),
				Active:    fastly.ToPointer(true),
				Locked:    fastly.ToPointer(true),
				Deployed:  fastly.ToPointer(true),
				Staging:   fastly.ToPointer(true),
				Testing:   fastly.ToPointer(false),
			},
			wantOutput: "\tNumber: 1\n\tService ID: example\n\tActive: true\n\tLocked: true\n\tDeployed: true\n\tStaging: true\n\tTesting: false\n",
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			var buf bytes.Buffer
			text.PrintVersion(&buf, testcase.prefix, testcase.version)
			testutil.AssertString(t, testcase.wantOutput, buf.String())
		})
	}
}
