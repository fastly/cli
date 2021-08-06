package compute_test

// NOTE: This file doesn't contain any tests. It only contains code that is
// shared across some of the other test files (mostly mocked API responses, but
// also a mocked HTTP client).

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v3/fastly"
)

func getServiceOK(i *fastly.GetServiceInput) (*fastly.Service, error) {
	return &fastly.Service{
		ID:   "12345",
		Name: "test",
	}, nil
}

func createDomainOK(i *fastly.CreateDomainInput) (*fastly.Domain, error) {
	return &fastly.Domain{
		ServiceID:      i.ServiceID,
		ServiceVersion: i.ServiceVersion,
		Name:           i.Name,
	}, nil
}

func createBackendOK(i *fastly.CreateBackendInput) (*fastly.Backend, error) {
	return &fastly.Backend{
		ServiceID:      i.ServiceID,
		ServiceVersion: i.ServiceVersion,
		Name:           i.Name,
	}, nil
}

func getPackageOk(i *fastly.GetPackageInput) (*fastly.Package, error) {
	return &fastly.Package{ServiceID: i.ServiceID, ServiceVersion: i.ServiceVersion}, nil
}

func updatePackageOk(i *fastly.UpdatePackageInput) (*fastly.Package, error) {
	return &fastly.Package{ServiceID: i.ServiceID, ServiceVersion: i.ServiceVersion}, nil
}

func updatePackageError(i *fastly.UpdatePackageInput) (*fastly.Package, error) {
	return nil, testutil.Err
}

func activateVersionOk(i *fastly.ActivateVersionInput) (*fastly.Version, error) {
	return &fastly.Version{ServiceID: i.ServiceID, Number: i.ServiceVersion}, nil
}

func updateVersionOk(i *fastly.UpdateVersionInput) (*fastly.Version, error) {
	return &fastly.Version{ServiceID: i.ServiceID, Number: i.ServiceVersion, Comment: *i.Comment}, nil
}

func listDomainsOk(i *fastly.ListDomainsInput) ([]*fastly.Domain, error) {
	return []*fastly.Domain{
		{Name: "https://directly-careful-coyote.edgecompute.app"},
	}, nil
}

func listBackendsOk(i *fastly.ListBackendsInput) ([]*fastly.Backend, error) {
	return []*fastly.Backend{
		{Name: "foo"},
		{Name: "bar"},
	}, nil
}

func listBackendsNone(i *fastly.ListBackendsInput) ([]*fastly.Backend, error) {
	return []*fastly.Backend{}, nil
}

type versionClient struct {
	fastlyVersions    []string
	fastlySysVersions []string
}

func (v versionClient) Do(req *http.Request) (*http.Response, error) {
	var vs []string

	if strings.Contains(req.URL.String(), "crates/fastly-sys/") {
		vs = v.fastlySysVersions
	}
	if strings.Contains(req.URL.String(), "crates/fastly/") {
		vs = v.fastlyVersions
	}

	rec := httptest.NewRecorder()

	var versions []string
	for _, vv := range vs {
		versions = append(versions, fmt.Sprintf(`{"num":"%s"}`, vv))
	}

	_, err := rec.Write([]byte(fmt.Sprintf(`{"versions":[%s]}`, strings.Join(versions, ","))))
	if err != nil {
		return nil, err
	}
	return rec.Result(), nil
}
