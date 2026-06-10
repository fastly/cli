package mock

import (
	"github.com/fastly/go-fastly/v15/fastly"
)

func NewFastlyClient(httpClient *HTTPClient) (*fastly.Client, error) {
	apiClient, err := fastly.NewClientForEndpoint("no-key", "https://api.example.com/")
	if err != nil {
		return nil, err
	}

	apiClient.HTTPClient = NewNetHTTPClientWithMockHTTPClient(httpClient)

	return apiClient, nil
}
