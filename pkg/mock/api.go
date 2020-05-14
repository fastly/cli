package mock

import (
	"github.com/fastly/go-fastly/fastly"
)

// API is a mock implementation of api.Interface that's used for testing.
// The zero value is useful, but will panic on all methods. Provide function
// implementations for the method(s) your test will call.
type API struct {
	GetTokenSelfFn func() (*fastly.Token, error)

	CreateServiceFn     func(*fastly.CreateServiceInput) (*fastly.Service, error)
	ListServicesFn      func(*fastly.ListServicesInput) ([]*fastly.Service, error)
	GetServiceFn        func(*fastly.GetServiceInput) (*fastly.Service, error)
	GetServiceDetailsFn func(*fastly.GetServiceInput) (*fastly.ServiceDetail, error)
	UpdateServiceFn     func(*fastly.UpdateServiceInput) (*fastly.Service, error)
	DeleteServiceFn     func(*fastly.DeleteServiceInput) error

	CloneVersionFn      func(*fastly.CloneVersionInput) (*fastly.Version, error)
	ListVersionsFn      func(*fastly.ListVersionsInput) ([]*fastly.Version, error)
	UpdateVersionFn     func(*fastly.UpdateVersionInput) (*fastly.Version, error)
	ActivateVersionFn   func(*fastly.ActivateVersionInput) (*fastly.Version, error)
	DeactivateVersionFn func(*fastly.DeactivateVersionInput) (*fastly.Version, error)
	LockVersionFn       func(*fastly.LockVersionInput) (*fastly.Version, error)
	LatestVersionFn     func(*fastly.LatestVersionInput) (*fastly.Version, error)

	CreateDomainFn func(*fastly.CreateDomainInput) (*fastly.Domain, error)
	ListDomainsFn  func(*fastly.ListDomainsInput) ([]*fastly.Domain, error)
	GetDomainFn    func(*fastly.GetDomainInput) (*fastly.Domain, error)
	UpdateDomainFn func(*fastly.UpdateDomainInput) (*fastly.Domain, error)
	DeleteDomainFn func(*fastly.DeleteDomainInput) error

	CreateBackendFn func(*fastly.CreateBackendInput) (*fastly.Backend, error)
	ListBackendsFn  func(*fastly.ListBackendsInput) ([]*fastly.Backend, error)
	GetBackendFn    func(*fastly.GetBackendInput) (*fastly.Backend, error)
	UpdateBackendFn func(*fastly.UpdateBackendInput) (*fastly.Backend, error)
	DeleteBackendFn func(*fastly.DeleteBackendInput) error

	CreateHealthCheckFn func(*fastly.CreateHealthCheckInput) (*fastly.HealthCheck, error)
	ListHealthChecksFn  func(*fastly.ListHealthChecksInput) ([]*fastly.HealthCheck, error)
	GetHealthCheckFn    func(*fastly.GetHealthCheckInput) (*fastly.HealthCheck, error)
	UpdateHealthCheckFn func(*fastly.UpdateHealthCheckInput) (*fastly.HealthCheck, error)
	DeleteHealthCheckFn func(*fastly.DeleteHealthCheckInput) error

	CreateBigQueryFn func(*fastly.CreateBigQueryInput) (*fastly.BigQuery, error)
	ListBigQueriesFn func(*fastly.ListBigQueriesInput) ([]*fastly.BigQuery, error)
	GetBigQueryFn    func(*fastly.GetBigQueryInput) (*fastly.BigQuery, error)
	UpdateBigQueryFn func(*fastly.UpdateBigQueryInput) (*fastly.BigQuery, error)
	DeleteBigQueryFn func(*fastly.DeleteBigQueryInput) error

	CreateS3Fn func(*fastly.CreateS3Input) (*fastly.S3, error)
	ListS3sFn  func(*fastly.ListS3sInput) ([]*fastly.S3, error)
	GetS3Fn    func(*fastly.GetS3Input) (*fastly.S3, error)
	UpdateS3Fn func(*fastly.UpdateS3Input) (*fastly.S3, error)
	DeleteS3Fn func(*fastly.DeleteS3Input) error

	CreateSyslogFn func(*fastly.CreateSyslogInput) (*fastly.Syslog, error)
	ListSyslogsFn  func(*fastly.ListSyslogsInput) ([]*fastly.Syslog, error)
	GetSyslogFn    func(*fastly.GetSyslogInput) (*fastly.Syslog, error)
	UpdateSyslogFn func(*fastly.UpdateSyslogInput) (*fastly.Syslog, error)
	DeleteSyslogFn func(*fastly.DeleteSyslogInput) error

	CreateLogentriesFn func(*fastly.CreateLogentriesInput) (*fastly.Logentries, error)
	ListLogentriesFn   func(*fastly.ListLogentriesInput) ([]*fastly.Logentries, error)
	GetLogentriesFn    func(*fastly.GetLogentriesInput) (*fastly.Logentries, error)
	UpdateLogentriesFn func(*fastly.UpdateLogentriesInput) (*fastly.Logentries, error)
	DeleteLogentriesFn func(*fastly.DeleteLogentriesInput) error

	CreatePapertrailFn func(*fastly.CreatePapertrailInput) (*fastly.Papertrail, error)
	ListPapertrailsFn  func(*fastly.ListPapertrailsInput) ([]*fastly.Papertrail, error)
	GetPapertrailFn    func(*fastly.GetPapertrailInput) (*fastly.Papertrail, error)
	UpdatePapertrailFn func(*fastly.UpdatePapertrailInput) (*fastly.Papertrail, error)
	DeletePapertrailFn func(*fastly.DeletePapertrailInput) error

	CreateSumologicFn func(*fastly.CreateSumologicInput) (*fastly.Sumologic, error)
	ListSumologicsFn  func(*fastly.ListSumologicsInput) ([]*fastly.Sumologic, error)
	GetSumologicFn    func(*fastly.GetSumologicInput) (*fastly.Sumologic, error)
	UpdateSumologicFn func(*fastly.UpdateSumologicInput) (*fastly.Sumologic, error)
	DeleteSumologicFn func(*fastly.DeleteSumologicInput) error

	CreateGCSFn func(*fastly.CreateGCSInput) (*fastly.GCS, error)
	ListGCSsFn  func(*fastly.ListGCSsInput) ([]*fastly.GCS, error)
	GetGCSFn    func(*fastly.GetGCSInput) (*fastly.GCS, error)
	UpdateGCSFn func(*fastly.UpdateGCSInput) (*fastly.GCS, error)
	DeleteGCSFn func(*fastly.DeleteGCSInput) error

	CreateFTPFn func(*fastly.CreateFTPInput) (*fastly.FTP, error)
	ListFTPsFn  func(*fastly.ListFTPsInput) ([]*fastly.FTP, error)
	GetFTPFn    func(*fastly.GetFTPInput) (*fastly.FTP, error)
	UpdateFTPFn func(*fastly.UpdateFTPInput) (*fastly.FTP, error)
	DeleteFTPFn func(*fastly.DeleteFTPInput) error

	CreateSplunkFn func(*fastly.CreateSplunkInput) (*fastly.Splunk, error)
	ListSplunksFn  func(*fastly.ListSplunksInput) ([]*fastly.Splunk, error)
	GetSplunkFn    func(*fastly.GetSplunkInput) (*fastly.Splunk, error)
	UpdateSplunkFn func(*fastly.UpdateSplunkInput) (*fastly.Splunk, error)
	DeleteSplunkFn func(*fastly.DeleteSplunkInput) error

	CreateScalyrFn func(*fastly.CreateScalyrInput) (*fastly.Scalyr, error)
	ListScalyrsFn  func(*fastly.ListScalyrsInput) ([]*fastly.Scalyr, error)
	GetScalyrFn    func(*fastly.GetScalyrInput) (*fastly.Scalyr, error)
	UpdateScalyrFn func(*fastly.UpdateScalyrInput) (*fastly.Scalyr, error)
	DeleteScalyrFn func(*fastly.DeleteScalyrInput) error

	CreateLogglyFn func(*fastly.CreateLogglyInput) (*fastly.Loggly, error)
	ListLogglyFn   func(*fastly.ListLogglyInput) ([]*fastly.Loggly, error)
	GetLogglyFn    func(*fastly.GetLogglyInput) (*fastly.Loggly, error)
	UpdateLogglyFn func(*fastly.UpdateLogglyInput) (*fastly.Loggly, error)
	DeleteLogglyFn func(*fastly.DeleteLogglyInput) error

	CreateHoneycombFn func(*fastly.CreateHoneycombInput) (*fastly.Honeycomb, error)
	ListHoneycombsFn  func(*fastly.ListHoneycombsInput) ([]*fastly.Honeycomb, error)
	GetHoneycombFn    func(*fastly.GetHoneycombInput) (*fastly.Honeycomb, error)
	UpdateHoneycombFn func(*fastly.UpdateHoneycombInput) (*fastly.Honeycomb, error)
	DeleteHoneycombFn func(*fastly.DeleteHoneycombInput) error

	CreateHerokuFn func(*fastly.CreateHerokuInput) (*fastly.Heroku, error)
	ListHerokusFn  func(*fastly.ListHerokusInput) ([]*fastly.Heroku, error)
	GetHerokuFn    func(*fastly.GetHerokuInput) (*fastly.Heroku, error)
	UpdateHerokuFn func(*fastly.UpdateHerokuInput) (*fastly.Heroku, error)
	DeleteHerokuFn func(*fastly.DeleteHerokuInput) error

	CreateSFTPFn func(*fastly.CreateSFTPInput) (*fastly.SFTP, error)
	ListSFTPsFn  func(*fastly.ListSFTPsInput) ([]*fastly.SFTP, error)
	GetSFTPFn    func(*fastly.GetSFTPInput) (*fastly.SFTP, error)
	UpdateSFTPFn func(*fastly.UpdateSFTPInput) (*fastly.SFTP, error)
	DeleteSFTPFn func(*fastly.DeleteSFTPInput) error

	CreateLogshuttleFn func(*fastly.CreateLogshuttleInput) (*fastly.Logshuttle, error)
	ListLogshuttlesFn  func(*fastly.ListLogshuttlesInput) ([]*fastly.Logshuttle, error)
	GetLogshuttleFn    func(*fastly.GetLogshuttleInput) (*fastly.Logshuttle, error)
	UpdateLogshuttleFn func(*fastly.UpdateLogshuttleInput) (*fastly.Logshuttle, error)
	DeleteLogshuttleFn func(*fastly.DeleteLogshuttleInput) error

	CreateCloudfilesFn func(*fastly.CreateCloudfilesInput) (*fastly.Cloudfiles, error)
	ListCloudfilesFn   func(*fastly.ListCloudfilesInput) ([]*fastly.Cloudfiles, error)
	GetCloudfilesFn    func(*fastly.GetCloudfilesInput) (*fastly.Cloudfiles, error)
	UpdateCloudfilesFn func(*fastly.UpdateCloudfilesInput) (*fastly.Cloudfiles, error)
	DeleteCloudfilesFn func(*fastly.DeleteCloudfilesInput) error

	GetUserFn func(*fastly.GetUserInput) (*fastly.User, error)

	GetRegionsFn   func() (*fastly.RegionsResponse, error)
	GetStatsJSONFn func(i *fastly.GetStatsInput, dst interface{}) error
}

