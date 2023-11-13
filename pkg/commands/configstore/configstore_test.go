package configstore_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/commands/configstore"
	fstfmt "github.com/fastly/cli/pkg/fmt"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestCreateStoreCommand(t *testing.T) {
	const (
		storeName = "test123"
		storeID   = "store-id-123"
	)
	now := time.Now()

	scenarios := []testutil.TestScenario{
		{
			Args:      testutil.Args(configstore.RootName + " create"),
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: testutil.Args(fmt.Sprintf("%s create --name %s", configstore.RootName, storeName)),
			API: mock.API{
				CreateConfigStoreFn: func(i *fastly.CreateConfigStoreInput) (*fastly.ConfigStore, error) {
					return nil, errors.New("invalid request")
				},
			},
			WantError: "invalid request",
		},
		{
			Args: testutil.Args(fmt.Sprintf("%s create --name %s", configstore.RootName, storeName)),
			API: mock.API{
				CreateConfigStoreFn: func(i *fastly.CreateConfigStoreInput) (*fastly.ConfigStore, error) {
					return &fastly.ConfigStore{
						ID:   storeID,
						Name: i.Name,
					}, nil
				},
			},
			WantOutput: fstfmt.Success("Created Config Store '%s' (%s)", storeName, storeID),
		},
		{
			Args: testutil.Args(fmt.Sprintf("%s create --name %s --json", configstore.RootName, storeName)),
			API: mock.API{
				CreateConfigStoreFn: func(i *fastly.CreateConfigStoreInput) (*fastly.ConfigStore, error) {
					return &fastly.ConfigStore{
						ID:        storeID,
						Name:      i.Name,
						CreatedAt: &now,
						UpdatedAt: &now,
					}, nil
				},
			},
			WantOutput: fstfmt.EncodeJSON(&fastly.ConfigStore{
				ID:        storeID,
				Name:      storeName,
				CreatedAt: &now,
				UpdatedAt: &now,
			}),
		},
	}

	for _, testcase := range scenarios {
		testcase := testcase
		t.Run(testcase.Name, func(t *testing.T) {
			var stdout bytes.Buffer
			app.Init = func(_ []string, _ io.Reader) (app.RunOpts, error) {
				opts := testutil.NewRunOpts(testcase.Args, &stdout)
				opts.APIClientFactory = mock.APIClient(testcase.API)
				return opts, nil
			}
			err := app.Run(testcase.Args, nil)
			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertString(t, testcase.WantOutput, stdout.String())
		})
	}
}

func TestDeleteStoreCommand(t *testing.T) {
	const storeID = "test123"
	errStoreNotFound := errors.New("store not found")

	scenarios := []testutil.TestScenario{
		{
			Args:      testutil.Args(configstore.RootName + " delete"),
			WantError: "error parsing arguments: required flag --store-id not provided",
		},
		{
			Args: testutil.Args(configstore.RootName + " delete --store-id DOES-NOT-EXIST"),
			API: mock.API{
				DeleteConfigStoreFn: func(i *fastly.DeleteConfigStoreInput) error {
					if i.ID != storeID {
						return errStoreNotFound
					}
					return nil
				},
			},
			WantError: errStoreNotFound.Error(),
		},
		{
			Args: testutil.Args(fmt.Sprintf("%s delete --store-id %s", configstore.RootName, storeID)),
			API: mock.API{
				DeleteConfigStoreFn: func(i *fastly.DeleteConfigStoreInput) error {
					if i.ID != storeID {
						return errStoreNotFound
					}
					return nil
				},
			},
			WantOutput: fstfmt.Success("Deleted Config Store '%s'\n", storeID),
		},
		{
			Args: testutil.Args(fmt.Sprintf("%s delete --store-id %s --json", configstore.RootName, storeID)),
			API: mock.API{
				DeleteConfigStoreFn: func(i *fastly.DeleteConfigStoreInput) error {
					if i.ID != storeID {
						return errStoreNotFound
					}
					return nil
				},
			},
			WantOutput: fstfmt.JSON(`{"id": %q, "deleted": true}`, storeID),
		},
	}

	for _, testcase := range scenarios {
		testcase := testcase
		t.Run(testcase.Name, func(t *testing.T) {
			var stdout bytes.Buffer
			app.Init = func(_ []string, _ io.Reader) (app.RunOpts, error) {
				opts := testutil.NewRunOpts(testcase.Args, &stdout)
				opts.APIClientFactory = mock.APIClient(testcase.API)
				return opts, nil
			}
			err := app.Run(testcase.Args, nil)
			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertString(t, testcase.WantOutput, stdout.String())
		})
	}
}

