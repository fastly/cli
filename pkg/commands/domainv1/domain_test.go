package domainv1_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	v1 "github.com/fastly/go-fastly/v9/fastly/domains/v1"

	root "github.com/fastly/cli/pkg/commands/domainv1"
	"github.com/fastly/cli/pkg/testutil"
)

func TestDomainV1Create(t *testing.T) {
	fqdn := "www.example.com"
	sid := "123"
	did := "domain-id"

	scenarios := []testutil.CLIScenario{
		{
			Args:      "",
			WantError: "error parsing arguments: required flag --fqdn not provided",
		},
		{
			Args: fmt.Sprintf("--fqdn %s --service-id %s", fqdn, sid),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(v1.Data{
							DomainID:  did,
							FQDN:      fqdn,
							ServiceID: &sid,
						}))),
					},
				},
			},
			WantOutput: fmt.Sprintf("SUCCESS: Created domain '%s' (domain-id: %s, service-id: %s)", fqdn, did, sid),
		},
		{
			Args: fmt.Sprintf("--fqdn %s", fqdn),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(v1.Data{
							DomainID: did,
							FQDN:     fqdn,
						}))),
					},
				},
			},
			WantOutput: fmt.Sprintf("SUCCESS: Created domain '%s' (domain-id: %s)", fqdn, did),
		},
		{
			Args: fmt.Sprintf("--fqdn %s", fqdn),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusBadRequest,
						Status:     http.StatusText(http.StatusBadRequest),
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(`
							{
							  "errors":[
								{
								  "title":"Invalid value for fqdn",
								  "detail":"fqdn has already been taken"
								}
							  ]
							}
						`))),
					},
				},
			},
			WantError: "400 - Bad Request",
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, "create"}, scenarios)
}

func TestDomainV1List(t *testing.T) {
	fqdn := "www.example.com"
	sid := "123"
	did := "domain-id"

	resp := testutil.GenJSON(v1.Collection{
		Data: []v1.Data{
			{
				DomainID:  did,
				FQDN:      fqdn,
				ServiceID: &sid,
			},
		},
	})

	scenarios := []testutil.CLIScenario{
		{
			Args:      "--verbose --json",
			WantError: "invalid flag combination, --verbose and --json",
		},
		{
			Args: "--json",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(resp)),
					},
				},
			},
			WantOutput: string(resp),
		},
		{
			Args: "",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusBadRequest,
						Status:     http.StatusText(http.StatusBadRequest),
						Body:       io.NopCloser(strings.NewReader(`{"error": "whoops"}`)),
					},
				},
			},
			WantError: "400 - Bad Request",
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, "list"}, scenarios)
}

func TestDomainV1Describe(t *testing.T) {
	fqdn := "www.example.com"
	sid := "123"
	did := "domain-id"

	resp := testutil.GenJSON(v1.Data{
		DomainID:  did,
		FQDN:      fqdn,
		ServiceID: &sid,
	})

	scenarios := []testutil.CLIScenario{
		{
			Args:      "",
			WantError: "error parsing arguments: required flag --domain-id not provided",
		},
		{
			Args: fmt.Sprintf("--domain-id %s --json", did),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(resp)),
					},
				},
			},
			WantOutput: string(resp),
		},
		{
			Args: fmt.Sprintf("--domain-id %s --json", did),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusBadRequest,
						Status:     http.StatusText(http.StatusBadRequest),
						Body:       io.NopCloser(strings.NewReader(`{"error": "whoops"}`)),
					},
				},
			},
			WantError: "400 - Bad Request",
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, "describe"}, scenarios)
}

func TestDomainV1Update(t *testing.T) {
	fqdn := "www.example.com"
	sid := "123"
	did := "domain-id"

	scenarios := []testutil.CLIScenario{
		{
			Args:      "",
			WantError: "error parsing arguments: required flag --domain-id not provided",
		},
		{
			Args: fmt.Sprintf("--domain-id %s --service-id %s", did, sid),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(v1.Data{
							DomainID:  did,
							FQDN:      fqdn,
							ServiceID: &sid,
						}))),
					},
				},
			},
			WantOutput: fmt.Sprintf("SUCCESS: Updated domain '%s' (domain-id: %s, service-id: %s)", fqdn, did, sid),
		},
		{
			Args: fmt.Sprintf("--domain-id %s", did),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(v1.Data{
							DomainID: did,
							FQDN:     fqdn,
						}))),
					},
				},
			},
			WantOutput: fmt.Sprintf("SUCCESS: Updated domain '%s' (domain-id: %s)", fqdn, did),
		},
		{
			Args: fmt.Sprintf("--domain-id %s", did),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusBadRequest,
						Status:     http.StatusText(http.StatusBadRequest),
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(`
							{
							  "errors":[
								{
								  "title":"Invalid value for domain-id",
								  "detail":"whoops"
								}
							  ]
							}
						`))),
					},
				},
			},
			WantError: "400 - Bad Request",
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, "update"}, scenarios)
}

func TestDomainV1Delete(t *testing.T) {
	did := "domain-id"

	scenarios := []testutil.CLIScenario{
		{
			Args:      "",
			WantError: "error parsing arguments: required flag --domain-id not provided",
		},
		{
			Args: fmt.Sprintf("--domain-id %s", did),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusNoContent,
						Status:     http.StatusText(http.StatusNoContent),
					},
				},
			},
			WantOutput: fmt.Sprintf("SUCCESS: Deleted domain (domain-id: %s)", did),
		},
		{
			Args: fmt.Sprintf("--domain-id %s", did),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusBadRequest,
						Status:     http.StatusText(http.StatusBadRequest),
						Body:       io.NopCloser(strings.NewReader(`{"error": "whoops"}`)),
					},
				},
			},
			WantError: "400 - Bad Request",
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, "delete"}, scenarios)
}