// GetTokenSelf implements Interface.
func (m API) GetTokenSelf() (*fastly.Token, error) {
	return m.GetTokenSelfFn()
}

// CreateService implements Interface.
func (m API) CreateService(i *fastly.CreateServiceInput) (*fastly.Service, error) {
	return m.CreateServiceFn(i)
}

// ListServices implements Interface.
func (m API) ListServices(i *fastly.ListServicesInput) ([]*fastly.Service, error) {
	return m.ListServicesFn(i)
}

// GetService implements Interface.
func (m API) GetService(i *fastly.GetServiceInput) (*fastly.Service, error) {
	return m.GetServiceFn(i)
}

// GetServiceDetails implements Interface.
func (m API) GetServiceDetails(i *fastly.GetServiceInput) (*fastly.ServiceDetail, error) {
	return m.GetServiceDetailsFn(i)
}

// UpdateService implements Interface.
func (m API) UpdateService(i *fastly.UpdateServiceInput) (*fastly.Service, error) {
	return m.UpdateServiceFn(i)
}

// DeleteService implements Interface.
func (m API) DeleteService(i *fastly.DeleteServiceInput) error {
	return m.DeleteServiceFn(i)
}

// CloneVersion implements Interface.
func (m API) CloneVersion(i *fastly.CloneVersionInput) (*fastly.Version, error) {
	return m.CloneVersionFn(i)
}

