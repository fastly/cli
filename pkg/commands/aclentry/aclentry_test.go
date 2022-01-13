package aclentry_test

import (
	"bytes"
	"testing"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v6/fastly"
)

func TestACLEntryCreate(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --acl-id flag",
			Args:      args("acl-entry create --ip 127.0.0.1"),
			WantError: "error parsing arguments: required flag --acl-id not provided",
		},
		{
			Name:      "validate missing --ip flag",
			Args:      args("acl-entry create --acl-id 123"),
			WantError: "error parsing arguments: required flag --ip not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Args:      args("acl-entry create --acl-id 123 --ip 127.0.0.1"),
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate CreateACLEntry API error",
			API: mock.API{
				CreateACLEntryFn: func(i *fastly.CreateACLEntryInput) (*fastly.ACLEntry, error) {
					return nil, testutil.Err
				},
			},
			Args:      args("acl-entry create --acl-id 123 --ip 127.0.0.1 --service-id 123"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate CreateACLEntry API success",
			API: mock.API{
				CreateACLEntryFn: func(i *fastly.CreateACLEntryInput) (*fastly.ACLEntry, error) {
					return &fastly.ACLEntry{
						ACLID:     i.ACLID,
						ID:        "456",
						IP:        i.IP,
						ServiceID: i.ServiceID,
					}, nil
				},
			},
			Args:       args("acl-entry create --acl-id 123 --ip 127.0.0.1 --service-id 123"),
			WantOutput: "Created ACL entry '456' (ip: 127.0.0.1, negated: false, service: 123)",
		},
		{
			Name: "validate CreateACLEntry API success with negated IP",
			API: mock.API{
				CreateACLEntryFn: func(i *fastly.CreateACLEntryInput) (*fastly.ACLEntry, error) {
					return &fastly.ACLEntry{
						ACLID:     i.ACLID,
						ID:        "456",
						IP:        i.IP,
						ServiceID: i.ServiceID,
						Negated:   bool(i.Negated),
					}, nil
				},
			},
			Args:       args("acl-entry create --acl-id 123 --ip 127.0.0.1 --negated --service-id 123"),
			WantOutput: "Created ACL entry '456' (ip: 127.0.0.1, negated: true, service: 123)",
		},
	}

	for _, testcase := range scenarios {
		t.Run(testcase.Name, func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.Args, &stdout)
			opts.APIClient = mock.APIClient(testcase.API)
			err := app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.WantOutput)
		})
	}
}

func TestACLEntryDelete(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --acl-id flag",
			Args:      args("acl-entry delete --id 456"),
			WantError: "error parsing arguments: required flag --acl-id not provided",
		},
		{
			Name:      "validate missing --id flag",
			Args:      args("acl-entry delete --acl-id 123"),
			WantError: "error parsing arguments: required flag --id not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Args:      args("acl-entry delete --acl-id 123 --id 456"),
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate DeleteACL API error",
			API: mock.API{
				DeleteACLEntryFn: func(i *fastly.DeleteACLEntryInput) error {
					return testutil.Err
				},
			},
			Args:      args("acl-entry delete --acl-id 123 --id 456 --service-id 123"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate DeleteACL API success",
			API: mock.API{
				DeleteACLEntryFn: func(i *fastly.DeleteACLEntryInput) error {
					return nil
				},
			},
			Args:       args("acl-entry delete --acl-id 123 --id 456 --service-id 123"),
			WantOutput: "Deleted ACL entry '456' (service: 123)",
		},
	}

	for _, testcase := range scenarios {
		t.Run(testcase.Name, func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.Args, &stdout)
			opts.APIClient = mock.APIClient(testcase.API)
			err := app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.WantOutput)
		})
	}
}

