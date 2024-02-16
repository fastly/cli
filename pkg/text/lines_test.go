package text_test

import (
	"bytes"
	"testing"

	"github.com/fastly/cli/v10/pkg/testutil"
	"github.com/fastly/cli/v10/pkg/text"
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
			wantOutput: "\nitem: value\n",
		},
		{
			name:       "number",
			mapItem:    text.Lines{"number": 2},
			wantOutput: "\nnumber: 2\n",
		},
		{
			name:       "sort",
			mapItem:    text.Lines{"b": 2, "a": 1, "c": 3},
			wantOutput: "\na: 1\nb: 2\nc: 3\n",
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			var buf bytes.Buffer
			text.PrintLines(&buf, testcase.mapItem)
			testutil.AssertString(t, testcase.wantOutput, buf.String())
		})
	}
}
