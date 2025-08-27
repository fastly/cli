package mock

import (
	"context"
	"crypto/ed25519"

	"github.com/fastly/go-fastly/v11/fastly"
)

// API is a mock implementation of api.Interface that's used for testing.
// The zero value is useful, but will panic on all methods. Provide function
// implementations for the method(s) your test will call.
type API struct {
	AllDatacentersFn func(context.Context) ([]fastly.Datacenter, error)
	AllIPsFn         func(context.Context) (fastly.IPAddrs, fastly.IPAddrs, error)

	CreateServiceFn     func(context.Context, *fastly.CreateServiceInput) (*fastly.Service, error)
	GetServicesFn       func(context.Context, *fastly.GetServicesInput) *fastly.ListPaginator[fastly.Service]
	ListServicesFn      func(context.Context, *fastly.ListServicesInput) ([]*fastly.Service, error)
	GetServiceFn        func(context.Context, *fastly.GetServiceInput) (*fastly.Service, error)
	GetServiceDetailsFn func(context.Context, *fastly.GetServiceInput) (*fastly.ServiceDetail, error)
	UpdateServiceFn     func(context.Context, *fastly.UpdateServiceInput) (*fastly.Service, error)
	DeleteServiceFn     func(context.Context, *fastly.DeleteServiceInput) error
	SearchServiceFn     func(context.Context, *fastly.SearchServiceInput) (*fastly.Service, error)

	CloneVersionFn      func(context.Context, *fastly.CloneVersionInput) (*fastly.Version, error)
	ListVersionsFn      func(context.Context, *fastly.ListVersionsInput) ([]*fastly.Version, error)
	GetVersionFn        func(context.Context, *fastly.GetVersionInput) (*fastly.Version, error)
	UpdateVersionFn     func(context.Context, *fastly.UpdateVersionInput) (*fastly.Version, error)
	ActivateVersionFn   func(context.Context, *fastly.ActivateVersionInput) (*fastly.Version, error)
	DeactivateVersionFn func(context.Context, *fastly.DeactivateVersionInput) (*fastly.Version, error)
	LockVersionFn       func(context.Context, *fastly.LockVersionInput) (*fastly.Version, error)
	LatestVersionFn     func(context.Context, *fastly.LatestVersionInput) (*fastly.Version, error)

	CreateDomainFn       func(context.Context, *fastly.CreateDomainInput) (*fastly.Domain, error)
	ListDomainsFn        func(context.Context, *fastly.ListDomainsInput) ([]*fastly.Domain, error)
	GetDomainFn          func(context.Context, *fastly.GetDomainInput) (*fastly.Domain, error)
	UpdateDomainFn       func(context.Context, *fastly.UpdateDomainInput) (*fastly.Domain, error)
	DeleteDomainFn       func(context.Context, *fastly.DeleteDomainInput) error
	ValidateDomainFn     func(context.Context, *fastly.ValidateDomainInput) (*fastly.DomainValidationResult, error)
	ValidateAllDomainsFn func(context.Context, *fastly.ValidateAllDomainsInput) ([]*fastly.DomainValidationResult, error)

	CreateBackendFn func(context.Context, *fastly.CreateBackendInput) (*fastly.Backend, error)
	ListBackendsFn  func(context.Context, *fastly.ListBackendsInput) ([]*fastly.Backend, error)
	GetBackendFn    func(context.Context, *fastly.GetBackendInput) (*fastly.Backend, error)
	UpdateBackendFn func(context.Context, *fastly.UpdateBackendInput) (*fastly.Backend, error)
	DeleteBackendFn func(context.Context, *fastly.DeleteBackendInput) error

	CreateHealthCheckFn func(context.Context, *fastly.CreateHealthCheckInput) (*fastly.HealthCheck, error)
	ListHealthChecksFn  func(context.Context, *fastly.ListHealthChecksInput) ([]*fastly.HealthCheck, error)
	GetHealthCheckFn    func(context.Context, *fastly.GetHealthCheckInput) (*fastly.HealthCheck, error)
	UpdateHealthCheckFn func(context.Context, *fastly.UpdateHealthCheckInput) (*fastly.HealthCheck, error)
	DeleteHealthCheckFn func(context.Context, *fastly.DeleteHealthCheckInput) error

	GetPackageFn    func(context.Context, *fastly.GetPackageInput) (*fastly.Package, error)
	UpdatePackageFn func(context.Context, *fastly.UpdatePackageInput) (*fastly.Package, error)

	CreateDictionaryFn func(context.Context, *fastly.CreateDictionaryInput) (*fastly.Dictionary, error)
	GetDictionaryFn    func(context.Context, *fastly.GetDictionaryInput) (*fastly.Dictionary, error)
	DeleteDictionaryFn func(context.Context, *fastly.DeleteDictionaryInput) error
	ListDictionariesFn func(context.Context, *fastly.ListDictionariesInput) ([]*fastly.Dictionary, error)
	UpdateDictionaryFn func(context.Context, *fastly.UpdateDictionaryInput) (*fastly.Dictionary, error)

	GetDictionaryItemsFn         func(context.Context, *fastly.GetDictionaryItemsInput) *fastly.ListPaginator[fastly.DictionaryItem]
	ListDictionaryItemsFn        func(context.Context, *fastly.ListDictionaryItemsInput) ([]*fastly.DictionaryItem, error)
	GetDictionaryItemFn          func(context.Context, *fastly.GetDictionaryItemInput) (*fastly.DictionaryItem, error)
	CreateDictionaryItemFn       func(context.Context, *fastly.CreateDictionaryItemInput) (*fastly.DictionaryItem, error)
	UpdateDictionaryItemFn       func(context.Context, *fastly.UpdateDictionaryItemInput) (*fastly.DictionaryItem, error)
	DeleteDictionaryItemFn       func(context.Context, *fastly.DeleteDictionaryItemInput) error
	BatchModifyDictionaryItemsFn func(context.Context, *fastly.BatchModifyDictionaryItemsInput) error

	GetDictionaryInfoFn func(context.Context, *fastly.GetDictionaryInfoInput) (*fastly.DictionaryInfo, error)

	CreateBigQueryFn func(context.Context, *fastly.CreateBigQueryInput) (*fastly.BigQuery, error)
	ListBigQueriesFn func(context.Context, *fastly.ListBigQueriesInput) ([]*fastly.BigQuery, error)
	GetBigQueryFn    func(context.Context, *fastly.GetBigQueryInput) (*fastly.BigQuery, error)
	UpdateBigQueryFn func(context.Context, *fastly.UpdateBigQueryInput) (*fastly.BigQuery, error)
	DeleteBigQueryFn func(context.Context, *fastly.DeleteBigQueryInput) error

	CreateS3Fn func(context.Context, *fastly.CreateS3Input) (*fastly.S3, error)
	ListS3sFn  func(context.Context, *fastly.ListS3sInput) ([]*fastly.S3, error)
	GetS3Fn    func(context.Context, *fastly.GetS3Input) (*fastly.S3, error)
	UpdateS3Fn func(context.Context, *fastly.UpdateS3Input) (*fastly.S3, error)
	DeleteS3Fn func(context.Context, *fastly.DeleteS3Input) error

	CreateKinesisFn func(context.Context, *fastly.CreateKinesisInput) (*fastly.Kinesis, error)
	ListKinesisFn   func(context.Context, *fastly.ListKinesisInput) ([]*fastly.Kinesis, error)
	GetKinesisFn    func(context.Context, *fastly.GetKinesisInput) (*fastly.Kinesis, error)
	UpdateKinesisFn func(context.Context, *fastly.UpdateKinesisInput) (*fastly.Kinesis, error)
	DeleteKinesisFn func(context.Context, *fastly.DeleteKinesisInput) error

	CreateSyslogFn func(context.Context, *fastly.CreateSyslogInput) (*fastly.Syslog, error)
	ListSyslogsFn  func(context.Context, *fastly.ListSyslogsInput) ([]*fastly.Syslog, error)
	GetSyslogFn    func(context.Context, *fastly.GetSyslogInput) (*fastly.Syslog, error)
	UpdateSyslogFn func(context.Context, *fastly.UpdateSyslogInput) (*fastly.Syslog, error)
	DeleteSyslogFn func(context.Context, *fastly.DeleteSyslogInput) error

	CreateLogentriesFn func(context.Context, *fastly.CreateLogentriesInput) (*fastly.Logentries, error)
	ListLogentriesFn   func(context.Context, *fastly.ListLogentriesInput) ([]*fastly.Logentries, error)
	GetLogentriesFn    func(context.Context, *fastly.GetLogentriesInput) (*fastly.Logentries, error)
	UpdateLogentriesFn func(context.Context, *fastly.UpdateLogentriesInput) (*fastly.Logentries, error)
	DeleteLogentriesFn func(context.Context, *fastly.DeleteLogentriesInput) error

	CreatePapertrailFn func(context.Context, *fastly.CreatePapertrailInput) (*fastly.Papertrail, error)
	ListPapertrailsFn  func(context.Context, *fastly.ListPapertrailsInput) ([]*fastly.Papertrail, error)
	GetPapertrailFn    func(context.Context, *fastly.GetPapertrailInput) (*fastly.Papertrail, error)
	UpdatePapertrailFn func(context.Context, *fastly.UpdatePapertrailInput) (*fastly.Papertrail, error)
	DeletePapertrailFn func(context.Context, *fastly.DeletePapertrailInput) error

	CreateSumologicFn func(context.Context, *fastly.CreateSumologicInput) (*fastly.Sumologic, error)
	ListSumologicsFn  func(context.Context, *fastly.ListSumologicsInput) ([]*fastly.Sumologic, error)
	GetSumologicFn    func(context.Context, *fastly.GetSumologicInput) (*fastly.Sumologic, error)
	UpdateSumologicFn func(context.Context, *fastly.UpdateSumologicInput) (*fastly.Sumologic, error)
	DeleteSumologicFn func(context.Context, *fastly.DeleteSumologicInput) error

	CreateGCSFn func(context.Context, *fastly.CreateGCSInput) (*fastly.GCS, error)
	ListGCSsFn  func(context.Context, *fastly.ListGCSsInput) ([]*fastly.GCS, error)
	GetGCSFn    func(context.Context, *fastly.GetGCSInput) (*fastly.GCS, error)
	UpdateGCSFn func(context.Context, *fastly.UpdateGCSInput) (*fastly.GCS, error)
	DeleteGCSFn func(context.Context, *fastly.DeleteGCSInput) error

	CreateGrafanaCloudLogsFn func(context.Context, *fastly.CreateGrafanaCloudLogsInput) (*fastly.GrafanaCloudLogs, error)
	ListGrafanaCloudLogsFn   func(context.Context, *fastly.ListGrafanaCloudLogsInput) ([]*fastly.GrafanaCloudLogs, error)
	GetGrafanaCloudLogsFn    func(context.Context, *fastly.GetGrafanaCloudLogsInput) (*fastly.GrafanaCloudLogs, error)
	UpdateGrafanaCloudLogsFn func(context.Context, *fastly.UpdateGrafanaCloudLogsInput) (*fastly.GrafanaCloudLogs, error)
	DeleteGrafanaCloudLogsFn func(context.Context, *fastly.DeleteGrafanaCloudLogsInput) error

	CreateFTPFn func(context.Context, *fastly.CreateFTPInput) (*fastly.FTP, error)
	ListFTPsFn  func(context.Context, *fastly.ListFTPsInput) ([]*fastly.FTP, error)
	GetFTPFn    func(context.Context, *fastly.GetFTPInput) (*fastly.FTP, error)
	UpdateFTPFn func(context.Context, *fastly.UpdateFTPInput) (*fastly.FTP, error)
	DeleteFTPFn func(context.Context, *fastly.DeleteFTPInput) error

	CreateSplunkFn func(context.Context, *fastly.CreateSplunkInput) (*fastly.Splunk, error)
	ListSplunksFn  func(context.Context, *fastly.ListSplunksInput) ([]*fastly.Splunk, error)
	GetSplunkFn    func(context.Context, *fastly.GetSplunkInput) (*fastly.Splunk, error)
	UpdateSplunkFn func(context.Context, *fastly.UpdateSplunkInput) (*fastly.Splunk, error)
	DeleteSplunkFn func(context.Context, *fastly.DeleteSplunkInput) error

	CreateScalyrFn func(context.Context, *fastly.CreateScalyrInput) (*fastly.Scalyr, error)
	ListScalyrsFn  func(context.Context, *fastly.ListScalyrsInput) ([]*fastly.Scalyr, error)
	GetScalyrFn    func(context.Context, *fastly.GetScalyrInput) (*fastly.Scalyr, error)
	UpdateScalyrFn func(context.Context, *fastly.UpdateScalyrInput) (*fastly.Scalyr, error)
	DeleteScalyrFn func(context.Context, *fastly.DeleteScalyrInput) error

	CreateLogglyFn func(context.Context, *fastly.CreateLogglyInput) (*fastly.Loggly, error)
	ListLogglyFn   func(context.Context, *fastly.ListLogglyInput) ([]*fastly.Loggly, error)
	GetLogglyFn    func(context.Context, *fastly.GetLogglyInput) (*fastly.Loggly, error)
	UpdateLogglyFn func(context.Context, *fastly.UpdateLogglyInput) (*fastly.Loggly, error)
	DeleteLogglyFn func(context.Context, *fastly.DeleteLogglyInput) error

	CreateHoneycombFn func(context.Context, *fastly.CreateHoneycombInput) (*fastly.Honeycomb, error)
	ListHoneycombsFn  func(context.Context, *fastly.ListHoneycombsInput) ([]*fastly.Honeycomb, error)
	GetHoneycombFn    func(context.Context, *fastly.GetHoneycombInput) (*fastly.Honeycomb, error)
	UpdateHoneycombFn func(context.Context, *fastly.UpdateHoneycombInput) (*fastly.Honeycomb, error)
	DeleteHoneycombFn func(context.Context, *fastly.DeleteHoneycombInput) error

	CreateHerokuFn func(context.Context, *fastly.CreateHerokuInput) (*fastly.Heroku, error)
	ListHerokusFn  func(context.Context, *fastly.ListHerokusInput) ([]*fastly.Heroku, error)
	GetHerokuFn    func(context.Context, *fastly.GetHerokuInput) (*fastly.Heroku, error)
	UpdateHerokuFn func(context.Context, *fastly.UpdateHerokuInput) (*fastly.Heroku, error)
	DeleteHerokuFn func(context.Context, *fastly.DeleteHerokuInput) error

	CreateSFTPFn func(context.Context, *fastly.CreateSFTPInput) (*fastly.SFTP, error)
	ListSFTPsFn  func(context.Context, *fastly.ListSFTPsInput) ([]*fastly.SFTP, error)
	GetSFTPFn    func(context.Context, *fastly.GetSFTPInput) (*fastly.SFTP, error)
	UpdateSFTPFn func(context.Context, *fastly.UpdateSFTPInput) (*fastly.SFTP, error)
	DeleteSFTPFn func(context.Context, *fastly.DeleteSFTPInput) error

	CreateLogshuttleFn func(context.Context, *fastly.CreateLogshuttleInput) (*fastly.Logshuttle, error)
	ListLogshuttlesFn  func(context.Context, *fastly.ListLogshuttlesInput) ([]*fastly.Logshuttle, error)
	GetLogshuttleFn    func(context.Context, *fastly.GetLogshuttleInput) (*fastly.Logshuttle, error)
	UpdateLogshuttleFn func(context.Context, *fastly.UpdateLogshuttleInput) (*fastly.Logshuttle, error)
	DeleteLogshuttleFn func(context.Context, *fastly.DeleteLogshuttleInput) error

	CreateCloudfilesFn func(context.Context, *fastly.CreateCloudfilesInput) (*fastly.Cloudfiles, error)
	ListCloudfilesFn   func(context.Context, *fastly.ListCloudfilesInput) ([]*fastly.Cloudfiles, error)
	GetCloudfilesFn    func(context.Context, *fastly.GetCloudfilesInput) (*fastly.Cloudfiles, error)
	UpdateCloudfilesFn func(context.Context, *fastly.UpdateCloudfilesInput) (*fastly.Cloudfiles, error)
	DeleteCloudfilesFn func(context.Context, *fastly.DeleteCloudfilesInput) error

	CreateDigitalOceanFn func(context.Context, *fastly.CreateDigitalOceanInput) (*fastly.DigitalOcean, error)
	ListDigitalOceansFn  func(context.Context, *fastly.ListDigitalOceansInput) ([]*fastly.DigitalOcean, error)
	GetDigitalOceanFn    func(context.Context, *fastly.GetDigitalOceanInput) (*fastly.DigitalOcean, error)
	UpdateDigitalOceanFn func(context.Context, *fastly.UpdateDigitalOceanInput) (*fastly.DigitalOcean, error)
	DeleteDigitalOceanFn func(context.Context, *fastly.DeleteDigitalOceanInput) error

	CreateElasticsearchFn func(context.Context, *fastly.CreateElasticsearchInput) (*fastly.Elasticsearch, error)
	ListElasticsearchFn   func(context.Context, *fastly.ListElasticsearchInput) ([]*fastly.Elasticsearch, error)
	GetElasticsearchFn    func(context.Context, *fastly.GetElasticsearchInput) (*fastly.Elasticsearch, error)
	UpdateElasticsearchFn func(context.Context, *fastly.UpdateElasticsearchInput) (*fastly.Elasticsearch, error)
	DeleteElasticsearchFn func(context.Context, *fastly.DeleteElasticsearchInput) error

	CreateBlobStorageFn func(context.Context, *fastly.CreateBlobStorageInput) (*fastly.BlobStorage, error)
	ListBlobStoragesFn  func(context.Context, *fastly.ListBlobStoragesInput) ([]*fastly.BlobStorage, error)
	GetBlobStorageFn    func(context.Context, *fastly.GetBlobStorageInput) (*fastly.BlobStorage, error)
	UpdateBlobStorageFn func(context.Context, *fastly.UpdateBlobStorageInput) (*fastly.BlobStorage, error)
	DeleteBlobStorageFn func(context.Context, *fastly.DeleteBlobStorageInput) error

	CreateDatadogFn func(context.Context, *fastly.CreateDatadogInput) (*fastly.Datadog, error)
	ListDatadogFn   func(context.Context, *fastly.ListDatadogInput) ([]*fastly.Datadog, error)
	GetDatadogFn    func(context.Context, *fastly.GetDatadogInput) (*fastly.Datadog, error)
	UpdateDatadogFn func(context.Context, *fastly.UpdateDatadogInput) (*fastly.Datadog, error)
	DeleteDatadogFn func(context.Context, *fastly.DeleteDatadogInput) error

	CreateHTTPSFn func(context.Context, *fastly.CreateHTTPSInput) (*fastly.HTTPS, error)
	ListHTTPSFn   func(context.Context, *fastly.ListHTTPSInput) ([]*fastly.HTTPS, error)
	GetHTTPSFn    func(context.Context, *fastly.GetHTTPSInput) (*fastly.HTTPS, error)
	UpdateHTTPSFn func(context.Context, *fastly.UpdateHTTPSInput) (*fastly.HTTPS, error)
	DeleteHTTPSFn func(context.Context, *fastly.DeleteHTTPSInput) error

	CreateKafkaFn func(context.Context, *fastly.CreateKafkaInput) (*fastly.Kafka, error)
	ListKafkasFn  func(context.Context, *fastly.ListKafkasInput) ([]*fastly.Kafka, error)
	GetKafkaFn    func(context.Context, *fastly.GetKafkaInput) (*fastly.Kafka, error)
	UpdateKafkaFn func(context.Context, *fastly.UpdateKafkaInput) (*fastly.Kafka, error)
	DeleteKafkaFn func(context.Context, *fastly.DeleteKafkaInput) error

	CreatePubsubFn func(context.Context, *fastly.CreatePubsubInput) (*fastly.Pubsub, error)
	ListPubsubsFn  func(context.Context, *fastly.ListPubsubsInput) ([]*fastly.Pubsub, error)
	GetPubsubFn    func(context.Context, *fastly.GetPubsubInput) (*fastly.Pubsub, error)
	UpdatePubsubFn func(context.Context, *fastly.UpdatePubsubInput) (*fastly.Pubsub, error)
	DeletePubsubFn func(context.Context, *fastly.DeletePubsubInput) error

	CreateOpenstackFn func(context.Context, *fastly.CreateOpenstackInput) (*fastly.Openstack, error)
	ListOpenstacksFn  func(context.Context, *fastly.ListOpenstackInput) ([]*fastly.Openstack, error)
	GetOpenstackFn    func(context.Context, *fastly.GetOpenstackInput) (*fastly.Openstack, error)
	UpdateOpenstackFn func(context.Context, *fastly.UpdateOpenstackInput) (*fastly.Openstack, error)
	DeleteOpenstackFn func(context.Context, *fastly.DeleteOpenstackInput) error

	GetRegionsFn   func(context.Context) (*fastly.RegionsResponse, error)
	GetStatsJSONFn func(context.Context, *fastly.GetStatsInput, any) error

	CreateManagedLoggingFn func(context.Context, *fastly.CreateManagedLoggingInput) (*fastly.ManagedLogging, error)

	GetGeneratedVCLFn func(context.Context, *fastly.GetGeneratedVCLInput) (*fastly.VCL, error)

	CreateVCLFn func(context.Context, *fastly.CreateVCLInput) (*fastly.VCL, error)
	ListVCLsFn  func(context.Context, *fastly.ListVCLsInput) ([]*fastly.VCL, error)
	GetVCLFn    func(context.Context, *fastly.GetVCLInput) (*fastly.VCL, error)
	UpdateVCLFn func(context.Context, *fastly.UpdateVCLInput) (*fastly.VCL, error)
	DeleteVCLFn func(context.Context, *fastly.DeleteVCLInput) error

	CreateSnippetFn        func(context.Context, *fastly.CreateSnippetInput) (*fastly.Snippet, error)
	ListSnippetsFn         func(context.Context, *fastly.ListSnippetsInput) ([]*fastly.Snippet, error)
	GetSnippetFn           func(context.Context, *fastly.GetSnippetInput) (*fastly.Snippet, error)
	GetDynamicSnippetFn    func(context.Context, *fastly.GetDynamicSnippetInput) (*fastly.DynamicSnippet, error)
	UpdateSnippetFn        func(context.Context, *fastly.UpdateSnippetInput) (*fastly.Snippet, error)
	UpdateDynamicSnippetFn func(context.Context, *fastly.UpdateDynamicSnippetInput) (*fastly.DynamicSnippet, error)
	DeleteSnippetFn        func(context.Context, *fastly.DeleteSnippetInput) error

	PurgeFn     func(context.Context, *fastly.PurgeInput) (*fastly.Purge, error)
	PurgeKeyFn  func(context.Context, *fastly.PurgeKeyInput) (*fastly.Purge, error)
	PurgeKeysFn func(context.Context, *fastly.PurgeKeysInput) (map[string]string, error)
	PurgeAllFn  func(context.Context, *fastly.PurgeAllInput) (*fastly.Purge, error)

	CreateACLFn func(context.Context, *fastly.CreateACLInput) (*fastly.ACL, error)
	DeleteACLFn func(context.Context, *fastly.DeleteACLInput) error
	GetACLFn    func(context.Context, *fastly.GetACLInput) (*fastly.ACL, error)
	ListACLsFn  func(context.Context, *fastly.ListACLsInput) ([]*fastly.ACL, error)
	UpdateACLFn func(context.Context, *fastly.UpdateACLInput) (*fastly.ACL, error)

	CreateACLEntryFn        func(context.Context, *fastly.CreateACLEntryInput) (*fastly.ACLEntry, error)
	DeleteACLEntryFn        func(context.Context, *fastly.DeleteACLEntryInput) error
	GetACLEntryFn           func(context.Context, *fastly.GetACLEntryInput) (*fastly.ACLEntry, error)
	GetACLEntriesFn         func(context.Context, *fastly.GetACLEntriesInput) *fastly.ListPaginator[fastly.ACLEntry]
	ListACLEntriesFn        func(context.Context, *fastly.ListACLEntriesInput) ([]*fastly.ACLEntry, error)
	UpdateACLEntryFn        func(context.Context, *fastly.UpdateACLEntryInput) (*fastly.ACLEntry, error)
	BatchModifyACLEntriesFn func(context.Context, *fastly.BatchModifyACLEntriesInput) error

	CreateNewRelicFn func(context.Context, *fastly.CreateNewRelicInput) (*fastly.NewRelic, error)
	DeleteNewRelicFn func(context.Context, *fastly.DeleteNewRelicInput) error
	GetNewRelicFn    func(context.Context, *fastly.GetNewRelicInput) (*fastly.NewRelic, error)
	ListNewRelicFn   func(context.Context, *fastly.ListNewRelicInput) ([]*fastly.NewRelic, error)
	UpdateNewRelicFn func(context.Context, *fastly.UpdateNewRelicInput) (*fastly.NewRelic, error)

	CreateNewRelicOTLPFn func(context.Context, *fastly.CreateNewRelicOTLPInput) (*fastly.NewRelicOTLP, error)
	DeleteNewRelicOTLPFn func(context.Context, *fastly.DeleteNewRelicOTLPInput) error
	GetNewRelicOTLPFn    func(context.Context, *fastly.GetNewRelicOTLPInput) (*fastly.NewRelicOTLP, error)
	ListNewRelicOTLPFn   func(context.Context, *fastly.ListNewRelicOTLPInput) ([]*fastly.NewRelicOTLP, error)
	UpdateNewRelicOTLPFn func(context.Context, *fastly.UpdateNewRelicOTLPInput) (*fastly.NewRelicOTLP, error)

	CreateUserFn        func(context.Context, *fastly.CreateUserInput) (*fastly.User, error)
	DeleteUserFn        func(context.Context, *fastly.DeleteUserInput) error
	GetCurrentUserFn    func(context.Context) (*fastly.User, error)
	GetUserFn           func(context.Context, *fastly.GetUserInput) (*fastly.User, error)
	ListCustomerUsersFn func(context.Context, *fastly.ListCustomerUsersInput) ([]*fastly.User, error)
	UpdateUserFn        func(context.Context, *fastly.UpdateUserInput) (*fastly.User, error)
	ResetUserPasswordFn func(context.Context, *fastly.ResetUserPasswordInput) error

	BatchDeleteTokensFn  func(context.Context, *fastly.BatchDeleteTokensInput) error
	CreateTokenFn        func(context.Context, *fastly.CreateTokenInput) (*fastly.Token, error)
	DeleteTokenFn        func(context.Context, *fastly.DeleteTokenInput) error
	DeleteTokenSelfFn    func(context.Context) error
	GetTokenSelfFn       func(context.Context) (*fastly.Token, error)
	ListCustomerTokensFn func(context.Context, *fastly.ListCustomerTokensInput) ([]*fastly.Token, error)
	ListTokensFn         func(context.Context, *fastly.ListTokensInput) ([]*fastly.Token, error)

	NewListKVStoreKeysPaginatorFn func(context.Context, *fastly.ListKVStoreKeysInput) fastly.PaginatorKVStoreEntries

	GetCustomTLSConfigurationFn    func(context.Context, *fastly.GetCustomTLSConfigurationInput) (*fastly.CustomTLSConfiguration, error)
	ListCustomTLSConfigurationsFn  func(context.Context, *fastly.ListCustomTLSConfigurationsInput) ([]*fastly.CustomTLSConfiguration, error)
	UpdateCustomTLSConfigurationFn func(context.Context, *fastly.UpdateCustomTLSConfigurationInput) (*fastly.CustomTLSConfiguration, error)
	GetTLSActivationFn             func(context.Context, *fastly.GetTLSActivationInput) (*fastly.TLSActivation, error)
	ListTLSActivationsFn           func(context.Context, *fastly.ListTLSActivationsInput) ([]*fastly.TLSActivation, error)
	UpdateTLSActivationFn          func(context.Context, *fastly.UpdateTLSActivationInput) (*fastly.TLSActivation, error)
	CreateTLSActivationFn          func(context.Context, *fastly.CreateTLSActivationInput) (*fastly.TLSActivation, error)
	DeleteTLSActivationFn          func(context.Context, *fastly.DeleteTLSActivationInput) error

	CreateCustomTLSCertificateFn func(context.Context, *fastly.CreateCustomTLSCertificateInput) (*fastly.CustomTLSCertificate, error)
	DeleteCustomTLSCertificateFn func(context.Context, *fastly.DeleteCustomTLSCertificateInput) error
	GetCustomTLSCertificateFn    func(context.Context, *fastly.GetCustomTLSCertificateInput) (*fastly.CustomTLSCertificate, error)
	ListCustomTLSCertificatesFn  func(context.Context, *fastly.ListCustomTLSCertificatesInput) ([]*fastly.CustomTLSCertificate, error)
	UpdateCustomTLSCertificateFn func(context.Context, *fastly.UpdateCustomTLSCertificateInput) (*fastly.CustomTLSCertificate, error)

	ListTLSDomainsFn func(context.Context, *fastly.ListTLSDomainsInput) ([]*fastly.TLSDomain, error)

	CreatePrivateKeyFn func(context.Context, *fastly.CreatePrivateKeyInput) (*fastly.PrivateKey, error)
	DeletePrivateKeyFn func(context.Context, *fastly.DeletePrivateKeyInput) error
	GetPrivateKeyFn    func(context.Context, *fastly.GetPrivateKeyInput) (*fastly.PrivateKey, error)
	ListPrivateKeysFn  func(context.Context, *fastly.ListPrivateKeysInput) ([]*fastly.PrivateKey, error)

	CreateBulkCertificateFn func(context.Context, *fastly.CreateBulkCertificateInput) (*fastly.BulkCertificate, error)
	DeleteBulkCertificateFn func(context.Context, *fastly.DeleteBulkCertificateInput) error
	GetBulkCertificateFn    func(context.Context, *fastly.GetBulkCertificateInput) (*fastly.BulkCertificate, error)
	ListBulkCertificatesFn  func(context.Context, *fastly.ListBulkCertificatesInput) ([]*fastly.BulkCertificate, error)
	UpdateBulkCertificateFn func(context.Context, *fastly.UpdateBulkCertificateInput) (*fastly.BulkCertificate, error)

	CreateTLSSubscriptionFn func(context.Context, *fastly.CreateTLSSubscriptionInput) (*fastly.TLSSubscription, error)
	DeleteTLSSubscriptionFn func(context.Context, *fastly.DeleteTLSSubscriptionInput) error
	GetTLSSubscriptionFn    func(context.Context, *fastly.GetTLSSubscriptionInput) (*fastly.TLSSubscription, error)
	ListTLSSubscriptionsFn  func(context.Context, *fastly.ListTLSSubscriptionsInput) ([]*fastly.TLSSubscription, error)
	UpdateTLSSubscriptionFn func(context.Context, *fastly.UpdateTLSSubscriptionInput) (*fastly.TLSSubscription, error)

	ListServiceAuthorizationsFn  func(context.Context, *fastly.ListServiceAuthorizationsInput) (*fastly.ServiceAuthorizations, error)
	GetServiceAuthorizationFn    func(context.Context, *fastly.GetServiceAuthorizationInput) (*fastly.ServiceAuthorization, error)
	CreateServiceAuthorizationFn func(context.Context, *fastly.CreateServiceAuthorizationInput) (*fastly.ServiceAuthorization, error)
	UpdateServiceAuthorizationFn func(context.Context, *fastly.UpdateServiceAuthorizationInput) (*fastly.ServiceAuthorization, error)
	DeleteServiceAuthorizationFn func(context.Context, *fastly.DeleteServiceAuthorizationInput) error

	CreateConfigStoreFn       func(context.Context, *fastly.CreateConfigStoreInput) (*fastly.ConfigStore, error)
	DeleteConfigStoreFn       func(context.Context, *fastly.DeleteConfigStoreInput) error
	GetConfigStoreFn          func(context.Context, *fastly.GetConfigStoreInput) (*fastly.ConfigStore, error)
	GetConfigStoreMetadataFn  func(context.Context, *fastly.GetConfigStoreMetadataInput) (*fastly.ConfigStoreMetadata, error)
	ListConfigStoresFn        func(context.Context, *fastly.ListConfigStoresInput) ([]*fastly.ConfigStore, error)
	ListConfigStoreServicesFn func(context.Context, *fastly.ListConfigStoreServicesInput) ([]*fastly.Service, error)
	UpdateConfigStoreFn       func(context.Context, *fastly.UpdateConfigStoreInput) (*fastly.ConfigStore, error)

	CreateConfigStoreItemFn func(context.Context, *fastly.CreateConfigStoreItemInput) (*fastly.ConfigStoreItem, error)
	DeleteConfigStoreItemFn func(context.Context, *fastly.DeleteConfigStoreItemInput) error
	GetConfigStoreItemFn    func(context.Context, *fastly.GetConfigStoreItemInput) (*fastly.ConfigStoreItem, error)
	ListConfigStoreItemsFn  func(context.Context, *fastly.ListConfigStoreItemsInput) ([]*fastly.ConfigStoreItem, error)
	UpdateConfigStoreItemFn func(context.Context, *fastly.UpdateConfigStoreItemInput) (*fastly.ConfigStoreItem, error)

	CreateKVStoreFn         func(context.Context, *fastly.CreateKVStoreInput) (*fastly.KVStore, error)
	GetKVStoreFn            func(context.Context, *fastly.GetKVStoreInput) (*fastly.KVStore, error)
	ListKVStoresFn          func(context.Context, *fastly.ListKVStoresInput) (*fastly.ListKVStoresResponse, error)
	DeleteKVStoreFn         func(context.Context, *fastly.DeleteKVStoreInput) error
	ListKVStoreKeysFn       func(context.Context, *fastly.ListKVStoreKeysInput) (*fastly.ListKVStoreKeysResponse, error)
	GetKVStoreKeyFn         func(context.Context, *fastly.GetKVStoreKeyInput) (string, error)
	GetKVStoreItemFn        func(context.Context, *fastly.GetKVStoreItemInput) (fastly.GetKVStoreItemOutput, error)
	InsertKVStoreKeyFn      func(context.Context, *fastly.InsertKVStoreKeyInput) error
	DeleteKVStoreKeyFn      func(context.Context, *fastly.DeleteKVStoreKeyInput) error
	BatchModifyKVStoreKeyFn func(context.Context, *fastly.BatchModifyKVStoreKeyInput) error

	CreateSecretStoreFn func(context.Context, *fastly.CreateSecretStoreInput) (*fastly.SecretStore, error)
	GetSecretStoreFn    func(context.Context, *fastly.GetSecretStoreInput) (*fastly.SecretStore, error)
	DeleteSecretStoreFn func(context.Context, *fastly.DeleteSecretStoreInput) error
	ListSecretStoresFn  func(context.Context, *fastly.ListSecretStoresInput) (*fastly.SecretStores, error)
	CreateSecretFn      func(context.Context, *fastly.CreateSecretInput) (*fastly.Secret, error)
	GetSecretFn         func(context.Context, *fastly.GetSecretInput) (*fastly.Secret, error)
	DeleteSecretFn      func(context.Context, *fastly.DeleteSecretInput) error
	ListSecretsFn       func(context.Context, *fastly.ListSecretsInput) (*fastly.Secrets, error)
	CreateClientKeyFn   func(context.Context) (*fastly.ClientKey, error)
	GetSigningKeyFn     func(context.Context) (ed25519.PublicKey, error)

	CreateResourceFn func(context.Context, *fastly.CreateResourceInput) (*fastly.Resource, error)
	DeleteResourceFn func(context.Context, *fastly.DeleteResourceInput) error
	GetResourceFn    func(context.Context, *fastly.GetResourceInput) (*fastly.Resource, error)
	ListResourcesFn  func(context.Context, *fastly.ListResourcesInput) ([]*fastly.Resource, error)
	UpdateResourceFn func(context.Context, *fastly.UpdateResourceInput) (*fastly.Resource, error)

	CreateERLFn func(context.Context, *fastly.CreateERLInput) (*fastly.ERL, error)
	DeleteERLFn func(context.Context, *fastly.DeleteERLInput) error
	GetERLFn    func(context.Context, *fastly.GetERLInput) (*fastly.ERL, error)
	ListERLsFn  func(context.Context, *fastly.ListERLsInput) ([]*fastly.ERL, error)
	UpdateERLFn func(context.Context, *fastly.UpdateERLInput) (*fastly.ERL, error)

	CreateConditionFn func(context.Context, *fastly.CreateConditionInput) (*fastly.Condition, error)
	DeleteConditionFn func(context.Context, *fastly.DeleteConditionInput) error
	GetConditionFn    func(context.Context, *fastly.GetConditionInput) (*fastly.Condition, error)
	ListConditionsFn  func(context.Context, *fastly.ListConditionsInput) ([]*fastly.Condition, error)
	UpdateConditionFn func(context.Context, *fastly.UpdateConditionInput) (*fastly.Condition, error)

	ListAlertDefinitionsFn  func(context.Context, *fastly.ListAlertDefinitionsInput) (*fastly.AlertDefinitionsResponse, error)
	CreateAlertDefinitionFn func(context.Context, *fastly.CreateAlertDefinitionInput) (*fastly.AlertDefinition, error)
	GetAlertDefinitionFn    func(context.Context, *fastly.GetAlertDefinitionInput) (*fastly.AlertDefinition, error)
	UpdateAlertDefinitionFn func(context.Context, *fastly.UpdateAlertDefinitionInput) (*fastly.AlertDefinition, error)
	DeleteAlertDefinitionFn func(context.Context, *fastly.DeleteAlertDefinitionInput) error
	TestAlertDefinitionFn   func(context.Context, *fastly.TestAlertDefinitionInput) error
	ListAlertHistoryFn      func(context.Context, *fastly.ListAlertHistoryInput) (*fastly.AlertHistoryResponse, error)

	ListObservabilityCustomDashboardsFn  func(context.Context, *fastly.ListObservabilityCustomDashboardsInput) (*fastly.ListDashboardsResponse, error)
	CreateObservabilityCustomDashboardFn func(context.Context, *fastly.CreateObservabilityCustomDashboardInput) (*fastly.ObservabilityCustomDashboard, error)
	GetObservabilityCustomDashboardFn    func(context.Context, *fastly.GetObservabilityCustomDashboardInput) (*fastly.ObservabilityCustomDashboard, error)
	UpdateObservabilityCustomDashboardFn func(context.Context, *fastly.UpdateObservabilityCustomDashboardInput) (*fastly.ObservabilityCustomDashboard, error)
	DeleteObservabilityCustomDashboardFn func(context.Context, *fastly.DeleteObservabilityCustomDashboardInput) error

	GetImageOptimizerDefaultSettingsFn    func(context.Context, *fastly.GetImageOptimizerDefaultSettingsInput) (*fastly.ImageOptimizerDefaultSettings, error)
	UpdateImageOptimizerDefaultSettingsFn func(context.Context, *fastly.UpdateImageOptimizerDefaultSettingsInput) (*fastly.ImageOptimizerDefaultSettings, error)
}

