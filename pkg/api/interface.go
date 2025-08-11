package api

import (
	"context"
	"crypto/ed25519"
	"net/http"

	"github.com/fastly/go-fastly/v11/fastly"
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
	AllIPs(context.Context) (v4, v6 fastly.IPAddrs, err error)
	AllDatacenters(context.Context) (datacenters []fastly.Datacenter, err error)

	CreateService(context.Context, *fastly.CreateServiceInput) (*fastly.Service, error)
	GetServices(context.Context, *fastly.GetServicesInput) *fastly.ListPaginator[fastly.Service]
	ListServices(context.Context, *fastly.ListServicesInput) ([]*fastly.Service, error)
	GetService(context.Context, *fastly.GetServiceInput) (*fastly.Service, error)
	GetServiceDetails(context.Context, *fastly.GetServiceInput) (*fastly.ServiceDetail, error)
	UpdateService(context.Context, *fastly.UpdateServiceInput) (*fastly.Service, error)
	DeleteService(context.Context, *fastly.DeleteServiceInput) error
	SearchService(context.Context, *fastly.SearchServiceInput) (*fastly.Service, error)

	CloneVersion(context.Context, *fastly.CloneVersionInput) (*fastly.Version, error)
	ListVersions(context.Context, *fastly.ListVersionsInput) ([]*fastly.Version, error)
	GetVersion(context.Context, *fastly.GetVersionInput) (*fastly.Version, error)
	UpdateVersion(context.Context, *fastly.UpdateVersionInput) (*fastly.Version, error)
	ActivateVersion(context.Context, *fastly.ActivateVersionInput) (*fastly.Version, error)
	DeactivateVersion(context.Context, *fastly.DeactivateVersionInput) (*fastly.Version, error)
	LockVersion(context.Context, *fastly.LockVersionInput) (*fastly.Version, error)
	LatestVersion(context.Context, *fastly.LatestVersionInput) (*fastly.Version, error)

	CreateDomain(context.Context, *fastly.CreateDomainInput) (*fastly.Domain, error)
	ListDomains(context.Context, *fastly.ListDomainsInput) ([]*fastly.Domain, error)
	GetDomain(context.Context, *fastly.GetDomainInput) (*fastly.Domain, error)
	UpdateDomain(context.Context, *fastly.UpdateDomainInput) (*fastly.Domain, error)
	DeleteDomain(context.Context, *fastly.DeleteDomainInput) error
	ValidateDomain(context.Context, *fastly.ValidateDomainInput) (*fastly.DomainValidationResult, error)
	ValidateAllDomains(context.Context, *fastly.ValidateAllDomainsInput) ([]*fastly.DomainValidationResult, error)

	CreateBackend(context.Context, *fastly.CreateBackendInput) (*fastly.Backend, error)
	ListBackends(context.Context, *fastly.ListBackendsInput) ([]*fastly.Backend, error)
	GetBackend(context.Context, *fastly.GetBackendInput) (*fastly.Backend, error)
	UpdateBackend(context.Context, *fastly.UpdateBackendInput) (*fastly.Backend, error)
	DeleteBackend(context.Context, *fastly.DeleteBackendInput) error

	CreateHealthCheck(context.Context, *fastly.CreateHealthCheckInput) (*fastly.HealthCheck, error)
	ListHealthChecks(context.Context, *fastly.ListHealthChecksInput) ([]*fastly.HealthCheck, error)
	GetHealthCheck(context.Context, *fastly.GetHealthCheckInput) (*fastly.HealthCheck, error)
	UpdateHealthCheck(context.Context, *fastly.UpdateHealthCheckInput) (*fastly.HealthCheck, error)
	DeleteHealthCheck(context.Context, *fastly.DeleteHealthCheckInput) error

	GetPackage(context.Context, *fastly.GetPackageInput) (*fastly.Package, error)
	UpdatePackage(context.Context, *fastly.UpdatePackageInput) (*fastly.Package, error)

	CreateDictionary(context.Context, *fastly.CreateDictionaryInput) (*fastly.Dictionary, error)
	GetDictionary(context.Context, *fastly.GetDictionaryInput) (*fastly.Dictionary, error)
	DeleteDictionary(context.Context, *fastly.DeleteDictionaryInput) error
	ListDictionaries(context.Context, *fastly.ListDictionariesInput) ([]*fastly.Dictionary, error)
	UpdateDictionary(context.Context, *fastly.UpdateDictionaryInput) (*fastly.Dictionary, error)

	GetDictionaryItems(context.Context, *fastly.GetDictionaryItemsInput) *fastly.ListPaginator[fastly.DictionaryItem]
	ListDictionaryItems(context.Context, *fastly.ListDictionaryItemsInput) ([]*fastly.DictionaryItem, error)
	GetDictionaryItem(context.Context, *fastly.GetDictionaryItemInput) (*fastly.DictionaryItem, error)
	CreateDictionaryItem(context.Context, *fastly.CreateDictionaryItemInput) (*fastly.DictionaryItem, error)
	UpdateDictionaryItem(context.Context, *fastly.UpdateDictionaryItemInput) (*fastly.DictionaryItem, error)
	DeleteDictionaryItem(context.Context, *fastly.DeleteDictionaryItemInput) error
	BatchModifyDictionaryItems(context.Context, *fastly.BatchModifyDictionaryItemsInput) error

	GetDictionaryInfo(context.Context, *fastly.GetDictionaryInfoInput) (*fastly.DictionaryInfo, error)

	CreateBigQuery(context.Context, *fastly.CreateBigQueryInput) (*fastly.BigQuery, error)
	ListBigQueries(context.Context, *fastly.ListBigQueriesInput) ([]*fastly.BigQuery, error)
	GetBigQuery(context.Context, *fastly.GetBigQueryInput) (*fastly.BigQuery, error)
	UpdateBigQuery(context.Context, *fastly.UpdateBigQueryInput) (*fastly.BigQuery, error)
	DeleteBigQuery(context.Context, *fastly.DeleteBigQueryInput) error

	CreateS3(context.Context, *fastly.CreateS3Input) (*fastly.S3, error)
	ListS3s(context.Context, *fastly.ListS3sInput) ([]*fastly.S3, error)
	GetS3(context.Context, *fastly.GetS3Input) (*fastly.S3, error)
	UpdateS3(context.Context, *fastly.UpdateS3Input) (*fastly.S3, error)
	DeleteS3(context.Context, *fastly.DeleteS3Input) error

	CreateKinesis(context.Context, *fastly.CreateKinesisInput) (*fastly.Kinesis, error)
	ListKinesis(context.Context, *fastly.ListKinesisInput) ([]*fastly.Kinesis, error)
	GetKinesis(context.Context, *fastly.GetKinesisInput) (*fastly.Kinesis, error)
	UpdateKinesis(context.Context, *fastly.UpdateKinesisInput) (*fastly.Kinesis, error)
	DeleteKinesis(context.Context, *fastly.DeleteKinesisInput) error

	CreateSyslog(context.Context, *fastly.CreateSyslogInput) (*fastly.Syslog, error)
	ListSyslogs(context.Context, *fastly.ListSyslogsInput) ([]*fastly.Syslog, error)
	GetSyslog(context.Context, *fastly.GetSyslogInput) (*fastly.Syslog, error)
	UpdateSyslog(context.Context, *fastly.UpdateSyslogInput) (*fastly.Syslog, error)
	DeleteSyslog(context.Context, *fastly.DeleteSyslogInput) error

	CreateLogentries(context.Context, *fastly.CreateLogentriesInput) (*fastly.Logentries, error)
	ListLogentries(context.Context, *fastly.ListLogentriesInput) ([]*fastly.Logentries, error)
	GetLogentries(context.Context, *fastly.GetLogentriesInput) (*fastly.Logentries, error)
	UpdateLogentries(context.Context, *fastly.UpdateLogentriesInput) (*fastly.Logentries, error)
	DeleteLogentries(context.Context, *fastly.DeleteLogentriesInput) error

	CreatePapertrail(context.Context, *fastly.CreatePapertrailInput) (*fastly.Papertrail, error)
	ListPapertrails(context.Context, *fastly.ListPapertrailsInput) ([]*fastly.Papertrail, error)
	GetPapertrail(context.Context, *fastly.GetPapertrailInput) (*fastly.Papertrail, error)
	UpdatePapertrail(context.Context, *fastly.UpdatePapertrailInput) (*fastly.Papertrail, error)
	DeletePapertrail(context.Context, *fastly.DeletePapertrailInput) error

	CreateSumologic(context.Context, *fastly.CreateSumologicInput) (*fastly.Sumologic, error)
	ListSumologics(context.Context, *fastly.ListSumologicsInput) ([]*fastly.Sumologic, error)
	GetSumologic(context.Context, *fastly.GetSumologicInput) (*fastly.Sumologic, error)
	UpdateSumologic(context.Context, *fastly.UpdateSumologicInput) (*fastly.Sumologic, error)
	DeleteSumologic(context.Context, *fastly.DeleteSumologicInput) error

	CreateGCS(context.Context, *fastly.CreateGCSInput) (*fastly.GCS, error)
	ListGCSs(context.Context, *fastly.ListGCSsInput) ([]*fastly.GCS, error)
	GetGCS(context.Context, *fastly.GetGCSInput) (*fastly.GCS, error)
	UpdateGCS(context.Context, *fastly.UpdateGCSInput) (*fastly.GCS, error)
	DeleteGCS(context.Context, *fastly.DeleteGCSInput) error

	CreateGrafanaCloudLogs(context.Context, *fastly.CreateGrafanaCloudLogsInput) (*fastly.GrafanaCloudLogs, error)
	ListGrafanaCloudLogs(context.Context, *fastly.ListGrafanaCloudLogsInput) ([]*fastly.GrafanaCloudLogs, error)
	GetGrafanaCloudLogs(context.Context, *fastly.GetGrafanaCloudLogsInput) (*fastly.GrafanaCloudLogs, error)
	UpdateGrafanaCloudLogs(context.Context, *fastly.UpdateGrafanaCloudLogsInput) (*fastly.GrafanaCloudLogs, error)
	DeleteGrafanaCloudLogs(context.Context, *fastly.DeleteGrafanaCloudLogsInput) error

	CreateFTP(context.Context, *fastly.CreateFTPInput) (*fastly.FTP, error)
	ListFTPs(context.Context, *fastly.ListFTPsInput) ([]*fastly.FTP, error)
	GetFTP(context.Context, *fastly.GetFTPInput) (*fastly.FTP, error)
	UpdateFTP(context.Context, *fastly.UpdateFTPInput) (*fastly.FTP, error)
	DeleteFTP(context.Context, *fastly.DeleteFTPInput) error

	CreateSplunk(context.Context, *fastly.CreateSplunkInput) (*fastly.Splunk, error)
	ListSplunks(context.Context, *fastly.ListSplunksInput) ([]*fastly.Splunk, error)
	GetSplunk(context.Context, *fastly.GetSplunkInput) (*fastly.Splunk, error)
	UpdateSplunk(context.Context, *fastly.UpdateSplunkInput) (*fastly.Splunk, error)
	DeleteSplunk(context.Context, *fastly.DeleteSplunkInput) error

	CreateScalyr(context.Context, *fastly.CreateScalyrInput) (*fastly.Scalyr, error)
	ListScalyrs(context.Context, *fastly.ListScalyrsInput) ([]*fastly.Scalyr, error)
	GetScalyr(context.Context, *fastly.GetScalyrInput) (*fastly.Scalyr, error)
	UpdateScalyr(context.Context, *fastly.UpdateScalyrInput) (*fastly.Scalyr, error)
	DeleteScalyr(context.Context, *fastly.DeleteScalyrInput) error

	CreateLoggly(context.Context, *fastly.CreateLogglyInput) (*fastly.Loggly, error)
	ListLoggly(context.Context, *fastly.ListLogglyInput) ([]*fastly.Loggly, error)
	GetLoggly(context.Context, *fastly.GetLogglyInput) (*fastly.Loggly, error)
	UpdateLoggly(context.Context, *fastly.UpdateLogglyInput) (*fastly.Loggly, error)
	DeleteLoggly(context.Context, *fastly.DeleteLogglyInput) error

	CreateHoneycomb(context.Context, *fastly.CreateHoneycombInput) (*fastly.Honeycomb, error)
	ListHoneycombs(context.Context, *fastly.ListHoneycombsInput) ([]*fastly.Honeycomb, error)
	GetHoneycomb(context.Context, *fastly.GetHoneycombInput) (*fastly.Honeycomb, error)
	UpdateHoneycomb(context.Context, *fastly.UpdateHoneycombInput) (*fastly.Honeycomb, error)
	DeleteHoneycomb(context.Context, *fastly.DeleteHoneycombInput) error

	CreateHeroku(context.Context, *fastly.CreateHerokuInput) (*fastly.Heroku, error)
	ListHerokus(context.Context, *fastly.ListHerokusInput) ([]*fastly.Heroku, error)
	GetHeroku(context.Context, *fastly.GetHerokuInput) (*fastly.Heroku, error)
	UpdateHeroku(context.Context, *fastly.UpdateHerokuInput) (*fastly.Heroku, error)
	DeleteHeroku(context.Context, *fastly.DeleteHerokuInput) error

	CreateSFTP(context.Context, *fastly.CreateSFTPInput) (*fastly.SFTP, error)
	ListSFTPs(context.Context, *fastly.ListSFTPsInput) ([]*fastly.SFTP, error)
	GetSFTP(context.Context, *fastly.GetSFTPInput) (*fastly.SFTP, error)
	UpdateSFTP(context.Context, *fastly.UpdateSFTPInput) (*fastly.SFTP, error)
	DeleteSFTP(context.Context, *fastly.DeleteSFTPInput) error

	CreateLogshuttle(context.Context, *fastly.CreateLogshuttleInput) (*fastly.Logshuttle, error)
	ListLogshuttles(context.Context, *fastly.ListLogshuttlesInput) ([]*fastly.Logshuttle, error)
	GetLogshuttle(context.Context, *fastly.GetLogshuttleInput) (*fastly.Logshuttle, error)
	UpdateLogshuttle(context.Context, *fastly.UpdateLogshuttleInput) (*fastly.Logshuttle, error)
	DeleteLogshuttle(context.Context, *fastly.DeleteLogshuttleInput) error

	CreateCloudfiles(context.Context, *fastly.CreateCloudfilesInput) (*fastly.Cloudfiles, error)
	ListCloudfiles(context.Context, *fastly.ListCloudfilesInput) ([]*fastly.Cloudfiles, error)
	GetCloudfiles(context.Context, *fastly.GetCloudfilesInput) (*fastly.Cloudfiles, error)
	UpdateCloudfiles(context.Context, *fastly.UpdateCloudfilesInput) (*fastly.Cloudfiles, error)
	DeleteCloudfiles(context.Context, *fastly.DeleteCloudfilesInput) error

	CreateDigitalOcean(context.Context, *fastly.CreateDigitalOceanInput) (*fastly.DigitalOcean, error)
	ListDigitalOceans(context.Context, *fastly.ListDigitalOceansInput) ([]*fastly.DigitalOcean, error)
	GetDigitalOcean(context.Context, *fastly.GetDigitalOceanInput) (*fastly.DigitalOcean, error)
	UpdateDigitalOcean(context.Context, *fastly.UpdateDigitalOceanInput) (*fastly.DigitalOcean, error)
	DeleteDigitalOcean(context.Context, *fastly.DeleteDigitalOceanInput) error

	CreateElasticsearch(context.Context, *fastly.CreateElasticsearchInput) (*fastly.Elasticsearch, error)
	ListElasticsearch(context.Context, *fastly.ListElasticsearchInput) ([]*fastly.Elasticsearch, error)
	GetElasticsearch(context.Context, *fastly.GetElasticsearchInput) (*fastly.Elasticsearch, error)
	UpdateElasticsearch(context.Context, *fastly.UpdateElasticsearchInput) (*fastly.Elasticsearch, error)
	DeleteElasticsearch(context.Context, *fastly.DeleteElasticsearchInput) error

	CreateBlobStorage(context.Context, *fastly.CreateBlobStorageInput) (*fastly.BlobStorage, error)
	ListBlobStorages(context.Context, *fastly.ListBlobStoragesInput) ([]*fastly.BlobStorage, error)
	GetBlobStorage(context.Context, *fastly.GetBlobStorageInput) (*fastly.BlobStorage, error)
	UpdateBlobStorage(context.Context, *fastly.UpdateBlobStorageInput) (*fastly.BlobStorage, error)
	DeleteBlobStorage(context.Context, *fastly.DeleteBlobStorageInput) error

	CreateDatadog(context.Context, *fastly.CreateDatadogInput) (*fastly.Datadog, error)
	ListDatadog(context.Context, *fastly.ListDatadogInput) ([]*fastly.Datadog, error)
	GetDatadog(context.Context, *fastly.GetDatadogInput) (*fastly.Datadog, error)
	UpdateDatadog(context.Context, *fastly.UpdateDatadogInput) (*fastly.Datadog, error)
	DeleteDatadog(context.Context, *fastly.DeleteDatadogInput) error

	CreateHTTPS(context.Context, *fastly.CreateHTTPSInput) (*fastly.HTTPS, error)
	ListHTTPS(context.Context, *fastly.ListHTTPSInput) ([]*fastly.HTTPS, error)
	GetHTTPS(context.Context, *fastly.GetHTTPSInput) (*fastly.HTTPS, error)
	UpdateHTTPS(context.Context, *fastly.UpdateHTTPSInput) (*fastly.HTTPS, error)
	DeleteHTTPS(context.Context, *fastly.DeleteHTTPSInput) error

	CreateKafka(context.Context, *fastly.CreateKafkaInput) (*fastly.Kafka, error)
	ListKafkas(context.Context, *fastly.ListKafkasInput) ([]*fastly.Kafka, error)
	GetKafka(context.Context, *fastly.GetKafkaInput) (*fastly.Kafka, error)
	UpdateKafka(context.Context, *fastly.UpdateKafkaInput) (*fastly.Kafka, error)
	DeleteKafka(context.Context, *fastly.DeleteKafkaInput) error

	CreatePubsub(context.Context, *fastly.CreatePubsubInput) (*fastly.Pubsub, error)
	ListPubsubs(context.Context, *fastly.ListPubsubsInput) ([]*fastly.Pubsub, error)
	GetPubsub(context.Context, *fastly.GetPubsubInput) (*fastly.Pubsub, error)
	UpdatePubsub(context.Context, *fastly.UpdatePubsubInput) (*fastly.Pubsub, error)
	DeletePubsub(context.Context, *fastly.DeletePubsubInput) error

	CreateOpenstack(context.Context, *fastly.CreateOpenstackInput) (*fastly.Openstack, error)
	ListOpenstack(context.Context, *fastly.ListOpenstackInput) ([]*fastly.Openstack, error)
	GetOpenstack(context.Context, *fastly.GetOpenstackInput) (*fastly.Openstack, error)
	UpdateOpenstack(context.Context, *fastly.UpdateOpenstackInput) (*fastly.Openstack, error)
	DeleteOpenstack(context.Context, *fastly.DeleteOpenstackInput) error

	GetRegions(context.Context) (*fastly.RegionsResponse, error)
	GetStatsJSON(context.Context, *fastly.GetStatsInput, any) error

	CreateManagedLogging(context.Context, *fastly.CreateManagedLoggingInput) (*fastly.ManagedLogging, error)

	GetGeneratedVCL(context.Context, *fastly.GetGeneratedVCLInput) (*fastly.VCL, error)

	CreateVCL(context.Context, *fastly.CreateVCLInput) (*fastly.VCL, error)
	ListVCLs(context.Context, *fastly.ListVCLsInput) ([]*fastly.VCL, error)
	GetVCL(context.Context, *fastly.GetVCLInput) (*fastly.VCL, error)
	UpdateVCL(context.Context, *fastly.UpdateVCLInput) (*fastly.VCL, error)
	DeleteVCL(context.Context, *fastly.DeleteVCLInput) error

	CreateSnippet(context.Context, *fastly.CreateSnippetInput) (*fastly.Snippet, error)
	ListSnippets(context.Context, *fastly.ListSnippetsInput) ([]*fastly.Snippet, error)
	GetSnippet(context.Context, *fastly.GetSnippetInput) (*fastly.Snippet, error)
	GetDynamicSnippet(context.Context, *fastly.GetDynamicSnippetInput) (*fastly.DynamicSnippet, error)
	UpdateSnippet(context.Context, *fastly.UpdateSnippetInput) (*fastly.Snippet, error)
	UpdateDynamicSnippet(context.Context, *fastly.UpdateDynamicSnippetInput) (*fastly.DynamicSnippet, error)
	DeleteSnippet(context.Context, *fastly.DeleteSnippetInput) error

	Purge(context.Context, *fastly.PurgeInput) (*fastly.Purge, error)
	PurgeKey(context.Context, *fastly.PurgeKeyInput) (*fastly.Purge, error)
	PurgeKeys(context.Context, *fastly.PurgeKeysInput) (map[string]string, error)
	PurgeAll(context.Context, *fastly.PurgeAllInput) (*fastly.Purge, error)

	CreateACL(context.Context, *fastly.CreateACLInput) (*fastly.ACL, error)
	DeleteACL(context.Context, *fastly.DeleteACLInput) error
	GetACL(context.Context, *fastly.GetACLInput) (*fastly.ACL, error)
	ListACLs(context.Context, *fastly.ListACLsInput) ([]*fastly.ACL, error)
	UpdateACL(context.Context, *fastly.UpdateACLInput) (*fastly.ACL, error)

	CreateACLEntry(context.Context, *fastly.CreateACLEntryInput) (*fastly.ACLEntry, error)
	DeleteACLEntry(context.Context, *fastly.DeleteACLEntryInput) error
	GetACLEntry(context.Context, *fastly.GetACLEntryInput) (*fastly.ACLEntry, error)
	GetACLEntries(context.Context, *fastly.GetACLEntriesInput) *fastly.ListPaginator[fastly.ACLEntry]
	ListACLEntries(context.Context, *fastly.ListACLEntriesInput) ([]*fastly.ACLEntry, error)
	UpdateACLEntry(context.Context, *fastly.UpdateACLEntryInput) (*fastly.ACLEntry, error)
	BatchModifyACLEntries(context.Context, *fastly.BatchModifyACLEntriesInput) error

	CreateNewRelic(context.Context, *fastly.CreateNewRelicInput) (*fastly.NewRelic, error)
	DeleteNewRelic(context.Context, *fastly.DeleteNewRelicInput) error
	GetNewRelic(context.Context, *fastly.GetNewRelicInput) (*fastly.NewRelic, error)
	ListNewRelic(context.Context, *fastly.ListNewRelicInput) ([]*fastly.NewRelic, error)
	UpdateNewRelic(context.Context, *fastly.UpdateNewRelicInput) (*fastly.NewRelic, error)

	CreateNewRelicOTLP(context.Context, *fastly.CreateNewRelicOTLPInput) (*fastly.NewRelicOTLP, error)
	DeleteNewRelicOTLP(context.Context, *fastly.DeleteNewRelicOTLPInput) error
	GetNewRelicOTLP(context.Context, *fastly.GetNewRelicOTLPInput) (*fastly.NewRelicOTLP, error)
	ListNewRelicOTLP(context.Context, *fastly.ListNewRelicOTLPInput) ([]*fastly.NewRelicOTLP, error)
	UpdateNewRelicOTLP(context.Context, *fastly.UpdateNewRelicOTLPInput) (*fastly.NewRelicOTLP, error)

	CreateUser(context.Context, *fastly.CreateUserInput) (*fastly.User, error)
	DeleteUser(context.Context, *fastly.DeleteUserInput) error
	GetCurrentUser(context.Context) (*fastly.User, error)
	GetUser(context.Context, *fastly.GetUserInput) (*fastly.User, error)
	ListCustomerUsers(context.Context, *fastly.ListCustomerUsersInput) ([]*fastly.User, error)
	UpdateUser(context.Context, *fastly.UpdateUserInput) (*fastly.User, error)
	ResetUserPassword(context.Context, *fastly.ResetUserPasswordInput) error

	BatchDeleteTokens(context.Context, *fastly.BatchDeleteTokensInput) error
	CreateToken(context.Context, *fastly.CreateTokenInput) (*fastly.Token, error)
	DeleteToken(context.Context, *fastly.DeleteTokenInput) error
	DeleteTokenSelf(context.Context) error
	GetTokenSelf(context.Context) (*fastly.Token, error)
	ListCustomerTokens(context.Context, *fastly.ListCustomerTokensInput) ([]*fastly.Token, error)
	ListTokens(context.Context, *fastly.ListTokensInput) ([]*fastly.Token, error)

	NewListKVStoreKeysPaginator(context.Context, *fastly.ListKVStoreKeysInput) fastly.PaginatorKVStoreEntries

	GetCustomTLSConfiguration(context.Context, *fastly.GetCustomTLSConfigurationInput) (*fastly.CustomTLSConfiguration, error)
	ListCustomTLSConfigurations(context.Context, *fastly.ListCustomTLSConfigurationsInput) ([]*fastly.CustomTLSConfiguration, error)
	UpdateCustomTLSConfiguration(context.Context, *fastly.UpdateCustomTLSConfigurationInput) (*fastly.CustomTLSConfiguration, error)
	GetTLSActivation(context.Context, *fastly.GetTLSActivationInput) (*fastly.TLSActivation, error)
	ListTLSActivations(context.Context, *fastly.ListTLSActivationsInput) ([]*fastly.TLSActivation, error)
	UpdateTLSActivation(context.Context, *fastly.UpdateTLSActivationInput) (*fastly.TLSActivation, error)
	CreateTLSActivation(context.Context, *fastly.CreateTLSActivationInput) (*fastly.TLSActivation, error)
	DeleteTLSActivation(context.Context, *fastly.DeleteTLSActivationInput) error

	CreateCustomTLSCertificate(context.Context, *fastly.CreateCustomTLSCertificateInput) (*fastly.CustomTLSCertificate, error)
	DeleteCustomTLSCertificate(context.Context, *fastly.DeleteCustomTLSCertificateInput) error
	GetCustomTLSCertificate(context.Context, *fastly.GetCustomTLSCertificateInput) (*fastly.CustomTLSCertificate, error)
	ListCustomTLSCertificates(context.Context, *fastly.ListCustomTLSCertificatesInput) ([]*fastly.CustomTLSCertificate, error)
	UpdateCustomTLSCertificate(context.Context, *fastly.UpdateCustomTLSCertificateInput) (*fastly.CustomTLSCertificate, error)

	ListTLSDomains(context.Context, *fastly.ListTLSDomainsInput) ([]*fastly.TLSDomain, error)

	CreatePrivateKey(context.Context, *fastly.CreatePrivateKeyInput) (*fastly.PrivateKey, error)
	DeletePrivateKey(context.Context, *fastly.DeletePrivateKeyInput) error
	GetPrivateKey(context.Context, *fastly.GetPrivateKeyInput) (*fastly.PrivateKey, error)
	ListPrivateKeys(context.Context, *fastly.ListPrivateKeysInput) ([]*fastly.PrivateKey, error)

	CreateBulkCertificate(context.Context, *fastly.CreateBulkCertificateInput) (*fastly.BulkCertificate, error)
	DeleteBulkCertificate(context.Context, *fastly.DeleteBulkCertificateInput) error
	GetBulkCertificate(context.Context, *fastly.GetBulkCertificateInput) (*fastly.BulkCertificate, error)
	ListBulkCertificates(context.Context, *fastly.ListBulkCertificatesInput) ([]*fastly.BulkCertificate, error)
	UpdateBulkCertificate(context.Context, *fastly.UpdateBulkCertificateInput) (*fastly.BulkCertificate, error)

	CreateTLSSubscription(context.Context, *fastly.CreateTLSSubscriptionInput) (*fastly.TLSSubscription, error)
	DeleteTLSSubscription(context.Context, *fastly.DeleteTLSSubscriptionInput) error
	GetTLSSubscription(context.Context, *fastly.GetTLSSubscriptionInput) (*fastly.TLSSubscription, error)
	ListTLSSubscriptions(context.Context, *fastly.ListTLSSubscriptionsInput) ([]*fastly.TLSSubscription, error)
	UpdateTLSSubscription(context.Context, *fastly.UpdateTLSSubscriptionInput) (*fastly.TLSSubscription, error)

	ListServiceAuthorizations(context.Context, *fastly.ListServiceAuthorizationsInput) (*fastly.ServiceAuthorizations, error)
	GetServiceAuthorization(context.Context, *fastly.GetServiceAuthorizationInput) (*fastly.ServiceAuthorization, error)
	CreateServiceAuthorization(context.Context, *fastly.CreateServiceAuthorizationInput) (*fastly.ServiceAuthorization, error)
	UpdateServiceAuthorization(context.Context, *fastly.UpdateServiceAuthorizationInput) (*fastly.ServiceAuthorization, error)
	DeleteServiceAuthorization(context.Context, *fastly.DeleteServiceAuthorizationInput) error

	CreateConfigStore(context.Context, *fastly.CreateConfigStoreInput) (*fastly.ConfigStore, error)
	DeleteConfigStore(context.Context, *fastly.DeleteConfigStoreInput) error
	GetConfigStore(context.Context, *fastly.GetConfigStoreInput) (*fastly.ConfigStore, error)
	GetConfigStoreMetadata(context.Context, *fastly.GetConfigStoreMetadataInput) (*fastly.ConfigStoreMetadata, error)
	ListConfigStores(context.Context, *fastly.ListConfigStoresInput) ([]*fastly.ConfigStore, error)
	ListConfigStoreServices(context.Context, *fastly.ListConfigStoreServicesInput) ([]*fastly.Service, error)
	UpdateConfigStore(context.Context, *fastly.UpdateConfigStoreInput) (*fastly.ConfigStore, error)

	CreateConfigStoreItem(context.Context, *fastly.CreateConfigStoreItemInput) (*fastly.ConfigStoreItem, error)
	DeleteConfigStoreItem(context.Context, *fastly.DeleteConfigStoreItemInput) error
	GetConfigStoreItem(context.Context, *fastly.GetConfigStoreItemInput) (*fastly.ConfigStoreItem, error)
	ListConfigStoreItems(context.Context, *fastly.ListConfigStoreItemsInput) ([]*fastly.ConfigStoreItem, error)
	UpdateConfigStoreItem(context.Context, *fastly.UpdateConfigStoreItemInput) (*fastly.ConfigStoreItem, error)

	CreateKVStore(context.Context, *fastly.CreateKVStoreInput) (*fastly.KVStore, error)
	ListKVStores(context.Context, *fastly.ListKVStoresInput) (*fastly.ListKVStoresResponse, error)
	DeleteKVStore(context.Context, *fastly.DeleteKVStoreInput) error
	GetKVStore(context.Context, *fastly.GetKVStoreInput) (*fastly.KVStore, error)
	ListKVStoreKeys(context.Context, *fastly.ListKVStoreKeysInput) (*fastly.ListKVStoreKeysResponse, error)
	GetKVStoreKey(context.Context, *fastly.GetKVStoreKeyInput) (string, error)
	DeleteKVStoreKey(context.Context, *fastly.DeleteKVStoreKeyInput) error
	InsertKVStoreKey(context.Context, *fastly.InsertKVStoreKeyInput) error
	BatchModifyKVStoreKey(context.Context, *fastly.BatchModifyKVStoreKeyInput) error

	CreateSecretStore(context.Context, *fastly.CreateSecretStoreInput) (*fastly.SecretStore, error)
	GetSecretStore(context.Context, *fastly.GetSecretStoreInput) (*fastly.SecretStore, error)
	DeleteSecretStore(context.Context, *fastly.DeleteSecretStoreInput) error
	ListSecretStores(context.Context, *fastly.ListSecretStoresInput) (*fastly.SecretStores, error)
	CreateSecret(context.Context, *fastly.CreateSecretInput) (*fastly.Secret, error)
	GetSecret(context.Context, *fastly.GetSecretInput) (*fastly.Secret, error)
	DeleteSecret(context.Context, *fastly.DeleteSecretInput) error
	ListSecrets(context.Context, *fastly.ListSecretsInput) (*fastly.Secrets, error)
	CreateClientKey(context.Context) (*fastly.ClientKey, error)
	GetSigningKey(context.Context) (ed25519.PublicKey, error)

	CreateResource(context.Context, *fastly.CreateResourceInput) (*fastly.Resource, error)
	DeleteResource(context.Context, *fastly.DeleteResourceInput) error
	GetResource(context.Context, *fastly.GetResourceInput) (*fastly.Resource, error)
	ListResources(context.Context, *fastly.ListResourcesInput) ([]*fastly.Resource, error)
	UpdateResource(context.Context, *fastly.UpdateResourceInput) (*fastly.Resource, error)

	CreateERL(context.Context, *fastly.CreateERLInput) (*fastly.ERL, error)
	DeleteERL(context.Context, *fastly.DeleteERLInput) error
	GetERL(context.Context, *fastly.GetERLInput) (*fastly.ERL, error)
	ListERLs(context.Context, *fastly.ListERLsInput) ([]*fastly.ERL, error)
	UpdateERL(context.Context, *fastly.UpdateERLInput) (*fastly.ERL, error)

	CreateCondition(context.Context, *fastly.CreateConditionInput) (*fastly.Condition, error)
	DeleteCondition(context.Context, *fastly.DeleteConditionInput) error
	GetCondition(context.Context, *fastly.GetConditionInput) (*fastly.Condition, error)
	ListConditions(context.Context, *fastly.ListConditionsInput) ([]*fastly.Condition, error)
	UpdateCondition(context.Context, *fastly.UpdateConditionInput) (*fastly.Condition, error)

	ListAlertDefinitions(context.Context, *fastly.ListAlertDefinitionsInput) (*fastly.AlertDefinitionsResponse, error)
	CreateAlertDefinition(context.Context, *fastly.CreateAlertDefinitionInput) (*fastly.AlertDefinition, error)
	GetAlertDefinition(context.Context, *fastly.GetAlertDefinitionInput) (*fastly.AlertDefinition, error)
	UpdateAlertDefinition(context.Context, *fastly.UpdateAlertDefinitionInput) (*fastly.AlertDefinition, error)
	DeleteAlertDefinition(context.Context, *fastly.DeleteAlertDefinitionInput) error
	TestAlertDefinition(context.Context, *fastly.TestAlertDefinitionInput) error
	ListAlertHistory(context.Context, *fastly.ListAlertHistoryInput) (*fastly.AlertHistoryResponse, error)

	ListObservabilityCustomDashboards(context.Context, *fastly.ListObservabilityCustomDashboardsInput) (*fastly.ListDashboardsResponse, error)
	CreateObservabilityCustomDashboard(context.Context, *fastly.CreateObservabilityCustomDashboardInput) (*fastly.ObservabilityCustomDashboard, error)
	GetObservabilityCustomDashboard(context.Context, *fastly.GetObservabilityCustomDashboardInput) (*fastly.ObservabilityCustomDashboard, error)
	UpdateObservabilityCustomDashboard(context.Context, *fastly.UpdateObservabilityCustomDashboardInput) (*fastly.ObservabilityCustomDashboard, error)
	DeleteObservabilityCustomDashboard(context.Context, *fastly.DeleteObservabilityCustomDashboardInput) error
}

// RealtimeStatsInterface is the subset of go-fastly's realtime stats API used here.
type RealtimeStatsInterface interface {
	GetRealtimeStatsJSON(context.Context, *fastly.GetRealtimeStatsInput, any) error
}

// Ensure that fastly.Client satisfies Interface.
var _ Interface = (*fastly.Client)(nil)

// Ensure that fastly.RTSClient satisfies RealtimeStatsInterface.
var _ RealtimeStatsInterface = (*fastly.RTSClient)(nil)
