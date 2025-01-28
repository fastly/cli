package testutil

import "net/http"

// MockRoundTripper implements [http.RoundTripper] for mocking HTTP responses
type MockRoundTripper struct {
	Response *http.Response
	Err      error
}

// RoundTrip executes a single HTTP transaction, returning a Response for the
// provided Request.
func (m *MockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.Response, m.Err
}