// AllDatacenters implements Interface.
func (m API) AllDatacenters(ctx context.Context) ([]fastly.Datacenter, error) {
	return m.AllDatacentersFn(ctx)
}

// AllIPs implements Interface.
func (m API) AllIPs(ctx context.Context) (fastly.IPAddrs, fastly.IPAddrs, error) {
	return m.AllIPsFn(ctx)
}

// CreateService implements Interface.
func (m API) CreateService(ctx context.Context, i *fastly.CreateServiceInput) (*fastly.Service, error) {
	return m.CreateServiceFn(ctx, i)
}

// GetServices implements Interface.
func (m API) GetServices(ctx context.Context, i *fastly.GetServicesInput) *fastly.ListPaginator[fastly.Service] {
	return m.GetServicesFn(ctx, i)
}

// ListServices implements Interface.
func (m API) ListServices(ctx context.Context, i *fastly.ListServicesInput) ([]*fastly.Service, error) {
	return m.ListServicesFn(ctx, i)
}

// GetService implements Interface.
func (m API) GetService(ctx context.Context, i *fastly.GetServiceInput) (*fastly.Service, error) {
	return m.GetServiceFn(ctx, i)
}

// GetServiceDetails implements Interface.
func (m API) GetServiceDetails(ctx context.Context, i *fastly.GetServiceInput) (*fastly.ServiceDetail, error) {
	return m.GetServiceDetailsFn(ctx, i)
}