func TestDescribeStoreCommand(t *testing.T) {
	const (
		storeName = "test123"
		storeID   = "store-id-123"
	)

	now := time.Now()

	scenarios := []testutil.TestScenario{
		{
			Args:      testutil.Args(configstore.RootName + " get"),
			WantError: "error parsing arguments: required flag --store-id not provided",
		},
		{
			Args: testutil.Args(fmt.Sprintf("%s get --store-id %s", configstore.RootName, storeID)),
			API: mock.API{
				GetConfigStoreFn: func(i *fastly.GetConfigStoreInput) (*fastly.ConfigStore, error) {
					return nil, errors.New("invalid request")
				},
			},
			WantError: "invalid request",
		},
		{
			Args: testutil.Args(fmt.Sprintf("%s get --store-id %s", configstore.RootName, storeID)),
			API: mock.API{
				GetConfigStoreFn: func(i *fastly.GetConfigStoreInput) (*fastly.ConfigStore, error) {
					return &fastly.ConfigStore{
						ID:        i.ID,
						Name:      storeName,
						CreatedAt: &now,
					}, nil
				},
			},
			WantOutput: fmtStore(
				&fastly.ConfigStore{
					ID:        storeID,
					Name:      storeName,
					CreatedAt: &now,
				},
				nil,
			),
		},
		{
			Args: testutil.Args(fmt.Sprintf("%s get --store-id %s --metadata", configstore.RootName, storeID)),
			API: mock.API{
				GetConfigStoreFn: func(i *fastly.GetConfigStoreInput) (*fastly.ConfigStore, error) {
					return &fastly.ConfigStore{
						ID:        i.ID,
						Name:      storeName,
						CreatedAt: &now,
					}, nil
				},
				GetConfigStoreMetadataFn: func(i *fastly.GetConfigStoreMetadataInput) (*fastly.ConfigStoreMetadata, error) {
					return &fastly.ConfigStoreMetadata{
						ItemCount: 42,
					}, nil
				},
			},
			WantOutput: fmtStore(
				&fastly.ConfigStore{
					ID:        storeID,
					Name:      storeName,
					CreatedAt: &now,
				},
				&fastly.ConfigStoreMetadata{
					ItemCount: 42,
				},
			),
		},
		{
			Args: testutil.Args(fmt.Sprintf("%s get --store-id %s --json", configstore.RootName, storeID)),
			API: mock.API{
				GetConfigStoreFn: func(i *fastly.GetConfigStoreInput) (*fastly.ConfigStore, error) {
					return &fastly.ConfigStore{
						ID:        i.ID,
						Name:      storeName,
						CreatedAt: &now,
					}, nil
				},
			},
			WantOutput: fstfmt.EncodeJSON(&fastly.ConfigStore{
				ID:        storeID,
				Name:      storeName,
				CreatedAt: &now,
			}),
		},
	}

	for _, testcase := range scenarios {
		testcase := testcase
		t.Run(testcase.Name, func(t *testing.T) {
			var stdout bytes.Buffer
			app.Init = func(_ []string, _ io.Reader) (app.RunOpts, error) {
				opts := testutil.NewRunOpts(testcase.Args, &stdout)
				opts.APIClientFactory = mock.APIClient(testcase.API)
				return opts, nil
			}
			err := app.Run(testcase.Args, nil)
			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertString(t, testcase.WantOutput, stdout.String())
		})
	}
}

func TestListStoresCommand(t *testing.T) {
	const (
		storeName = "test123"
		storeID   = "store-id-123"
	)

	now := time.Now()

	stores := []*fastly.ConfigStore{
		{ID: storeID, Name: storeName, CreatedAt: &now},
		{ID: storeID + "+1", Name: storeName + "+1", CreatedAt: &now},
	}

	scenarios := []testutil.TestScenario{
		{
			Args: testutil.Args(configstore.RootName + " list"),
			API: mock.API{
				ListConfigStoresFn: func() ([]*fastly.ConfigStore, error) {
					return nil, nil
				},
			},
			WantOutput: fmtStores(nil),
		},
		{
			Args: testutil.Args(configstore.RootName + " list"),
			API: mock.API{
				ListConfigStoresFn: func() ([]*fastly.ConfigStore, error) {
					return nil, errors.New("unknown error")
				},
			},
			WantError: "unknown error",
		},
		{
			Args: testutil.Args(configstore.RootName + " list"),
			API: mock.API{
				ListConfigStoresFn: func() ([]*fastly.ConfigStore, error) {
					return stores, nil
				},
			},
			WantOutput: fmtStores(stores),
		},
		{
			Args: testutil.Args(configstore.RootName + " list --json"),
			API: mock.API{
				ListConfigStoresFn: func() ([]*fastly.ConfigStore, error) {
					return stores, nil
				},
			},
			WantOutput: fstfmt.EncodeJSON(stores),
		},
	}

	for _, testcase := range scenarios {
		testcase := testcase
		t.Run(testcase.Name, func(t *testing.T) {
			var stdout bytes.Buffer
			app.Init = func(_ []string, _ io.Reader) (app.RunOpts, error) {
				opts := testutil.NewRunOpts(testcase.Args, &stdout)
				opts.APIClientFactory = mock.APIClient(testcase.API)
				return opts, nil
			}
			err := app.Run(testcase.Args, nil)
			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertString(t, testcase.WantOutput, stdout.String())
		})
	}
}

