package aclentry_test

import (
	"bytes"
	"testing"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
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
			WantError: "error reading service: no service ID found",
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
						IP:        *i.IP,
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
						IP:        *i.IP,
						ServiceID: i.ServiceID,
						Negated:   bool(*i.Negated),
					}, nil
				},
			},
			Args:       args("acl-entry create --acl-id 123 --ip 127.0.0.1 --negated --service-id 123"),
			WantOutput: "Created ACL entry '456' (ip: 127.0.0.1, negated: true, service: 123)",
		},
	}

	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
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

	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
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

	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
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
	count         int
	maxPages      int
	numOfPages    int
	requestedPage int
	returnErr     bool
}

func (p *mockACLPaginator) HasNext() bool {
	if p.count > p.maxPages {
		return false
	}
	p.count++
	return true
}

func (p mockACLPaginator) Remaining() int {
	return 1
}

func (p *mockACLPaginator) GetNext() (as []*fastly.ACLEntry, err error) {
	if p.returnErr {
		err = testutil.Err
	}
	t := testutil.Date
	pageOne := fastly.ACLEntry{
		ACLID:     "123",
		Comment:   "foo",
		CreatedAt: &t,
		DeletedAt: &t,
		ID:        "456",
		IP:        "127.0.0.1",
		ServiceID: "123",
		UpdatedAt: &t,
	}
	pageTwo := fastly.ACLEntry{
		ACLID:     "123",
		Comment:   "bar",
		CreatedAt: &t,
		DeletedAt: &t,
		ID:        "789",
		IP:        "127.0.0.2",
		Negated:   true,
		ServiceID: "123",
		UpdatedAt: &t,
	}
	if p.count == 1 {
		as = append(as, &pageOne)
	}
	if p.count == 2 {
		as = append(as, &pageTwo)
	}
	if p.requestedPage > 0 && p.numOfPages == 1 {
		p.count = p.maxPages + 1 // forces only one result to be displayed
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
					return &mockACLPaginator{returnErr: true}
				},
			},
			Args:      args("acl-entry list --acl-id 123 --service-id 123"),
			WantError: testutil.Err.Error(),
		},
		// NOTE: Our mock paginator defines two ACL entries, and so even when
		// setting --per-page 1 we expect the final output to display both items.
		{
			Name: "validate ListACLEntries API success",
			API: mock.API{
				NewListACLEntriesPaginatorFn: func(i *fastly.ListACLEntriesInput) fastly.PaginatorACLEntries {
					return &mockACLPaginator{numOfPages: i.PerPage, maxPages: 2}
				},
			},
			Args:       args("acl-entry list --acl-id 123 --per-page 1 --service-id 123"),
			WantOutput: listACLEntriesOutput,
		},
		// In the following test, we set --page 1 and as there's only one record
		// displayed per page we expect only the first record to be displayed.
		{
			Name: "validate all results displayed even when page is set",
			API: mock.API{
				NewListACLEntriesPaginatorFn: func(i *fastly.ListACLEntriesInput) fastly.PaginatorACLEntries {
					return &mockACLPaginator{count: i.Page - 1, requestedPage: i.Page, numOfPages: i.PerPage, maxPages: 2}
				},
			},
			Args:       args("acl-entry list --acl-id 123 --page 1 --per-page 1 --service-id 123"),
			WantOutput: listACLEntriesOutputPageOne,
		},
		// In the following test, we set --page 2 and as there's only one record
		// displayed per page we expect only the second record to be displayed.
		{
			Name: "validate only page two of the results are displayed",
			API: mock.API{
				NewListACLEntriesPaginatorFn: func(i *fastly.ListACLEntriesInput) fastly.PaginatorACLEntries {
					return &mockACLPaginator{count: i.Page - 1, requestedPage: i.Page, numOfPages: i.PerPage, maxPages: 2}
				},
			},
			Args:       args("acl-entry list --acl-id 123 --page 2 --per-page 1 --service-id 123"),
			WantOutput: listACLEntriesOutputPageTwo,
		},
		{
			Name: "validate --verbose flag",
			API: mock.API{
				NewListACLEntriesPaginatorFn: func(i *fastly.ListACLEntriesInput) fastly.PaginatorACLEntries {
					return &mockACLPaginator{numOfPages: i.PerPage, maxPages: 2}
				},
			},
			Args:       args("acl-entry list --acl-id 123 --per-page 1 --service-id 123 --verbose"),
			WantOutput: listACLEntriesOutputVerbose,
		},
	}

	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
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

var listACLEntriesOutput = `SERVICE ID  ID   IP         SUBNET  NEGATED
123         456  127.0.0.1  0       false
123         789  127.0.0.2  0       true
`

var listACLEntriesOutputPageOne = `SERVICE ID  ID   IP         SUBNET  NEGATED
123         456  127.0.0.1  0       false
`

var listACLEntriesOutputPageTwo = `SERVICE ID  ID   IP         SUBNET  NEGATED
123         789  127.0.0.2  0       true
`

var listACLEntriesOutputVerbose = `Fastly API token provided via config file (profile: user)
Fastly API endpoint: https://api.fastly.com

Service ID (via --service-id): 123

ACL ID: 123
ID: 456
IP: 127.0.0.1
Subnet: 0
Negated: false
Comment: foo

Created at: 2021-06-15 23:00:00 +0000 UTC
Updated at: 2021-06-15 23:00:00 +0000 UTC
Deleted at: 2021-06-15 23:00:00 +0000 UTC

ACL ID: 123
ID: 789
IP: 127.0.0.2
Subnet: 0
Negated: true
Comment: bar

Created at: 2021-06-15 23:00:00 +0000 UTC
Updated at: 2021-06-15 23:00:00 +0000 UTC
Deleted at: 2021-06-15 23:00:00 +0000 UTC

`

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

	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
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