// ListVersions implements Interface.
func (m API) ListVersions(i *fastly.ListVersionsInput) ([]*fastly.Version, error) {
	return m.ListVersionsFn(i)
}

// UpdateVersion implements Interface.
func (m API) UpdateVersion(i *fastly.UpdateVersionInput) (*fastly.Version, error) {
	return m.UpdateVersionFn(i)
}

// ActivateVersion implements Interface.
func (m API) ActivateVersion(i *fastly.ActivateVersionInput) (*fastly.Version, error) {
	return m.ActivateVersionFn(i)
}

// DeactivateVersion implements Interface.
func (m API) DeactivateVersion(i *fastly.DeactivateVersionInput) (*fastly.Version, error) {
	return m.DeactivateVersionFn(i)
}

// LockVersion implements Interface.
func (m API) LockVersion(i *fastly.LockVersionInput) (*fastly.Version, error) {
	return m.LockVersionFn(i)
}

// LatestVersion implements Interface.
func (m API) LatestVersion(i *fastly.LatestVersionInput) (*fastly.Version, error) {
	return m.LatestVersionFn(i)
}

// CreateDomain implements Interface.
func (m API) CreateDomain(i *fastly.CreateDomainInput) (*fastly.Domain, error) {
	return m.CreateDomainFn(i)
}

