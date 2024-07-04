// Package undocumented provides abstractions for talking to undocumented Fastly
// API endpoints.
package undocumented

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/debug"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/useragent"
)

// EdgeComputeTrial is the API endpoint for activating a compute trial.
const EdgeComputeTrial = "/customer/%s/edge-compute-trial"

// RequestTimeout is the timeout for the API network request.
const RequestTimeout = 5 * time.Second

// APIError models a custom error for undocumented API calls.
type APIError struct {
	Err        error
	StatusCode int
}

// Error implements the error interface.
func (e APIError) Error() string {
	return e.Err.Error()
}

// NewError returns an APIError.
func NewError(err error, statusCode int) APIError {
	return APIError{
		Err:        err,
		StatusCode: statusCode,
	}
}

// HTTPHeader represents a HTTP request header.
type HTTPHeader struct {
	Key   string
	Value string
}

// CallOptions is used as input to Call().
type CallOptions struct {
	APIEndpoint string
	Body        io.Reader
	Debug       bool
	HTTPClient  api.HTTPClient
	HTTPHeaders []HTTPHeader
	Method      string
	Path        string
	Token       string
}

// Call calls the given API endpoint and returns its response data.
//
// WARNING: Loads entire response body into memory.
func Call(opts CallOptions) (data []byte, err error) {
	host := strings.TrimSuffix(opts.APIEndpoint, "/")
	endpoint := fmt.Sprintf("%s%s", host, opts.Path)

	req, err := http.NewRequest(opts.Method, endpoint, opts.Body)
	if err != nil {
		return data, NewError(err, 0)
	}

	if opts.Token != "" {
		req.Header.Set("Fastly-Key", opts.Token)
	}
	req.Header.Set("User-Agent", useragent.Name)
	for _, header := range opts.HTTPHeaders {
		req.Header.Set(header.Key, header.Value)
	}

	if opts.Debug {
		debug.DumpHTTPRequest(req)
	}
	res, err := opts.HTTPClient.Do(req)
	if opts.Debug {
		debug.DumpHTTPResponse(res)
	}

	if err != nil {
		if urlErr, ok := err.(*url.Error); ok && urlErr.Timeout() {
			return data, fsterr.RemediationError{
				Inner:       err,
				Remediation: fsterr.NetworkRemediation,
			}
		}
		return data, NewError(err, 0)
	}
	defer res.Body.Close() // #nosec G307

	data, err = io.ReadAll(res.Body)
	if err != nil {
		return []byte{}, NewError(err, res.StatusCode)
	}

	if res.StatusCode >= 400 {
		return data, NewError(fmt.Errorf("error response: %q", data), res.StatusCode)
	}

	return data, nil
}
