package compute_test

// NOTE: This file doesn't contain any tests. It only contains code that is
// shared across some of the other test files (mostly mocked API responses, but
// also a mocked HTTP client).

import (
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v8/fastly"
)

func getServiceOK(i *fastly.GetServiceInput) (*fastly.Service, error) {
	return &fastly.Service{
		ID:   "12345",
		Name: "test",
	}, nil
}

func createDomainOK(i *fastly.CreateDomainInput) (*fastly.Domain, error) {
	return &fastly.Domain{
		ServiceID:      i.ServiceID,
		ServiceVersion: i.ServiceVersion,
		Name:           *i.Name,
	}, nil
}

func createBackendOK(i *fastly.CreateBackendInput) (*fastly.Backend, error) {
	return &fastly.Backend{
		ServiceID:      i.ServiceID,
		ServiceVersion: i.ServiceVersion,
		Name:           *i.Name,
	}, nil
}

func createDictionaryOK(i *fastly.CreateDictionaryInput) (*fastly.Dictionary, error) {
	return &fastly.Dictionary{
		ServiceID:      i.ServiceID,
		ServiceVersion: i.ServiceVersion,
		Name:           *i.Name,
	}, nil
}

func createDictionaryItemOK(i *fastly.CreateDictionaryItemInput) (*fastly.DictionaryItem, error) {
	return &fastly.DictionaryItem{
		ServiceID:    i.ServiceID,
		DictionaryID: i.DictionaryID,
		ItemKey:      i.ItemKey,
		ItemValue:    i.ItemValue,
	}, nil
}

func createKVStoreOK(i *fastly.CreateKVStoreInput) (*fastly.KVStore, error) {
	return &fastly.KVStore{
		ID:   "example-store",
		Name: i.Name,
	}, nil
}

func createKVStoreItemOK(i *fastly.InsertKVStoreKeyInput) error {
	return nil
}

func createResourceOK(i *fastly.CreateResourceInput) (*fastly.Resource, error) {
	return nil, nil
}

func getPackageOk(i *fastly.GetPackageInput) (*fastly.Package, error) {
	return &fastly.Package{ServiceID: i.ServiceID, ServiceVersion: i.ServiceVersion}, nil
}

func updatePackageOk(i *fastly.UpdatePackageInput) (*fastly.Package, error) {
	return &fastly.Package{ServiceID: i.ServiceID, ServiceVersion: i.ServiceVersion}, nil
}

func updatePackageError(i *fastly.UpdatePackageInput) (*fastly.Package, error) {
	return nil, testutil.Err
}

func activateVersionOk(i *fastly.ActivateVersionInput) (*fastly.Version, error) {
	return &fastly.Version{ServiceID: i.ServiceID, Number: i.ServiceVersion}, nil
}

func updateVersionOk(i *fastly.UpdateVersionInput) (*fastly.Version, error) {
	return &fastly.Version{ServiceID: i.ServiceID, Number: i.ServiceVersion, Comment: *i.Comment}, nil
}

func listDomainsOk(i *fastly.ListDomainsInput) ([]*fastly.Domain, error) {
	return []*fastly.Domain{
		{Name: "https://directly-careful-coyote.edgecompute.app"},
	}, nil
}

func getServiceDetailsWasm(i *fastly.GetServiceInput) (*fastly.ServiceDetail, error) {
	return &fastly.ServiceDetail{
		Type: "wasm",
	}, nil
}