// ListDomains implements Interface.
func (m API) ListDomains(i *fastly.ListDomainsInput) ([]*fastly.Domain, error) {
	return m.ListDomainsFn(i)
}

// GetDomain implements Interface.
func (m API) GetDomain(i *fastly.GetDomainInput) (*fastly.Domain, error) {
	return m.GetDomainFn(i)
}

// UpdateDomain implements Interface.
func (m API) UpdateDomain(i *fastly.UpdateDomainInput) (*fastly.Domain, error) {
	return m.UpdateDomainFn(i)
}

// DeleteDomain implements Interface.
func (m API) DeleteDomain(i *fastly.DeleteDomainInput) error {
	return m.DeleteDomainFn(i)
}

// CreateBackend implements Interface.
func (m API) CreateBackend(i *fastly.CreateBackendInput) (*fastly.Backend, error) {
	return m.CreateBackendFn(i)
}

// ListBackends implements Interface.
func (m API) ListBackends(i *fastly.ListBackendsInput) ([]*fastly.Backend, error) {
	return m.ListBackendsFn(i)
}

// GetBackend implements Interface.
func (m API) GetBackend(i *fastly.GetBackendInput) (*fastly.Backend, error) {
	return m.GetBackendFn(i)
}

// UpdateBackend implements Interface.
func (m API) UpdateBackend(i *fastly.UpdateBackendInput) (*fastly.Backend, error) {
	return m.UpdateBackendFn(i)
}

// DeleteBackend implements Interface.
func (m API) DeleteBackend(i *fastly.DeleteBackendInput) error {
	return m.DeleteBackendFn(i)
}

// CreateHealthCheck implements Interface.
func (m API) CreateHealthCheck(i *fastly.CreateHealthCheckInput) (*fastly.HealthCheck, error) {
	return m.CreateHealthCheckFn(i)
}

// ListHealthChecks implements Interface.
func (m API) ListHealthChecks(i *fastly.ListHealthChecksInput) ([]*fastly.HealthCheck, error) {
	return m.ListHealthChecksFn(i)
}

// GetHealthCheck implements Interface.
func (m API) GetHealthCheck(i *fastly.GetHealthCheckInput) (*fastly.HealthCheck, error) {
	return m.GetHealthCheckFn(i)
}

