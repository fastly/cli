package setup_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/fastly/cli/pkg/commands/compute/setup"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/cli/pkg/text"
)

// TestProductsCreate tests the `Create` method of the `Products` struct.
func TestProductsCreate(t *testing.T) {
	scenarios := []struct {
		name             string
		setup            *manifest.SetupProducts
		client           *mock.HTTPClient
		wantOutput       string
		wantError        string
		expectedRequests []testutil.ExpectedRequest
	}{
		{
			name: "successfully enables a single product",
			setup: &manifest.SetupProducts{
				APIDiscovery: &manifest.SetupProduct{
					Enable: true,
				},
			},
			client: mock.NewHTTPClientWithResponses([]*http.Response{
				{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(`{"id":"api_discovery","name":"API Discovery"}`)),
				},
			}),
			expectedRequests: []testutil.ExpectedRequest{
				{
					Method: http.MethodPut,
					Path:   "/enabled-products/v1/api_discovery/services/123",
				},
			},
			wantOutput: "Enabling product 'api_discovery'...",
		},
		{
			name: "successfully enables multiple products",
			setup: &manifest.SetupProducts{
				APIDiscovery: &manifest.SetupProduct{
					Enable: true,
				},
				OriginInspector: &manifest.SetupProduct{
					Enable: true,
				},
			},
			client: mock.NewHTTPClientWithResponses([]*http.Response{
				{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(`{"id":"some_product","name":"Some Product"}`)),
				},
				{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(`{"id":"some_product","name":"Some Product"}`)),
				},
			}),
			expectedRequests: []testutil.ExpectedRequest{
				{
					Method: http.MethodPut,
					Path:   "/enabled-products/v1/api_discovery/services/123",
				},
				{
					Method: http.MethodPut,
					Path:   "/enabled-products/v1/origin_inspector/services/123",
				},
			},
		},
		{
			name: "handles API error when enabling a product",
			setup: &manifest.SetupProducts{
				APIDiscovery: &manifest.SetupProduct{
					Enable: true,
				},
			},
			client: mock.NewHTTPClientWithErrors([]error{
				testutil.Err,
			}),
			expectedRequests: []testutil.ExpectedRequest{
				{
					Method: http.MethodPut,
					Path:   "/enabled-products/v1/api_discovery/services/123",
				},
			},
			wantError: "error enabling product [api_discovery]: Put \"https://api.example.com/enabled-products/v1/api_discovery/services/123\": test error",
		},
		{
			name:  "no API calls when no products are configured",
			setup: &manifest.SetupProducts{},
			client: mock.NewHTTPClientWithResponses([]*http.Response{
				{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(`{"id":"some_product","name":"Some Product"}`)),
				},
			}),
		},
		{
			name: "successfully enables ngwaf with workspace id",
			setup: &manifest.SetupProducts{
				NGWAF: &manifest.SetupProductNGWAF{
					SetupProduct: manifest.SetupProduct{
						Enable: true,
					},
					WorkspaceID: "w-123",
				},
			},
			client: mock.NewHTTPClientWithResponses([]*http.Response{
				{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(`{"id":"ngwaf","name":"Next-Gen WAF"}`)),
				},
				{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(`{"id":"some_product","name":"Some Product"}`)),
				},
			}),
			wantOutput: "Enabling product 'ngwaf'...",
			expectedRequests: []testutil.ExpectedRequest{
				{
					Method:   http.MethodPut,
					Path:     "/enabled-products/v1/ngwaf/services/123",
					WantJSON: testutil.StrPtr("{\"workspace_id\":\"w-123\"}"),
				},
			},
		},
	}

	for _, testcase := range scenarios {
		t.Run(testcase.name, func(t *testing.T) {
			apiClient, err := mock.NewFastlyClient(testcase.client)
			if err != nil {
				t.Fatal(fmt.Errorf("failed to mock fastly.client: %w", err))
			}

			var out bytes.Buffer
			spinner, err := text.NewSpinner(&out)
			if err != nil {
				t.Fatal(err)
			}

			products := setup.Products{
				APIClient: apiClient,
				ServiceID: "123",
				Spinner:   spinner,
				Stdout:    &out,
				Setup:     testcase.setup,
			}

			err = products.Configure()
			testutil.AssertNoError(t, err)

			err = products.Create()

			if testcase.wantError != "" {
				testutil.AssertErrorContains(t, err, testcase.wantError)
			} else {
				testutil.AssertNoError(t, err)
				if testcase.wantOutput != "" {
					testutil.AssertStringContains(t, out.String(), testcase.wantOutput)
				}
			}

			if len(testcase.expectedRequests) != len(testcase.client.Requests) {
				t.Errorf("expected %d API calls, but got %d", len(testcase.expectedRequests), len(testcase.client.Requests))
			}

			for i, expectedRequest := range testcase.expectedRequests {
				testutil.AssertRequest(t, &testcase.client.Requests[i], expectedRequest)
			}
		})
	}
}
