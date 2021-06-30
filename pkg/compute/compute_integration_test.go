package compute_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v3/fastly"
)

func createServiceOK(i *fastly.CreateServiceInput) (*fastly.Service, error) {
	return &fastly.Service{
		ID:   "12345",
		Name: i.Name,
		Type: i.Type,
	}, nil
}

func getServiceOK(i *fastly.GetServiceInput) (*fastly.Service, error) {
	return &fastly.Service{
		ID:   "12345",
		Name: "test",
	}, nil
}

func createServiceError(*fastly.CreateServiceInput) (*fastly.Service, error) {
	return nil, testutil.Err
}

func deleteServiceOK(i *fastly.DeleteServiceInput) error {
	return nil
}

func createDomainOK(i *fastly.CreateDomainInput) (*fastly.Domain, error) {
	return &fastly.Domain{
		ServiceID:      i.ServiceID,
		ServiceVersion: i.ServiceVersion,
		Name:           i.Name,
	}, nil
}

func createDomainError(i *fastly.CreateDomainInput) (*fastly.Domain, error) {
	return nil, testutil.Err
}

func deleteDomainOK(i *fastly.DeleteDomainInput) error {
	return nil
}

func createBackendOK(i *fastly.CreateBackendInput) (*fastly.Backend, error) {
	return &fastly.Backend{
		ServiceID:      i.ServiceID,
		ServiceVersion: i.ServiceVersion,
		Name:           i.Name,
	}, nil
}

func createBackendError(i *fastly.CreateBackendInput) (*fastly.Backend, error) {
	return nil, testutil.Err
}

func deleteBackendOK(i *fastly.DeleteBackendInput) error {
	return nil
}

func getPackageOk(i *fastly.GetPackageInput) (*fastly.Package, error) {
	return &fastly.Package{ServiceID: i.ServiceID, ServiceVersion: i.ServiceVersion}, nil
}

func getPackageIdentical(i *fastly.GetPackageInput) (*fastly.Package, error) {
	return &fastly.Package{
		ServiceID:      i.ServiceID,
		ServiceVersion: i.ServiceVersion,
		Metadata: fastly.PackageMetadata{
			HashSum: "2b742f99854df7e024c287e36fb0fdfc5414942e012be717e52148ea0d6800d66fc659563f6f11105815051e82b14b61edc84b33b49789b790db1ed3446fb483",
		},
	}, nil
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

func activateVersionError(i *fastly.ActivateVersionInput) (*fastly.Version, error) {
	return nil, testutil.Err
}

func listDomainsOk(i *fastly.ListDomainsInput) ([]*fastly.Domain, error) {
	return []*fastly.Domain{
		{Name: "https://directly-careful-coyote.edgecompute.app"},
	}, nil
}

func listDomainsError(i *fastly.ListDomainsInput) ([]*fastly.Domain, error) {
	return nil, testutil.Err
}

func listBackendsOk(i *fastly.ListBackendsInput) ([]*fastly.Backend, error) {
	return []*fastly.Backend{
		{Name: "foobar"},
	}, nil
}

func listBackendsError(i *fastly.ListBackendsInput) ([]*fastly.Backend, error) {
	return nil, testutil.Err
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
