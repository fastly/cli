package domainv1_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
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
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, "create"}, scenarios)
}
