package computeacl_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v13/fastly"

	"github.com/stretchr/testify/require"

	root "github.com/fastly/cli/pkg/commands/compute"
	sub "github.com/fastly/cli/pkg/commands/compute/computeacl"
	fstfmt "github.com/fastly/cli/pkg/fmt"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v13/fastly/computeacls"
)

func TestComputeACLCreate(t *testing.T) {
	const (
		aclName = "foo"
		aclID   = "bar"
	)

	acl := computeacls.ComputeACL{
		Name:         aclName,
		ComputeACLID: aclID,
	}

	scenarios := []testutil.CLIScenario[sub.APICreateFunc]{
		{
			Name:      "validate missing --name flag",
			Args:      "",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Name: "validate API success (without mock)",
			Args: fmt.Sprintf("--name %s", aclName),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader((testutil.GenJSON(acl)))),
					},
				},
			},
			WantOutput: fstfmt.Success("Created compute ACL '%s' (id: %s)", aclName, aclID),
		},
		{
			Name: "validate API success",
			Args: fmt.Sprintf("--name %s", aclName),
			APIFuncMock: func() sub.APICreateFunc {
				return func(_ context.Context, _ *fastly.Client, i *computeacls.CreateInput) (*computeacls.ComputeACL, error) {
					require.Equal(t, aclName, *i.Name, "unexpected ACL name")

					return &acl, nil
				}
			},
			WantOutput: fstfmt.Success("Created compute ACL '%s' (id: %s)", aclName, aclID),
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--name %s --json", aclName),
			APIFuncMock: func() sub.APICreateFunc {
				return func(_ context.Context, _ *fastly.Client, _ *computeacls.CreateInput) (*computeacls.ComputeACL, error) {
					return &acl, nil
				}
			},
			WantOutput: fstfmt.EncodeJSON(acl),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "create"}, scenarios)
}

func TestComputeACLDelete(t *testing.T) {
	const aclID = "foo"

	scenarios := []testutil.CLIScenario[sub.APIDeleteFunc]{
		{
			Name:      "validate missing --acl-id flag",
			Args:      "",
			WantError: "error parsing arguments: required flag --acl-id not provided",
		},
		{
			Name: "validate API success (without mock)",
			Args: fmt.Sprintf("--acl-id %s", aclID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
					},
				},
			},
			WantOutput: fstfmt.Success("Deleted compute ACL (id: %s)", aclID),
		},
		{
			Name: "validate API success",
			Args: fmt.Sprintf("--acl-id %s", aclID),
			APIFuncMock: func() sub.APIDeleteFunc {
				return func(_ context.Context, _ *fastly.Client, i *computeacls.DeleteInput) error {
					require.Equal(t, aclID, *i.ComputeACLID, "unexpected ACL ID")

					return nil
				}
			},
			WantOutput: fstfmt.Success("Deleted compute ACL (id: %s)", aclID),
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--acl-id %s --json", aclID),
			APIFuncMock: func() sub.APIDeleteFunc {
				return func(_ context.Context, _ *fastly.Client, _ *computeacls.DeleteInput) error {
					return nil
				}
			},
			WantOutput: fstfmt.JSON(`{"id": %q, "deleted": true}`, aclID),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "delete"}, scenarios)
}

func TestComputeACLDescribe(t *testing.T) {
	const (
		aclName = "foo"
		aclID   = "bar"
	)

	acl := computeacls.ComputeACL{
		Name:         aclName,
		ComputeACLID: aclID,
	}

	scenarios := []testutil.CLIScenario[sub.APIDescribeFunc]{
		{
			Name:      "validate missing --acl-id flag",
			Args:      "",
			WantError: "error parsing arguments: required flag --acl-id not provided",
		},
		{
			Name: "validate API success (without mock)",
			Args: fmt.Sprintf("--acl-id %s", aclID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader((testutil.GenJSON(acl)))),
					},
				},
			},
			WantOutput: computeACL,
		},
		{
			Name: "validate API success",
			Args: fmt.Sprintf("--acl-id %s", aclID),
			APIFuncMock: func() sub.APIDescribeFunc {
				return func(_ context.Context, _ *fastly.Client, i *computeacls.DescribeInput) (*computeacls.ComputeACL, error) {
					require.Equal(t, aclID, *i.ComputeACLID, "unexpected ACL ID")

					return &acl, nil
				}
			},
			WantOutput: computeACL,
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--acl-id %s --json", aclID),
			APIFuncMock: func() sub.APIDescribeFunc {
				return func(_ context.Context, _ *fastly.Client, _ *computeacls.DescribeInput) (*computeacls.ComputeACL, error) {
					return &acl, nil
				}
			},
			WantOutput: fstfmt.EncodeJSON(acl),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "describe"}, scenarios)
}

