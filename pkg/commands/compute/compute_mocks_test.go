package compute_test

// NOTE: This file doesn't contain any tests. It only contains code that is
// shared across some of the other test files (mostly mocked API responses, but
// also a mocked HTTP client).

import (
	"context"

	"github.com/fastly/go-fastly/v13/fastly"

	"github.com/fastly/cli/pkg/testutil"
)

func getServiceOK(_ context.Context, _ *fastly.GetServiceInput) (*fastly.Service, error) {
	return &fastly.Service{
		ServiceID: fastly.ToPointer("12345"),
		Name:      fastly.ToPointer("test"),
	}, nil
}

func createDomainOK(_ context.Context, i *fastly.CreateDomainInput) (*fastly.Domain, error) {
	return &fastly.Domain{
		ServiceID:      fastly.ToPointer(i.ServiceID),
		ServiceVersion: fastly.ToPointer(i.ServiceVersion),
		Name:           i.Name,
	}, nil
}

func createBackendOK(_ context.Context, i *fastly.CreateBackendInput) (*fastly.Backend, error) {
	return &fastly.Backend{
		ServiceID:      fastly.ToPointer(i.ServiceID),
		ServiceVersion: fastly.ToPointer(i.ServiceVersion),
		Name:           i.Name,
	}, nil
}

func createConfigStoreOK(_ context.Context, i *fastly.CreateConfigStoreInput) (*fastly.ConfigStore, error) {
	return &fastly.ConfigStore{
		Name: i.Name,
	}, nil
}

func updateConfigStoreItemOK(_ context.Context, i *fastly.UpdateConfigStoreItemInput) (*fastly.ConfigStoreItem, error) {
	return &fastly.ConfigStoreItem{
		Key:   i.Key,
		Value: i.Value,
	}, nil
}

func createDictionaryOK(_ context.Context, i *fastly.CreateDictionaryInput) (*fastly.Dictionary, error) {
	return &fastly.Dictionary{
		ServiceID:      fastly.ToPointer(i.ServiceID),
		ServiceVersion: fastly.ToPointer(i.ServiceVersion),
		Name:           i.Name,
	}, nil
}

func createDictionaryItemOK(_ context.Context, i *fastly.CreateDictionaryItemInput) (*fastly.DictionaryItem, error) {
	return &fastly.DictionaryItem{
		ServiceID:    fastly.ToPointer(i.ServiceID),
		DictionaryID: fastly.ToPointer(i.DictionaryID),
		ItemKey:      i.ItemKey,
		ItemValue:    i.ItemValue,
	}, nil
}

func createKVStoreOK(_ context.Context, i *fastly.CreateKVStoreInput) (*fastly.KVStore, error) {
	return &fastly.KVStore{
		StoreID: "example-store",
		Name:    i.Name,
	}, nil
}

func createKVStoreItemOK(_ context.Context, _ *fastly.InsertKVStoreKeyInput) error {
	return nil
}

func createResourceOK(_ context.Context, _ *fastly.CreateResourceInput) (*fastly.Resource, error) {
	return nil, nil
}

func getPackageOk(_ context.Context, i *fastly.GetPackageInput) (*fastly.Package, error) {
	return &fastly.Package{ServiceID: fastly.ToPointer(i.ServiceID), ServiceVersion: fastly.ToPointer(i.ServiceVersion)}, nil
}

func updatePackageOk(_ context.Context, i *fastly.UpdatePackageInput) (*fastly.Package, error) {
	return &fastly.Package{ServiceID: fastly.ToPointer(i.ServiceID), ServiceVersion: fastly.ToPointer(i.ServiceVersion)}, nil
}

func updatePackageError(_ context.Context, _ *fastly.UpdatePackageInput) (*fastly.Package, error) {
	return nil, testutil.Err
}

func activateVersionOk(_ context.Context, i *fastly.ActivateVersionInput) (*fastly.Version, error) {
	return &fastly.Version{ServiceID: fastly.ToPointer(i.ServiceID), Number: fastly.ToPointer(i.ServiceVersion)}, nil
}

func updateVersionOk(_ context.Context, i *fastly.UpdateVersionInput) (*fastly.Version, error) {
	return &fastly.Version{ServiceID: fastly.ToPointer(i.ServiceID), Number: fastly.ToPointer(i.ServiceVersion), Comment: i.Comment}, nil
}

func listDomainsOk(_ context.Context, _ *fastly.ListDomainsInput) ([]*fastly.Domain, error) {
	return []*fastly.Domain{
		{Name: fastly.ToPointer("https://directly-careful-coyote.edgecompute.app")},
	}, nil
}

func listKVStoresOk(_ context.Context, _ *fastly.ListKVStoresInput) (*fastly.ListKVStoresResponse, error) {
	return &fastly.ListKVStoresResponse{
		Data: []fastly.KVStore{
			{
				StoreID: "123",
				Name:    "store_one",
			},
			{
				StoreID: "456",
				Name:    "store_two",
			},
		},
	}, nil
}

func listKVStoresEmpty(_ context.Context, _ *fastly.ListKVStoresInput) (*fastly.ListKVStoresResponse, error) {
	return &fastly.ListKVStoresResponse{}, nil
}

func getKVStoreOk(_ context.Context, _ *fastly.GetKVStoreInput) (*fastly.KVStore, error) {
	return &fastly.KVStore{
		StoreID: "123",
		Name:    "store_one",
	}, nil
}

func listSecretStoresOk(_ context.Context, _ *fastly.ListSecretStoresInput) (*fastly.SecretStores, error) {
	return &fastly.SecretStores{
		Data: []fastly.SecretStore{
			{
				StoreID: "123",
				Name:    "store_one",
			},
			{
				StoreID: "456",
				Name:    "store_two",
			},
		},
	}, nil
}

func listSecretStoresEmpty(_ context.Context, _ *fastly.ListSecretStoresInput) (*fastly.SecretStores, error) {
	return &fastly.SecretStores{}, nil
}

func getSecretStoreOk(_ context.Context, _ *fastly.GetSecretStoreInput) (*fastly.SecretStore, error) {
	return &fastly.SecretStore{
		StoreID: "123",
		Name:    "store_one",
	}, nil
}

func createSecretStoreOk(_ context.Context, _ *fastly.CreateSecretStoreInput) (*fastly.SecretStore, error) {
	return &fastly.SecretStore{
		StoreID: "123",
		Name:    "store_one",
	}, nil
}

func createSecretOk(_ context.Context, _ *fastly.CreateSecretInput) (*fastly.Secret, error) {
	return &fastly.Secret{
		Digest: []byte("123"),
		Name:   "foo",
	}, nil
}

func listConfigStoresOk(_ context.Context, _ *fastly.ListConfigStoresInput) ([]*fastly.ConfigStore, error) {
	return []*fastly.ConfigStore{
		{
			StoreID: "123",
			Name:    "example",
		},
		{
			StoreID: "456",
			Name:    "example_two",
		},
	}, nil
}

func listConfigStoresEmpty(_ context.Context, _ *fastly.ListConfigStoresInput) ([]*fastly.ConfigStore, error) {
	return []*fastly.ConfigStore{}, nil
}

func getConfigStoreOk(_ context.Context, _ *fastly.GetConfigStoreInput) (*fastly.ConfigStore, error) {
	return &fastly.ConfigStore{
		StoreID: "123",
		Name:    "example",
	}, nil
}

func getServiceDetailsWasm(_ context.Context, _ *fastly.GetServiceInput) (*fastly.ServiceDetail, error) {
	return &fastly.ServiceDetail{
		Type: fastly.ToPointer("wasm"),
	}, nil
}
