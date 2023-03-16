package configstore_test

import (
	"bytes"

	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v7/fastly"
)

func fmtStore(cs *fastly.ConfigStore, csm *fastly.ConfigStoreMetadata) string {
	var b bytes.Buffer
	text.PrintConfigStore(&b, cs, csm)
	return b.String()
}

func fmtStores(s []*fastly.ConfigStore) string {
	var b bytes.Buffer
	text.PrintConfigStoresTbl(&b, s)
	return b.String()
}

func fmtServices(s []*fastly.Service) string {
	var b bytes.Buffer
	text.PrintConfigStoreServicesTbl(&b, s)
	return b.String()
}