func TestComputeACLList(t *testing.T) {
	acls := computeacls.ComputeACLs{
		Data: []computeacls.ComputeACL{
			{
				Name:         "foo",
				ComputeACLID: "bar",
			},
			{
				Name:         "foobar",
				ComputeACLID: "baz",
			},
		},
		Meta: computeacls.MetaACLs{
			Total: 2,
		},
	}

	scenarios := []testutil.CLIScenario[sub.APIListFunc]{
		{
			Name: "validate API success (zero compute ACLs, without mock)",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(computeacls.ComputeACLs{
							Data: []computeacls.ComputeACL{},
							Meta: computeacls.MetaACLs{
								Total: 0,
							},
						}))),
					},
				},
			},
			WantOutput: zeroComputeACLs,
		},
		{
			Name: "validate API success (zero compute ACLs)",
			APIFuncMock: func() sub.APIListFunc {
				return func(_ context.Context, _ *fastly.Client) (*computeacls.ComputeACLs, error) {
					return &computeacls.ComputeACLs{}, nil
				}
			},
			WantOutput: zeroComputeACLs,
		},
		{
			Name: "validate API success",
			APIFuncMock: func() sub.APIListFunc {
				return func(_ context.Context, _ *fastly.Client) (*computeacls.ComputeACLs, error) {
					return &acls, nil
				}
			},
			WantOutput: computeACLs,
		},
		{
			Name: "validate optional --json flag",
			Args: "--json",
			APIFuncMock: func() sub.APIListFunc {
				return func(_ context.Context, _ *fastly.Client) (*computeacls.ComputeACLs, error) {
					return &acls, nil
				}
			},
			WantOutput: fstfmt.EncodeJSON(acls),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "list-acls"}, scenarios)
}

func TestComputeACLLookup(t *testing.T) {
	const (
		aclID        = "foo"
		aclIP        = "1.2.3.4"
		aclNoMatchIP = "192.168.0.0"
	)

	entry := computeacls.ComputeACLEntry{
		Prefix: "1.2.3.4/32",
		Action: "ALLOW",
	}

	scenarios := []testutil.CLIScenario[sub.APILookupFunc]{
		{
			Name:      "validate missing --ip flag",
			Args:      fmt.Sprintf("--acl-id %s", aclID),
			WantError: "error parsing arguments: required flag --ip not provided",
		},
		{
			Name:      "validate missing --acl-id flag",
			Args:      fmt.Sprintf("--ip %s", aclIP),
			WantError: "error parsing arguments: required flag --acl-id not provided",
		},
		{
			Name: "validate 'no match found' (without mock)",
			Args: fmt.Sprintf("--acl-id %s --ip %s", aclID, aclNoMatchIP),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusNoContent,
						Status:     http.StatusText(http.StatusNoContent),
					},
				},
			},
			WantOutput: fstfmt.Info("Compute ACL (%s) has no entry with IP (%s)", aclID, aclNoMatchIP),
		},
		{
			Name: "validate 'no match found' with --json flag",
			Args: fmt.Sprintf("--acl-id %s --ip %s --json", aclID, aclNoMatchIP),
			APIFuncMock: func() sub.APILookupFunc {
				return func(_ context.Context, _ *fastly.Client, i *computeacls.LookupInput) (*computeacls.ComputeACLEntry, error) {
					require.Equal(t, aclID, *i.ComputeACLID, "unexpected ACL ID")

					require.Equal(t, aclNoMatchIP, *i.ComputeACLIP, "unexpected ACL match IP")

					return nil, nil
				}
			},
			WantOutput: fstfmt.EncodeJSON(nil),
		},
		{
			Name: "validate 'match found",
			Args: fmt.Sprintf("--acl-id %s --ip %s", aclID, aclIP),
			APIFuncMock: func() sub.APILookupFunc {
				return func(_ context.Context, _ *fastly.Client, _ *computeacls.LookupInput) (*computeacls.ComputeACLEntry, error) {
					return &entry, nil
				}
			},
			WantOutput: computeACLEntry,
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--acl-id %s --ip %s --json", aclID, aclIP),
			APIFuncMock: func() sub.APILookupFunc {
				return func(_ context.Context, _ *fastly.Client, _ *computeacls.LookupInput) (*computeacls.ComputeACLEntry, error) {
					return &entry, nil
				}
			},
			WantOutput: fstfmt.EncodeJSON(&entry),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "lookup"}, scenarios)
}

