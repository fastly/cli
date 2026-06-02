package zone_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v15/fastly"
	"github.com/fastly/go-fastly/v15/fastly/dns/v1/dnszones"

	dnsroot "github.com/fastly/cli/pkg/commands/dns"
	root "github.com/fastly/cli/pkg/commands/dns/zone"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/cli/pkg/text"
)

const (
	zoneID   = "zone-id-123"
	zoneName = "example.com."
	zoneType = "secondary"
)

var testZone = dnszones.Zone{
	ID:   fastly.ToPointer(zoneID),
	Name: fastly.ToPointer(zoneName),
	Type: fastly.ToPointer(zoneType),
}

func TestZoneCreate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: "--name example.com",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusCreated,
						Status:     http.StatusText(http.StatusCreated),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(testZone))),
					},
				},
			},
			WantOutput: fmt.Sprintf("SUCCESS: Created DNS zone '%s' (zone-id: %s)", zoneName, zoneID),
		},
		{
			Args:      "--name `example .com`",
			WantError: "zone names cannot contain spaces",
		},
		{
			Args:      "--name " + strings.Repeat("a", 256) + ".com",
			WantError: "zone names cannot exceed 255 characters",
		},
		{
			Args:      "--verbose --json --name example.com",
			WantError: "invalid flag combination, --verbose and --json",
		},
		{
			Args:      "--name example.com --primary-description foo",
			WantError: "--primary-description requires --primary-address",
		},
		{
			Args:      "--name example.com --primary-address 1.2.3.4 --primary-description foo --primary-description bar",
			WantError: "--primary-description cannot be provided more times than --primary-address",
		},
		{
			Args: "--name example.com --json",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusCreated,
						Status:     http.StatusText(http.StatusCreated),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(testZone))),
					},
				},
			},
			WantOutput: string(testutil.GenJSON(testZone)),
		},
		{
			Args: "--name example.com",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusBadRequest,
						Status:     http.StatusText(http.StatusBadRequest),
						Body:       io.NopCloser(strings.NewReader(`{"errors": [{"title": "Invalid parameters"}]}`)),
					},
				},
			},
			WantError: "400 - Bad Request",
		},
	}
	testutil.RunCLIScenarios(t, []string{dnsroot.CommandName, root.CommandName, "create"}, scenarios)
}

func TestZoneDescribe(t *testing.T) {
	resp := testutil.GenJSON(testZone)

	scenarios := []testutil.CLIScenario{
		{
			Args:      "",
			WantError: "error parsing arguments: required flag --zone-id not provided",
		},
		{
			Args:      "--verbose --json --zone-id " + zoneID,
			WantError: "invalid flag combination, --verbose and --json",
		},
		{
			Args: "--zone-id " + zoneID,
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(resp)),
					},
				},
			},
			WantOutput: fmtZone(&testZone),
		},
		{
			Args: "--zone-id " + zoneID + " --json",
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
			Args: "--zone-id " + zoneID,
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusNotFound,
						Status:     http.StatusText(http.StatusNotFound),
						Body:       io.NopCloser(strings.NewReader(`{"errors": [{"title": "Not Found"}]}`)),
					},
				},
			},
			WantError: "404 - Not Found",
		},
	}
	testutil.RunCLIScenarios(t, []string{dnsroot.CommandName, root.CommandName, "describe"}, scenarios)
}

func TestZoneDelete(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "",
			WantError: "error parsing arguments: required flag --zone-id not provided",
		},
		{
			Args: "--zone-id " + zoneID,
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusNoContent,
						Status:     http.StatusText(http.StatusNoContent),
						Body:       http.NoBody,
					},
				},
			},
			WantOutput: fmt.Sprintf("SUCCESS: Deleted DNS zone (zone-id: %s)", zoneID),
		},
		{
			Args: "--zone-id " + zoneID + " --json",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusNoContent,
						Status:     http.StatusText(http.StatusNoContent),
						Body:       http.NoBody,
					},
				},
			},
			WantOutput: fmt.Sprintf(`"id": %q`, zoneID),
		},
		{
			Args: "--zone-id " + zoneID,
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusNotFound,
						Status:     http.StatusText(http.StatusNotFound),
						Body:       io.NopCloser(strings.NewReader(`{"errors": [{"title": "Not Found"}]}`)),
					},
				},
			},
			WantError: "404 - Not Found",
		},
	}
	testutil.RunCLIScenarios(t, []string{dnsroot.CommandName, root.CommandName, "delete"}, scenarios)
}

func TestZoneList(t *testing.T) {
	zones := []dnszones.Zone{testZone}
	resp := testutil.GenJSON(dnszones.Zones{
		Data: zones,
		Meta: dnszones.MetaZones{},
	})

	scenarios := []testutil.CLIScenario{
		{
			Args:      "--verbose --json",
			WantError: "invalid flag combination, --verbose and --json",
		},
		{
			Args: "",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(resp)),
					},
				},
			},
			WantOutput: zoneName,
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
			WantOutput: string(testutil.GenJSON(zones)),
		},
		{
			Args: "",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusBadRequest,
						Status:     http.StatusText(http.StatusBadRequest),
						Body:       io.NopCloser(strings.NewReader(`{"errors": [{"title": "Bad Request"}]}`)),
					},
				},
			},
			WantError: "400 - Bad Request",
		},
	}
	testutil.RunCLIScenarios(t, []string{dnsroot.CommandName, root.CommandName, "list"}, scenarios)
}

func TestZoneUpdate(t *testing.T) {
	updated := dnszones.Zone{
		ID:          fastly.ToPointer(zoneID),
		Name:        fastly.ToPointer(zoneName),
		Type:        fastly.ToPointer(zoneType),
		Description: fastly.ToPointer("updated description"),
	}

	scenarios := []testutil.CLIScenario{
		{
			Args:      "",
			WantError: "error parsing arguments: required flag --zone-id not provided",
		},
		{
			Args:      "--verbose --json --zone-id " + zoneID,
			WantError: "invalid flag combination, --verbose and --json",
		},
		{
			Args:      "--zone-id " + zoneID + " --primary-description foo",
			WantError: "--primary-description requires --primary-address",
		},
		{
			Args:      "--zone-id " + zoneID + " --primary-address 1.2.3.4 --primary-description foo --primary-description bar",
			WantError: "--primary-description cannot be provided more times than --primary-address",
		},
		{
			Args: "--zone-id " + zoneID + " --description updated-description",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(updated))),
					},
				},
			},
			WantOutput: fmt.Sprintf("SUCCESS: Updated DNS zone '%s' (zone-id: %s)", zoneName, zoneID),
		},
		{
			Args: "--zone-id " + zoneID + " --json",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(updated))),
					},
				},
			},
			WantOutput: string(testutil.GenJSON(updated)),
		},
		{
			Args: "--zone-id " + zoneID,
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusBadRequest,
						Status:     http.StatusText(http.StatusBadRequest),
						Body:       io.NopCloser(strings.NewReader(`{"errors": [{"title": "Bad Request"}]}`)),
					},
				},
			},
			WantError: "400 - Bad Request",
		},
	}
	testutil.RunCLIScenarios(t, []string{dnsroot.CommandName, root.CommandName, "update"}, scenarios)
}

func fmtZone(z *dnszones.Zone) string {
	var b bytes.Buffer
	text.PrintDNSZone(&b, "", z)
	return b.String()
}
