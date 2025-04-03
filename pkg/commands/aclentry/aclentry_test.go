package aclentry_test

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v10/fastly"

	root "github.com/fastly/cli/pkg/commands/aclentry"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestACLEntryCreate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --acl-id flag",
			Args:      "--ip 127.0.0.1",
			WantError: "error parsing arguments: required flag --acl-id not provided",
		},
		{
			Name:      "validate missing --ip flag",
			Args:      "--acl-id 123",
			WantError: "error reading service: no service ID found",
		},
		{
			Name:      "validate missing --service-id flag",
			Args:      "--acl-id 123 --ip 127.0.0.1",
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate CreateACLEntry API error",
			API: mock.API{
				CreateACLEntryFn: func(i *fastly.CreateACLEntryInput) (*fastly.ACLEntry, error) {
					return nil, testutil.Err
				},
			},
			Args:      "--acl-id 123 --ip 127.0.0.1 --service-id 123",
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate CreateACLEntry API success",
			API: mock.API{
				CreateACLEntryFn: func(i *fastly.CreateACLEntryInput) (*fastly.ACLEntry, error) {
					return &fastly.ACLEntry{
						ACLID:     fastly.ToPointer(i.ACLID),
						EntryID:   fastly.ToPointer("456"),
						IP:        i.IP,
						ServiceID: fastly.ToPointer(i.ServiceID),
					}, nil
				},
			},
			Args:       "--acl-id 123 --ip 127.0.0.1 --service-id 123",
			WantOutput: "Created ACL entry '456' (ip: 127.0.0.1, negated: false, service: 123)",
		},
		{
			Name: "validate CreateACLEntry API success with negated IP",
			API: mock.API{
				CreateACLEntryFn: func(i *fastly.CreateACLEntryInput) (*fastly.ACLEntry, error) {
					return &fastly.ACLEntry{
						ACLID:     fastly.ToPointer(i.ACLID),
						EntryID:   fastly.ToPointer("456"),
						IP:        i.IP,
						ServiceID: fastly.ToPointer(i.ServiceID),
						Negated:   fastly.ToPointer(bool(fastly.ToValue(i.Negated))),
					}, nil
				},
			},
			Args:       "--acl-id 123 --ip 127.0.0.1 --negated --service-id 123",
			WantOutput: "Created ACL entry '456' (ip: 127.0.0.1, negated: true, service: 123)",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, "create"}, scenarios)
}

func TestACLEntryDelete(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --acl-id flag",
			Args:      "--id 456",
			WantError: "error parsing arguments: required flag --acl-id not provided",
		},
		{
			Name:      "validate missing --id flag",
			Args:      "--acl-id 123",
			WantError: "error parsing arguments: required flag --id not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Args:      "--acl-id 123 --id 456",
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate DeleteACL API error",
			API: mock.API{
				DeleteACLEntryFn: func(i *fastly.DeleteACLEntryInput) error {
					return testutil.Err
				},
			},
			Args:      "--acl-id 123 --id 456 --service-id 123",
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate DeleteACL API success",
			API: mock.API{
				DeleteACLEntryFn: func(i *fastly.DeleteACLEntryInput) error {
					return nil
				},
			},
			Args:       "--acl-id 123 --id 456 --service-id 123",
			WantOutput: "Deleted ACL entry '456' (service: 123)",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, "delete"}, scenarios)
}

func TestACLEntryDescribe(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --acl-id flag",
			Args:      "--id 456",
			WantError: "error parsing arguments: required flag --acl-id not provided",
		},
		{
			Name:      "validate missing --id flag",
			Args:      "--acl-id 123",
			WantError: "error parsing arguments: required flag --id not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Args:      "--acl-id 123 --id 456",
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate GetACL API error",
			API: mock.API{
				GetACLEntryFn: func(i *fastly.GetACLEntryInput) (*fastly.ACLEntry, error) {
					return nil, testutil.Err
				},
			},
			Args:      "--acl-id 123 --id 456 --service-id 123",
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate GetACL API success",
			API: mock.API{
				GetACLEntryFn: getACLEntry,
			},
			Args:       "--acl-id 123 --id 456 --service-id 123",
			WantOutput: "\nService ID: 123\nACL ID: 123\nID: 456\nIP: 127.0.0.1\nSubnet: 0\nNegated: false\nComment: \n\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\nDeleted at: 2021-06-15 23:00:00 +0000 UTC\n",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, "describe"}, scenarios)
}