// SearchService implements Interface.
func (m API) SearchService(ctx context.Context, i *fastly.SearchServiceInput) (*fastly.Service, error) {
	return m.SearchServiceFn(ctx, i)
}

// UpdateService implements Interface.
func (m API) UpdateService(ctx context.Context, i *fastly.UpdateServiceInput) (*fastly.Service, error) {
	return m.UpdateServiceFn(ctx, i)
}

// DeleteService implements Interface.
func (m API) DeleteService(ctx context.Context, i *fastly.DeleteServiceInput) error {
	return m.DeleteServiceFn(ctx, i)
}

// CloneVersion implements Interface.
func (m API) CloneVersion(ctx context.Context, i *fastly.CloneVersionInput) (*fastly.Version, error) {
	return m.CloneVersionFn(ctx, i)
}

// ListVersions implements Interface.
func (m API) ListVersions(ctx context.Context, i *fastly.ListVersionsInput) ([]*fastly.Version, error) {
	return m.ListVersionsFn(ctx, i)
}

// GetVersion implements Interface.
func (m API) GetVersion(ctx context.Context, i *fastly.GetVersionInput) (*fastly.Version, error) {
	return m.GetVersionFn(ctx, i)
}

// UpdateVersion implements Interface.
func (m API) UpdateVersion(ctx context.Context, i *fastly.UpdateVersionInput) (*fastly.Version, error) {
	return m.UpdateVersionFn(ctx, i)
}

// ActivateVersion implements Interface.
func (m API) ActivateVersion(ctx context.Context, i *fastly.ActivateVersionInput) (*fastly.Version, error) {
	return m.ActivateVersionFn(ctx, i)
}

// DeactivateVersion implements Interface.
func (m API) DeactivateVersion(ctx context.Context, i *fastly.DeactivateVersionInput) (*fastly.Version, error) {
	return m.DeactivateVersionFn(ctx, i)
}

// LockVersion implements Interface.
func (m API) LockVersion(ctx context.Context, i *fastly.LockVersionInput) (*fastly.Version, error) {
	return m.LockVersionFn(ctx, i)
}

// LatestVersion implements Interface.
func (m API) LatestVersion(ctx context.Context, i *fastly.LatestVersionInput) (*fastly.Version, error) {
	return m.LatestVersionFn(ctx, i)
}

// CreateDomain implements Interface.
func (m API) CreateDomain(ctx context.Context, i *fastly.CreateDomainInput) (*fastly.Domain, error) {
	return m.CreateDomainFn(ctx, i)
}

