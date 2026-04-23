package mock

import "net/http"

type RoundTripper struct {
	Client *HTTPClient
}

var _ http.RoundTripper = (*RoundTripper)(nil)

func NewRoundTripper(c *HTTPClient) *RoundTripper {
	return &RoundTripper{Client: c}
}

func (t *RoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	if t.Client == nil {
		return nil, &ErrMockMisconfigured{Msg: "mock.RoundTripper: Client is nil"}
	}

	// Use Client's Do() behavior
	resp, err := t.Client.Do(r)
	if err != nil {
		return nil, err
	}

	// Be defensive: avoid returning a nil response when err == nil.
	if resp == nil {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       http.NoBody,
			Request:    r,
		}, nil
	}
	if resp.Request == nil {
		resp.Request = r
	}

	return resp, nil
}

type ErrMockMisconfigured struct{ Msg string }

func (e *ErrMockMisconfigured) Error() string { return e.Msg }
