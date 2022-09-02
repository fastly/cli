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

// NewError returns an APIError
func NewError(err error, statusCode int) APIError {
	return APIError{
		Err:        err,
		StatusCode: statusCode,
	}
}

// Get calls the given API endpoint and returns its response data.
func Get(host, path, token string, c api.HTTPClient) (data []byte, err error) {
	host = strings.TrimSuffix(host, "/")
	endpoint := fmt.Sprintf("%s%s", host, path)

	req, err := http.NewRequest(http.MethodPost, endpoint, nil)
	if err != nil {
		return data, NewError(err, 0)
	}

	req.Header.Set("Fastly-Key", token)
	req.Header.Set("User-Agent", useragent.Name)

	res, err := c.Do(req)
	if err != nil {
		if urlErr, ok := err.(*url.Error); ok && urlErr.Timeout() {
			return data, fsterr.RemediationError{
				Inner:       err,
				Remediation: fsterr.NetworkRemediation,
			}
		}
		return data, NewError(err, 0)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return data, NewError(fmt.Errorf("non-2xx response"), res.StatusCode)
	}

	data, err = io.ReadAll(res.Body)
	if err != nil {
		return data, NewError(err, res.StatusCode)
	}

	return data, nil
}