// ListDomains implements Interface.
func (m API) ListDomains(ctx context.Context, i *fastly.ListDomainsInput) ([]*fastly.Domain, error) {
	return m.ListDomainsFn(ctx, i)
}

// GetDomain implements Interface.
func (m API) GetDomain(ctx context.Context, i *fastly.GetDomainInput) (*fastly.Domain, error) {
	return m.GetDomainFn(ctx, i)
}

// UpdateDomain implements Interface.
func (m API) UpdateDomain(ctx context.Context, i *fastly.UpdateDomainInput) (*fastly.Domain, error) {
	return m.UpdateDomainFn(ctx, i)
}

// DeleteDomain implements Interface.
func (m API) DeleteDomain(ctx context.Context, i *fastly.DeleteDomainInput) error {
	return m.DeleteDomainFn(ctx, i)
}

// ValidateDomain implements Interface.
func (m API) ValidateDomain(ctx context.Context, i *fastly.ValidateDomainInput) (*fastly.DomainValidationResult, error) {
	return m.ValidateDomainFn(ctx, i)
}

// ValidateAllDomains implements Interface.
func (m API) ValidateAllDomains(ctx context.Context, i *fastly.ValidateAllDomainsInput) (results []*fastly.DomainValidationResult, err error) {
	return m.ValidateAllDomainsFn(ctx, i)
}

// CreateBackend implements Interface.
func (m API) CreateBackend(ctx context.Context, i *fastly.CreateBackendInput) (*fastly.Backend, error) {
	return m.CreateBackendFn(ctx, i)
}

// ListBackends implements Interface.
func (m API) ListBackends(ctx context.Context, i *fastly.ListBackendsInput) ([]*fastly.Backend, error) {
	return m.ListBackendsFn(ctx, i)
}

// GetBackend implements Interface.
func (m API) GetBackend(ctx context.Context, i *fastly.GetBackendInput) (*fastly.Backend, error) {
	return m.GetBackendFn(ctx, i)
}

// UpdateBackend implements Interface.
func (m API) UpdateBackend(ctx context.Context, i *fastly.UpdateBackendInput) (*fastly.Backend, error) {
	return m.UpdateBackendFn(ctx, i)
}

// DeleteBackend implements Interface.
func (m API) DeleteBackend(ctx context.Context, i *fastly.DeleteBackendInput) error {
	return m.DeleteBackendFn(ctx, i)
}

// CreateHealthCheck implements Interface.
func (m API) CreateHealthCheck(ctx context.Context, i *fastly.CreateHealthCheckInput) (*fastly.HealthCheck, error) {
	return m.CreateHealthCheckFn(ctx, i)
}

// ListHealthChecks implements Interface.
func (m API) ListHealthChecks(ctx context.Context, i *fastly.ListHealthChecksInput) ([]*fastly.HealthCheck, error) {
	return m.ListHealthChecksFn(ctx, i)
}

// GetHealthCheck implements Interface.
func (m API) GetHealthCheck(ctx context.Context, i *fastly.GetHealthCheckInput) (*fastly.HealthCheck, error) {
	return m.GetHealthCheckFn(ctx, i)
}

// UpdateHealthCheck implements Interface.
func (m API) UpdateHealthCheck(ctx context.Context, i *fastly.UpdateHealthCheckInput) (*fastly.HealthCheck, error) {
	return m.UpdateHealthCheckFn(ctx, i)
}

// DeleteHealthCheck implements Interface.
func (m API) DeleteHealthCheck(ctx context.Context, i *fastly.DeleteHealthCheckInput) error {
	return m.DeleteHealthCheckFn(ctx, i)
}

// GetPackage implements Interface.
func (m API) GetPackage(ctx context.Context, i *fastly.GetPackageInput) (*fastly.Package, error) {
	return m.GetPackageFn(ctx, i)
}

// UpdatePackage implements Interface.
func (m API) UpdatePackage(ctx context.Context, i *fastly.UpdatePackageInput) (*fastly.Package, error) {
	return m.UpdatePackageFn(ctx, i)
}

// CreateDictionary implements Interface.
func (m API) CreateDictionary(ctx context.Context, i *fastly.CreateDictionaryInput) (*fastly.Dictionary, error) {
	return m.CreateDictionaryFn(ctx, i)
}

// GetDictionary implements Interface.
func (m API) GetDictionary(ctx context.Context, i *fastly.GetDictionaryInput) (*fastly.Dictionary, error) {
	return m.GetDictionaryFn(ctx, i)
}

// DeleteDictionary implements Interface.
func (m API) DeleteDictionary(ctx context.Context, i *fastly.DeleteDictionaryInput) error {
	return m.DeleteDictionaryFn(ctx, i)
}

// ListDictionaries implements Interface.
func (m API) ListDictionaries(ctx context.Context, i *fastly.ListDictionariesInput) ([]*fastly.Dictionary, error) {
	return m.ListDictionariesFn(ctx, i)
}

// UpdateDictionary implements Interface.
func (m API) UpdateDictionary(ctx context.Context, i *fastly.UpdateDictionaryInput) (*fastly.Dictionary, error) {
	return m.UpdateDictionaryFn(ctx, i)
}

// GetDictionaryItems implements Interface.
func (m API) GetDictionaryItems(ctx context.Context, i *fastly.GetDictionaryItemsInput) *fastly.ListPaginator[fastly.DictionaryItem] {
	return m.GetDictionaryItemsFn(ctx, i)
}

// ListDictionaryItems implements Interface.
func (m API) ListDictionaryItems(ctx context.Context, i *fastly.ListDictionaryItemsInput) ([]*fastly.DictionaryItem, error) {
	return m.ListDictionaryItemsFn(ctx, i)
}

// GetDictionaryItem implements Interface.
func (m API) GetDictionaryItem(ctx context.Context, i *fastly.GetDictionaryItemInput) (*fastly.DictionaryItem, error) {
	return m.GetDictionaryItemFn(ctx, i)
}

// CreateDictionaryItem implements Interface.
func (m API) CreateDictionaryItem(ctx context.Context, i *fastly.CreateDictionaryItemInput) (*fastly.DictionaryItem, error) {
	return m.CreateDictionaryItemFn(ctx, i)
}

// UpdateDictionaryItem implements Interface.
func (m API) UpdateDictionaryItem(ctx context.Context, i *fastly.UpdateDictionaryItemInput) (*fastly.DictionaryItem, error) {
	return m.UpdateDictionaryItemFn(ctx, i)
}

// DeleteDictionaryItem implements Interface.
func (m API) DeleteDictionaryItem(ctx context.Context, i *fastly.DeleteDictionaryItemInput) error {
	return m.DeleteDictionaryItemFn(ctx, i)
}

// BatchModifyDictionaryItems implements Interface.
func (m API) BatchModifyDictionaryItems(ctx context.Context, i *fastly.BatchModifyDictionaryItemsInput) error {
	return m.BatchModifyDictionaryItemsFn(ctx, i)
}

// GetDictionaryInfo implements Interface.
func (m API) GetDictionaryInfo(ctx context.Context, i *fastly.GetDictionaryInfoInput) (*fastly.DictionaryInfo, error) {
	return m.GetDictionaryInfoFn(ctx, i)
}

// CreateBigQuery implements Interface.
func (m API) CreateBigQuery(ctx context.Context, i *fastly.CreateBigQueryInput) (*fastly.BigQuery, error) {
	return m.CreateBigQueryFn(ctx, i)
}

// ListBigQueries implements Interface.
func (m API) ListBigQueries(ctx context.Context, i *fastly.ListBigQueriesInput) ([]*fastly.BigQuery, error) {
	return m.ListBigQueriesFn(ctx, i)
}

// GetBigQuery implements Interface.
func (m API) GetBigQuery(ctx context.Context, i *fastly.GetBigQueryInput) (*fastly.BigQuery, error) {
	return m.GetBigQueryFn(ctx, i)
}

// UpdateBigQuery implements Interface.
func (m API) UpdateBigQuery(ctx context.Context, i *fastly.UpdateBigQueryInput) (*fastly.BigQuery, error) {
	return m.UpdateBigQueryFn(ctx, i)
}

// DeleteBigQuery implements Interface.
func (m API) DeleteBigQuery(ctx context.Context, i *fastly.DeleteBigQueryInput) error {
	return m.DeleteBigQueryFn(ctx, i)
}

// CreateS3 implements Interface.
func (m API) CreateS3(ctx context.Context, i *fastly.CreateS3Input) (*fastly.S3, error) {
	return m.CreateS3Fn(ctx, i)
}

// ListS3s implements Interface.
func (m API) ListS3s(ctx context.Context, i *fastly.ListS3sInput) ([]*fastly.S3, error) {
	return m.ListS3sFn(ctx, i)
}

// GetS3 implements Interface.
func (m API) GetS3(ctx context.Context, i *fastly.GetS3Input) (*fastly.S3, error) {
	return m.GetS3Fn(ctx, i)
}

// UpdateS3 implements Interface.
func (m API) UpdateS3(ctx context.Context, i *fastly.UpdateS3Input) (*fastly.S3, error) {
	return m.UpdateS3Fn(ctx, i)
}

// DeleteS3 implements Interface.
func (m API) DeleteS3(ctx context.Context, i *fastly.DeleteS3Input) error {
	return m.DeleteS3Fn(ctx, i)
}

// CreateKinesis implements Interface.
func (m API) CreateKinesis(ctx context.Context, i *fastly.CreateKinesisInput) (*fastly.Kinesis, error) {
	return m.CreateKinesisFn(ctx, i)
}

// ListKinesis implements Interface.
func (m API) ListKinesis(ctx context.Context, i *fastly.ListKinesisInput) ([]*fastly.Kinesis, error) {
	return m.ListKinesisFn(ctx, i)
}

// GetKinesis implements Interface.
func (m API) GetKinesis(ctx context.Context, i *fastly.GetKinesisInput) (*fastly.Kinesis, error) {
	return m.GetKinesisFn(ctx, i)
}

// UpdateKinesis implements Interface.
func (m API) UpdateKinesis(ctx context.Context, i *fastly.UpdateKinesisInput) (*fastly.Kinesis, error) {
	return m.UpdateKinesisFn(ctx, i)
}

// DeleteKinesis implements Interface.
func (m API) DeleteKinesis(ctx context.Context, i *fastly.DeleteKinesisInput) error {
	return m.DeleteKinesisFn(ctx, i)
}

// CreateSyslog implements Interface.
func (m API) CreateSyslog(ctx context.Context, i *fastly.CreateSyslogInput) (*fastly.Syslog, error) {
	return m.CreateSyslogFn(ctx, i)
}

// ListSyslogs implements Interface.
func (m API) ListSyslogs(ctx context.Context, i *fastly.ListSyslogsInput) ([]*fastly.Syslog, error) {
	return m.ListSyslogsFn(ctx, i)
}

// GetSyslog implements Interface.
func (m API) GetSyslog(ctx context.Context, i *fastly.GetSyslogInput) (*fastly.Syslog, error) {
	return m.GetSyslogFn(ctx, i)
}

// UpdateSyslog implements Interface.
func (m API) UpdateSyslog(ctx context.Context, i *fastly.UpdateSyslogInput) (*fastly.Syslog, error) {
	return m.UpdateSyslogFn(ctx, i)
}

// DeleteSyslog implements Interface.
func (m API) DeleteSyslog(ctx context.Context, i *fastly.DeleteSyslogInput) error {
	return m.DeleteSyslogFn(ctx, i)
}

// CreateLogentries implements Interface.
func (m API) CreateLogentries(ctx context.Context, i *fastly.CreateLogentriesInput) (*fastly.Logentries, error) {
	return m.CreateLogentriesFn(ctx, i)
}

// ListLogentries implements Interface.
func (m API) ListLogentries(ctx context.Context, i *fastly.ListLogentriesInput) ([]*fastly.Logentries, error) {
	return m.ListLogentriesFn(ctx, i)
}

// GetLogentries implements Interface.
func (m API) GetLogentries(ctx context.Context, i *fastly.GetLogentriesInput) (*fastly.Logentries, error) {
	return m.GetLogentriesFn(ctx, i)
}

// UpdateLogentries implements Interface.
func (m API) UpdateLogentries(ctx context.Context, i *fastly.UpdateLogentriesInput) (*fastly.Logentries, error) {
	return m.UpdateLogentriesFn(ctx, i)
}

// DeleteLogentries implements Interface.
func (m API) DeleteLogentries(ctx context.Context, i *fastly.DeleteLogentriesInput) error {
	return m.DeleteLogentriesFn(ctx, i)
}

// CreatePapertrail implements Interface.
func (m API) CreatePapertrail(ctx context.Context, i *fastly.CreatePapertrailInput) (*fastly.Papertrail, error) {
	return m.CreatePapertrailFn(ctx, i)
}

// ListPapertrails implements Interface.
func (m API) ListPapertrails(ctx context.Context, i *fastly.ListPapertrailsInput) ([]*fastly.Papertrail, error) {
	return m.ListPapertrailsFn(ctx, i)
}

// GetPapertrail implements Interface.
func (m API) GetPapertrail(ctx context.Context, i *fastly.GetPapertrailInput) (*fastly.Papertrail, error) {
	return m.GetPapertrailFn(ctx, i)
}

// UpdatePapertrail implements Interface.
func (m API) UpdatePapertrail(ctx context.Context, i *fastly.UpdatePapertrailInput) (*fastly.Papertrail, error) {
	return m.UpdatePapertrailFn(ctx, i)
}

// DeletePapertrail implements Interface.
func (m API) DeletePapertrail(ctx context.Context, i *fastly.DeletePapertrailInput) error {
	return m.DeletePapertrailFn(ctx, i)
}

// CreateSumologic implements Interface.
func (m API) CreateSumologic(ctx context.Context, i *fastly.CreateSumologicInput) (*fastly.Sumologic, error) {
	return m.CreateSumologicFn(ctx, i)
}

// ListSumologics implements Interface.
func (m API) ListSumologics(ctx context.Context, i *fastly.ListSumologicsInput) ([]*fastly.Sumologic, error) {
	return m.ListSumologicsFn(ctx, i)
}

// GetSumologic implements Interface.
func (m API) GetSumologic(ctx context.Context, i *fastly.GetSumologicInput) (*fastly.Sumologic, error) {
	return m.GetSumologicFn(ctx, i)
}

// UpdateSumologic implements Interface.
func (m API) UpdateSumologic(ctx context.Context, i *fastly.UpdateSumologicInput) (*fastly.Sumologic, error) {
	return m.UpdateSumologicFn(ctx, i)
}

