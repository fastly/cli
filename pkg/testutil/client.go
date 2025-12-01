package testutil

import "net/http"

// MockRoundTripper implements [http.RoundTripper] for mocking HTTP responses.
type MockRoundTripper struct {
	Response *http.Response
	Err      error
}

// RoundTrip executes a single HTTP transaction, returning a Response for the
// provided Request.
func (m *MockRoundTripper) RoundTrip(_ *http.Request) (*http.Response, error) {
	return m.Response, m.Err
}

// MultiResponseRoundTripper implements [http.RoundTripper] for mocking multiple
// sequential HTTP responses. This is useful when the code under test makes
// multiple HTTP calls (e.g., GET then PATCH).
//
// When we perform a get and update in go-fastly operations (such as for alerts),
// we need to be able to parse multiple responses back from the API.
type MultiResponseRoundTripper struct {
	Responses []*http.Response
	index     int
}

// RoundTrip executes a single HTTP transaction, returning the next Response
// in sequence.
func (m *MultiResponseRoundTripper) RoundTrip(_ *http.Request) (*http.Response, error) {
	if m.index >= len(m.Responses) {
		return m.Responses[len(m.Responses)-1], nil
	}
	resp := m.Responses[m.index]
	m.index++
	return resp, nil
}