func TestListStoreServicesCommand(t *testing.T) {
	const (
		storeName = "test123"
		storeID   = "store-id-123"
	)

	services := []*fastly.Service{
		{ID: "abc1", Name: "test1", Type: "wasm"},
		{ID: "abc2", Name: "test2", Type: "vcl"},
	}

	scenarios := []testutil.TestScenario{
		{
			Args: testutil.Args(fmt.Sprintf("%s list-services --store-id %s", configstore.RootName, storeID)),
			API: mock.API{
				ListConfigStoreServicesFn: func(i *fastly.ListConfigStoreServicesInput) ([]*fastly.Service, error) {
					return nil, nil
				},
			},
			WantOutput: fmtServices(nil),
		},
		{
			Args: testutil.Args(fmt.Sprintf("%s list-services --store-id %s", configstore.RootName, storeID)),
			API: mock.API{
				ListConfigStoreServicesFn: func(i *fastly.ListConfigStoreServicesInput) ([]*fastly.Service, error) {
					return nil, errors.New("unknown error")
				},
			},
			WantError: "unknown error",
		},
		{
			Args: testutil.Args(fmt.Sprintf("%s list-services --store-id %s", configstore.RootName, storeID)),
			API: mock.API{
				ListConfigStoreServicesFn: func(i *fastly.ListConfigStoreServicesInput) ([]*fastly.Service, error) {
					return services, nil
				},
			},
			WantOutput: fmtServices(services),
		},
		{
			Args: testutil.Args(fmt.Sprintf("%s list-services --store-id %s --json", configstore.RootName, storeID)),
			API: mock.API{
				ListConfigStoreServicesFn: func(i *fastly.ListConfigStoreServicesInput) ([]*fastly.Service, error) {
					return services, nil
				},
			},
			WantOutput: fstfmt.EncodeJSON(services),
		},
	}

	for _, testcase := range scenarios {
		testcase := testcase
		t.Run(testcase.Name, func(t *testing.T) {
			var stdout bytes.Buffer
			app.Init = func(_ []string, _ io.Reader) (app.RunOpts, error) {
				opts := testutil.NewRunOpts(testcase.Args, &stdout)
				opts.APIClientFactory = mock.APIClient(testcase.API)
				return opts, nil
			}
			err := app.Run(testcase.Args, nil)
			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertString(t, testcase.WantOutput, stdout.String())
		})
	}
}

func TestUpdateStoreCommand(t *testing.T) {
	const (
		storeID   = "store-id-123"
		storeName = "test123"
	)
	now := time.Now()

	scenarios := []testutil.TestScenario{
		{
			Args:      testutil.Args(fmt.Sprintf("%s update --store-id %s", configstore.RootName, storeID)),
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: testutil.Args(fmt.Sprintf("%s update --store-id %s --name %s", configstore.RootName, storeID, storeName)),
			API: mock.API{
				UpdateConfigStoreFn: func(i *fastly.UpdateConfigStoreInput) (*fastly.ConfigStore, error) {
					return nil, errors.New("invalid request")
				},
			},
			WantError: "invalid request",
		},
		{
			Args: testutil.Args(fmt.Sprintf("%s update --store-id %s --name %s", configstore.RootName, storeID, storeName)),
			API: mock.API{
				UpdateConfigStoreFn: func(i *fastly.UpdateConfigStoreInput) (*fastly.ConfigStore, error) {
					return &fastly.ConfigStore{
						ID:        storeID,
						Name:      i.Name,
						CreatedAt: &now,
					}, nil
				},
			},
			WantOutput: fstfmt.Success("Updated Config Store '%s' (%s)", storeName, storeID),
		},
		{
			Args: testutil.Args(fmt.Sprintf("%s update --store-id %s --name %s --json", configstore.RootName, storeID, storeName)),
			API: mock.API{
				UpdateConfigStoreFn: func(i *fastly.UpdateConfigStoreInput) (*fastly.ConfigStore, error) {
					return &fastly.ConfigStore{
						ID:        storeID,
						Name:      i.Name,
						CreatedAt: &now,
						UpdatedAt: &now,
					}, nil
				},
			},
			WantOutput: fstfmt.EncodeJSON(&fastly.ConfigStore{
				ID:        storeID,
				Name:      storeName,
				CreatedAt: &now,
				UpdatedAt: &now,
			}),
		},
	}

	for _, testcase := range scenarios {
		testcase := testcase
		t.Run(testcase.Name, func(t *testing.T) {
			var stdout bytes.Buffer
			app.Init = func(_ []string, _ io.Reader) (app.RunOpts, error) {
				opts := testutil.NewRunOpts(testcase.Args, &stdout)
				opts.APIClientFactory = mock.APIClient(testcase.API)
				return opts, nil
			}
			err := app.Run(testcase.Args, nil)
			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertString(t, testcase.WantOutput, stdout.String())
		})
	}
}