// UpdateHealthCheck implements Interface.
func (m API) UpdateHealthCheck(i *fastly.UpdateHealthCheckInput) (*fastly.HealthCheck, error) {
	return m.UpdateHealthCheckFn(i)
}

// DeleteHealthCheck implements Interface.
func (m API) DeleteHealthCheck(i *fastly.DeleteHealthCheckInput) error {
	return m.DeleteHealthCheckFn(i)
}

// CreateBigQuery implements Interface.
func (m API) CreateBigQuery(i *fastly.CreateBigQueryInput) (*fastly.BigQuery, error) {
	return m.CreateBigQueryFn(i)
}

// ListBigQueries implements Interface.
func (m API) ListBigQueries(i *fastly.ListBigQueriesInput) ([]*fastly.BigQuery, error) {
	return m.ListBigQueriesFn(i)
}

// GetBigQuery implements Interface.
func (m API) GetBigQuery(i *fastly.GetBigQueryInput) (*fastly.BigQuery, error) {
	return m.GetBigQueryFn(i)
}

// UpdateBigQuery implements Interface.
func (m API) UpdateBigQuery(i *fastly.UpdateBigQueryInput) (*fastly.BigQuery, error) {
	return m.UpdateBigQueryFn(i)
}

// DeleteBigQuery implements Interface.
func (m API) DeleteBigQuery(i *fastly.DeleteBigQueryInput) error {
	return m.DeleteBigQueryFn(i)
}

// CreateS3 implements Interface.
func (m API) CreateS3(i *fastly.CreateS3Input) (*fastly.S3, error) {
	return m.CreateS3Fn(i)
}

// ListS3s implements Interface.
func (m API) ListS3s(i *fastly.ListS3sInput) ([]*fastly.S3, error) {
	return m.ListS3sFn(i)
}

// GetS3 implements Interface.
func (m API) GetS3(i *fastly.GetS3Input) (*fastly.S3, error) {
	return m.GetS3Fn(i)
}

// UpdateS3 implements Interface.
func (m API) UpdateS3(i *fastly.UpdateS3Input) (*fastly.S3, error) {
	return m.UpdateS3Fn(i)
}

// DeleteS3 implements Interface.
func (m API) DeleteS3(i *fastly.DeleteS3Input) error {
	return m.DeleteS3Fn(i)
}

// CreateSyslog implements Interface.
func (m API) CreateSyslog(i *fastly.CreateSyslogInput) (*fastly.Syslog, error) {
	return m.CreateSyslogFn(i)
}

// ListSyslogs implements Interface.
func (m API) ListSyslogs(i *fastly.ListSyslogsInput) ([]*fastly.Syslog, error) {
	return m.ListSyslogsFn(i)
}

// GetSyslog implements Interface.
func (m API) GetSyslog(i *fastly.GetSyslogInput) (*fastly.Syslog, error) {
	return m.GetSyslogFn(i)
}

// UpdateSyslog implements Interface.
func (m API) UpdateSyslog(i *fastly.UpdateSyslogInput) (*fastly.Syslog, error) {
	return m.UpdateSyslogFn(i)
}

// DeleteSyslog implements Interface.
func (m API) DeleteSyslog(i *fastly.DeleteSyslogInput) error {
	return m.DeleteSyslogFn(i)
}

// CreateLogentries implements Interface.
func (m API) CreateLogentries(i *fastly.CreateLogentriesInput) (*fastly.Logentries, error) {
	return m.CreateLogentriesFn(i)
}

// ListLogentries implements Interface.
func (m API) ListLogentries(i *fastly.ListLogentriesInput) ([]*fastly.Logentries, error) {
	return m.ListLogentriesFn(i)
}

// GetLogentries implements Interface.
func (m API) GetLogentries(i *fastly.GetLogentriesInput) (*fastly.Logentries, error) {
	return m.GetLogentriesFn(i)
}

// UpdateLogentries implements Interface.
func (m API) UpdateLogentries(i *fastly.UpdateLogentriesInput) (*fastly.Logentries, error) {
	return m.UpdateLogentriesFn(i)
}

// DeleteLogentries implements Interface.
func (m API) DeleteLogentries(i *fastly.DeleteLogentriesInput) error {
	return m.DeleteLogentriesFn(i)
}