// DeleteSumologic implements Interface.
func (m API) DeleteSumologic(ctx context.Context, i *fastly.DeleteSumologicInput) error {
	return m.DeleteSumologicFn(ctx, i)
}

// CreateGCS implements Interface.
func (m API) CreateGCS(ctx context.Context, i *fastly.CreateGCSInput) (*fastly.GCS, error) {
	return m.CreateGCSFn(ctx, i)
}

// ListGCSs implements Interface.
func (m API) ListGCSs(ctx context.Context, i *fastly.ListGCSsInput) ([]*fastly.GCS, error) {
	return m.ListGCSsFn(ctx, i)
}

// GetGCS implements Interface.
func (m API) GetGCS(ctx context.Context, i *fastly.GetGCSInput) (*fastly.GCS, error) {
	return m.GetGCSFn(ctx, i)
}

// UpdateGCS implements Interface.
func (m API) UpdateGCS(ctx context.Context, i *fastly.UpdateGCSInput) (*fastly.GCS, error) {
	return m.UpdateGCSFn(ctx, i)
}

// DeleteGCS implements Interface.
func (m API) DeleteGCS(ctx context.Context, i *fastly.DeleteGCSInput) error {
	return m.DeleteGCSFn(ctx, i)
}

// CreateGrafanaCloudLogs implements Interface.
func (m API) CreateGrafanaCloudLogs(ctx context.Context, i *fastly.CreateGrafanaCloudLogsInput) (*fastly.GrafanaCloudLogs, error) {
	return m.CreateGrafanaCloudLogsFn(ctx, i)
}

// ListGrafanaCloudLogs implements Interface.
func (m API) ListGrafanaCloudLogs(ctx context.Context, i *fastly.ListGrafanaCloudLogsInput) ([]*fastly.GrafanaCloudLogs, error) {
	return m.ListGrafanaCloudLogsFn(ctx, i)
}

// GetGrafanaCloudLogs implements Interface.
func (m API) GetGrafanaCloudLogs(ctx context.Context, i *fastly.GetGrafanaCloudLogsInput) (*fastly.GrafanaCloudLogs, error) {
	return m.GetGrafanaCloudLogsFn(ctx, i)
}

// UpdateGrafanaCloudLogs implements Interface.
func (m API) UpdateGrafanaCloudLogs(ctx context.Context, i *fastly.UpdateGrafanaCloudLogsInput) (*fastly.GrafanaCloudLogs, error) {
	return m.UpdateGrafanaCloudLogsFn(ctx, i)
}

// DeleteGrafanaCloudLogs implements Interface.
func (m API) DeleteGrafanaCloudLogs(ctx context.Context, i *fastly.DeleteGrafanaCloudLogsInput) error {
	return m.DeleteGrafanaCloudLogsFn(ctx, i)
}

// CreateFTP implements Interface.
func (m API) CreateFTP(ctx context.Context, i *fastly.CreateFTPInput) (*fastly.FTP, error) {
	return m.CreateFTPFn(ctx, i)
}

// ListFTPs implements Interface.
func (m API) ListFTPs(ctx context.Context, i *fastly.ListFTPsInput) ([]*fastly.FTP, error) {
	return m.ListFTPsFn(ctx, i)
}

// GetFTP implements Interface.
func (m API) GetFTP(ctx context.Context, i *fastly.GetFTPInput) (*fastly.FTP, error) {
	return m.GetFTPFn(ctx, i)
}

// UpdateFTP implements Interface.
func (m API) UpdateFTP(ctx context.Context, i *fastly.UpdateFTPInput) (*fastly.FTP, error) {
	return m.UpdateFTPFn(ctx, i)
}

// DeleteFTP implements Interface.
func (m API) DeleteFTP(ctx context.Context, i *fastly.DeleteFTPInput) error {
	return m.DeleteFTPFn(ctx, i)
}

// CreateSplunk implements Interface.
func (m API) CreateSplunk(ctx context.Context, i *fastly.CreateSplunkInput) (*fastly.Splunk, error) {
	return m.CreateSplunkFn(ctx, i)
}

// ListSplunks implements Interface.
func (m API) ListSplunks(ctx context.Context, i *fastly.ListSplunksInput) ([]*fastly.Splunk, error) {
	return m.ListSplunksFn(ctx, i)
}

// GetSplunk implements Interface.
func (m API) GetSplunk(ctx context.Context, i *fastly.GetSplunkInput) (*fastly.Splunk, error) {
	return m.GetSplunkFn(ctx, i)
}

// UpdateSplunk implements Interface.
func (m API) UpdateSplunk(ctx context.Context, i *fastly.UpdateSplunkInput) (*fastly.Splunk, error) {
	return m.UpdateSplunkFn(ctx, i)
}

// DeleteSplunk implements Interface.
func (m API) DeleteSplunk(ctx context.Context, i *fastly.DeleteSplunkInput) error {
	return m.DeleteSplunkFn(ctx, i)
}

// CreateScalyr implements Interface.
func (m API) CreateScalyr(ctx context.Context, i *fastly.CreateScalyrInput) (*fastly.Scalyr, error) {
	return m.CreateScalyrFn(ctx, i)
}

// ListScalyrs implements Interface.
func (m API) ListScalyrs(ctx context.Context, i *fastly.ListScalyrsInput) ([]*fastly.Scalyr, error) {
	return m.ListScalyrsFn(ctx, i)
}

// GetScalyr implements Interface.
func (m API) GetScalyr(ctx context.Context, i *fastly.GetScalyrInput) (*fastly.Scalyr, error) {
	return m.GetScalyrFn(ctx, i)
}

// UpdateScalyr implements Interface.
func (m API) UpdateScalyr(ctx context.Context, i *fastly.UpdateScalyrInput) (*fastly.Scalyr, error) {
	return m.UpdateScalyrFn(ctx, i)
}

// DeleteScalyr implements Interface.
func (m API) DeleteScalyr(ctx context.Context, i *fastly.DeleteScalyrInput) error {
	return m.DeleteScalyrFn(ctx, i)
}

// CreateLoggly implements Interface.
func (m API) CreateLoggly(ctx context.Context, i *fastly.CreateLogglyInput) (*fastly.Loggly, error) {
	return m.CreateLogglyFn(ctx, i)
}

// ListLoggly implements Interface.
func (m API) ListLoggly(ctx context.Context, i *fastly.ListLogglyInput) ([]*fastly.Loggly, error) {
	return m.ListLogglyFn(ctx, i)
}

// GetLoggly implements Interface.
func (m API) GetLoggly(ctx context.Context, i *fastly.GetLogglyInput) (*fastly.Loggly, error) {
	return m.GetLogglyFn(ctx, i)
}

// UpdateLoggly implements Interface.
func (m API) UpdateLoggly(ctx context.Context, i *fastly.UpdateLogglyInput) (*fastly.Loggly, error) {
	return m.UpdateLogglyFn(ctx, i)
}

// DeleteLoggly implements Interface.
func (m API) DeleteLoggly(ctx context.Context, i *fastly.DeleteLogglyInput) error {
	return m.DeleteLogglyFn(ctx, i)
}

// CreateHoneycomb implements Interface.
func (m API) CreateHoneycomb(ctx context.Context, i *fastly.CreateHoneycombInput) (*fastly.Honeycomb, error) {
	return m.CreateHoneycombFn(ctx, i)
}

// ListHoneycombs implements Interface.
func (m API) ListHoneycombs(ctx context.Context, i *fastly.ListHoneycombsInput) ([]*fastly.Honeycomb, error) {
	return m.ListHoneycombsFn(ctx, i)
}

// GetHoneycomb implements Interface.
func (m API) GetHoneycomb(ctx context.Context, i *fastly.GetHoneycombInput) (*fastly.Honeycomb, error) {
	return m.GetHoneycombFn(ctx, i)
}

// UpdateHoneycomb implements Interface.
func (m API) UpdateHoneycomb(ctx context.Context, i *fastly.UpdateHoneycombInput) (*fastly.Honeycomb, error) {
	return m.UpdateHoneycombFn(ctx, i)
}

// DeleteHoneycomb implements Interface.
func (m API) DeleteHoneycomb(ctx context.Context, i *fastly.DeleteHoneycombInput) error {
	return m.DeleteHoneycombFn(ctx, i)
}

// CreateHeroku implements Interface.
func (m API) CreateHeroku(ctx context.Context, i *fastly.CreateHerokuInput) (*fastly.Heroku, error) {
	return m.CreateHerokuFn(ctx, i)
}

// ListHerokus implements Interface.
func (m API) ListHerokus(ctx context.Context, i *fastly.ListHerokusInput) ([]*fastly.Heroku, error) {
	return m.ListHerokusFn(ctx, i)
}

// GetHeroku implements Interface.
func (m API) GetHeroku(ctx context.Context, i *fastly.GetHerokuInput) (*fastly.Heroku, error) {
	return m.GetHerokuFn(ctx, i)
}

// UpdateHeroku implements Interface.
func (m API) UpdateHeroku(ctx context.Context, i *fastly.UpdateHerokuInput) (*fastly.Heroku, error) {
	return m.UpdateHerokuFn(ctx, i)
}

// DeleteHeroku implements Interface.
func (m API) DeleteHeroku(ctx context.Context, i *fastly.DeleteHerokuInput) error {
	return m.DeleteHerokuFn(ctx, i)
}

// CreateSFTP implements Interface.
func (m API) CreateSFTP(ctx context.Context, i *fastly.CreateSFTPInput) (*fastly.SFTP, error) {
	return m.CreateSFTPFn(ctx, i)
}

// ListSFTPs implements Interface.
func (m API) ListSFTPs(ctx context.Context, i *fastly.ListSFTPsInput) ([]*fastly.SFTP, error) {
	return m.ListSFTPsFn(ctx, i)
}

// GetSFTP implements Interface.
func (m API) GetSFTP(ctx context.Context, i *fastly.GetSFTPInput) (*fastly.SFTP, error) {
	return m.GetSFTPFn(ctx, i)
}

// UpdateSFTP implements Interface.
func (m API) UpdateSFTP(ctx context.Context, i *fastly.UpdateSFTPInput) (*fastly.SFTP, error) {
	return m.UpdateSFTPFn(ctx, i)
}

// DeleteSFTP implements Interface.
func (m API) DeleteSFTP(ctx context.Context, i *fastly.DeleteSFTPInput) error {
	return m.DeleteSFTPFn(ctx, i)
}

// CreateLogshuttle implements Interface.
func (m API) CreateLogshuttle(ctx context.Context, i *fastly.CreateLogshuttleInput) (*fastly.Logshuttle, error) {
	return m.CreateLogshuttleFn(ctx, i)
}

// ListLogshuttles implements Interface.
func (m API) ListLogshuttles(ctx context.Context, i *fastly.ListLogshuttlesInput) ([]*fastly.Logshuttle, error) {
	return m.ListLogshuttlesFn(ctx, i)
}

// GetLogshuttle implements Interface.
func (m API) GetLogshuttle(ctx context.Context, i *fastly.GetLogshuttleInput) (*fastly.Logshuttle, error) {
	return m.GetLogshuttleFn(ctx, i)
}

// UpdateLogshuttle implements Interface.
func (m API) UpdateLogshuttle(ctx context.Context, i *fastly.UpdateLogshuttleInput) (*fastly.Logshuttle, error) {
	return m.UpdateLogshuttleFn(ctx, i)
}

// DeleteLogshuttle implements Interface.
func (m API) DeleteLogshuttle(ctx context.Context, i *fastly.DeleteLogshuttleInput) error {
	return m.DeleteLogshuttleFn(ctx, i)
}

// CreateCloudfiles implements Interface.
func (m API) CreateCloudfiles(ctx context.Context, i *fastly.CreateCloudfilesInput) (*fastly.Cloudfiles, error) {
	return m.CreateCloudfilesFn(ctx, i)
}

// ListCloudfiles implements Interface.
func (m API) ListCloudfiles(ctx context.Context, i *fastly.ListCloudfilesInput) ([]*fastly.Cloudfiles, error) {
	return m.ListCloudfilesFn(ctx, i)
}

// GetCloudfiles implements Interface.
func (m API) GetCloudfiles(ctx context.Context, i *fastly.GetCloudfilesInput) (*fastly.Cloudfiles, error) {
	return m.GetCloudfilesFn(ctx, i)
}

// UpdateCloudfiles implements Interface.
func (m API) UpdateCloudfiles(ctx context.Context, i *fastly.UpdateCloudfilesInput) (*fastly.Cloudfiles, error) {
	return m.UpdateCloudfilesFn(ctx, i)
}

// DeleteCloudfiles implements Interface.
func (m API) DeleteCloudfiles(ctx context.Context, i *fastly.DeleteCloudfilesInput) error {
	return m.DeleteCloudfilesFn(ctx, i)
}

// CreateDigitalOcean implements Interface.
func (m API) CreateDigitalOcean(ctx context.Context, i *fastly.CreateDigitalOceanInput) (*fastly.DigitalOcean, error) {
	return m.CreateDigitalOceanFn(ctx, i)
}

// ListDigitalOceans implements Interface.
func (m API) ListDigitalOceans(ctx context.Context, i *fastly.ListDigitalOceansInput) ([]*fastly.DigitalOcean, error) {
	return m.ListDigitalOceansFn(ctx, i)
}

// GetDigitalOcean implements Interface.
func (m API) GetDigitalOcean(ctx context.Context, i *fastly.GetDigitalOceanInput) (*fastly.DigitalOcean, error) {
	return m.GetDigitalOceanFn(ctx, i)
}

// UpdateDigitalOcean implements Interface.
func (m API) UpdateDigitalOcean(ctx context.Context, i *fastly.UpdateDigitalOceanInput) (*fastly.DigitalOcean, error) {
	return m.UpdateDigitalOceanFn(ctx, i)
}

// DeleteDigitalOcean implements Interface.
func (m API) DeleteDigitalOcean(ctx context.Context, i *fastly.DeleteDigitalOceanInput) error {
	return m.DeleteDigitalOceanFn(ctx, i)
}

// CreateElasticsearch implements Interface.
func (m API) CreateElasticsearch(ctx context.Context, i *fastly.CreateElasticsearchInput) (*fastly.Elasticsearch, error) {
	return m.CreateElasticsearchFn(ctx, i)
}

