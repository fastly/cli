package setup_test

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/fastly/cli/pkg/commands/compute/setup"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v12/fastly"
)

// TestProductsCreate tests the `Create` method of the `Products` struct.
func TestProductsCreate(t *testing.T) {
	scenarios := []struct {
		name             string
		setup            *manifest.SetupProducts
		mockHTTP         *mockHTTPClient
		expectedError    string
		expectedOutput   string
		expectedAPIPaths []string
	}{
		{
			name: "successfully enables a single product",
			setup: &manifest.SetupProducts{
				APIDiscovery: &manifest.SetupProduct{
					Enable: true,
				},
			},
			mockHTTP: &mockHTTPClient{
				response: &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(`{"id":"api_discovery","name":"API Discovery"}`)),
				},
			},
			expectedOutput:   "Enabling product 'api_discovery'...",
			expectedAPIPaths: []string{"/service/123/product/api_discovery/enable"},
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
			mockHTTP: &mockHTTPClient{
				response: &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(`{"id":"some_product","name":"Some Product"}`)),
				},
			},
			expectedAPIPaths: []string{
				"/service/123/product/api_discovery/enable",
				"/service/123/product/origin_inspector/enable",
			},
		},
		{
			name: "handles API error when enabling a product",
			setup: &manifest.SetupProducts{
				APIDiscovery: &manifest.SetupProduct{
					Enable: true,
				},
			},
			mockHTTP: &mockHTTPClient{
				err: errors.New("api error"),
			},
			expectedError: "error enabling product [api_discovery]: api error",
		},
		{
			name:  "no API calls when no products are configured",
			setup: &manifest.SetupProducts{},
			mockHTTP: &mockHTTPClient{
				response: &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(`{"id":"some_product","name":"Some Product"}`)),
				},
			},
			expectedAPIPaths: []string{},
		},
		{
			name: "successfully enables ngwaf with workspace id",
			setup: &manifest.SetupProducts{
				Ngwaf: &manifest.SetupProductNgwaf{
					SetupProduct: manifest.SetupProduct{
						Enable: true,
					},
					WorkspaceID: "w-123",
				},
			},
			mockHTTP: &mockHTTPClient{
				response: &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(`{"id":"ngwaf","name":"Next-Gen WAF"}`)),
				},
			},
			expectedOutput:   "Enabling product 'ngwaf'...",
			expectedAPIPaths: []string{"/service/123/product/ngwaf/enable"},
		},
	}

	for _, testcase := range scenarios {
		t.Run(testcase.name, func(t *testing.T) {
			apiClient, err := fastly.NewClient(testcase.mockHTTP)
			if err != nil {
				t.Fatal(err)
			}

			var out bytes.Buffer
			products := setup.Products{
				APIClient: apiClient,
				ServiceID: "123",
				Spinner:   testutil.NewSpinner(&out),
				Stdout:    &out,
				Setup:     testcase.setup,
			}

			err = products.Configure()
			testutil.AssertNoError(t, err)

			err = products.Create()

			if testcase.expectedError != "" {
				testutil.AssertErrorContains(t, err, testcase.expectedError)
			} else {
				testutil.AssertNoError(t, err)
				if testcase.expectedOutput != "" {
					testutil.AssertStringContains(t, out.String(), testcase.expectedOutput)
				}
			}

			if len(testcase.expectedAPIPaths) != len(testcase.mockHTTP.called) {
				t.Errorf("expected %d API calls, but got %d", len(testcase.expectedAPIPaths), len(testcase.mockHTTP.called))
			}

			for i, path := range testcase.expectedAPIPaths {
				if i >= len(testcase.mockHTTP.called) {
					t.Errorf("expected API path %s to be called, but it was not", path)
					continue
				}
				testutil.AssertStringContains(t, testcase.mockHTTP.called[i], path)
			}
		})
	}
}

type mockHTTPClient struct {
	response *http.Response
	err      error
	called   []string
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	m.called = append(m.called, req.URL.Path)
	if m.err != nil {
		return nil, m.err
	}
	return m.response, nil
}