// CreatePapertrail implements Interface.
func (m API) CreatePapertrail(i *fastly.CreatePapertrailInput) (*fastly.Papertrail, error) {
	return m.CreatePapertrailFn(i)
}

// ListPapertrails implements Interface.
func (m API) ListPapertrails(i *fastly.ListPapertrailsInput) ([]*fastly.Papertrail, error) {
	return m.ListPapertrailsFn(i)
}

// GetPapertrail implements Interface.
func (m API) GetPapertrail(i *fastly.GetPapertrailInput) (*fastly.Papertrail, error) {
	return m.GetPapertrailFn(i)
}

// UpdatePapertrail implements Interface.
func (m API) UpdatePapertrail(i *fastly.UpdatePapertrailInput) (*fastly.Papertrail, error) {
	return m.UpdatePapertrailFn(i)
}

// DeletePapertrail implements Interface.
func (m API) DeletePapertrail(i *fastly.DeletePapertrailInput) error {
	return m.DeletePapertrailFn(i)
}

// CreateSumologic implements Interface.
func (m API) CreateSumologic(i *fastly.CreateSumologicInput) (*fastly.Sumologic, error) {
	return m.CreateSumologicFn(i)
}

// ListSumologics implements Interface.
func (m API) ListSumologics(i *fastly.ListSumologicsInput) ([]*fastly.Sumologic, error) {
	return m.ListSumologicsFn(i)
}

// GetSumologic implements Interface.
func (m API) GetSumologic(i *fastly.GetSumologicInput) (*fastly.Sumologic, error) {
	return m.GetSumologicFn(i)
}

// UpdateSumologic implements Interface.
func (m API) UpdateSumologic(i *fastly.UpdateSumologicInput) (*fastly.Sumologic, error) {
	return m.UpdateSumologicFn(i)
}

// DeleteSumologic implements Interface.
func (m API) DeleteSumologic(i *fastly.DeleteSumologicInput) error {
	return m.DeleteSumologicFn(i)
}

// CreateGCS implements Interface.
func (m API) CreateGCS(i *fastly.CreateGCSInput) (*fastly.GCS, error) {
	return m.CreateGCSFn(i)
}

// ListGCSs implements Interface.
func (m API) ListGCSs(i *fastly.ListGCSsInput) ([]*fastly.GCS, error) {
	return m.ListGCSsFn(i)
}

// GetGCS implements Interface.
func (m API) GetGCS(i *fastly.GetGCSInput) (*fastly.GCS, error) {
	return m.GetGCSFn(i)
}

// UpdateGCS implements Interface.
func (m API) UpdateGCS(i *fastly.UpdateGCSInput) (*fastly.GCS, error) {
	return m.UpdateGCSFn(i)
}

// DeleteGCS implements Interface.
func (m API) DeleteGCS(i *fastly.DeleteGCSInput) error {
	return m.DeleteGCSFn(i)
}

// CreateFTP implements Interface.
func (m API) CreateFTP(i *fastly.CreateFTPInput) (*fastly.FTP, error) {
	return m.CreateFTPFn(i)
}

// ListFTPs implements Interface.
func (m API) ListFTPs(i *fastly.ListFTPsInput) ([]*fastly.FTP, error) {
	return m.ListFTPsFn(i)
}

// GetFTP implements Interface.
func (m API) GetFTP(i *fastly.GetFTPInput) (*fastly.FTP, error) {
	return m.GetFTPFn(i)
}

// UpdateFTP implements Interface.
func (m API) UpdateFTP(i *fastly.UpdateFTPInput) (*fastly.FTP, error) {
	return m.UpdateFTPFn(i)
}

// DeleteFTP implements Interface.
func (m API) DeleteFTP(i *fastly.DeleteFTPInput) error {
	return m.DeleteFTPFn(i)
}

// CreateSplunk implements Interface.
func (m API) CreateSplunk(i *fastly.CreateSplunkInput) (*fastly.Splunk, error) {
	return m.CreateSplunkFn(i)
}