// ListElasticsearch implements Interface.
func (m API) ListElasticsearch(ctx context.Context, i *fastly.ListElasticsearchInput) ([]*fastly.Elasticsearch, error) {
	return m.ListElasticsearchFn(ctx, i)
}

// GetElasticsearch implements Interface.
func (m API) GetElasticsearch(ctx context.Context, i *fastly.GetElasticsearchInput) (*fastly.Elasticsearch, error) {
	return m.GetElasticsearchFn(ctx, i)
}

// UpdateElasticsearch implements Interface.
func (m API) UpdateElasticsearch(ctx context.Context, i *fastly.UpdateElasticsearchInput) (*fastly.Elasticsearch, error) {
	return m.UpdateElasticsearchFn(ctx, i)
}

// DeleteElasticsearch implements Interface.
func (m API) DeleteElasticsearch(ctx context.Context, i *fastly.DeleteElasticsearchInput) error {
	return m.DeleteElasticsearchFn(ctx, i)
}

// CreateBlobStorage implements Interface.
func (m API) CreateBlobStorage(ctx context.Context, i *fastly.CreateBlobStorageInput) (*fastly.BlobStorage, error) {
	return m.CreateBlobStorageFn(ctx, i)
}

// ListBlobStorages implements Interface.
func (m API) ListBlobStorages(ctx context.Context, i *fastly.ListBlobStoragesInput) ([]*fastly.BlobStorage, error) {
	return m.ListBlobStoragesFn(ctx, i)
}

// GetBlobStorage implements Interface.
func (m API) GetBlobStorage(ctx context.Context, i *fastly.GetBlobStorageInput) (*fastly.BlobStorage, error) {
	return m.GetBlobStorageFn(ctx, i)
}

// UpdateBlobStorage implements Interface.
func (m API) UpdateBlobStorage(ctx context.Context, i *fastly.UpdateBlobStorageInput) (*fastly.BlobStorage, error) {
	return m.UpdateBlobStorageFn(ctx, i)
}

// DeleteBlobStorage implements Interface.
func (m API) DeleteBlobStorage(ctx context.Context, i *fastly.DeleteBlobStorageInput) error {
	return m.DeleteBlobStorageFn(ctx, i)
}

// CreateDatadog implements Interface.
func (m API) CreateDatadog(ctx context.Context, i *fastly.CreateDatadogInput) (*fastly.Datadog, error) {
	return m.CreateDatadogFn(ctx, i)
}

// ListDatadog implements Interface.
func (m API) ListDatadog(ctx context.Context, i *fastly.ListDatadogInput) ([]*fastly.Datadog, error) {
	return m.ListDatadogFn(ctx, i)
}

// GetDatadog implements Interface.
func (m API) GetDatadog(ctx context.Context, i *fastly.GetDatadogInput) (*fastly.Datadog, error) {
	return m.GetDatadogFn(ctx, i)
}

// UpdateDatadog implements Interface.
func (m API) UpdateDatadog(ctx context.Context, i *fastly.UpdateDatadogInput) (*fastly.Datadog, error) {
	return m.UpdateDatadogFn(ctx, i)
}

// DeleteDatadog implements Interface.
func (m API) DeleteDatadog(ctx context.Context, i *fastly.DeleteDatadogInput) error {
	return m.DeleteDatadogFn(ctx, i)
}

// CreateHTTPS implements Interface.
func (m API) CreateHTTPS(ctx context.Context, i *fastly.CreateHTTPSInput) (*fastly.HTTPS, error) {
	return m.CreateHTTPSFn(ctx, i)
}

// ListHTTPS implements Interface.
func (m API) ListHTTPS(ctx context.Context, i *fastly.ListHTTPSInput) ([]*fastly.HTTPS, error) {
	return m.ListHTTPSFn(ctx, i)
}

// GetHTTPS implements Interface.
func (m API) GetHTTPS(ctx context.Context, i *fastly.GetHTTPSInput) (*fastly.HTTPS, error) {
	return m.GetHTTPSFn(ctx, i)
}

// UpdateHTTPS implements Interface.
func (m API) UpdateHTTPS(ctx context.Context, i *fastly.UpdateHTTPSInput) (*fastly.HTTPS, error) {
	return m.UpdateHTTPSFn(ctx, i)
}

// DeleteHTTPS implements Interface.
func (m API) DeleteHTTPS(ctx context.Context, i *fastly.DeleteHTTPSInput) error {
	return m.DeleteHTTPSFn(ctx, i)
}

// CreateKafka implements Interface.
func (m API) CreateKafka(ctx context.Context, i *fastly.CreateKafkaInput) (*fastly.Kafka, error) {
	return m.CreateKafkaFn(ctx, i)
}

// ListKafkas implements Interface.
func (m API) ListKafkas(ctx context.Context, i *fastly.ListKafkasInput) ([]*fastly.Kafka, error) {
	return m.ListKafkasFn(ctx, i)
}

// GetKafka implements Interface.
func (m API) GetKafka(ctx context.Context, i *fastly.GetKafkaInput) (*fastly.Kafka, error) {
	return m.GetKafkaFn(ctx, i)
}

// UpdateKafka implements Interface.
func (m API) UpdateKafka(ctx context.Context, i *fastly.UpdateKafkaInput) (*fastly.Kafka, error) {
	return m.UpdateKafkaFn(ctx, i)
}

// DeleteKafka implements Interface.
func (m API) DeleteKafka(ctx context.Context, i *fastly.DeleteKafkaInput) error {
	return m.DeleteKafkaFn(ctx, i)
}

// CreatePubsub implements Interface.
func (m API) CreatePubsub(ctx context.Context, i *fastly.CreatePubsubInput) (*fastly.Pubsub, error) {
	return m.CreatePubsubFn(ctx, i)
}

// ListPubsubs implements Interface.
func (m API) ListPubsubs(ctx context.Context, i *fastly.ListPubsubsInput) ([]*fastly.Pubsub, error) {
	return m.ListPubsubsFn(ctx, i)
}

// GetPubsub implements Interface.
func (m API) GetPubsub(ctx context.Context, i *fastly.GetPubsubInput) (*fastly.Pubsub, error) {
	return m.GetPubsubFn(ctx, i)
}

// UpdatePubsub implements Interface.
func (m API) UpdatePubsub(ctx context.Context, i *fastly.UpdatePubsubInput) (*fastly.Pubsub, error) {
	return m.UpdatePubsubFn(ctx, i)
}

// DeletePubsub implements Interface.
func (m API) DeletePubsub(ctx context.Context, i *fastly.DeletePubsubInput) error {
	return m.DeletePubsubFn(ctx, i)
}

// CreateOpenstack implements Interface.
func (m API) CreateOpenstack(ctx context.Context, i *fastly.CreateOpenstackInput) (*fastly.Openstack, error) {
	return m.CreateOpenstackFn(ctx, i)
}

// ListOpenstack implements Interface.
func (m API) ListOpenstack(ctx context.Context, i *fastly.ListOpenstackInput) ([]*fastly.Openstack, error) {
	return m.ListOpenstacksFn(ctx, i)
}

// GetOpenstack implements Interface.
func (m API) GetOpenstack(ctx context.Context, i *fastly.GetOpenstackInput) (*fastly.Openstack, error) {
	return m.GetOpenstackFn(ctx, i)
}

// UpdateOpenstack implements Interface.
func (m API) UpdateOpenstack(ctx context.Context, i *fastly.UpdateOpenstackInput) (*fastly.Openstack, error) {
	return m.UpdateOpenstackFn(ctx, i)
}

// DeleteOpenstack implements Interface.
func (m API) DeleteOpenstack(ctx context.Context, i *fastly.DeleteOpenstackInput) error {
	return m.DeleteOpenstackFn(ctx, i)
}

// GetRegions implements Interface.
func (m API) GetRegions(ctx context.Context) (*fastly.RegionsResponse, error) {
	return m.GetRegionsFn(ctx)
}

// GetStatsJSON implements Interface.
func (m API) GetStatsJSON(ctx context.Context, i *fastly.GetStatsInput, dst any) error {
	return m.GetStatsJSONFn(ctx, i, dst)
}

// CreateManagedLogging implements Interface.
func (m API) CreateManagedLogging(ctx context.Context, i *fastly.CreateManagedLoggingInput) (*fastly.ManagedLogging, error) {
	return m.CreateManagedLoggingFn(ctx, i)
}

// GetGeneratedVCL implements Interface.
func (m API) GetGeneratedVCL(ctx context.Context, i *fastly.GetGeneratedVCLInput) (*fastly.VCL, error) {
	return m.GetGeneratedVCLFn(ctx, i)
}

// CreateVCL implements Interface.
func (m API) CreateVCL(ctx context.Context, i *fastly.CreateVCLInput) (*fastly.VCL, error) {
	return m.CreateVCLFn(ctx, i)
}

// ListVCLs implements Interface.
func (m API) ListVCLs(ctx context.Context, i *fastly.ListVCLsInput) ([]*fastly.VCL, error) {
	return m.ListVCLsFn(ctx, i)
}

// GetVCL implements Interface.
func (m API) GetVCL(ctx context.Context, i *fastly.GetVCLInput) (*fastly.VCL, error) {
	return m.GetVCLFn(ctx, i)
}

// UpdateVCL implements Interface.
func (m API) UpdateVCL(ctx context.Context, i *fastly.UpdateVCLInput) (*fastly.VCL, error) {
	return m.UpdateVCLFn(ctx, i)
}

// DeleteVCL implements Interface.
func (m API) DeleteVCL(ctx context.Context, i *fastly.DeleteVCLInput) error {
	return m.DeleteVCLFn(ctx, i)
}

// CreateSnippet implements Interface.
func (m API) CreateSnippet(ctx context.Context, i *fastly.CreateSnippetInput) (*fastly.Snippet, error) {
	return m.CreateSnippetFn(ctx, i)
}

// ListSnippets implements Interface.
func (m API) ListSnippets(ctx context.Context, i *fastly.ListSnippetsInput) ([]*fastly.Snippet, error) {
	return m.ListSnippetsFn(ctx, i)
}

// GetSnippet implements Interface.
func (m API) GetSnippet(ctx context.Context, i *fastly.GetSnippetInput) (*fastly.Snippet, error) {
	return m.GetSnippetFn(ctx, i)
}

// GetDynamicSnippet implements Interface.
func (m API) GetDynamicSnippet(ctx context.Context, i *fastly.GetDynamicSnippetInput) (*fastly.DynamicSnippet, error) {
	return m.GetDynamicSnippetFn(ctx, i)
}

// UpdateSnippet implements Interface.
func (m API) UpdateSnippet(ctx context.Context, i *fastly.UpdateSnippetInput) (*fastly.Snippet, error) {
	return m.UpdateSnippetFn(ctx, i)
}

// UpdateDynamicSnippet implements Interface.
func (m API) UpdateDynamicSnippet(ctx context.Context, i *fastly.UpdateDynamicSnippetInput) (*fastly.DynamicSnippet, error) {
	return m.UpdateDynamicSnippetFn(ctx, i)
}

// DeleteSnippet implements Interface.
func (m API) DeleteSnippet(ctx context.Context, i *fastly.DeleteSnippetInput) error {
	return m.DeleteSnippetFn(ctx, i)
}

// Purge implements Interface.
func (m API) Purge(ctx context.Context, i *fastly.PurgeInput) (*fastly.Purge, error) {
	return m.PurgeFn(ctx, i)
}

// PurgeKey implements Interface.
func (m API) PurgeKey(ctx context.Context, i *fastly.PurgeKeyInput) (*fastly.Purge, error) {
	return m.PurgeKeyFn(ctx, i)
}

// PurgeKeys implements Interface.
func (m API) PurgeKeys(ctx context.Context, i *fastly.PurgeKeysInput) (map[string]string, error) {
	return m.PurgeKeysFn(ctx, i)
}

// PurgeAll implements Interface.
func (m API) PurgeAll(ctx context.Context, i *fastly.PurgeAllInput) (*fastly.Purge, error) {
	return m.PurgeAllFn(ctx, i)
}

// CreateACL implements Interface.
func (m API) CreateACL(ctx context.Context, i *fastly.CreateACLInput) (*fastly.ACL, error) {
	return m.CreateACLFn(ctx, i)
}

// DeleteACL implements Interface.
func (m API) DeleteACL(ctx context.Context, i *fastly.DeleteACLInput) error {
	return m.DeleteACLFn(ctx, i)
}

// GetACL implements Interface.
func (m API) GetACL(ctx context.Context, i *fastly.GetACLInput) (*fastly.ACL, error) {
	return m.GetACLFn(ctx, i)
}

// ListACLs implements Interface.
func (m API) ListACLs(ctx context.Context, i *fastly.ListACLsInput) ([]*fastly.ACL, error) {
	return m.ListACLsFn(ctx, i)
}

// UpdateACL implements Interface.
func (m API) UpdateACL(ctx context.Context, i *fastly.UpdateACLInput) (*fastly.ACL, error) {
	return m.UpdateACLFn(ctx, i)
}

// CreateACLEntry implements Interface.
func (m API) CreateACLEntry(ctx context.Context, i *fastly.CreateACLEntryInput) (*fastly.ACLEntry, error) {
	return m.CreateACLEntryFn(ctx, i)
}

// DeleteACLEntry implements Interface.
func (m API) DeleteACLEntry(ctx context.Context, i *fastly.DeleteACLEntryInput) error {
	return m.DeleteACLEntryFn(ctx, i)
}

// GetACLEntry implements Interface.
func (m API) GetACLEntry(ctx context.Context, i *fastly.GetACLEntryInput) (*fastly.ACLEntry, error) {
	return m.GetACLEntryFn(ctx, i)
}

// GetACLEntries implements Interface.
func (m API) GetACLEntries(ctx context.Context, i *fastly.GetACLEntriesInput) *fastly.ListPaginator[fastly.ACLEntry] {
	return m.GetACLEntriesFn(ctx, i)
}

// ListACLEntries implements Interface.
func (m API) ListACLEntries(ctx context.Context, i *fastly.ListACLEntriesInput) ([]*fastly.ACLEntry, error) {
	return m.ListACLEntriesFn(ctx, i)
}

// UpdateACLEntry implements Interface.
func (m API) UpdateACLEntry(ctx context.Context, i *fastly.UpdateACLEntryInput) (*fastly.ACLEntry, error) {
	return m.UpdateACLEntryFn(ctx, i)
}

// BatchModifyACLEntries implements Interface.
func (m API) BatchModifyACLEntries(ctx context.Context, i *fastly.BatchModifyACLEntriesInput) error {
	return m.BatchModifyACLEntriesFn(ctx, i)
}