func TestComputeACLUpdate(t *testing.T) {
	const aclID = "foo"
	const aclOperation = "create"
	const aclPrefix = "1.2.3.0/24"
	const aclAction = "BLOCK"

	scenarios := []testutil.CLIScenario[sub.APIUpdateFunc]{
		{
			Name:      "validate missing --acl-id flag",
			WantError: "error parsing arguments: required flag --acl-id not provided",
		},
		{
			Name:      "validate invalid --file and --operation combination",
			Args:      fmt.Sprintf("--acl-id %s --file testdata/entries.json --operation %s", aclID, aclOperation),
			WantError: "invalid flag combination, --file and --operation",
		},
		{
			Name:      "validate invalid --file and --prefix combination",
			Args:      fmt.Sprintf("--acl-id %s --file testdata/entries.json --prefix %s", aclID, aclPrefix),
			WantError: "invalid flag combination, --file and --prefix",
		},
		{
			Name:      "validate invalid --file and --action combination",
			Args:      fmt.Sprintf("--acl-id %s --file testdata/entries.json --action %s", aclID, aclAction),
			WantError: "invalid flag combination, --file and --action",
		},
		{
			Name: "validate API success for updating a single entry (without mock)",
			Args: fmt.Sprintf("--acl-id %s --operation %s --prefix %s --action %s", aclID, aclOperation, aclPrefix, aclAction),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusAccepted,
						Status:     http.StatusText(http.StatusAccepted),
					},
				},
			},
			WantOutput: fstfmt.Success("Updated compute ACL entry (prefix: %s, id: %s)", aclPrefix, aclID),
		},
		{
			Name: "validate API success for updating a single entry",
			Args: fmt.Sprintf("--acl-id %s --operation %s --prefix %s --action %s", aclID, aclOperation, aclPrefix, aclAction),
			APIFuncMock: func() sub.APIUpdateFunc {
				return func(_ context.Context, _ *fastly.Client, i *computeacls.UpdateInput) error {
					require.Equal(t, aclID, *i.ComputeACLID, "unexpected ACL ID")

					require.Len(t, i.Entries, 1, "unexpected number of ACL entries")
					require.Equal(t, aclOperation, *i.Entries[0].Operation, "unexpected ACL operation")
					require.Equal(t, aclPrefix, *i.Entries[0].Prefix, "unexpected ACL prefix")
					require.Equal(t, aclAction, *i.Entries[0].Action, "unexpected ACL action")

					return nil
				}
			},
			WantOutput: fstfmt.Success("Updated compute ACL entry (prefix: %s, id: %s)", aclPrefix, aclID),
		},
		{
			Name:      "validate error from --file set with invalid json",
			Args:      fmt.Sprintf(`--acl-id %s --file {"foo":"bar"}`, aclID),
			WantError: "missing 'entries' {\"foo\":\"bar\"}",
		},
		{
			Name:      "validate error from --file set with zero json entries",
			Args:      fmt.Sprintf(`--acl-id %s --file {"entries":[]}`, aclID),
			WantError: "missing 'entries' {\"entries\":[]}",
		},
		{
			Name: "validate success with --file",
			Args: fmt.Sprintf("--acl-id %s --file testdata/entries.json", aclID),
			APIFuncMock: func() sub.APIUpdateFunc {
				return func(_ context.Context, _ *fastly.Client, i *computeacls.UpdateInput) error {
					require.Equal(t, aclID, *i.ComputeACLID, "unexpected ACL ID")

					require.Len(t, i.Entries, 4, "unexpected number of ACL entries")

					require.Equal(t, "create", *i.Entries[0].Operation, "unexpected ACL operation for entry 1")
					require.Equal(t, "1.2.3.0/24", *i.Entries[0].Prefix, "unexpected ACL prefix for entry 1")
					require.Equal(t, "BLOCK", *i.Entries[0].Action, "unexpected ACL action for entry 1")

					require.Equal(t, "update", *i.Entries[1].Operation, "unexpected ACL operation for entry 2")
					require.Equal(t, "192.168.0.0/16", *i.Entries[1].Prefix, "unexpected ACL prefix for entry 2")
					require.Equal(t, "BLOCK", *i.Entries[1].Action, "unexpected ACL action for entry 2")

					require.Equal(t, "create", *i.Entries[2].Operation, "unexpected ACL operation for entry 3")
					require.Equal(t, "23.23.23.23/32", *i.Entries[2].Prefix, "unexpected ACL prefix for entry 3")
					require.Equal(t, "ALLOW", *i.Entries[2].Action, "unexpected ACL action for entry 3")

					require.Equal(t, "update", *i.Entries[3].Operation, "unexpected ACL operation for entry 4")
					require.Equal(t, "1.2.3.4/32", *i.Entries[3].Prefix, "unexpected ACL prefix for entry 4")
					require.Equal(t, "ALLOW", *i.Entries[3].Action, "unexpected ACL action for entry 4")

					return nil
				}
			},
			WantOutput: fstfmt.Success("Updated %d compute ACL entries (id: %s)", 4, aclID),
		},
		{
			Name: "validate success with --file as inline json",
			Args: fmt.Sprintf(`--acl-id %s --file {"entries":[{"op":"%s","prefix":"%s","action":"%s"}]}`, aclID, aclOperation, aclPrefix, aclAction),
			APIFuncMock: func() sub.APIUpdateFunc {
				return func(_ context.Context, _ *fastly.Client, i *computeacls.UpdateInput) error {
					require.Equal(t, aclID, *i.ComputeACLID, "unexpected ACL ID")

					require.Len(t, i.Entries, 1, "unexpected number of ACL entries")
					require.Equal(t, aclOperation, *i.Entries[0].Operation, "unexpected ACL operation")
					require.Equal(t, aclPrefix, *i.Entries[0].Prefix, "unexpected ACL prefix")
					require.Equal(t, aclAction, *i.Entries[0].Action, "unexpected ACL action")

					return nil
				}
			},
			WantOutput: fstfmt.Success("Updated %d compute ACL entries (id: %s)", 1, aclID),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "update"}, scenarios)
}