// ListSplunks implements Interface.
func (m API) ListSplunks(i *fastly.ListSplunksInput) ([]*fastly.Splunk, error) {
	return m.ListSplunksFn(i)
}

// GetSplunk implements Interface.
func (m API) GetSplunk(i *fastly.GetSplunkInput) (*fastly.Splunk, error) {
	return m.GetSplunkFn(i)
}

// UpdateSplunk implements Interface.
func (m API) UpdateSplunk(i *fastly.UpdateSplunkInput) (*fastly.Splunk, error) {
	return m.UpdateSplunkFn(i)
}

// DeleteSplunk implements Interface.
func (m API) DeleteSplunk(i *fastly.DeleteSplunkInput) error {
	return m.DeleteSplunkFn(i)
}

// CreateScalyr implements Interface.
func (m API) CreateScalyr(i *fastly.CreateScalyrInput) (*fastly.Scalyr, error) {
	return m.CreateScalyrFn(i)
}

// ListScalyrs implements Interface.
func (m API) ListScalyrs(i *fastly.ListScalyrsInput) ([]*fastly.Scalyr, error) {
	return m.ListScalyrsFn(i)
}

// GetScalyr implements Interface.
func (m API) GetScalyr(i *fastly.GetScalyrInput) (*fastly.Scalyr, error) {
	return m.GetScalyrFn(i)
}

// UpdateScalyr implements Interface.
func (m API) UpdateScalyr(i *fastly.UpdateScalyrInput) (*fastly.Scalyr, error) {
	return m.UpdateScalyrFn(i)
}

// DeleteScalyr implements Interface.
func (m API) DeleteScalyr(i *fastly.DeleteScalyrInput) error {
	return m.DeleteScalyrFn(i)
}

// CreateLoggly implements Interface.
func (m API) CreateLoggly(i *fastly.CreateLogglyInput) (*fastly.Loggly, error) {
	return m.CreateLogglyFn(i)
}

// ListLoggly implements Interface.
func (m API) ListLoggly(i *fastly.ListLogglyInput) ([]*fastly.Loggly, error) {
	return m.ListLogglyFn(i)
}

// GetLoggly implements Interface.
func (m API) GetLoggly(i *fastly.GetLogglyInput) (*fastly.Loggly, error) {
	return m.GetLogglyFn(i)
}

// UpdateLoggly implements Interface.
func (m API) UpdateLoggly(i *fastly.UpdateLogglyInput) (*fastly.Loggly, error) {
	return m.UpdateLogglyFn(i)
}

// DeleteLoggly implements Interface.
func (m API) DeleteLoggly(i *fastly.DeleteLogglyInput) error {
	return m.DeleteLogglyFn(i)
}

// CreateHoneycomb implements Interface.
func (m API) CreateHoneycomb(i *fastly.CreateHoneycombInput) (*fastly.Honeycomb, error) {
	return m.CreateHoneycombFn(i)
}

// ListHoneycombs implements Interface.
func (m API) ListHoneycombs(i *fastly.ListHoneycombsInput) ([]*fastly.Honeycomb, error) {
	return m.ListHoneycombsFn(i)
}

// GetHoneycomb implements Interface.
func (m API) GetHoneycomb(i *fastly.GetHoneycombInput) (*fastly.Honeycomb, error) {
	return m.GetHoneycombFn(i)
}

// UpdateHoneycomb implements Interface.
func (m API) UpdateHoneycomb(i *fastly.UpdateHoneycombInput) (*fastly.Honeycomb, error) {
	return m.UpdateHoneycombFn(i)
}

// DeleteHoneycomb implements Interface.
func (m API) DeleteHoneycomb(i *fastly.DeleteHoneycombInput) error {
	return m.DeleteHoneycombFn(i)
}

// CreateHeroku implements Interface.
func (m API) CreateHeroku(i *fastly.CreateHerokuInput) (*fastly.Heroku, error) {
	return m.CreateHerokuFn(i)
}

// ListHerokus implements Interface.
func (m API) ListHerokus(i *fastly.ListHerokusInput) ([]*fastly.Heroku, error) {
	return m.ListHerokusFn(i)
}