func TestACLEntryDescribe(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --acl-id flag",
			Args:      args("acl-entry describe --id 456"),
			WantError: "error parsing arguments: required flag --acl-id not provided",
		},
		{
			Name:      "validate missing --id flag",
			Args:      args("acl-entry describe --acl-id 123"),
			WantError: "error parsing arguments: required flag --id not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Args:      args("acl-entry describe --acl-id 123 --id 456"),
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate GetACL API error",
			API: mock.API{
				GetACLEntryFn: func(i *fastly.GetACLEntryInput) (*fastly.ACLEntry, error) {
					return nil, testutil.Err
				},
			},
			Args:      args("acl-entry describe --acl-id 123 --id 456 --service-id 123"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate GetACL API success",
			API: mock.API{
				GetACLEntryFn: getACLEntry,
			},
			Args:       args("acl-entry describe --acl-id 123 --id 456 --service-id 123"),
			WantOutput: "\nService ID: 123\nACL ID: 123\nID: 456\nIP: 127.0.0.1\nSubnet: 0\nNegated: false\nComment: \n\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\nDeleted at: 2021-06-15 23:00:00 +0000 UTC\n",
		},
	}

	for _, testcase := range scenarios {
		t.Run(testcase.Name, func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.Args, &stdout)
			opts.APIClient = mock.APIClient(testcase.API)
			err := app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.WantOutput)
		})
	}
}

// TODO: Use generics support in go 1.18 to remove the need for multiple types.
//
// e.g. replace mockACLPaginator, mockDictionaryItemPaginator, mockServicesPaginator

type mockACLPaginator struct {
	returnErr bool
	hasNext   bool
}

func (p *mockACLPaginator) HasNext() bool {
	if p.hasNext {
		p.hasNext = false
		return true
	}
	return false
}

func (p mockACLPaginator) Remaining() int {
	return 1
}

func (p mockACLPaginator) GetNext() (as []*fastly.ACLEntry, err error) {
	if p.returnErr {
		err = testutil.Err
	}
	t := testutil.Date
	as = []*fastly.ACLEntry{
		{
			ACLID:     "123",
			Comment:   "foo",
			CreatedAt: &t,
			DeletedAt: &t,
			ID:        "456",
			IP:        "127.0.0.1",
			ServiceID: "123",
			UpdatedAt: &t,
		},
		{
			ACLID:     "123",
			Comment:   "bar",
			CreatedAt: &t,
			DeletedAt: &t,
			ID:        "789",
			IP:        "127.0.0.2",
			Negated:   true,
			ServiceID: "123",
			UpdatedAt: &t,
		},
	}
	return as, err
}

func TestACLEntryList(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --acl-id flag",
			Args:      args("acl-entry list"),
			WantError: "error parsing arguments: required flag --acl-id not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Args:      args("acl-entry list --acl-id 123"),
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate ListACLEntries API error (via GetNext() call)",
			API: mock.API{
				NewListACLEntriesPaginatorFn: func(i *fastly.ListACLEntriesInput) fastly.PaginatorACLEntries {
					return &mockACLPaginator{returnErr: true, hasNext: true}
				},
			},
			Args:      args("acl-entry list --acl-id 123 --service-id 123"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate ListACLEntries API success",
			API: mock.API{
				NewListACLEntriesPaginatorFn: func(i *fastly.ListACLEntriesInput) fastly.PaginatorACLEntries {
					return &mockACLPaginator{hasNext: true}
				},
			},
			Args:       args("acl-entry list --acl-id 123 --service-id 123"),
			WantOutput: "SERVICE ID  ID   IP         SUBNET  NEGATED\n123         456  127.0.0.1  0       false\n123         789  127.0.0.2  0       true\n",
		},
		{
			Name: "validate --verbose flag",
			API: mock.API{
				NewListACLEntriesPaginatorFn: func(i *fastly.ListACLEntriesInput) fastly.PaginatorACLEntries {
					return &mockACLPaginator{hasNext: true}
				},
			},
			Args:       args("acl-entry list --acl-id 123 --service-id 123 --verbose"),
			WantOutput: "Fastly API token not provided\nFastly API endpoint: https://api.fastly.com\nService ID (via --service-id): 123\n\nACL ID: 123\nID: 456\nIP: 127.0.0.1\nSubnet: 0\nNegated: false\nComment: foo\n\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\nDeleted at: 2021-06-15 23:00:00 +0000 UTC\n\nACL ID: 123\nID: 789\nIP: 127.0.0.2\nSubnet: 0\nNegated: true\nComment: bar\n\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\nDeleted at: 2021-06-15 23:00:00 +0000 UTC\n",
		},
	}

	for _, testcase := range scenarios {
		t.Run(testcase.Name, func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.Args, &stdout)
			opts.APIClient = mock.APIClient(testcase.API)
			err := app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.WantOutput)
		})
	}
}

