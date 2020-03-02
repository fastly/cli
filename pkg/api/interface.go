package api

import (
	"net/http"

	"github.com/fastly/go-fastly/fastly"
)

// HTTPClient models a concrete http.Client. It's a consumer contract for some
// commands which need to make direct HTTP requests to the API, because the
// official Fastly client library lacks certain endpoints, so we call the API
// directly.
type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

// Interface models the methods of the Fastly API client that we use.
// It exists to allow for easier testing, in combination with Mock.
type Interface interface {
	GetTokenSelf() (*fastly.Token, error)

	CreateService(*fastly.CreateServiceInput) (*fastly.Service, error)
	ListServices(*fastly.ListServicesInput) ([]*fastly.Service, error)
	GetService(*fastly.GetServiceInput) (*fastly.Service, error)
	GetServiceDetails(*fastly.GetServiceInput) (*fastly.ServiceDetail, error)
	UpdateService(*fastly.UpdateServiceInput) (*fastly.Service, error)
	DeleteService(*fastly.DeleteServiceInput) error

	CloneVersion(*fastly.CloneVersionInput) (*fastly.Version, error)
	ListVersions(*fastly.ListVersionsInput) ([]*fastly.Version, error)
	UpdateVersion(*fastly.UpdateVersionInput) (*fastly.Version, error)
	ActivateVersion(*fastly.ActivateVersionInput) (*fastly.Version, error)
	DeactivateVersion(*fastly.DeactivateVersionInput) (*fastly.Version, error)
	LockVersion(*fastly.LockVersionInput) (*fastly.Version, error)
	LatestVersion(*fastly.LatestVersionInput) (*fastly.Version, error)

	CreateDomain(*fastly.CreateDomainInput) (*fastly.Domain, error)
	ListDomains(*fastly.ListDomainsInput) ([]*fastly.Domain, error)
	GetDomain(*fastly.GetDomainInput) (*fastly.Domain, error)
	UpdateDomain(*fastly.UpdateDomainInput) (*fastly.Domain, error)
	DeleteDomain(*fastly.DeleteDomainInput) error

	CreateBackend(*fastly.CreateBackendInput) (*fastly.Backend, error)
	ListBackends(*fastly.ListBackendsInput) ([]*fastly.Backend, error)
	GetBackend(*fastly.GetBackendInput) (*fastly.Backend, error)
	UpdateBackend(*fastly.UpdateBackendInput) (*fastly.Backend, error)
	DeleteBackend(*fastly.DeleteBackendInput) error

	CreateHealthCheck(*fastly.CreateHealthCheckInput) (*fastly.HealthCheck, error)
	ListHealthChecks(*fastly.ListHealthChecksInput) ([]*fastly.HealthCheck, error)
	GetHealthCheck(*fastly.GetHealthCheckInput) (*fastly.HealthCheck, error)
	UpdateHealthCheck(*fastly.UpdateHealthCheckInput) (*fastly.HealthCheck, error)
	DeleteHealthCheck(*fastly.DeleteHealthCheckInput) error

	CreateBigQuery(*fastly.CreateBigQueryInput) (*fastly.BigQuery, error)
	ListBigQueries(*fastly.ListBigQueriesInput) ([]*fastly.BigQuery, error)
	GetBigQuery(*fastly.GetBigQueryInput) (*fastly.BigQuery, error)
	UpdateBigQuery(*fastly.UpdateBigQueryInput) (*fastly.BigQuery, error)
	DeleteBigQuery(*fastly.DeleteBigQueryInput) error

	GetUser(*fastly.GetUserInput) (*fastly.User, error)
}

// Interface assertion, to catch mismatches early.
var _ Interface = (*fastly.Client)(nil)