// GetHeroku implements Interface.
func (m API) GetHeroku(i *fastly.GetHerokuInput) (*fastly.Heroku, error) {
	return m.GetHerokuFn(i)
}

// UpdateHeroku implements Interface.
func (m API) UpdateHeroku(i *fastly.UpdateHerokuInput) (*fastly.Heroku, error) {
	return m.UpdateHerokuFn(i)
}

// DeleteHeroku implements Interface.
func (m API) DeleteHeroku(i *fastly.DeleteHerokuInput) error {
	return m.DeleteHerokuFn(i)
}

// CreateSFTP implements Interface.
func (m API) CreateSFTP(i *fastly.CreateSFTPInput) (*fastly.SFTP, error) {
	return m.CreateSFTPFn(i)
}

// ListSFTPs implements Interface.
func (m API) ListSFTPs(i *fastly.ListSFTPsInput) ([]*fastly.SFTP, error) {
	return m.ListSFTPsFn(i)
}

// GetSFTP implements Interface.
func (m API) GetSFTP(i *fastly.GetSFTPInput) (*fastly.SFTP, error) {
	return m.GetSFTPFn(i)
}

// UpdateSFTP implements Interface.
func (m API) UpdateSFTP(i *fastly.UpdateSFTPInput) (*fastly.SFTP, error) {
	return m.UpdateSFTPFn(i)
}

// DeleteSFTP implements Interface.
func (m API) DeleteSFTP(i *fastly.DeleteSFTPInput) error {
	return m.DeleteSFTPFn(i)
}

// CreateLogshuttle implements Interface.
func (m API) CreateLogshuttle(i *fastly.CreateLogshuttleInput) (*fastly.Logshuttle, error) {
	return m.CreateLogshuttleFn(i)
}

// ListLogshuttles implements Interface.
func (m API) ListLogshuttles(i *fastly.ListLogshuttlesInput) ([]*fastly.Logshuttle, error) {
	return m.ListLogshuttlesFn(i)
}

// GetLogshuttle implements Interface.
func (m API) GetLogshuttle(i *fastly.GetLogshuttleInput) (*fastly.Logshuttle, error) {
	return m.GetLogshuttleFn(i)
}

// UpdateLogshuttle implements Interface.
func (m API) UpdateLogshuttle(i *fastly.UpdateLogshuttleInput) (*fastly.Logshuttle, error) {
	return m.UpdateLogshuttleFn(i)
}

// DeleteLogshuttle implements Interface.
func (m API) DeleteLogshuttle(i *fastly.DeleteLogshuttleInput) error {
	return m.DeleteLogshuttleFn(i)
}

// CreateCloudfiles implements Interface.
func (m API) CreateCloudfiles(i *fastly.CreateCloudfilesInput) (*fastly.Cloudfiles, error) {
	return m.CreateCloudfilesFn(i)
}

// ListCloudfiles implements Interface.
func (m API) ListCloudfiles(i *fastly.ListCloudfilesInput) ([]*fastly.Cloudfiles, error) {
	return m.ListCloudfilesFn(i)
}

// GetCloudfiles implements Interface.
func (m API) GetCloudfiles(i *fastly.GetCloudfilesInput) (*fastly.Cloudfiles, error) {
	return m.GetCloudfilesFn(i)
}

// UpdateCloudfiles implements Interface.
func (m API) UpdateCloudfiles(i *fastly.UpdateCloudfilesInput) (*fastly.Cloudfiles, error) {
	return m.UpdateCloudfilesFn(i)
}

// DeleteCloudfiles implements Interface.
func (m API) DeleteCloudfiles(i *fastly.DeleteCloudfilesInput) error {
	return m.DeleteCloudfilesFn(i)
}

// GetUser implements Interface.
func (m API) GetUser(i *fastly.GetUserInput) (*fastly.User, error) {
	return m.GetUserFn(i)
}

// GetRegions implements Interface.
func (m API) GetRegions() (*fastly.RegionsResponse, error) {
	return m.GetRegionsFn()
}

// GetStatsJSON implements Interface.
func (m API) GetStatsJSON(i *fastly.GetStatsInput, dst interface{}) error {
	return m.GetStatsJSONFn(i, dst)
}