func TestACLEntryUpdate(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --acl-id flag",
			Args:      args("acl-entry update"),
			WantError: "error parsing arguments: required flag --acl-id not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Args:      args("acl-entry update --acl-id 123"),
			WantError: "error reading service: no service ID found",
		},
		{
			Name:      "validate missing --id flag for single entry update",
			Args:      args("acl-entry update --acl-id 123 --service-id 123"),
			WantError: "no ID found",
		},
		{
			Name: "validate UpdateACL API error",
			API: mock.API{
				UpdateACLEntryFn: func(i *fastly.UpdateACLEntryInput) (*fastly.ACLEntry, error) {
					return nil, testutil.Err
				},
			},
			Args:      args("acl-entry update --acl-id 123 --id 456 --service-id 123"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate error from --file set with invalid json",
			API: mock.API{
				BatchModifyACLEntriesFn: func(i *fastly.BatchModifyACLEntriesInput) error {
					return nil
				},
			},
			Args:      args(`acl-entry update --acl-id 123 --file {"foo":"bar"} --id 456 --service-id 123`),
			WantError: "missing 'entries'",
		},
		{
			Name: "validate error from --file set with zero json entries",
			API: mock.API{
				BatchModifyACLEntriesFn: func(i *fastly.BatchModifyACLEntriesInput) error {
					return nil
				},
			},
			Args:      args(`acl-entry update --acl-id 123 --file {"entries":[]} --id 456 --service-id 123`),
			WantError: "missing 'entries'",
		},
		{
			Name: "validate success with --file",
			API: mock.API{
				BatchModifyACLEntriesFn: func(i *fastly.BatchModifyACLEntriesInput) error {
					return nil
				},
			},
			Args:       args("acl-entry update --acl-id 123 --file testdata/batch.json --id 456 --service-id 123"),
			WantOutput: "Updated 3 ACL entries (service: 123)",
		},
		// NOTE: When specifying JSON inline be sure not to have any spaces, and don't
		// try to side-step it by wrapping in single quotes as the CLI parser will
		// get confused (it will consider the single quotes as being part of the
		// string it has parsed, e.g. "'{...}'" which means a json.Unmarshal error).
		{
			Name: "validate success with --file as inline json",
			API: mock.API{
				BatchModifyACLEntriesFn: func(i *fastly.BatchModifyACLEntriesInput) error {
					return nil
				},
			},
			Args:       args(`acl-entry update --acl-id 123 --file {"entries":[{"op":"create","ip":"127.0.0.1","subnet":8},{"op":"update"},{"op":"upsert"}]} --id 456 --service-id 123`),
			WantOutput: "Updated 3 ACL entries (service: 123)",
		},
	}

	for _, testcase := range scenarios {
		t.Run(testcase.Name, func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.Args, &stdout)
			opts.APIClient = mock.APIClient(testcase.API)
			err := app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.WantOutput)
		})
	}
}

func getACLEntry(i *fastly.GetACLEntryInput) (*fastly.ACLEntry, error) {
	t := testutil.Date

	return &fastly.ACLEntry{
		ACLID:     i.ACLID,
		ID:        i.ID,
		IP:        "127.0.0.1",
		ServiceID: i.ServiceID,

		CreatedAt: &t,
		DeletedAt: &t,
		UpdatedAt: &t,
	}, nil
}