// CreateNewRelic implements Interface.
func (m API) CreateNewRelic(ctx context.Context, i *fastly.CreateNewRelicInput) (*fastly.NewRelic, error) {
	return m.CreateNewRelicFn(ctx, i)
}

// DeleteNewRelic implements Interface.
func (m API) DeleteNewRelic(ctx context.Context, i *fastly.DeleteNewRelicInput) error {
	return m.DeleteNewRelicFn(ctx, i)
}

// GetNewRelic implements Interface.
func (m API) GetNewRelic(ctx context.Context, i *fastly.GetNewRelicInput) (*fastly.NewRelic, error) {
	return m.GetNewRelicFn(ctx, i)
}

// ListNewRelic implements Interface.
func (m API) ListNewRelic(ctx context.Context, i *fastly.ListNewRelicInput) ([]*fastly.NewRelic, error) {
	return m.ListNewRelicFn(ctx, i)
}

// UpdateNewRelic implements Interface.
func (m API) UpdateNewRelic(ctx context.Context, i *fastly.UpdateNewRelicInput) (*fastly.NewRelic, error) {
	return m.UpdateNewRelicFn(ctx, i)
}

// CreateNewRelicOTLP implements Interface.
func (m API) CreateNewRelicOTLP(ctx context.Context, i *fastly.CreateNewRelicOTLPInput) (*fastly.NewRelicOTLP, error) {
	return m.CreateNewRelicOTLPFn(ctx, i)
}

// DeleteNewRelicOTLP implements Interface.
func (m API) DeleteNewRelicOTLP(ctx context.Context, i *fastly.DeleteNewRelicOTLPInput) error {
	return m.DeleteNewRelicOTLPFn(ctx, i)
}

// GetNewRelicOTLP implements Interface.
func (m API) GetNewRelicOTLP(ctx context.Context, i *fastly.GetNewRelicOTLPInput) (*fastly.NewRelicOTLP, error) {
	return m.GetNewRelicOTLPFn(ctx, i)
}

// ListNewRelicOTLP implements Interface.
func (m API) ListNewRelicOTLP(ctx context.Context, i *fastly.ListNewRelicOTLPInput) ([]*fastly.NewRelicOTLP, error) {
	return m.ListNewRelicOTLPFn(ctx, i)
}

// UpdateNewRelicOTLP implements Interface.
func (m API) UpdateNewRelicOTLP(ctx context.Context, i *fastly.UpdateNewRelicOTLPInput) (*fastly.NewRelicOTLP, error) {
	return m.UpdateNewRelicOTLPFn(ctx, i)
}

// CreateUser implements Interface.
func (m API) CreateUser(ctx context.Context, i *fastly.CreateUserInput) (*fastly.User, error) {
	return m.CreateUserFn(ctx, i)
}

// DeleteUser implements Interface.
func (m API) DeleteUser(ctx context.Context, i *fastly.DeleteUserInput) error {
	return m.DeleteUserFn(ctx, i)
}

// GetCurrentUser implements Interface.
func (m API) GetCurrentUser(ctx context.Context) (*fastly.User, error) {
	return m.GetCurrentUserFn(ctx)
}

// GetUser implements Interface.
func (m API) GetUser(ctx context.Context, i *fastly.GetUserInput) (*fastly.User, error) {
	return m.GetUserFn(ctx, i)
}

// ListCustomerUsers implements Interface.
func (m API) ListCustomerUsers(ctx context.Context, i *fastly.ListCustomerUsersInput) ([]*fastly.User, error) {
	return m.ListCustomerUsersFn(ctx, i)
}

// UpdateUser implements Interface.
func (m API) UpdateUser(ctx context.Context, i *fastly.UpdateUserInput) (*fastly.User, error) {
	return m.UpdateUserFn(ctx, i)
}

// ResetUserPassword implements Interface.
func (m API) ResetUserPassword(ctx context.Context, i *fastly.ResetUserPasswordInput) error {
	return m.ResetUserPasswordFn(ctx, i)
}

// BatchDeleteTokens implements Interface.
func (m API) BatchDeleteTokens(ctx context.Context, i *fastly.BatchDeleteTokensInput) error {
	return m.BatchDeleteTokensFn(ctx, i)
}

// CreateToken implements Interface.
func (m API) CreateToken(ctx context.Context, i *fastly.CreateTokenInput) (*fastly.Token, error) {
	return m.CreateTokenFn(ctx, i)
}

// DeleteToken implements Interface.
func (m API) DeleteToken(ctx context.Context, i *fastly.DeleteTokenInput) error {
	return m.DeleteTokenFn(ctx, i)
}

// DeleteTokenSelf implements Interface.
func (m API) DeleteTokenSelf(ctx context.Context) error {
	return m.DeleteTokenSelfFn(ctx)
}

// GetTokenSelf implements Interface.
func (m API) GetTokenSelf(ctx context.Context) (*fastly.Token, error) {
	return m.GetTokenSelfFn(ctx)
}

// ListCustomerTokens implements Interface.
func (m API) ListCustomerTokens(ctx context.Context, i *fastly.ListCustomerTokensInput) ([]*fastly.Token, error) {
	return m.ListCustomerTokensFn(ctx, i)
}

// ListTokens implements Interface.
func (m API) ListTokens(ctx context.Context, i *fastly.ListTokensInput) ([]*fastly.Token, error) {
	return m.ListTokensFn(ctx, i)
}

// NewListKVStoreKeysPaginator implements Interface.
func (m API) NewListKVStoreKeysPaginator(ctx context.Context, i *fastly.ListKVStoreKeysInput) fastly.PaginatorKVStoreEntries {
	return m.NewListKVStoreKeysPaginatorFn(ctx, i)
}

// GetKVStoreItem implements Interface.
func (m API) GetKVStoreItem(ctx context.Context, i *fastly.GetKVStoreItemInput) (fastly.GetKVStoreItemOutput, error) {
	return m.GetKVStoreItemFn(ctx, i)
}

// GetCustomTLSConfiguration implements Interface.
func (m API) GetCustomTLSConfiguration(ctx context.Context, i *fastly.GetCustomTLSConfigurationInput) (*fastly.CustomTLSConfiguration, error) {
	return m.GetCustomTLSConfigurationFn(ctx, i)
}

// ListCustomTLSConfigurations implements Interface.
func (m API) ListCustomTLSConfigurations(ctx context.Context, i *fastly.ListCustomTLSConfigurationsInput) ([]*fastly.CustomTLSConfiguration, error) {
	return m.ListCustomTLSConfigurationsFn(ctx, i)
}

// UpdateCustomTLSConfiguration implements Interface.
func (m API) UpdateCustomTLSConfiguration(ctx context.Context, i *fastly.UpdateCustomTLSConfigurationInput) (*fastly.CustomTLSConfiguration, error) {
	return m.UpdateCustomTLSConfigurationFn(ctx, i)
}

// GetTLSActivation implements Interface.
func (m API) GetTLSActivation(ctx context.Context, i *fastly.GetTLSActivationInput) (*fastly.TLSActivation, error) {
	return m.GetTLSActivationFn(ctx, i)
}

// ListTLSActivations implements Interface.
func (m API) ListTLSActivations(ctx context.Context, i *fastly.ListTLSActivationsInput) ([]*fastly.TLSActivation, error) {
	return m.ListTLSActivationsFn(ctx, i)
}

// UpdateTLSActivation implements Interface.
func (m API) UpdateTLSActivation(ctx context.Context, i *fastly.UpdateTLSActivationInput) (*fastly.TLSActivation, error) {
	return m.UpdateTLSActivationFn(ctx, i)
}

// CreateTLSActivation implements Interface.
func (m API) CreateTLSActivation(ctx context.Context, i *fastly.CreateTLSActivationInput) (*fastly.TLSActivation, error) {
	return m.CreateTLSActivationFn(ctx, i)
}

// DeleteTLSActivation implements Interface.
func (m API) DeleteTLSActivation(ctx context.Context, i *fastly.DeleteTLSActivationInput) error {
	return m.DeleteTLSActivationFn(ctx, i)
}

// CreateCustomTLSCertificate implements Interface.
func (m API) CreateCustomTLSCertificate(ctx context.Context, i *fastly.CreateCustomTLSCertificateInput) (*fastly.CustomTLSCertificate, error) {
	return m.CreateCustomTLSCertificateFn(ctx, i)
}

// DeleteCustomTLSCertificate implements Interface.
func (m API) DeleteCustomTLSCertificate(ctx context.Context, i *fastly.DeleteCustomTLSCertificateInput) error {
	return m.DeleteCustomTLSCertificateFn(ctx, i)
}

// GetCustomTLSCertificate implements Interface.
func (m API) GetCustomTLSCertificate(ctx context.Context, i *fastly.GetCustomTLSCertificateInput) (*fastly.CustomTLSCertificate, error) {
	return m.GetCustomTLSCertificateFn(ctx, i)
}

// ListCustomTLSCertificates implements Interface.
func (m API) ListCustomTLSCertificates(ctx context.Context, i *fastly.ListCustomTLSCertificatesInput) ([]*fastly.CustomTLSCertificate, error) {
	return m.ListCustomTLSCertificatesFn(ctx, i)
}

// UpdateCustomTLSCertificate implements Interface.
func (m API) UpdateCustomTLSCertificate(ctx context.Context, i *fastly.UpdateCustomTLSCertificateInput) (*fastly.CustomTLSCertificate, error) {
	return m.UpdateCustomTLSCertificateFn(ctx, i)
}

// ListTLSDomains implements Interface.
func (m API) ListTLSDomains(ctx context.Context, i *fastly.ListTLSDomainsInput) ([]*fastly.TLSDomain, error) {
	return m.ListTLSDomainsFn(ctx, i)
}

// CreatePrivateKey implements Interface.
func (m API) CreatePrivateKey(ctx context.Context, i *fastly.CreatePrivateKeyInput) (*fastly.PrivateKey, error) {
	return m.CreatePrivateKeyFn(ctx, i)
}

// DeletePrivateKey implements Interface.
func (m API) DeletePrivateKey(ctx context.Context, i *fastly.DeletePrivateKeyInput) error {
	return m.DeletePrivateKeyFn(ctx, i)
}

// GetPrivateKey implements Interface.
func (m API) GetPrivateKey(ctx context.Context, i *fastly.GetPrivateKeyInput) (*fastly.PrivateKey, error) {
	return m.GetPrivateKeyFn(ctx, i)
}

// ListPrivateKeys implements Interface.
func (m API) ListPrivateKeys(ctx context.Context, i *fastly.ListPrivateKeysInput) ([]*fastly.PrivateKey, error) {
	return m.ListPrivateKeysFn(ctx, i)
}

// CreateBulkCertificate implements Interface.
func (m API) CreateBulkCertificate(ctx context.Context, i *fastly.CreateBulkCertificateInput) (*fastly.BulkCertificate, error) {
	return m.CreateBulkCertificateFn(ctx, i)
}

// DeleteBulkCertificate implements Interface.
func (m API) DeleteBulkCertificate(ctx context.Context, i *fastly.DeleteBulkCertificateInput) error {
	return m.DeleteBulkCertificateFn(ctx, i)
}

// GetBulkCertificate implements Interface.
func (m API) GetBulkCertificate(ctx context.Context, i *fastly.GetBulkCertificateInput) (*fastly.BulkCertificate, error) {
	return m.GetBulkCertificateFn(ctx, i)
}

// ListBulkCertificates implements Interface.
func (m API) ListBulkCertificates(ctx context.Context, i *fastly.ListBulkCertificatesInput) ([]*fastly.BulkCertificate, error) {
	return m.ListBulkCertificatesFn(ctx, i)
}

// UpdateBulkCertificate implements Interface.
func (m API) UpdateBulkCertificate(ctx context.Context, i *fastly.UpdateBulkCertificateInput) (*fastly.BulkCertificate, error) {
	return m.UpdateBulkCertificateFn(ctx, i)
}

// CreateTLSSubscription implements Interface.
func (m API) CreateTLSSubscription(ctx context.Context, i *fastly.CreateTLSSubscriptionInput) (*fastly.TLSSubscription, error) {
	return m.CreateTLSSubscriptionFn(ctx, i)
}

// DeleteTLSSubscription implements Interface.
func (m API) DeleteTLSSubscription(ctx context.Context, i *fastly.DeleteTLSSubscriptionInput) error {
	return m.DeleteTLSSubscriptionFn(ctx, i)
}

// GetTLSSubscription implements Interface.
func (m API) GetTLSSubscription(ctx context.Context, i *fastly.GetTLSSubscriptionInput) (*fastly.TLSSubscription, error) {
	return m.GetTLSSubscriptionFn(ctx, i)
}

// ListTLSSubscriptions implements Interface.
func (m API) ListTLSSubscriptions(ctx context.Context, i *fastly.ListTLSSubscriptionsInput) ([]*fastly.TLSSubscription, error) {
	return m.ListTLSSubscriptionsFn(ctx, i)
}

// UpdateTLSSubscription implements Interface.
func (m API) UpdateTLSSubscription(ctx context.Context, i *fastly.UpdateTLSSubscriptionInput) (*fastly.TLSSubscription, error) {
	return m.UpdateTLSSubscriptionFn(ctx, i)
}

// ListServiceAuthorizations implements Interface.
func (m API) ListServiceAuthorizations(ctx context.Context, i *fastly.ListServiceAuthorizationsInput) (*fastly.ServiceAuthorizations, error) {
	return m.ListServiceAuthorizationsFn(ctx, i)
}

// GetServiceAuthorization implements Interface.
func (m API) GetServiceAuthorization(ctx context.Context, i *fastly.GetServiceAuthorizationInput) (*fastly.ServiceAuthorization, error) {
	return m.GetServiceAuthorizationFn(ctx, i)
}

// CreateServiceAuthorization implements Interface.
func (m API) CreateServiceAuthorization(ctx context.Context, i *fastly.CreateServiceAuthorizationInput) (*fastly.ServiceAuthorization, error) {
	return m.CreateServiceAuthorizationFn(ctx, i)
}

// UpdateServiceAuthorization implements Interface.
func (m API) UpdateServiceAuthorization(ctx context.Context, i *fastly.UpdateServiceAuthorizationInput) (*fastly.ServiceAuthorization, error) {
	return m.UpdateServiceAuthorizationFn(ctx, i)
}