func TestACLEntryList(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --acl-id flag",
			WantError: "error parsing arguments: required flag --acl-id not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Args:      "--acl-id 123",
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate ListACLEntries API error (via GetNext() call)",
			API: mock.API{
				GetACLEntriesFn: func(i *fastly.GetACLEntriesInput) *fastly.ListPaginator[fastly.ACLEntry] {
					return fastly.NewPaginator[fastly.ACLEntry](&mock.HTTPClient{
						Errors: []error{
							testutil.Err,
						},
						Responses: []*http.Response{nil},
					}, fastly.ListOpts{}, "/example")
				},
			},
			Args:      "--acl-id 123 --service-id 123",
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate ListACLEntries API success",
			API: mock.API{
				GetACLEntriesFn: func(i *fastly.GetACLEntriesInput) *fastly.ListPaginator[fastly.ACLEntry] {
					return fastly.NewPaginator[fastly.ACLEntry](&mock.HTTPClient{
						Errors: []error{nil},
						Responses: []*http.Response{
							{
								Body: io.NopCloser(strings.NewReader(`[
                  {
                    "id": "456",
                    "service_id": "123",
                    "acl_id": "xyz",
                    "ip": "127.0.0.1",
                    "negated": 0,
                    "subnet": 0,
                    "comment": "",
                    "created_at": "2020-04-21T18:14:32+00:00",
                    "updated_at": "2020-04-21T18:14:32+00:00",
                    "deleted_at": null
                  },
                  {
                    "id": "789",
                    "service_id": "123",
                    "acl_id": "xyz",
                    "ip": "127.0.0.2",
                    "negated": 1,
                    "subnet": 0,
                    "comment": "",
                    "created_at": "2020-04-21T18:14:32+00:00",
                    "updated_at": "2020-04-21T18:14:32+00:00",
                    "deleted_at": null
                  }
                ]`)),
							},
						},
					}, fastly.ListOpts{}, "/example")
				},
			},
			Args:       "--acl-id 123 --service-id 123",
			WantOutput: listACLEntriesOutput,
		},
		{
			Name: "validate --verbose flag",
			API: mock.API{
				GetACLEntriesFn: func(i *fastly.GetACLEntriesInput) *fastly.ListPaginator[fastly.ACLEntry] {
					return fastly.NewPaginator[fastly.ACLEntry](&mock.HTTPClient{
						Errors: []error{nil},
						Responses: []*http.Response{
							{
								Body: io.NopCloser(strings.NewReader(`[
                  {
                    "id": "456",
                    "service_id": "123",
                    "acl_id": "123",
                    "ip": "127.0.0.1",
                    "negated": 0,
                    "subnet": 0,
                    "comment": "foo",
                    "created_at": "2021-06-15T23:00:00Z",
                    "updated_at": "2021-06-15T23:00:00Z",
                    "deleted_at": "2021-06-15T23:00:00Z"
                  },
                  {
                    "id": "789",
                    "service_id": "123",
                    "acl_id": "123",
                    "ip": "127.0.0.2",
                    "negated": 1,
                    "subnet": 0,
                    "comment": "bar",
                    "created_at": "2021-06-15T23:00:00Z",
                    "updated_at": "2021-06-15T23:00:00Z",
                    "deleted_at": "2021-06-15T23:00:00Z"
                  }
                ]`)),
							},
						},
					}, fastly.ListOpts{}, "/example")
				},
			},
			Args:       "--acl-id 123 --per-page 1 --service-id 123 --verbose",
			WantOutput: listACLEntriesOutputVerbose,
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, "list"}, scenarios)
}

var listACLEntriesOutput = `SERVICE ID  ID   IP         SUBNET  NEGATED
123         456  127.0.0.1  0       false
123         789  127.0.0.2  0       true
`

var listACLEntriesOutputVerbose = `Fastly API endpoint: https://api.fastly.com
Fastly API token provided via config file (profile: user)

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
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --acl-id flag",
			WantError: "error parsing arguments: required flag --acl-id not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Args:      "--acl-id 123",
			WantError: "error reading service: no service ID found",
		},
		{
			Name:      "validate missing --id flag for single entry update",
			Args:      "--acl-id 123 --service-id 123",
			WantError: "no ID found",
		},
		{
			Name: "validate UpdateACL API error",
			API: mock.API{
				UpdateACLEntryFn: func(i *fastly.UpdateACLEntryInput) (*fastly.ACLEntry, error) {
					return nil, testutil.Err
				},
			},
			Args:      "--acl-id 123 --id 456 --service-id 123",
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate error from --file set with invalid json",
			API: mock.API{
				BatchModifyACLEntriesFn: func(i *fastly.BatchModifyACLEntriesInput) error {
					return nil
				},
			},
			Args:      `--acl-id 123 --file {"foo":"bar"} --id 456 --service-id 123`,
			WantError: "missing 'entries'",
		},
		{
			Name: "validate error from --file set with zero json entries",
			API: mock.API{
				BatchModifyACLEntriesFn: func(i *fastly.BatchModifyACLEntriesInput) error {
					return nil
				},
			},
			Args:      `--acl-id 123 --file {"entries":[]} --id 456 --service-id 123`,
			WantError: "missing 'entries'",
		},
		{
			Name: "validate success with --file",
			API: mock.API{
				BatchModifyACLEntriesFn: func(i *fastly.BatchModifyACLEntriesInput) error {
					return nil
				},
			},
			Args:       "--acl-id 123 --file testdata/batch.json --id 456 --service-id 123",
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
			Args:       `--acl-id 123 --file {"entries":[{"op":"create","ip":"127.0.0.1","subnet":8},{"op":"update"},{"op":"upsert"}]} --id 456 --service-id 123`,
			WantOutput: "Updated 3 ACL entries (service: 123)",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, "update"}, scenarios)
}

func getACLEntry(i *fastly.GetACLEntryInput) (*fastly.ACLEntry, error) {
	t := testutil.Date

	return &fastly.ACLEntry{
		ACLID:     fastly.ToPointer(i.ACLID),
		EntryID:   fastly.ToPointer(i.EntryID),
		IP:        fastly.ToPointer("127.0.0.1"),
		ServiceID: fastly.ToPointer(i.ServiceID),

		CreatedAt: &t,
		DeletedAt: &t,
		UpdatedAt: &t,
	}, nil
}
