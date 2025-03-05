package accesskeys_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	root "github.com/fastly/cli/pkg/commands/objectstorage"
	sub "github.com/fastly/cli/pkg/commands/objectstorage/accesskeys"
	fstfmt "github.com/fastly/cli/pkg/fmt"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v9/fastly/objectstorage/accesskeys"
)

const (
	akID          = "accessKeyId"
	akSecret      = "accessKeySecret"
	akDescription = "accessKeyDescription"
	akPermission  = "read-only-objects"
)

var ak = accesskeys.AccessKey{
	AccessKeyID: akID,
	SecretKey:   akSecret,
	Description: akDescription,
	Permission:  akPermission,
	CreatedAt:   testutil.Date,
}

func TestAccessKeysCreate(t *testing.T) {

	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --description flag",
			Args:      fmt.Sprintf("--permission %s", akPermission),
			WantError: "error parsing arguments: required flag --description not provided",
		},
		{
			Name:      "validate missing --permission flag",
			Args:      fmt.Sprintf("--description %s", akDescription),
			WantError: "error parsing arguments: required flag --permission not provided",
		},
		{
			Name: "validate internal server error",
			Args: fmt.Sprintf("--description %s --permission %s", akDescription, akPermission),
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
			Args: fmt.Sprintf("--description %s --permission %s", akDescription, akPermission),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader((testutil.GenJSON(ak)))),
					},
				},
			},
			WantOutput: fstfmt.Success("Created access key (id: %s, secret: %s)", akID, ak.SecretKey),
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--description %s --permission %s --json", akDescription, akPermission),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(ak))),
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(ak),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "create"}, scenarios)
}

func TestAccessKeysDelete(t *testing.T) {
	const accessKeyID = "accessKeyID"

	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --ak-id flag",
			Args:      "",
			WantError: "error parsing arguments: required flag --ak-id not provided",
		},
		{
			Name: "validate bad request",
			Args: "--ak-id bar",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusBadRequest,
						Status:     http.StatusText(http.StatusBadRequest),
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(`
							{
    							"title": "invalid Acess Key ID",
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
			Args: fmt.Sprintf("--ak-id %s", accessKeyID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusNoContent,
						Status:     http.StatusText(http.StatusNoContent),
					},
				},
			},
			WantOutput: fstfmt.Success("Deleted access key (id: %s)", accessKeyID),
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--ak-id %s --json", accessKeyID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusNoContent,
						Status:     http.StatusText(http.StatusNoContent),
					},
				},
			},
			WantOutput: fstfmt.JSON(`{"id": %q, "deleted": true}`, accessKeyID),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "delete"}, scenarios)
}

func TestAccessKeysGet(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --ak-id flag",
			Args:      "",
			WantError: "error parsing arguments: required flag --ak-id not provided",
		},
		{
			Name: "validate bad request",
			Args: "--ak-id baz",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusBadRequest,
						Status:     http.StatusText(http.StatusBadRequest),
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(`
							{
    							"title": "invalid Acess Key ID",
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
			Args: fmt.Sprintf("--ak-id %s", akID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader((testutil.GenJSON(ak)))),
					},
				},
			},
			WantOutput: akString,
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--ak-id %s --json", akID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader((testutil.GenJSON(ak)))),
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(ak),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "get"}, scenarios)
}

func TestAccessKeysList(t *testing.T) {
	acesskeysobject := accesskeys.AccessKeys{
		Data: []accesskeys.AccessKey{
			{
				AccessKeyID: "foo",
				SecretKey:   "bar",
				Description: "bat",
				Permission:  akPermission,
			},
			{
				AccessKeyID: "foobar",
				SecretKey:   "baz",
				Description: "bizz",
				Permission:  akPermission,
			},
		},
		Meta: accesskeys.MetaAccessKeys{},
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
			Name: "validate API success (zero access keys)",
			Args: "",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(accesskeys.AccessKeys{
							Data: []accesskeys.AccessKey{},
							Meta: accesskeys.MetaAccessKeys{},
						}))),
					},
				},
			},
			WantOutput: zeroListAccessKeysString,
		},
		{
			Name: "validate API success",
			Args: "",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(acesskeysobject))),
					},
				},
			},
			WantOutput: listAccessKeysString,
		},
		{
			Name: "validate optional --json flag",
			Args: "--json",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(acesskeysobject))),
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(acesskeysobject),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "list-access-keys"}, scenarios)
}

var akString = strings.TrimSpace(`
ID: accessKeyId
Secret: accessKeySecret
Description: accessKeyDescription
Permission: read-only-objects
Buckets: []
Created (UTC): 2021-06-15 23:00
`) + "\n"

var listAccessKeysString = strings.TrimSpace(`
ID      Secret  Description  Permssion          Buckets  Created At
foo     bar     bat          read-only-objects  []       0001-01-01 00:00:00 +0000 UTC
foobar  baz     bizz         read-only-objects  []       0001-01-01 00:00:00 +0000 UTC
`) + "\n"

var zeroListAccessKeysString = strings.TrimSpace(`
ID  Secret  Description  Permssion  Buckets  Created At
`) + "\n"
