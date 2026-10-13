package text_test

import (
	"bytes"
	"testing"

	"github.com/fastly/go-fastly/v17/fastly"

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
				ServiceID:     new("1"),
				Name:          new("2"),
				Type:          new("3"),
				CustomerID:    new("4"),
				ActiveVersion: new(5),
			},
			wantOutput: "ID: 1\nName: 2\nType: 3\nCustomer ID: 4\nActive version: 5\nVersions: 0\n",
		},
		{
			name:   "with prefix",
			prefix: "\t",
			service: &fastly.Service{
				ServiceID:     new("1"),
				Name:          new("2"),
				Type:          new("3"),
				CustomerID:    new("4"),
				ActiveVersion: new(5),
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
				Number:    new(1),
				ServiceID: new("example"),
				Active:    new(true),
				Locked:    new(true),
				Deployed:  new(true),
				Staging:   new(true),
				Testing:   new(false),
			},
			wantOutput: "Number: 1\nService ID: example\nActive: true\nLocked: true\nDeployed: true\nStaged: true\nTesting: false\n",
		},
		{
			name:   "with",
			prefix: "\t",
			version: &fastly.Version{
				Number:    new(1),
				ServiceID: new("example"),
				Active:    new(true),
				Locked:    new(true),
				Deployed:  new(true),
				Staging:   new(true),
				Testing:   new(false),
			},
			wantOutput: "\tNumber: 1\n\tService ID: example\n\tActive: true\n\tLocked: true\n\tDeployed: true\n\tStaged: true\n\tTesting: false\n",
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			var buf bytes.Buffer
			text.PrintVersion(&buf, testcase.prefix, testcase.version)
			testutil.AssertString(t, testcase.wantOutput, buf.String())
		})
	}
}

func TestIsFastlyID(t *testing.T) {
	for _, testcase := range []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "looks like an ID",
			input: "XkblwIHmR01sOnDHxusu6a",
			want:  true,
		},
		{
			name:  "looks like a URL",
			input: "https://github.com/fastly/cli",
			want:  false,
		},
		{
			name:  "too short",
			input: "Vkzj9WNseT1XN0",
			want:  false,
		},
		{
			name:  "too long",
			input: "Vkzj9WNseT1XN0GqjYrgQGVkzj9WNseT1",
			want:  false,
		},
		{
			name:  "invalid characters",
			input: "GLql1:uzgoC-tEK7bdobt5",
			want:  false,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			testutil.AssertBool(t, testcase.want, text.IsFastlyID(testcase.input))
		})
	}
}
