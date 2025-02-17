package computeacl_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"testing"

	root "github.com/fastly/cli/pkg/commands/compute"
	sub "github.com/fastly/cli/pkg/commands/compute/computeacl"
	fstfmt "github.com/fastly/cli/pkg/fmt"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v9/fastly/computeacls"
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

	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --name flag",
			Args:      "",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Name: "validate internal server error",
			Args: fmt.Sprintf("--name %s", aclName),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusInternalServerError,
						Status:     http.StatusText(http.StatusInternalServerError),
					},
				},
			},
			WantError: "500 - Internal Server Error",
		},
		{
			Name: "validate API success",
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
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--name %s --json", aclName),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(acl))),
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(acl),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "create"}, scenarios)
}

func TestComputeACLDelete(t *testing.T) {
	const aclID = "foo"

	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --acl-id flag",
			Args:      "",
			WantError: "error parsing arguments: required flag --acl-id not provided",
		},
		{
			Name: "validate bad request",
			Args: "--acl-id bar",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusBadRequest,
						Status:     http.StatusText(http.StatusBadRequest),
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(`
							{
    							"title": "invalid ACL ID",
    							"status": 400
							}
						`))),
					},
				},
			},
			WantError: "400 - Bad Request",
		},
		{
			Name: "validate API success",
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
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--acl-id %s --json", aclID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
					},
				},
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

	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --acl-id flag",
			Args:      "",
			WantError: "error parsing arguments: required flag --acl-id not provided",
		},
		{
			Name: "validate bad request",
			Args: "--acl-id baz",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusBadRequest,
						Status:     http.StatusText(http.StatusBadRequest),
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(`
							{
    							"title": "invalid ACL ID",
    							"status": 400
							}
						`))),
					},
				},
			},
			WantError: "400 - Bad Request",
		},
		{
			Name: "validate API success",
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
			WantOutput: fmtComputeACL(&acl),
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--acl-id %s --json", aclID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader((testutil.GenJSON(acl)))),
					},
				},
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

	scenarios := []testutil.CLIScenario{
		{
			Name: "validate internal server error",
			Args: "",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusInternalServerError,
						Status:     http.StatusText(http.StatusInternalServerError),
					},
				},
			},
			WantError: "500 - Internal Server Error",
		},
		{
			Name: "validate API success (empty list)",
			Args: "",
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
			WantOutput: fmtComputeACLs(nil),
		},
		{
			Name: "validate API success",
			Args: "",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(acls))),
					},
				},
			},
			WantOutput: fmtComputeACLs(acls.Data),
		},
		{
			Name: "validate optional --json flag",
			Args: "--json",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(acls))),
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(acls),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "list-acls"}, scenarios)
}

func TestComputeACLLookup(t *testing.T) {
	const (
		aclID = "foo"
		aclIP = "1.2.3.4"
	)

	entry := computeacls.ComputeACLEntry{
		Prefix: "1.2.3.4/32",
		Action: "ALLOW",
	}

	scenarios := []testutil.CLIScenario{
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
			Name: "validate bad request",
			Args: fmt.Sprintf("--acl-id baz --ip %s", aclIP),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusBadRequest,
						Status:     http.StatusText(http.StatusBadRequest),
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(`
							{
    							"title": "invalid ACL ID",
    							"status": 400
							}
						`))),
					},
				},
			},
			WantError: "400 - Bad Request",
		},
		{
			Name: "validate API status 204 - no content",
			Args: fmt.Sprintf("--acl-id %s --ip 192.168.0.0", aclID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusNoContent,
						Status:     http.StatusText(http.StatusNoContent),
					},
				},
			},
			WantOutput: fstfmt.Warning("Compute ACL (%s) has no entry with IP (192.168.0.0)", aclID),
		},
		{
			Name: "validate API success",
			Args: fmt.Sprintf("--acl-id %s --ip %s", aclID, aclIP),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader((testutil.GenJSON(entry)))),
					},
				},
			},
			WantOutput: fmtComputeACLEntry(&entry),
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--acl-id %s --ip %s --json", aclID, aclIP),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader((testutil.GenJSON(entry)))),
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(&entry),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "lookup"}, scenarios)
}

