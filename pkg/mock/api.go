package mock

import (
	"github.com/fastly/go-fastly/v6/fastly"
)

// API is a mock implementation of api.Interface that's used for testing.
// The zero value is useful, but will panic on all methods. Provide function
// implementations for the method(s) your test will call.
type API struct {
	AllDatacentersFn func() (datacenters []fastly.Datacenter, err error)
	AllIPsFn         func() (v4, v6 fastly.IPAddrs, err error)

	CreateServiceFn     func(*fastly.CreateServiceInput) (*fastly.Service, error)
	ListServicesFn      func(*fastly.ListServicesInput) ([]*fastly.Service, error)
	GetServiceFn        func(*fastly.GetServiceInput) (*fastly.Service, error)
	GetServiceDetailsFn func(*fastly.GetServiceInput) (*fastly.ServiceDetail, error)
	UpdateServiceFn     func(*fastly.UpdateServiceInput) (*fastly.Service, error)
	DeleteServiceFn     func(*fastly.DeleteServiceInput) error
	SearchServiceFn     func(*fastly.SearchServiceInput) (*fastly.Service, error)

	CloneVersionFn      func(*fastly.CloneVersionInput) (*fastly.Version, error)
	ListVersionsFn      func(*fastly.ListVersionsInput) ([]*fastly.Version, error)
	GetVersionFn        func(*fastly.GetVersionInput) (*fastly.Version, error)
	UpdateVersionFn     func(*fastly.UpdateVersionInput) (*fastly.Version, error)
	ActivateVersionFn   func(*fastly.ActivateVersionInput) (*fastly.Version, error)
	DeactivateVersionFn func(*fastly.DeactivateVersionInput) (*fastly.Version, error)
	LockVersionFn       func(*fastly.LockVersionInput) (*fastly.Version, error)
	LatestVersionFn     func(*fastly.LatestVersionInput) (*fastly.Version, error)

	CreateDomainFn       func(*fastly.CreateDomainInput) (*fastly.Domain, error)
	ListDomainsFn        func(*fastly.ListDomainsInput) ([]*fastly.Domain, error)
	GetDomainFn          func(*fastly.GetDomainInput) (*fastly.Domain, error)
	UpdateDomainFn       func(*fastly.UpdateDomainInput) (*fastly.Domain, error)
	DeleteDomainFn       func(*fastly.DeleteDomainInput) error
	ValidateDomainFn     func(i *fastly.ValidateDomainInput) (*fastly.DomainValidationResult, error)
	ValidateAllDomainsFn func(i *fastly.ValidateAllDomainsInput) (results []*fastly.DomainValidationResult, err error)

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

	GetPackageFn    func(*fastly.GetPackageInput) (*fastly.Package, error)
	UpdatePackageFn func(*fastly.UpdatePackageInput) (*fastly.Package, error)

	CreateDictionaryFn func(*fastly.CreateDictionaryInput) (*fastly.Dictionary, error)
	GetDictionaryFn    func(*fastly.GetDictionaryInput) (*fastly.Dictionary, error)
	DeleteDictionaryFn func(*fastly.DeleteDictionaryInput) error
	ListDictionariesFn func(*fastly.ListDictionariesInput) ([]*fastly.Dictionary, error)
	UpdateDictionaryFn func(*fastly.UpdateDictionaryInput) (*fastly.Dictionary, error)

	ListDictionaryItemsFn        func(*fastly.ListDictionaryItemsInput) ([]*fastly.DictionaryItem, error)
	GetDictionaryItemFn          func(*fastly.GetDictionaryItemInput) (*fastly.DictionaryItem, error)
	CreateDictionaryItemFn       func(*fastly.CreateDictionaryItemInput) (*fastly.DictionaryItem, error)
	UpdateDictionaryItemFn       func(*fastly.UpdateDictionaryItemInput) (*fastly.DictionaryItem, error)
	DeleteDictionaryItemFn       func(*fastly.DeleteDictionaryItemInput) error
	BatchModifyDictionaryItemsFn func(*fastly.BatchModifyDictionaryItemsInput) error

	GetDictionaryInfoFn func(*fastly.GetDictionaryInfoInput) (*fastly.DictionaryInfo, error)

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

	CreateKinesisFn func(*fastly.CreateKinesisInput) (*fastly.Kinesis, error)
	ListKinesisFn   func(*fastly.ListKinesisInput) ([]*fastly.Kinesis, error)
	GetKinesisFn    func(*fastly.GetKinesisInput) (*fastly.Kinesis, error)
	UpdateKinesisFn func(*fastly.UpdateKinesisInput) (*fastly.Kinesis, error)
	DeleteKinesisFn func(*fastly.DeleteKinesisInput) error

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

	CreateDigitalOceanFn func(*fastly.CreateDigitalOceanInput) (*fastly.DigitalOcean, error)
	ListDigitalOceansFn  func(*fastly.ListDigitalOceansInput) ([]*fastly.DigitalOcean, error)
	GetDigitalOceanFn    func(*fastly.GetDigitalOceanInput) (*fastly.DigitalOcean, error)
	UpdateDigitalOceanFn func(*fastly.UpdateDigitalOceanInput) (*fastly.DigitalOcean, error)
	DeleteDigitalOceanFn func(*fastly.DeleteDigitalOceanInput) error

	CreateElasticsearchFn func(*fastly.CreateElasticsearchInput) (*fastly.Elasticsearch, error)
	ListElasticsearchFn   func(*fastly.ListElasticsearchInput) ([]*fastly.Elasticsearch, error)
	GetElasticsearchFn    func(*fastly.GetElasticsearchInput) (*fastly.Elasticsearch, error)
	UpdateElasticsearchFn func(*fastly.UpdateElasticsearchInput) (*fastly.Elasticsearch, error)
	DeleteElasticsearchFn func(*fastly.DeleteElasticsearchInput) error

	CreateBlobStorageFn func(*fastly.CreateBlobStorageInput) (*fastly.BlobStorage, error)
	ListBlobStoragesFn  func(*fastly.ListBlobStoragesInput) ([]*fastly.BlobStorage, error)
	GetBlobStorageFn    func(*fastly.GetBlobStorageInput) (*fastly.BlobStorage, error)
	UpdateBlobStorageFn func(*fastly.UpdateBlobStorageInput) (*fastly.BlobStorage, error)
	DeleteBlobStorageFn func(*fastly.DeleteBlobStorageInput) error

	CreateDatadogFn func(*fastly.CreateDatadogInput) (*fastly.Datadog, error)
	ListDatadogFn   func(*fastly.ListDatadogInput) ([]*fastly.Datadog, error)
	GetDatadogFn    func(*fastly.GetDatadogInput) (*fastly.Datadog, error)
	UpdateDatadogFn func(*fastly.UpdateDatadogInput) (*fastly.Datadog, error)
	DeleteDatadogFn func(*fastly.DeleteDatadogInput) error

	CreateHTTPSFn func(*fastly.CreateHTTPSInput) (*fastly.HTTPS, error)
	ListHTTPSFn   func(*fastly.ListHTTPSInput) ([]*fastly.HTTPS, error)
	GetHTTPSFn    func(*fastly.GetHTTPSInput) (*fastly.HTTPS, error)
	UpdateHTTPSFn func(*fastly.UpdateHTTPSInput) (*fastly.HTTPS, error)
	DeleteHTTPSFn func(*fastly.DeleteHTTPSInput) error

	CreateKafkaFn func(*fastly.CreateKafkaInput) (*fastly.Kafka, error)
	ListKafkasFn  func(*fastly.ListKafkasInput) ([]*fastly.Kafka, error)
	GetKafkaFn    func(*fastly.GetKafkaInput) (*fastly.Kafka, error)
	UpdateKafkaFn func(*fastly.UpdateKafkaInput) (*fastly.Kafka, error)
	DeleteKafkaFn func(*fastly.DeleteKafkaInput) error

	CreatePubsubFn func(*fastly.CreatePubsubInput) (*fastly.Pubsub, error)
	ListPubsubsFn  func(*fastly.ListPubsubsInput) ([]*fastly.Pubsub, error)
	GetPubsubFn    func(*fastly.GetPubsubInput) (*fastly.Pubsub, error)
	UpdatePubsubFn func(*fastly.UpdatePubsubInput) (*fastly.Pubsub, error)
	DeletePubsubFn func(*fastly.DeletePubsubInput) error

	CreateOpenstackFn func(*fastly.CreateOpenstackInput) (*fastly.Openstack, error)
	ListOpenstacksFn  func(*fastly.ListOpenstackInput) ([]*fastly.Openstack, error)
	GetOpenstackFn    func(*fastly.GetOpenstackInput) (*fastly.Openstack, error)
	UpdateOpenstackFn func(*fastly.UpdateOpenstackInput) (*fastly.Openstack, error)
	DeleteOpenstackFn func(*fastly.DeleteOpenstackInput) error

	GetRegionsFn   func() (*fastly.RegionsResponse, error)
	GetStatsJSONFn func(i *fastly.GetStatsInput, dst interface{}) error

	CreateManagedLoggingFn func(*fastly.CreateManagedLoggingInput) (*fastly.ManagedLogging, error)

	CreateVCLFn func(*fastly.CreateVCLInput) (*fastly.VCL, error)
	ListVCLsFn  func(*fastly.ListVCLsInput) ([]*fastly.VCL, error)
	GetVCLFn    func(*fastly.GetVCLInput) (*fastly.VCL, error)
	UpdateVCLFn func(*fastly.UpdateVCLInput) (*fastly.VCL, error)
	DeleteVCLFn func(*fastly.DeleteVCLInput) error

	CreateSnippetFn        func(i *fastly.CreateSnippetInput) (*fastly.Snippet, error)
	ListSnippetsFn         func(i *fastly.ListSnippetsInput) ([]*fastly.Snippet, error)
	GetSnippetFn           func(i *fastly.GetSnippetInput) (*fastly.Snippet, error)
	GetDynamicSnippetFn    func(i *fastly.GetDynamicSnippetInput) (*fastly.DynamicSnippet, error)
	UpdateSnippetFn        func(i *fastly.UpdateSnippetInput) (*fastly.Snippet, error)
	UpdateDynamicSnippetFn func(i *fastly.UpdateDynamicSnippetInput) (*fastly.DynamicSnippet, error)
	DeleteSnippetFn        func(i *fastly.DeleteSnippetInput) error

	PurgeFn     func(i *fastly.PurgeInput) (*fastly.Purge, error)
	PurgeKeyFn  func(i *fastly.PurgeKeyInput) (*fastly.Purge, error)
	PurgeKeysFn func(i *fastly.PurgeKeysInput) (map[string]string, error)
	PurgeAllFn  func(i *fastly.PurgeAllInput) (*fastly.Purge, error)

	CreateACLFn func(i *fastly.CreateACLInput) (*fastly.ACL, error)
	DeleteACLFn func(i *fastly.DeleteACLInput) error
	GetACLFn    func(i *fastly.GetACLInput) (*fastly.ACL, error)
	ListACLsFn  func(i *fastly.ListACLsInput) ([]*fastly.ACL, error)
	UpdateACLFn func(i *fastly.UpdateACLInput) (*fastly.ACL, error)

	CreateACLEntryFn        func(i *fastly.CreateACLEntryInput) (*fastly.ACLEntry, error)
	DeleteACLEntryFn        func(i *fastly.DeleteACLEntryInput) error
	GetACLEntryFn           func(i *fastly.GetACLEntryInput) (*fastly.ACLEntry, error)
	ListACLEntriesFn        func(i *fastly.ListACLEntriesInput) ([]*fastly.ACLEntry, error)
	UpdateACLEntryFn        func(i *fastly.UpdateACLEntryInput) (*fastly.ACLEntry, error)
	BatchModifyACLEntriesFn func(i *fastly.BatchModifyACLEntriesInput) error

	CreateNewRelicFn func(i *fastly.CreateNewRelicInput) (*fastly.NewRelic, error)
	DeleteNewRelicFn func(i *fastly.DeleteNewRelicInput) error
	GetNewRelicFn    func(i *fastly.GetNewRelicInput) (*fastly.NewRelic, error)
	ListNewRelicFn   func(i *fastly.ListNewRelicInput) ([]*fastly.NewRelic, error)
	UpdateNewRelicFn func(i *fastly.UpdateNewRelicInput) (*fastly.NewRelic, error)

	CreateUserFn        func(i *fastly.CreateUserInput) (*fastly.User, error)
	DeleteUserFn        func(i *fastly.DeleteUserInput) error
	GetCurrentUserFn    func() (*fastly.User, error)
	GetUserFn           func(i *fastly.GetUserInput) (*fastly.User, error)
	ListCustomerUsersFn func(i *fastly.ListCustomerUsersInput) ([]*fastly.User, error)
	UpdateUserFn        func(i *fastly.UpdateUserInput) (*fastly.User, error)
	ResetUserPasswordFn func(i *fastly.ResetUserPasswordInput) error

	BatchDeleteTokensFn  func(i *fastly.BatchDeleteTokensInput) error
	CreateTokenFn        func(i *fastly.CreateTokenInput) (*fastly.Token, error)
	DeleteTokenFn        func(i *fastly.DeleteTokenInput) error
	DeleteTokenSelfFn    func() error
	GetTokenSelfFn       func() (*fastly.Token, error)
	ListCustomerTokensFn func(i *fastly.ListCustomerTokensInput) ([]*fastly.Token, error)
	ListTokensFn         func() ([]*fastly.Token, error)

	NewListACLEntriesPaginatorFn      func(i *fastly.ListACLEntriesInput) fastly.PaginatorACLEntries
	NewListDictionaryItemsPaginatorFn func(i *fastly.ListDictionaryItemsInput) fastly.PaginatorDictionaryItems
	NewListServicesPaginatorFn        func(i *fastly.ListServicesInput) fastly.PaginatorServices

	GetCustomTLSConfigurationFn    func(i *fastly.GetCustomTLSConfigurationInput) (*fastly.CustomTLSConfiguration, error)
	ListCustomTLSConfigurationsFn  func(i *fastly.ListCustomTLSConfigurationsInput) ([]*fastly.CustomTLSConfiguration, error)
	UpdateCustomTLSConfigurationFn func(i *fastly.UpdateCustomTLSConfigurationInput) (*fastly.CustomTLSConfiguration, error)
	GetTLSActivationFn             func(i *fastly.GetTLSActivationInput) (*fastly.TLSActivation, error)
	ListTLSActivationsFn           func(i *fastly.ListTLSActivationsInput) ([]*fastly.TLSActivation, error)
	UpdateTLSActivationFn          func(i *fastly.UpdateTLSActivationInput) (*fastly.TLSActivation, error)
	CreateTLSActivationFn          func(i *fastly.CreateTLSActivationInput) (*fastly.TLSActivation, error)
}

