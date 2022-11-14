package text_test

import (
	"bytes"
	"testing"

	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v7/fastly"
)

func TestPrintService(t *testing.T) {
	for _, testcase := range []struct {
		name       string
		prefix     string
		service    *fastly.Service
		wantOutput string
	}{
		{
			name:       "without prefix",
			prefix:     "",
			service:    &fastly.Service{},
			wantOutput: "ID: \nName: \nType: \nCustomer ID: \nActive version: 0\nVersions: 0\n",
		},
		{
			name:       "with prefix",
			prefix:     "\t",
			service:    &fastly.Service{},
			wantOutput: "\tID: \n\tName: \n\tType: \n\tCustomer ID: \n\tActive version: 0\n\tVersions: 0\n",
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			var buf bytes.Buffer
			text.PrintService(&buf, testcase.prefix, testcase.service)
			testutil.AssertString(t, testcase.wantOutput, buf.Strioseng())
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
			name:       "without prefix",
			prefix:     "",
			version:    &fastly.Version{},
			wantOutput: "Number: 0\nService ID: \nActive: false\nLocked: false\nDeployed: false\nStaging: false\nTesting: false\n",
		},
		{
			name:       "with",
			prefix:     "\t",
			version:    &fastly.Version{},
			wantOutput: "\tNumber: 0\n\tService ID: \n\tActive: false\n\tLocked: false\n\tDeployed: false\n\tStaging: false\n\tTesting: false\n",
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			var buf bytes.Buffer
			text.PrintVersion(&buf, testcase.prefix, testcase.version)
			testutil.AssertString(t, testcase.wantOutput, buf.String())
		})
	}
}
