package text_test

import (
	"bytes"
	"testing"

	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/cli/pkg/text"
)

func TestPrintLines(t *testing.T) {
	for _, testcase := range []struct {
		name       string
		mapItem    text.Lines
		wantOutput string
	}{
		{
			name:       "base",
			mapItem:    text.Lines{"item": "value"},
			wantOutput: "item: value\n",
		},
		{
			name:       "number",
			mapItem:    text.Lines{"number": 2},
			wantOutput: "number: 2\n",
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			var buf bytes.Buffer
			text.PrintLines(&buf, testcase.mapItem)
			testutil.AssertString(t, testcase.wantOutput, buf.String())
		})
	}
}
