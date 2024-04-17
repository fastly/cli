// Package undocumented provides abstractions for talking to undocumented Fastly
// API endpoints.
package undocumented

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/fastly/cli/pkg/api"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/useragent"
)

// EntitledProductCheck is the API endpoint for checking whether a user already
// has paid access to the specified product.
const EntitledProductCheck = "/entitled-products/%s"

// EntitledProductMessageCompute is shown to a user who doesn't yet have paid
// access to the Compute product.
const EntitledProductMessageCompute = "By creating this Compute service, you acknowledge that the service is a trial service for evaluation purposes subject to Fastly’s terms of service (www.fastly.com/terms)."

// ProductCompute is the ID for the Compute product.
const ProductCompute = "compute"

// RequestTimeout is the timeout for the API network request.
const RequestTimeout = 5 * time.Second

// EntitledProductResponse represents the API response for requesting a
// customer's entitlement data.
type EntitledProductResponse struct {
	AccessLevel      string `json:"access_level"`
	CustomerID       string `json:"customer_id"`
	HasAccess        bool   `json:"has_access"`
	HasPermToDisable bool   `json:"has_permission_to_disable"`
	HasPermToEnable  bool   `json:"has_permission_to_enable"`
}

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
		rc := req.Clone(context.Background())
		rc.Header.Set("Fastly-Key", "REDACTED")
		dump, _ := httputil.DumpRequest(rc, true)
		fmt.Printf("undocumented.Call request dump:\n\n%#v\n\n", string(dump))
	}

	res, err := opts.HTTPClient.Do(req)

	if opts.Debug && res != nil {
		dump, _ := httputil.DumpResponse(res, true)
		fmt.Printf("undocumented.Call response dump:\n\n%#v\n\n", string(dump))
	}

	if err != nil {
		if urlErr, ok := err.(*url.Error); ok && urlErr.Timeout() {
			return data, fsterr.RemediationError{
				Inner:       err,
				Remediation: fsterr.NetworkRemediation,
			}
		}
		statusCode := http.StatusInternalServerError
		if res != nil {
			statusCode = res.StatusCode
		}
		return data, NewError(err, statusCode)
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