func TestComputeACLListEntries(t *testing.T) {
	const aclID = "foo"

	entries := &computeacls.ComputeACLEntries{
		Entries: []computeacls.ComputeACLEntry{
			{
				Prefix: "1.2.3.0/24",
				Action: "BLOCK",
			},
			{
				Prefix: "1.2.3.4/32",
				Action: "ALLOW",
			},
			{
				Prefix: "23.23.23.23/32",
				Action: "ALLOW",
			},
			{
				Prefix: "192.168.0.0/16",
				Action: "BLOCK",
			},
		},
		Meta: computeacls.MetaEntries{
			Limit: 100,
		},
	}

	scenarios := []testutil.CLIScenario[sub.APIListEntriesFunc]{
		{
			Name:      "validate missing --acl-id flag",
			WantError: "error parsing arguments: required flag --acl-id not provided",
		},
		{
			Name: "validate API success (zero compute ACL entries, without mock)",
			Args: fmt.Sprintf("--acl-id %s", aclID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(computeacls.ComputeACLEntries{
							Entries: []computeacls.ComputeACLEntry{},
							Meta: computeacls.MetaEntries{
								Limit: 100,
							},
						}))),
					},
				},
			},
			WantOutput: zeroComputeACLEntries,
		},
		{
			Name: "validate API success (zero compute ACL entries)",
			Args: fmt.Sprintf("--acl-id %s", aclID),
			APIFuncMock: func() sub.APIListEntriesFunc {
				return func(_ context.Context, _ *fastly.Client, i *computeacls.ListEntriesInput) (*computeacls.ComputeACLEntries, error) {
					require.Equal(t, aclID, *i.ComputeACLID, "unexpected ACL ID")

					return &computeacls.ComputeACLEntries{}, nil
				}
			},
			WantOutput: zeroComputeACLEntries,
		},
		{
			Name: "validate API success",
			Args: fmt.Sprintf("--acl-id %s", aclID),
			APIFuncMock: func() sub.APIListEntriesFunc {
				return func(_ context.Context, _ *fastly.Client, _ *computeacls.ListEntriesInput) (*computeacls.ComputeACLEntries, error) {
					return entries, nil
				}
			},
			WantOutput: computeACLEntries,
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--acl-id %s --json", aclID),
			APIFuncMock: func() sub.APIListEntriesFunc {
				return func(_ context.Context, _ *fastly.Client, _ *computeacls.ListEntriesInput) (*computeacls.ComputeACLEntries, error) {
					return entries, nil
				}
			},
			WantOutput: fstfmt.EncodeJSON(entries.Entries),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "list-entries"}, scenarios)
}

var computeACL = strings.TrimSpace(`
ID: bar
Name: foo
`) + "\n"

var computeACLs = strings.TrimSpace(`
Name    ID
foo     bar
foobar  baz
`) + "\n"

var zeroComputeACLs = strings.TrimSpace(`
Name  ID
`) + "\n"

var computeACLEntry = strings.TrimSpace(`
Prefix: 1.2.3.4/32
Action: ALLOW
`) + "\n"

var computeACLEntries = strings.TrimSpace(`
Prefix          Action
1.2.3.0/24      BLOCK
1.2.3.4/32      ALLOW
23.23.23.23/32  ALLOW
192.168.0.0/16  BLOCK
`) + "\n"

var zeroComputeACLEntries = strings.TrimSpace(`
Prefix  Action
`) + "\n"
