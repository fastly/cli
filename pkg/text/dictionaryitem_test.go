package text_test

import (
	"bytes"
	"testing"

	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v4/fastly"
)

func TestPrintDictionaryItem(t *testing.T) {
	for _, testcase := range []struct {
		name           string
		dictionaryItem *fastly.DictionaryItem
		wantOutput     string
	}{
		{
			name:           "base",
			dictionaryItem: &fastly.DictionaryItem{},
			wantOutput:     "Dictionary ID: \nItem Key: \nItem Value: \n",
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			var buf bytes.Buffer
			text.PrintDictionaryItem(&buf, "", testcase.dictionaryItem)
			testutil.AssertString(t, testcase.wantOutput, buf.String())
		})
	}
}
