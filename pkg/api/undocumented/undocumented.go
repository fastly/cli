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

// Get calls the given API endpoint and returns its response data.
func Get(host, path, token string, c api.HTTPClient) (data []byte, statusCode int, err error) {
	host = strings.TrimSuffix(host, "/")
	endpoint := fmt.Sprintf("%s%s", host, path)

	req, err := http.NewRequest(http.MethodPost, endpoint, nil)
	if err != nil {
		return data, statusCode, err
	}

	req.Header.Set("Fastly-Key", token)
	req.Header.Set("User-Agent", useragent.Name)

	res, err := c.Do(req)
	if err != nil {
		if urlErr, ok := err.(*url.Error); ok && urlErr.Timeout() {
			return data, statusCode, fsterr.RemediationError{
				Inner:       err,
				Remediation: fsterr.NetworkRemediation,
			}
		}
		return data, statusCode, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return data, res.StatusCode, fmt.Errorf("error from API: '%s'", res.Status)
	}

	data, err = io.ReadAll(res.Body)
	if err != nil {
		return data, res.StatusCode, err
	}

	return data, res.StatusCode, nil
}