func TestComputeACLUpdate(t *testing.T) {
	const aclID = "foo"

	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --acl-id flag",
			Args:      "--file testdata/batch.json",
			WantError: "error parsing arguments: required flag --acl-id not provided",
		},
		{
			Name: "validate bad request",
			Args: "--acl-id bar --file testdata/entries.json",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusBadRequest,
						Status:     http.StatusText(http.StatusBadRequest),
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(`
							{
    							"title": "invalid ACL ID",
    							"status": 400
							}
						`))),
					},
				},
			},
			WantError: "400 - Bad Request",
		},
		{
			Name: "validate error from --file set with invalid json",
			Args: fmt.Sprintf(`--acl-id %s --file {"foo":"bar"}`, aclID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusBadRequest,
						Status:     http.StatusText(http.StatusBadRequest),
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(`
							{
								"title": "can't parse body",
								"status": 400,
								"detail": "missing field 'entries' at line 1 column 13"
							}
						`))),
					},
				},
			},
			WantError: "missing 'entries' {\"foo\":\"bar\"}",
		},
		{
			Name: "validate error from --file set with zero json entries",
			Args: fmt.Sprintf(`--acl-id %s --file {"entries":[]}`, aclID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusAccepted,
						Status:     http.StatusText(http.StatusAccepted),
					},
				},
			},
			WantError: "missing 'entries' {\"entries\":[]}",
		},
		{
			Name: "validate success with --file",
			Args: fmt.Sprintf("--acl-id %s --file testdata/entries.json", aclID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusAccepted,
						Status:     http.StatusText(http.StatusAccepted),
					},
				},
			},
			WantOutput: fstfmt.Success("Updated %d compute ACL entries (id: %s)", 4, aclID),
		},
		{
			Name: "validate success with --file as inline json",
			Args: fmt.Sprintf(`--acl-id %s --file {"entries":[{"op":"create","prefix":"1.2.3.0/24","action":"BLOCK"}]}`, aclID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusAccepted,
						Status:     http.StatusText(http.StatusAccepted),
					},
				},
			},
			WantOutput: fstfmt.Success("Updated %d compute ACL entries (id: %s)", 1, aclID),
		},
		{
			Name: "validate success for updating a single entry with --operation, --prefix, and --action",
			Args: fmt.Sprintf("--acl-id %s --operation create --prefix 1.2.3.0/24 --action BLOCK", aclID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusAccepted,
						Status:     http.StatusText(http.StatusAccepted),
					},
				},
			},
			WantOutput: fstfmt.Success("Updated compute ACL entry (prefix: 1.2.3.0/24, id: %s)", aclID),
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

	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --acl-id flag",
			Args:      "",
			WantError: "error parsing arguments: required flag --acl-id not provided",
		},
		{
			Name: "validate bad request",
			Args: "--acl-id bar",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusBadRequest,
						Status:     http.StatusText(http.StatusBadRequest),
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(`
							{
    							"title": "invalid ACL ID",
    							"status": 400
							}
						`))),
					},
				},
			},
			WantError: "400 - Bad Request",
		},
		{
			Name: "validate API success (empty list)",
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
			WantOutput: fmtComputeACLEntries(nil),
		},
		{
			Name: "validate API success",
			Args: fmt.Sprintf("--acl-id %s", aclID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(entries))),
					},
				},
			},
			WantOutput: fmtComputeACLEntries(entries.Entries),
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--acl-id %s --json", aclID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(entries))),
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(entries.Entries),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "list-entries"}, scenarios)
}

func fmtComputeACL(acl *computeacls.ComputeACL) string {
	var b bytes.Buffer
	text.PrintComputeACL(&b, "", acl)
	return b.String()
}

func fmtComputeACLs(acls []computeacls.ComputeACL) string {
	var b bytes.Buffer
	text.PrintComputeACLsTbl(&b, acls)
	return b.String()
}

func fmtComputeACLEntry(entry *computeacls.ComputeACLEntry) string {
	var b bytes.Buffer
	text.PrintComputeACLEntry(&b, "", entry)
	return b.String()
}

func fmtComputeACLEntries(entries []computeacls.ComputeACLEntry) string {
	var b bytes.Buffer
	text.PrintComputeACLEntriesTbl(&b, entries)
	return b.String()
}