// AllDatacenters implements Interface.
func (m API) AllDatacenters() ([]fastly.Datacenter, error) {
	return m.AllDatacentersFn()
}

// AllIPs implements Interface.
func (m API) AllIPs() (fastly.IPAddrs, fastly.IPAddrs, error) {
	return m.AllIPsFn()
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

// SearchService implements Interface.
func (m API) SearchService(i *fastly.SearchServiceInput) (*fastly.Service, error) {
	return m.SearchServiceFn(i)
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

// GetVersion implements Interface.
func (m API) GetVersion(i *fastly.GetVersionInput) (*fastly.Version, error) {
	return m.GetVersionFn(i)
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

// ValidateDomain implements Interface.
func (m API) ValidateDomain(i *fastly.ValidateDomainInput) (*fastly.DomainValidationResult, error) {
	return m.ValidateDomainFn(i)
}

// ValidateAllDomains implements Interface.
func (m API) ValidateAllDomains(i *fastly.ValidateAllDomainsInput) (results []*fastly.DomainValidationResult, err error) {
	return m.ValidateAllDomainsFn(i)
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

// GetPackage implements Interface.
func (m API) GetPackage(i *fastly.GetPackageInput) (*fastly.Package, error) {
	return m.GetPackageFn(i)
}

// UpdatePackage implements Interface.
func (m API) UpdatePackage(i *fastly.UpdatePackageInput) (*fastly.Package, error) {
	return m.UpdatePackageFn(i)
}

// CreateDictionary implements Interface.
func (m API) CreateDictionary(i *fastly.CreateDictionaryInput) (*fastly.Dictionary, error) {
	return m.CreateDictionaryFn(i)
}

// GetDictionary implements Interface.
func (m API) GetDictionary(i *fastly.GetDictionaryInput) (*fastly.Dictionary, error) {
	return m.GetDictionaryFn(i)
}

// DeleteDictionary implements Interface.
func (m API) DeleteDictionary(i *fastly.DeleteDictionaryInput) error {
	return m.DeleteDictionaryFn(i)
}

// ListDictionaries implements Interface.
func (m API) ListDictionaries(i *fastly.ListDictionariesInput) ([]*fastly.Dictionary, error) {
	return m.ListDictionariesFn(i)
}

// UpdateDictionary implements Interface.
func (m API) UpdateDictionary(i *fastly.UpdateDictionaryInput) (*fastly.Dictionary, error) {
	return m.UpdateDictionaryFn(i)
}

// ListDictionaryItems implements Interface.
func (m API) ListDictionaryItems(i *fastly.ListDictionaryItemsInput) ([]*fastly.DictionaryItem, error) {
	return m.ListDictionaryItemsFn(i)
}

// GetDictionaryItem implements Interface.
func (m API) GetDictionaryItem(i *fastly.GetDictionaryItemInput) (*fastly.DictionaryItem, error) {
	return m.GetDictionaryItemFn(i)
}

// CreateDictionaryItem implements Interface.
func (m API) CreateDictionaryItem(i *fastly.CreateDictionaryItemInput) (*fastly.DictionaryItem, error) {
	return m.CreateDictionaryItemFn(i)
}

// UpdateDictionaryItem implements Interface.
func (m API) UpdateDictionaryItem(i *fastly.UpdateDictionaryItemInput) (*fastly.DictionaryItem, error) {
	return m.UpdateDictionaryItemFn(i)
}

// DeleteDictionaryItem implements Interface.
func (m API) DeleteDictionaryItem(i *fastly.DeleteDictionaryItemInput) error {
	return m.DeleteDictionaryItemFn(i)
}

// BatchModifyDictionaryItems implements Interface.
func (m API) BatchModifyDictionaryItems(i *fastly.BatchModifyDictionaryItemsInput) error {
	return m.BatchModifyDictionaryItemsFn(i)
}

// GetDictionaryInfo implements Interface.
func (m API) GetDictionaryInfo(i *fastly.GetDictionaryInfoInput) (*fastly.DictionaryInfo, error) {
	return m.GetDictionaryInfoFn(i)
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

// CreateKinesis implements Interface.
func (m API) CreateKinesis(i *fastly.CreateKinesisInput) (*fastly.Kinesis, error) {
	return m.CreateKinesisFn(i)
}

// ListKinesis implements Interface.
func (m API) ListKinesis(i *fastly.ListKinesisInput) ([]*fastly.Kinesis, error) {
	return m.ListKinesisFn(i)
}

// GetKinesis implements Interface.
func (m API) GetKinesis(i *fastly.GetKinesisInput) (*fastly.Kinesis, error) {
	return m.GetKinesisFn(i)
}

// UpdateKinesis implements Interface.
func (m API) UpdateKinesis(i *fastly.UpdateKinesisInput) (*fastly.Kinesis, error) {
	return m.UpdateKinesisFn(i)
}

// DeleteKinesis implements Interface.
func (m API) DeleteKinesis(i *fastly.DeleteKinesisInput) error {
	return m.DeleteKinesisFn(i)
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

// CreateDigitalOcean implements Interface.
func (m API) CreateDigitalOcean(i *fastly.CreateDigitalOceanInput) (*fastly.DigitalOcean, error) {
	return m.CreateDigitalOceanFn(i)
}

// ListDigitalOceans implements Interface.
func (m API) ListDigitalOceans(i *fastly.ListDigitalOceansInput) ([]*fastly.DigitalOcean, error) {
	return m.ListDigitalOceansFn(i)
}

// GetDigitalOcean implements Interface.
func (m API) GetDigitalOcean(i *fastly.GetDigitalOceanInput) (*fastly.DigitalOcean, error) {
	return m.GetDigitalOceanFn(i)
}

// UpdateDigitalOcean implements Interface.
func (m API) UpdateDigitalOcean(i *fastly.UpdateDigitalOceanInput) (*fastly.DigitalOcean, error) {
	return m.UpdateDigitalOceanFn(i)
}

// DeleteDigitalOcean implements Interface.
func (m API) DeleteDigitalOcean(i *fastly.DeleteDigitalOceanInput) error {
	return m.DeleteDigitalOceanFn(i)
}

// CreateElasticsearch implements Interface.
func (m API) CreateElasticsearch(i *fastly.CreateElasticsearchInput) (*fastly.Elasticsearch, error) {
	return m.CreateElasticsearchFn(i)
}

// ListElasticsearch implements Interface.
func (m API) ListElasticsearch(i *fastly.ListElasticsearchInput) ([]*fastly.Elasticsearch, error) {
	return m.ListElasticsearchFn(i)
}

// GetElasticsearch implements Interface.
func (m API) GetElasticsearch(i *fastly.GetElasticsearchInput) (*fastly.Elasticsearch, error) {
	return m.GetElasticsearchFn(i)
}

// UpdateElasticsearch implements Interface.
func (m API) UpdateElasticsearch(i *fastly.UpdateElasticsearchInput) (*fastly.Elasticsearch, error) {
	return m.UpdateElasticsearchFn(i)
}

// DeleteElasticsearch implements Interface.
func (m API) DeleteElasticsearch(i *fastly.DeleteElasticsearchInput) error {
	return m.DeleteElasticsearchFn(i)
}

// CreateBlobStorage implements Interface.
func (m API) CreateBlobStorage(i *fastly.CreateBlobStorageInput) (*fastly.BlobStorage, error) {
	return m.CreateBlobStorageFn(i)
}

// ListBlobStorages implements Interface.
func (m API) ListBlobStorages(i *fastly.ListBlobStoragesInput) ([]*fastly.BlobStorage, error) {
	return m.ListBlobStoragesFn(i)
}

// GetBlobStorage implements Interface.
func (m API) GetBlobStorage(i *fastly.GetBlobStorageInput) (*fastly.BlobStorage, error) {
	return m.GetBlobStorageFn(i)
}

// UpdateBlobStorage implements Interface.
func (m API) UpdateBlobStorage(i *fastly.UpdateBlobStorageInput) (*fastly.BlobStorage, error) {
	return m.UpdateBlobStorageFn(i)
}

// DeleteBlobStorage implements Interface.
func (m API) DeleteBlobStorage(i *fastly.DeleteBlobStorageInput) error {
	return m.DeleteBlobStorageFn(i)
}

// CreateDatadog implements Interface.
func (m API) CreateDatadog(i *fastly.CreateDatadogInput) (*fastly.Datadog, error) {
	return m.CreateDatadogFn(i)
}

// ListDatadog implements Interface.
func (m API) ListDatadog(i *fastly.ListDatadogInput) ([]*fastly.Datadog, error) {
	return m.ListDatadogFn(i)
}

// GetDatadog implements Interface.
func (m API) GetDatadog(i *fastly.GetDatadogInput) (*fastly.Datadog, error) {
	return m.GetDatadogFn(i)
}

// UpdateDatadog implements Interface.
func (m API) UpdateDatadog(i *fastly.UpdateDatadogInput) (*fastly.Datadog, error) {
	return m.UpdateDatadogFn(i)
}

// DeleteDatadog implements Interface.
func (m API) DeleteDatadog(i *fastly.DeleteDatadogInput) error {
	return m.DeleteDatadogFn(i)
}

// CreateHTTPS implements Interface.
func (m API) CreateHTTPS(i *fastly.CreateHTTPSInput) (*fastly.HTTPS, error) {
	return m.CreateHTTPSFn(i)
}

// ListHTTPS implements Interface.
func (m API) ListHTTPS(i *fastly.ListHTTPSInput) ([]*fastly.HTTPS, error) {
	return m.ListHTTPSFn(i)
}

// GetHTTPS implements Interface.
func (m API) GetHTTPS(i *fastly.GetHTTPSInput) (*fastly.HTTPS, error) {
	return m.GetHTTPSFn(i)
}

// UpdateHTTPS implements Interface.
func (m API) UpdateHTTPS(i *fastly.UpdateHTTPSInput) (*fastly.HTTPS, error) {
	return m.UpdateHTTPSFn(i)
}

// DeleteHTTPS implements Interface.
func (m API) DeleteHTTPS(i *fastly.DeleteHTTPSInput) error {
	return m.DeleteHTTPSFn(i)
}

// CreateKafka implements Interface.
func (m API) CreateKafka(i *fastly.CreateKafkaInput) (*fastly.Kafka, error) {
	return m.CreateKafkaFn(i)
}

// ListKafkas implements Interface.
func (m API) ListKafkas(i *fastly.ListKafkasInput) ([]*fastly.Kafka, error) {
	return m.ListKafkasFn(i)
}

// GetKafka implements Interface.
func (m API) GetKafka(i *fastly.GetKafkaInput) (*fastly.Kafka, error) {
	return m.GetKafkaFn(i)
}

// UpdateKafka implements Interface.
func (m API) UpdateKafka(i *fastly.UpdateKafkaInput) (*fastly.Kafka, error) {
	return m.UpdateKafkaFn(i)
}

// DeleteKafka implements Interface.
func (m API) DeleteKafka(i *fastly.DeleteKafkaInput) error {
	return m.DeleteKafkaFn(i)
}

// CreatePubsub implements Interface.
func (m API) CreatePubsub(i *fastly.CreatePubsubInput) (*fastly.Pubsub, error) {
	return m.CreatePubsubFn(i)
}

// ListPubsubs implements Interface.
func (m API) ListPubsubs(i *fastly.ListPubsubsInput) ([]*fastly.Pubsub, error) {
	return m.ListPubsubsFn(i)
}

// GetPubsub implements Interface.
func (m API) GetPubsub(i *fastly.GetPubsubInput) (*fastly.Pubsub, error) {
	return m.GetPubsubFn(i)
}

// UpdatePubsub implements Interface.
func (m API) UpdatePubsub(i *fastly.UpdatePubsubInput) (*fastly.Pubsub, error) {
	return m.UpdatePubsubFn(i)
}

// DeletePubsub implements Interface.
func (m API) DeletePubsub(i *fastly.DeletePubsubInput) error {
	return m.DeletePubsubFn(i)
}

// CreateOpenstack implements Interface.
func (m API) CreateOpenstack(i *fastly.CreateOpenstackInput) (*fastly.Openstack, error) {
	return m.CreateOpenstackFn(i)
}

// ListOpenstack implements Interface.
func (m API) ListOpenstack(i *fastly.ListOpenstackInput) ([]*fastly.Openstack, error) {
	return m.ListOpenstacksFn(i)
}

// GetOpenstack implements Interface.
func (m API) GetOpenstack(i *fastly.GetOpenstackInput) (*fastly.Openstack, error) {
	return m.GetOpenstackFn(i)
}

// UpdateOpenstack implements Interface.
func (m API) UpdateOpenstack(i *fastly.UpdateOpenstackInput) (*fastly.Openstack, error) {
	return m.UpdateOpenstackFn(i)
}

// DeleteOpenstack implements Interface.
func (m API) DeleteOpenstack(i *fastly.DeleteOpenstackInput) error {
	return m.DeleteOpenstackFn(i)
}

// GetRegions implements Interface.
func (m API) GetRegions() (*fastly.RegionsResponse, error) {
	return m.GetRegionsFn()
}

// GetStatsJSON implements Interface.
func (m API) GetStatsJSON(i *fastly.GetStatsInput, dst interface{}) error {
	return m.GetStatsJSONFn(i, dst)
}

// CreateManagedLogging implements Interface.
func (m API) CreateManagedLogging(i *fastly.CreateManagedLoggingInput) (*fastly.ManagedLogging, error) {
	return m.CreateManagedLoggingFn(i)
}

// CreateVCL implements Interface.
func (m API) CreateVCL(i *fastly.CreateVCLInput) (*fastly.VCL, error) {
	return m.CreateVCLFn(i)
}

// ListVCLs implements Interface.
func (m API) ListVCLs(i *fastly.ListVCLsInput) ([]*fastly.VCL, error) {
	return m.ListVCLsFn(i)
}

// GetVCL implements Interface.
func (m API) GetVCL(i *fastly.GetVCLInput) (*fastly.VCL, error) {
	return m.GetVCLFn(i)
}

// UpdateVCL implements Interface.
func (m API) UpdateVCL(i *fastly.UpdateVCLInput) (*fastly.VCL, error) {
	return m.UpdateVCLFn(i)
}

// DeleteVCL implements Interface.
func (m API) DeleteVCL(i *fastly.DeleteVCLInput) error {
	return m.DeleteVCLFn(i)
}

// CreateSnippet implements Interface.
func (m API) CreateSnippet(i *fastly.CreateSnippetInput) (*fastly.Snippet, error) {
	return m.CreateSnippetFn(i)
}

// ListSnippets implements Interface.
func (m API) ListSnippets(i *fastly.ListSnippetsInput) ([]*fastly.Snippet, error) {
	return m.ListSnippetsFn(i)
}

// GetSnippet implements Interface.
func (m API) GetSnippet(i *fastly.GetSnippetInput) (*fastly.Snippet, error) {
	return m.GetSnippetFn(i)
}

// GetDynamicSnippet implements Interface.
func (m API) GetDynamicSnippet(i *fastly.GetDynamicSnippetInput) (*fastly.DynamicSnippet, error) {
	return m.GetDynamicSnippetFn(i)
}

// UpdateSnippet implements Interface.
func (m API) UpdateSnippet(i *fastly.UpdateSnippetInput) (*fastly.Snippet, error) {
	return m.UpdateSnippetFn(i)
}

// UpdateDynamicSnippet implements Interface.
func (m API) UpdateDynamicSnippet(i *fastly.UpdateDynamicSnippetInput) (*fastly.DynamicSnippet, error) {
	return m.UpdateDynamicSnippetFn(i)
}

// DeleteSnippet implements Interface.
func (m API) DeleteSnippet(i *fastly.DeleteSnippetInput) error {
	return m.DeleteSnippetFn(i)
}

// Purge implements Interface.
func (m API) Purge(i *fastly.PurgeInput) (*fastly.Purge, error) {
	return m.PurgeFn(i)
}

// PurgeKey implements Interface.
func (m API) PurgeKey(i *fastly.PurgeKeyInput) (*fastly.Purge, error) {
	return m.PurgeKeyFn(i)
}

// PurgeKeys implements Interface.
func (m API) PurgeKeys(i *fastly.PurgeKeysInput) (map[string]string, error) {
	return m.PurgeKeysFn(i)
}

// PurgeAll implements Interface.
func (m API) PurgeAll(i *fastly.PurgeAllInput) (*fastly.Purge, error) {
	return m.PurgeAllFn(i)
}

// CreateACL implements Interface.
func (m API) CreateACL(i *fastly.CreateACLInput) (*fastly.ACL, error) {
	return m.CreateACLFn(i)
}

// DeleteACL implements Interface.
func (m API) DeleteACL(i *fastly.DeleteACLInput) error {
	return m.DeleteACLFn(i)
}

// GetACL implements Interface.
func (m API) GetACL(i *fastly.GetACLInput) (*fastly.ACL, error) {
	return m.GetACLFn(i)
}

// ListACLs implements Interface.
func (m API) ListACLs(i *fastly.ListACLsInput) ([]*fastly.ACL, error) {
	return m.ListACLsFn(i)
}

// UpdateACL implements Interface.
func (m API) UpdateACL(i *fastly.UpdateACLInput) (*fastly.ACL, error) {
	return m.UpdateACLFn(i)
}

// CreateACLEntry implements Interface.
func (m API) CreateACLEntry(i *fastly.CreateACLEntryInput) (*fastly.ACLEntry, error) {
	return m.CreateACLEntryFn(i)
}

// DeleteACLEntry implements Interface.
func (m API) DeleteACLEntry(i *fastly.DeleteACLEntryInput) error {
	return m.DeleteACLEntryFn(i)
}

// GetACLEntry implements Interface.
func (m API) GetACLEntry(i *fastly.GetACLEntryInput) (*fastly.ACLEntry, error) {
	return m.GetACLEntryFn(i)
}

// ListACLEntries implements Interface.
func (m API) ListACLEntries(i *fastly.ListACLEntriesInput) ([]*fastly.ACLEntry, error) {
	return m.ListACLEntriesFn(i)
}

// UpdateACLEntry implements Interface.
func (m API) UpdateACLEntry(i *fastly.UpdateACLEntryInput) (*fastly.ACLEntry, error) {
	return m.UpdateACLEntryFn(i)
}

// BatchModifyACLEntries implements Interface.
func (m API) BatchModifyACLEntries(i *fastly.BatchModifyACLEntriesInput) error {
	return m.BatchModifyACLEntriesFn(i)
}

// CreateNewRelic implements Interface.
func (m API) CreateNewRelic(i *fastly.CreateNewRelicInput) (*fastly.NewRelic, error) {
	return m.CreateNewRelicFn(i)
}

// DeleteNewRelic implements Interface.
func (m API) DeleteNewRelic(i *fastly.DeleteNewRelicInput) error {
	return m.DeleteNewRelicFn(i)
}

// GetNewRelic implements Interface.
func (m API) GetNewRelic(i *fastly.GetNewRelicInput) (*fastly.NewRelic, error) {
	return m.GetNewRelicFn(i)
}

// ListNewRelic implements Interface.
func (m API) ListNewRelic(i *fastly.ListNewRelicInput) ([]*fastly.NewRelic, error) {
	return m.ListNewRelicFn(i)
}

// UpdateNewRelic implements Interface.
func (m API) UpdateNewRelic(i *fastly.UpdateNewRelicInput) (*fastly.NewRelic, error) {
	return m.UpdateNewRelicFn(i)
}

// CreateUser implements Interface.
func (m API) CreateUser(i *fastly.CreateUserInput) (*fastly.User, error) {
	return m.CreateUserFn(i)
}

// DeleteUser implements Interface.
func (m API) DeleteUser(i *fastly.DeleteUserInput) error {
	return m.DeleteUserFn(i)
}

// GetCurrentUser implements Interface.
func (m API) GetCurrentUser() (*fastly.User, error) {
	return m.GetCurrentUserFn()
}

// GetUser implements Interface.
func (m API) GetUser(i *fastly.GetUserInput) (*fastly.User, error) {
	return m.GetUserFn(i)
}

// ListCustomerUsers implements Interface.
func (m API) ListCustomerUsers(i *fastly.ListCustomerUsersInput) ([]*fastly.User, error) {
	return m.ListCustomerUsersFn(i)
}

// UpdateUser implements Interface.
func (m API) UpdateUser(i *fastly.UpdateUserInput) (*fastly.User, error) {
	return m.UpdateUserFn(i)
}

// ResetUserPassword implements Interface.
func (m API) ResetUserPassword(i *fastly.ResetUserPasswordInput) error {
	return m.ResetUserPasswordFn(i)
}

// BatchDeleteTokens implements Interface.
func (m API) BatchDeleteTokens(i *fastly.BatchDeleteTokensInput) error {
	return m.BatchDeleteTokensFn(i)
}

// CreateToken implements Interface.
func (m API) CreateToken(i *fastly.CreateTokenInput) (*fastly.Token, error) {
	return m.CreateTokenFn(i)
}

// DeleteToken implements Interface.
func (m API) DeleteToken(i *fastly.DeleteTokenInput) error {
	return m.DeleteTokenFn(i)
}

// DeleteTokenSelf implements Interface.
func (m API) DeleteTokenSelf() error {
	return m.DeleteTokenSelfFn()
}

// GetTokenSelf implements Interface.
func (m API) GetTokenSelf() (*fastly.Token, error) {
	return m.GetTokenSelfFn()
}

// ListCustomerTokens implements Interface.
func (m API) ListCustomerTokens(i *fastly.ListCustomerTokensInput) ([]*fastly.Token, error) {
	return m.ListCustomerTokensFn(i)
}

// ListTokens implements Interface.
func (m API) ListTokens() ([]*fastly.Token, error) {
	return m.ListTokensFn()
}

// NewListACLEntriesPaginator implements Interface.
func (m API) NewListACLEntriesPaginator(i *fastly.ListACLEntriesInput) fastly.PaginatorACLEntries {
	return m.NewListACLEntriesPaginatorFn(i)
}

// NewListDictionaryItemsPaginator implements Interface.
func (m API) NewListDictionaryItemsPaginator(i *fastly.ListDictionaryItemsInput) fastly.PaginatorDictionaryItems {
	return m.NewListDictionaryItemsPaginatorFn(i)
}

// NewListServicesPaginator implements Interface.
func (m API) NewListServicesPaginator(i *fastly.ListServicesInput) fastly.PaginatorServices {
	return m.NewListServicesPaginatorFn(i)
}

// GetCustomTLSConfiguration implements Interface.
func (m API) GetCustomTLSConfiguration(i *fastly.GetCustomTLSConfigurationInput) (*fastly.CustomTLSConfiguration, error) {
	return m.GetCustomTLSConfigurationFn(i)
}

// ListCustomTLSConfigurations implements Interface.
func (m API) ListCustomTLSConfigurations(i *fastly.ListCustomTLSConfigurationsInput) ([]*fastly.CustomTLSConfiguration, error) {
	return m.ListCustomTLSConfigurationsFn(i)
}

// UpdateCustomTLSConfiguration implements Interface.
func (m API) UpdateCustomTLSConfiguration(i *fastly.UpdateCustomTLSConfigurationInput) (*fastly.CustomTLSConfiguration, error) {
	return m.UpdateCustomTLSConfigurationFn(i)
}

// GetTLSActivation implements Interface.
func (m API) GetTLSActivation(i *fastly.GetTLSActivationInput) (*fastly.TLSActivation, error) {
	return m.GetTLSActivationFn(i)
}

// ListTLSActivations implements Interface.
func (m API) ListTLSActivations(i *fastly.ListTLSActivationsInput) ([]*fastly.TLSActivation, error) {
	return m.ListTLSActivationsFn(i)
}

// UpdateTLSActivation implements Interface.
func (m API) UpdateTLSActivation(i *fastly.UpdateTLSActivationInput) (*fastly.TLSActivation, error) {
	return m.UpdateTLSActivationFn(i)
}

// CreateTLSActivation implements Interface.
func (m API) CreateTLSActivation(i *fastly.CreateTLSActivationInput) (*fastly.TLSActivation, error) {
	return m.CreateTLSActivationFn(i)
}