// DeleteServiceAuthorization implements Interface.
func (m API) DeleteServiceAuthorization(ctx context.Context, i *fastly.DeleteServiceAuthorizationInput) error {
	return m.DeleteServiceAuthorizationFn(ctx, i)
}

// CreateConfigStore implements Interface.
func (m API) CreateConfigStore(ctx context.Context, i *fastly.CreateConfigStoreInput) (*fastly.ConfigStore, error) {
	return m.CreateConfigStoreFn(ctx, i)
}

// DeleteConfigStore implements Interface.
func (m API) DeleteConfigStore(ctx context.Context, i *fastly.DeleteConfigStoreInput) error {
	return m.DeleteConfigStoreFn(ctx, i)
}

// GetConfigStore implements Interface.
func (m API) GetConfigStore(ctx context.Context, i *fastly.GetConfigStoreInput) (*fastly.ConfigStore, error) {
	return m.GetConfigStoreFn(ctx, i)
}

// GetConfigStoreMetadata implements Interface.
func (m API) GetConfigStoreMetadata(ctx context.Context, i *fastly.GetConfigStoreMetadataInput) (*fastly.ConfigStoreMetadata, error) {
	return m.GetConfigStoreMetadataFn(ctx, i)
}

// ListConfigStores implements Interface.
func (m API) ListConfigStores(ctx context.Context, i *fastly.ListConfigStoresInput) ([]*fastly.ConfigStore, error) {
	return m.ListConfigStoresFn(ctx, i)
}

// ListConfigStoreServices implements Interface.
func (m API) ListConfigStoreServices(ctx context.Context, i *fastly.ListConfigStoreServicesInput) ([]*fastly.Service, error) {
	return m.ListConfigStoreServicesFn(ctx, i)
}

// UpdateConfigStore implements Interface.
func (m API) UpdateConfigStore(ctx context.Context, i *fastly.UpdateConfigStoreInput) (*fastly.ConfigStore, error) {
	return m.UpdateConfigStoreFn(ctx, i)
}

// CreateConfigStoreItem implements Interface.
func (m API) CreateConfigStoreItem(ctx context.Context, i *fastly.CreateConfigStoreItemInput) (*fastly.ConfigStoreItem, error) {
	return m.CreateConfigStoreItemFn(ctx, i)
}

// DeleteConfigStoreItem implements Interface.
func (m API) DeleteConfigStoreItem(ctx context.Context, i *fastly.DeleteConfigStoreItemInput) error {
	return m.DeleteConfigStoreItemFn(ctx, i)
}

// GetConfigStoreItem implements Interface.
func (m API) GetConfigStoreItem(ctx context.Context, i *fastly.GetConfigStoreItemInput) (*fastly.ConfigStoreItem, error) {
	return m.GetConfigStoreItemFn(ctx, i)
}

// ListConfigStoreItems implements Interface.
func (m API) ListConfigStoreItems(ctx context.Context, i *fastly.ListConfigStoreItemsInput) ([]*fastly.ConfigStoreItem, error) {
	return m.ListConfigStoreItemsFn(ctx, i)
}

// UpdateConfigStoreItem implements Interface.
func (m API) UpdateConfigStoreItem(ctx context.Context, i *fastly.UpdateConfigStoreItemInput) (*fastly.ConfigStoreItem, error) {
	return m.UpdateConfigStoreItemFn(ctx, i)
}

// CreateKVStore implements Interface.
func (m API) CreateKVStore(ctx context.Context, i *fastly.CreateKVStoreInput) (*fastly.KVStore, error) {
	return m.CreateKVStoreFn(ctx, i)
}

// GetKVStore implements Interface.
func (m API) GetKVStore(ctx context.Context, i *fastly.GetKVStoreInput) (*fastly.KVStore, error) {
	return m.GetKVStoreFn(ctx, i)
}

// ListKVStores implements Interface.
func (m API) ListKVStores(ctx context.Context, i *fastly.ListKVStoresInput) (*fastly.ListKVStoresResponse, error) {
	return m.ListKVStoresFn(ctx, i)
}

// DeleteKVStore implements Interface.
func (m API) DeleteKVStore(ctx context.Context, i *fastly.DeleteKVStoreInput) error {
	return m.DeleteKVStoreFn(ctx, i)
}

// ListKVStoreKeys implements Interface.
func (m API) ListKVStoreKeys(ctx context.Context, i *fastly.ListKVStoreKeysInput) (*fastly.ListKVStoreKeysResponse, error) {
	return m.ListKVStoreKeysFn(ctx, i)
}

// GetKVStoreKey implements Interface.
func (m API) GetKVStoreKey(ctx context.Context, i *fastly.GetKVStoreKeyInput) (string, error) {
	return m.GetKVStoreKeyFn(ctx, i)
}

// InsertKVStoreKey implements Interface.
func (m API) InsertKVStoreKey(ctx context.Context, i *fastly.InsertKVStoreKeyInput) error {
	return m.InsertKVStoreKeyFn(ctx, i)
}

// DeleteKVStoreKey implements Interface.
func (m API) DeleteKVStoreKey(ctx context.Context, i *fastly.DeleteKVStoreKeyInput) error {
	return m.DeleteKVStoreKeyFn(ctx, i)
}

// BatchModifyKVStoreKey implements Interface.
func (m API) BatchModifyKVStoreKey(ctx context.Context, i *fastly.BatchModifyKVStoreKeyInput) error {
	return m.BatchModifyKVStoreKeyFn(ctx, i)
}

// CreateSecretStore implements Interface.
func (m API) CreateSecretStore(ctx context.Context, i *fastly.CreateSecretStoreInput) (*fastly.SecretStore, error) {
	return m.CreateSecretStoreFn(ctx, i)
}

// GetSecretStore implements Interface.
func (m API) GetSecretStore(ctx context.Context, i *fastly.GetSecretStoreInput) (*fastly.SecretStore, error) {
	return m.GetSecretStoreFn(ctx, i)
}

// DeleteSecretStore implements Interface.
func (m API) DeleteSecretStore(ctx context.Context, i *fastly.DeleteSecretStoreInput) error {
	return m.DeleteSecretStoreFn(ctx, i)
}

// ListSecretStores implements Interface.
func (m API) ListSecretStores(ctx context.Context, i *fastly.ListSecretStoresInput) (*fastly.SecretStores, error) {
	return m.ListSecretStoresFn(ctx, i)
}

// CreateSecret implements Interface.
func (m API) CreateSecret(ctx context.Context, i *fastly.CreateSecretInput) (*fastly.Secret, error) {
	return m.CreateSecretFn(ctx, i)
}

// GetSecret implements Interface.
func (m API) GetSecret(ctx context.Context, i *fastly.GetSecretInput) (*fastly.Secret, error) {
	return m.GetSecretFn(ctx, i)
}

// DeleteSecret implements Interface.
func (m API) DeleteSecret(ctx context.Context, i *fastly.DeleteSecretInput) error {
	return m.DeleteSecretFn(ctx, i)
}

// ListSecrets implements Interface.
func (m API) ListSecrets(ctx context.Context, i *fastly.ListSecretsInput) (*fastly.Secrets, error) {
	return m.ListSecretsFn(ctx, i)
}

// CreateClientKey implements Interface.
func (m API) CreateClientKey(ctx context.Context) (*fastly.ClientKey, error) {
	return m.CreateClientKeyFn(ctx)
}

// GetSigningKey implements Interface.
func (m API) GetSigningKey(ctx context.Context) (ed25519.PublicKey, error) {
	return m.GetSigningKeyFn(ctx)
}

// CreateResource implements Interface.
func (m API) CreateResource(ctx context.Context, i *fastly.CreateResourceInput) (*fastly.Resource, error) {
	return m.CreateResourceFn(ctx, i)
}

// DeleteResource implements Interface.
func (m API) DeleteResource(ctx context.Context, i *fastly.DeleteResourceInput) error {
	return m.DeleteResourceFn(ctx, i)
}

// GetResource implements Interface.
func (m API) GetResource(ctx context.Context, i *fastly.GetResourceInput) (*fastly.Resource, error) {
	return m.GetResourceFn(ctx, i)
}

// ListResources implements Interface.
func (m API) ListResources(ctx context.Context, i *fastly.ListResourcesInput) ([]*fastly.Resource, error) {
	return m.ListResourcesFn(ctx, i)
}

// UpdateResource implements Interface.
func (m API) UpdateResource(ctx context.Context, i *fastly.UpdateResourceInput) (*fastly.Resource, error) {
	return m.UpdateResourceFn(ctx, i)
}

// CreateERL implements Interface.
func (m API) CreateERL(ctx context.Context, i *fastly.CreateERLInput) (*fastly.ERL, error) {
	return m.CreateERLFn(ctx, i)
}

// DeleteERL implements Interface.
func (m API) DeleteERL(ctx context.Context, i *fastly.DeleteERLInput) error {
	return m.DeleteERLFn(ctx, i)
}

// GetERL implements Interface.
func (m API) GetERL(ctx context.Context, i *fastly.GetERLInput) (*fastly.ERL, error) {
	return m.GetERLFn(ctx, i)
}

// ListERLs implements Interface.
func (m API) ListERLs(ctx context.Context, i *fastly.ListERLsInput) ([]*fastly.ERL, error) {
	return m.ListERLsFn(ctx, i)
}

// UpdateERL implements Interface.
func (m API) UpdateERL(ctx context.Context, i *fastly.UpdateERLInput) (*fastly.ERL, error) {
	return m.UpdateERLFn(ctx, i)
}

// CreateCondition implements Interface.
func (m API) CreateCondition(ctx context.Context, i *fastly.CreateConditionInput) (*fastly.Condition, error) {
	return m.CreateConditionFn(ctx, i)
}

// DeleteCondition implements Interface.
func (m API) DeleteCondition(ctx context.Context, i *fastly.DeleteConditionInput) error {
	return m.DeleteConditionFn(ctx, i)
}

// GetCondition implements Interface.
func (m API) GetCondition(ctx context.Context, i *fastly.GetConditionInput) (*fastly.Condition, error) {
	return m.GetConditionFn(ctx, i)
}

// ListConditions implements Interface.
func (m API) ListConditions(ctx context.Context, i *fastly.ListConditionsInput) ([]*fastly.Condition, error) {
	return m.ListConditionsFn(ctx, i)
}

// UpdateCondition implements Interface.
func (m API) UpdateCondition(ctx context.Context, i *fastly.UpdateConditionInput) (*fastly.Condition, error) {
	return m.UpdateConditionFn(ctx, i)
}

// ListAlertDefinitions implements Interface.
func (m API) ListAlertDefinitions(ctx context.Context, i *fastly.ListAlertDefinitionsInput) (*fastly.AlertDefinitionsResponse, error) {
	return m.ListAlertDefinitionsFn(ctx, i)
}

// CreateAlertDefinition implements Interface.
func (m API) CreateAlertDefinition(ctx context.Context, i *fastly.CreateAlertDefinitionInput) (*fastly.AlertDefinition, error) {
	return m.CreateAlertDefinitionFn(ctx, i)
}

// GetAlertDefinition implements Interface.
func (m API) GetAlertDefinition(ctx context.Context, i *fastly.GetAlertDefinitionInput) (*fastly.AlertDefinition, error) {
	return m.GetAlertDefinitionFn(ctx, i)
}

// UpdateAlertDefinition implements Interface.
func (m API) UpdateAlertDefinition(ctx context.Context, i *fastly.UpdateAlertDefinitionInput) (*fastly.AlertDefinition, error) {
	return m.UpdateAlertDefinitionFn(ctx, i)
}

// DeleteAlertDefinition implements Interface.
func (m API) DeleteAlertDefinition(ctx context.Context, i *fastly.DeleteAlertDefinitionInput) error {
	return m.DeleteAlertDefinitionFn(ctx, i)
}

// TestAlertDefinition implements Interface.
func (m API) TestAlertDefinition(ctx context.Context, i *fastly.TestAlertDefinitionInput) error {
	return m.TestAlertDefinitionFn(ctx, i)
}

// ListAlertHistory implements Interface.
func (m API) ListAlertHistory(ctx context.Context, i *fastly.ListAlertHistoryInput) (*fastly.AlertHistoryResponse, error) {
	return m.ListAlertHistoryFn(ctx, i)
}

// CreateObservabilityCustomDashboard implements Interface.
func (m API) CreateObservabilityCustomDashboard(ctx context.Context, i *fastly.CreateObservabilityCustomDashboardInput) (*fastly.ObservabilityCustomDashboard, error) {
	return m.CreateObservabilityCustomDashboardFn(ctx, i)
}

// DeleteObservabilityCustomDashboard implements Interface.
func (m API) DeleteObservabilityCustomDashboard(ctx context.Context, i *fastly.DeleteObservabilityCustomDashboardInput) error {
	return m.DeleteObservabilityCustomDashboardFn(ctx, i)
}

// GetObservabilityCustomDashboard implements Interface.
func (m API) GetObservabilityCustomDashboard(ctx context.Context, i *fastly.GetObservabilityCustomDashboardInput) (*fastly.ObservabilityCustomDashboard, error) {
	return m.GetObservabilityCustomDashboardFn(ctx, i)
}

// ListObservabilityCustomDashboards implements Interface.
func (m API) ListObservabilityCustomDashboards(ctx context.Context, i *fastly.ListObservabilityCustomDashboardsInput) (*fastly.ListDashboardsResponse, error) {
	return m.ListObservabilityCustomDashboardsFn(ctx, i)
}

// UpdateObservabilityCustomDashboard implements Interface.
func (m API) UpdateObservabilityCustomDashboard(ctx context.Context, i *fastly.UpdateObservabilityCustomDashboardInput) (*fastly.ObservabilityCustomDashboard, error) {
	return m.UpdateObservabilityCustomDashboardFn(ctx, i)
}

// GetImageOptimizerDefaultSettings implements Interface.
func (m API) GetImageOptimizerDefaultSettings(ctx context.Context, i *fastly.GetImageOptimizerDefaultSettingsInput) (*fastly.ImageOptimizerDefaultSettings, error) {
	return m.GetImageOptimizerDefaultSettingsFn(ctx, i)
}

// UpdateImageOptimizerDefaultSettings implements Interface.
func (m API) UpdateImageOptimizerDefaultSettings(ctx context.Context, i *fastly.UpdateImageOptimizerDefaultSettingsInput) (*fastly.ImageOptimizerDefaultSettings, error) {
	return m.UpdateImageOptimizerDefaultSettingsFn(ctx, i)
}
