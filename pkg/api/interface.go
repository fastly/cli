package api

import (
	"crypto/ed25519"
	"net/http"

	"github.com/fastly/go-fastly/v8/fastly"
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
	AllIPs() (v4, v6 fastly.IPAddrs, err error)
	AllDatacenters() (datacenters []fastly.Datacenter, err error)

	CreateService(*fastly.CreateServiceInput) (*fastly.Service, error)
	GetServices(*fastly.GetServicesInput) *fastly.ListPaginator[fastly.Service]
	ListServices(*fastly.ListServicesInput) ([]*fastly.Service, error)
	GetService(*fastly.GetServiceInput) (*fastly.Service, error)
	GetServiceDetails(*fastly.GetServiceInput) (*fastly.ServiceDetail, error)
	UpdateService(*fastly.UpdateServiceInput) (*fastly.Service, error)
	DeleteService(*fastly.DeleteServiceInput) error
	SearchService(*fastly.SearchServiceInput) (*fastly.Service, error)

	CloneVersion(*fastly.CloneVersionInput) (*fastly.Version, error)
	ListVersions(*fastly.ListVersionsInput) ([]*fastly.Version, error)
	GetVersion(*fastly.GetVersionInput) (*fastly.Version, error)
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
	ValidateDomain(i *fastly.ValidateDomainInput) (*fastly.DomainValidationResult, error)
	ValidateAllDomains(i *fastly.ValidateAllDomainsInput) (results []*fastly.DomainValidationResult, err error)

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

	GetPackage(*fastly.GetPackageInput) (*fastly.Package, error)
	UpdatePackage(*fastly.UpdatePackageInput) (*fastly.Package, error)

	CreateDictionary(*fastly.CreateDictionaryInput) (*fastly.Dictionary, error)
	GetDictionary(*fastly.GetDictionaryInput) (*fastly.Dictionary, error)
	DeleteDictionary(*fastly.DeleteDictionaryInput) error
	ListDictionaries(*fastly.ListDictionariesInput) ([]*fastly.Dictionary, error)
	UpdateDictionary(*fastly.UpdateDictionaryInput) (*fastly.Dictionary, error)

	GetDictionaryItems(*fastly.GetDictionaryItemsInput) *fastly.ListPaginator[fastly.DictionaryItem]
	ListDictionaryItems(*fastly.ListDictionaryItemsInput) ([]*fastly.DictionaryItem, error)
	GetDictionaryItem(*fastly.GetDictionaryItemInput) (*fastly.DictionaryItem, error)
	CreateDictionaryItem(*fastly.CreateDictionaryItemInput) (*fastly.DictionaryItem, error)
	UpdateDictionaryItem(*fastly.UpdateDictionaryItemInput) (*fastly.DictionaryItem, error)
	DeleteDictionaryItem(*fastly.DeleteDictionaryItemInput) error
	BatchModifyDictionaryItems(*fastly.BatchModifyDictionaryItemsInput) error

	GetDictionaryInfo(*fastly.GetDictionaryInfoInput) (*fastly.DictionaryInfo, error)

	CreateBigQuery(*fastly.CreateBigQueryInput) (*fastly.BigQuery, error)
	ListBigQueries(*fastly.ListBigQueriesInput) ([]*fastly.BigQuery, error)
	GetBigQuery(*fastly.GetBigQueryInput) (*fastly.BigQuery, error)
	UpdateBigQuery(*fastly.UpdateBigQueryInput) (*fastly.BigQuery, error)
	DeleteBigQuery(*fastly.DeleteBigQueryInput) error

	CreateS3(*fastly.CreateS3Input) (*fastly.S3, error)
	ListS3s(*fastly.ListS3sInput) ([]*fastly.S3, error)
	GetS3(*fastly.GetS3Input) (*fastly.S3, error)
	UpdateS3(*fastly.UpdateS3Input) (*fastly.S3, error)
	DeleteS3(*fastly.DeleteS3Input) error

	CreateKinesis(*fastly.CreateKinesisInput) (*fastly.Kinesis, error)
	ListKinesis(*fastly.ListKinesisInput) ([]*fastly.Kinesis, error)
	GetKinesis(*fastly.GetKinesisInput) (*fastly.Kinesis, error)
	UpdateKinesis(*fastly.UpdateKinesisInput) (*fastly.Kinesis, error)
	DeleteKinesis(*fastly.DeleteKinesisInput) error

	CreateSyslog(*fastly.CreateSyslogInput) (*fastly.Syslog, error)
	ListSyslogs(*fastly.ListSyslogsInput) ([]*fastly.Syslog, error)
	GetSyslog(*fastly.GetSyslogInput) (*fastly.Syslog, error)
	UpdateSyslog(*fastly.UpdateSyslogInput) (*fastly.Syslog, error)
	DeleteSyslog(*fastly.DeleteSyslogInput) error

	CreateLogentries(*fastly.CreateLogentriesInput) (*fastly.Logentries, error)
	ListLogentries(*fastly.ListLogentriesInput) ([]*fastly.Logentries, error)
	GetLogentries(*fastly.GetLogentriesInput) (*fastly.Logentries, error)
	UpdateLogentries(*fastly.UpdateLogentriesInput) (*fastly.Logentries, error)
	DeleteLogentries(*fastly.DeleteLogentriesInput) error

	CreatePapertrail(*fastly.CreatePapertrailInput) (*fastly.Papertrail, error)
	ListPapertrails(*fastly.ListPapertrailsInput) ([]*fastly.Papertrail, error)
	GetPapertrail(*fastly.GetPapertrailInput) (*fastly.Papertrail, error)
	UpdatePapertrail(*fastly.UpdatePapertrailInput) (*fastly.Papertrail, error)
	DeletePapertrail(*fastly.DeletePapertrailInput) error

	CreateSumologic(*fastly.CreateSumologicInput) (*fastly.Sumologic, error)
	ListSumologics(*fastly.ListSumologicsInput) ([]*fastly.Sumologic, error)
	GetSumologic(*fastly.GetSumologicInput) (*fastly.Sumologic, error)
	UpdateSumologic(*fastly.UpdateSumologicInput) (*fastly.Sumologic, error)
	DeleteSumologic(*fastly.DeleteSumologicInput) error

	CreateGCS(*fastly.CreateGCSInput) (*fastly.GCS, error)
	ListGCSs(*fastly.ListGCSsInput) ([]*fastly.GCS, error)
	GetGCS(*fastly.GetGCSInput) (*fastly.GCS, error)
	UpdateGCS(*fastly.UpdateGCSInput) (*fastly.GCS, error)
	DeleteGCS(*fastly.DeleteGCSInput) error

	CreateFTP(*fastly.CreateFTPInput) (*fastly.FTP, error)
	ListFTPs(*fastly.ListFTPsInput) ([]*fastly.FTP, error)
	GetFTP(*fastly.GetFTPInput) (*fastly.FTP, error)
	UpdateFTP(*fastly.UpdateFTPInput) (*fastly.FTP, error)
	DeleteFTP(*fastly.DeleteFTPInput) error

	CreateSplunk(*fastly.CreateSplunkInput) (*fastly.Splunk, error)
	ListSplunks(*fastly.ListSplunksInput) ([]*fastly.Splunk, error)
	GetSplunk(*fastly.GetSplunkInput) (*fastly.Splunk, error)
	UpdateSplunk(*fastly.UpdateSplunkInput) (*fastly.Splunk, error)
	DeleteSplunk(*fastly.DeleteSplunkInput) error

	CreateScalyr(*fastly.CreateScalyrInput) (*fastly.Scalyr, error)
	ListScalyrs(*fastly.ListScalyrsInput) ([]*fastly.Scalyr, error)
	GetScalyr(*fastly.GetScalyrInput) (*fastly.Scalyr, error)
	UpdateScalyr(*fastly.UpdateScalyrInput) (*fastly.Scalyr, error)
	DeleteScalyr(*fastly.DeleteScalyrInput) error

	CreateLoggly(*fastly.CreateLogglyInput) (*fastly.Loggly, error)
	ListLoggly(*fastly.ListLogglyInput) ([]*fastly.Loggly, error)
	GetLoggly(*fastly.GetLogglyInput) (*fastly.Loggly, error)
	UpdateLoggly(*fastly.UpdateLogglyInput) (*fastly.Loggly, error)
	DeleteLoggly(*fastly.DeleteLogglyInput) error

	CreateHoneycomb(*fastly.CreateHoneycombInput) (*fastly.Honeycomb, error)
	ListHoneycombs(*fastly.ListHoneycombsInput) ([]*fastly.Honeycomb, error)
	GetHoneycomb(*fastly.GetHoneycombInput) (*fastly.Honeycomb, error)
	UpdateHoneycomb(*fastly.UpdateHoneycombInput) (*fastly.Honeycomb, error)
	DeleteHoneycomb(*fastly.DeleteHoneycombInput) error

	CreateHeroku(*fastly.CreateHerokuInput) (*fastly.Heroku, error)
	ListHerokus(*fastly.ListHerokusInput) ([]*fastly.Heroku, error)
	GetHeroku(*fastly.GetHerokuInput) (*fastly.Heroku, error)
	UpdateHeroku(*fastly.UpdateHerokuInput) (*fastly.Heroku, error)
	DeleteHeroku(*fastly.DeleteHerokuInput) error

	CreateSFTP(*fastly.CreateSFTPInput) (*fastly.SFTP, error)
	ListSFTPs(*fastly.ListSFTPsInput) ([]*fastly.SFTP, error)
	GetSFTP(*fastly.GetSFTPInput) (*fastly.SFTP, error)
	UpdateSFTP(*fastly.UpdateSFTPInput) (*fastly.SFTP, error)
	DeleteSFTP(*fastly.DeleteSFTPInput) error

	CreateLogshuttle(*fastly.CreateLogshuttleInput) (*fastly.Logshuttle, error)
	ListLogshuttles(*fastly.ListLogshuttlesInput) ([]*fastly.Logshuttle, error)
	GetLogshuttle(*fastly.GetLogshuttleInput) (*fastly.Logshuttle, error)
	UpdateLogshuttle(*fastly.UpdateLogshuttleInput) (*fastly.Logshuttle, error)
	DeleteLogshuttle(*fastly.DeleteLogshuttleInput) error

	CreateCloudfiles(*fastly.CreateCloudfilesInput) (*fastly.Cloudfiles, error)
	ListCloudfiles(*fastly.ListCloudfilesInput) ([]*fastly.Cloudfiles, error)
	GetCloudfiles(*fastly.GetCloudfilesInput) (*fastly.Cloudfiles, error)
	UpdateCloudfiles(*fastly.UpdateCloudfilesInput) (*fastly.Cloudfiles, error)
	DeleteCloudfiles(*fastly.DeleteCloudfilesInput) error

	CreateDigitalOcean(*fastly.CreateDigitalOceanInput) (*fastly.DigitalOcean, error)
	ListDigitalOceans(*fastly.ListDigitalOceansInput) ([]*fastly.DigitalOcean, error)
	GetDigitalOcean(*fastly.GetDigitalOceanInput) (*fastly.DigitalOcean, error)
	UpdateDigitalOcean(*fastly.UpdateDigitalOceanInput) (*fastly.DigitalOcean, error)
	DeleteDigitalOcean(*fastly.DeleteDigitalOceanInput) error

	CreateElasticsearch(*fastly.CreateElasticsearchInput) (*fastly.Elasticsearch, error)
	ListElasticsearch(*fastly.ListElasticsearchInput) ([]*fastly.Elasticsearch, error)
	GetElasticsearch(*fastly.GetElasticsearchInput) (*fastly.Elasticsearch, error)
	UpdateElasticsearch(*fastly.UpdateElasticsearchInput) (*fastly.Elasticsearch, error)
	DeleteElasticsearch(*fastly.DeleteElasticsearchInput) error

	CreateBlobStorage(*fastly.CreateBlobStorageInput) (*fastly.BlobStorage, error)
	ListBlobStorages(*fastly.ListBlobStoragesInput) ([]*fastly.BlobStorage, error)
	GetBlobStorage(*fastly.GetBlobStorageInput) (*fastly.BlobStorage, error)
	UpdateBlobStorage(*fastly.UpdateBlobStorageInput) (*fastly.BlobStorage, error)
	DeleteBlobStorage(*fastly.DeleteBlobStorageInput) error

	CreateDatadog(*fastly.CreateDatadogInput) (*fastly.Datadog, error)
	ListDatadog(*fastly.ListDatadogInput) ([]*fastly.Datadog, error)
	GetDatadog(*fastly.GetDatadogInput) (*fastly.Datadog, error)
	UpdateDatadog(*fastly.UpdateDatadogInput) (*fastly.Datadog, error)
	DeleteDatadog(*fastly.DeleteDatadogInput) error

	CreateHTTPS(*fastly.CreateHTTPSInput) (*fastly.HTTPS, error)
	ListHTTPS(*fastly.ListHTTPSInput) ([]*fastly.HTTPS, error)
	GetHTTPS(*fastly.GetHTTPSInput) (*fastly.HTTPS, error)
	UpdateHTTPS(*fastly.UpdateHTTPSInput) (*fastly.HTTPS, error)
	DeleteHTTPS(*fastly.DeleteHTTPSInput) error

	CreateKafka(*fastly.CreateKafkaInput) (*fastly.Kafka, error)
	ListKafkas(*fastly.ListKafkasInput) ([]*fastly.Kafka, error)
	GetKafka(*fastly.GetKafkaInput) (*fastly.Kafka, error)
	UpdateKafka(*fastly.UpdateKafkaInput) (*fastly.Kafka, error)
	DeleteKafka(*fastly.DeleteKafkaInput) error

	CreatePubsub(*fastly.CreatePubsubInput) (*fastly.Pubsub, error)
	ListPubsubs(*fastly.ListPubsubsInput) ([]*fastly.Pubsub, error)
	GetPubsub(*fastly.GetPubsubInput) (*fastly.Pubsub, error)
	UpdatePubsub(*fastly.UpdatePubsubInput) (*fastly.Pubsub, error)
	DeletePubsub(*fastly.DeletePubsubInput) error

	CreateOpenstack(*fastly.CreateOpenstackInput) (*fastly.Openstack, error)
	ListOpenstack(*fastly.ListOpenstackInput) ([]*fastly.Openstack, error)
	GetOpenstack(*fastly.GetOpenstackInput) (*fastly.Openstack, error)
	UpdateOpenstack(*fastly.UpdateOpenstackInput) (*fastly.Openstack, error)
	DeleteOpenstack(*fastly.DeleteOpenstackInput) error

	GetRegions() (*fastly.RegionsResponse, error)
	GetStatsJSON(*fastly.GetStatsInput, any) error

	CreateManagedLogging(*fastly.CreateManagedLoggingInput) (*fastly.ManagedLogging, error)

	CreateVCL(*fastly.CreateVCLInput) (*fastly.VCL, error)
	ListVCLs(*fastly.ListVCLsInput) ([]*fastly.VCL, error)
	GetVCL(*fastly.GetVCLInput) (*fastly.VCL, error)
	UpdateVCL(*fastly.UpdateVCLInput) (*fastly.VCL, error)
	DeleteVCL(*fastly.DeleteVCLInput) error

	CreateSnippet(i *fastly.CreateSnippetInput) (*fastly.Snippet, error)
	ListSnippets(i *fastly.ListSnippetsInput) ([]*fastly.Snippet, error)
	GetSnippet(i *fastly.GetSnippetInput) (*fastly.Snippet, error)
	GetDynamicSnippet(i *fastly.GetDynamicSnippetInput) (*fastly.DynamicSnippet, error)
	UpdateSnippet(i *fastly.UpdateSnippetInput) (*fastly.Snippet, error)
	UpdateDynamicSnippet(i *fastly.UpdateDynamicSnippetInput) (*fastly.DynamicSnippet, error)
	DeleteSnippet(i *fastly.DeleteSnippetInput) error

	Purge(i *fastly.PurgeInput) (*fastly.Purge, error)
	PurgeKey(i *fastly.PurgeKeyInput) (*fastly.Purge, error)
	PurgeKeys(i *fastly.PurgeKeysInput) (map[string]string, error)
	PurgeAll(i *fastly.PurgeAllInput) (*fastly.Purge, error)

	CreateACL(i *fastly.CreateACLInput) (*fastly.ACL, error)
	DeleteACL(i *fastly.DeleteACLInput) error
	GetACL(i *fastly.GetACLInput) (*fastly.ACL, error)
	ListACLs(i *fastly.ListACLsInput) ([]*fastly.ACL, error)
	UpdateACL(i *fastly.UpdateACLInput) (*fastly.ACL, error)

	CreateACLEntry(i *fastly.CreateACLEntryInput) (*fastly.ACLEntry, error)
	DeleteACLEntry(i *fastly.DeleteACLEntryInput) error
	GetACLEntry(i *fastly.GetACLEntryInput) (*fastly.ACLEntry, error)
	GetACLEntries(*fastly.GetACLEntriesInput) *fastly.ListPaginator[fastly.ACLEntry]
	ListACLEntries(i *fastly.ListACLEntriesInput) ([]*fastly.ACLEntry, error)
	UpdateACLEntry(i *fastly.UpdateACLEntryInput) (*fastly.ACLEntry, error)
	BatchModifyACLEntries(i *fastly.BatchModifyACLEntriesInput) error

	CreateNewRelic(i *fastly.CreateNewRelicInput) (*fastly.NewRelic, error)
	DeleteNewRelic(i *fastly.DeleteNewRelicInput) error
	GetNewRelic(i *fastly.GetNewRelicInput) (*fastly.NewRelic, error)
	ListNewRelic(i *fastly.ListNewRelicInput) ([]*fastly.NewRelic, error)
	UpdateNewRelic(i *fastly.UpdateNewRelicInput) (*fastly.NewRelic, error)

	CreateNewRelicOTLP(i *fastly.CreateNewRelicOTLPInput) (*fastly.NewRelicOTLP, error)
	DeleteNewRelicOTLP(i *fastly.DeleteNewRelicOTLPInput) error
	GetNewRelicOTLP(i *fastly.GetNewRelicOTLPInput) (*fastly.NewRelicOTLP, error)
	ListNewRelicOTLP(i *fastly.ListNewRelicOTLPInput) ([]*fastly.NewRelicOTLP, error)
	UpdateNewRelicOTLP(i *fastly.UpdateNewRelicOTLPInput) (*fastly.NewRelicOTLP, error)

	CreateUser(i *fastly.CreateUserInput) (*fastly.User, error)
	DeleteUser(i *fastly.DeleteUserInput) error
	GetCurrentUser() (*fastly.User, error)
	GetUser(i *fastly.GetUserInput) (*fastly.User, error)
	ListCustomerUsers(i *fastly.ListCustomerUsersInput) ([]*fastly.User, error)
	UpdateUser(i *fastly.UpdateUserInput) (*fastly.User, error)
	ResetUserPassword(i *fastly.ResetUserPasswordInput) error

	BatchDeleteTokens(i *fastly.BatchDeleteTokensInput) error
	CreateToken(i *fastly.CreateTokenInput) (*fastly.Token, error)
	DeleteToken(i *fastly.DeleteTokenInput) error
	DeleteTokenSelf() error
	GetTokenSelf() (*fastly.Token, error)
	ListCustomerTokens(i *fastly.ListCustomerTokensInput) ([]*fastly.Token, error)
	ListTokens(i *fastly.ListTokensInput) ([]*fastly.Token, error)

	NewListKVStoreKeysPaginator(i *fastly.ListKVStoreKeysInput) fastly.PaginatorKVStoreEntries

	GetCustomTLSConfiguration(i *fastly.GetCustomTLSConfigurationInput) (*fastly.CustomTLSConfiguration, error)
	ListCustomTLSConfigurations(i *fastly.ListCustomTLSConfigurationsInput) ([]*fastly.CustomTLSConfiguration, error)
	UpdateCustomTLSConfiguration(i *fastly.UpdateCustomTLSConfigurationInput) (*fastly.CustomTLSConfiguration, error)
	GetTLSActivation(i *fastly.GetTLSActivationInput) (*fastly.TLSActivation, error)
	ListTLSActivations(i *fastly.ListTLSActivationsInput) ([]*fastly.TLSActivation, error)
	UpdateTLSActivation(i *fastly.UpdateTLSActivationInput) (*fastly.TLSActivation, error)
	CreateTLSActivation(i *fastly.CreateTLSActivationInput) (*fastly.TLSActivation, error)
	DeleteTLSActivation(i *fastly.DeleteTLSActivationInput) error

	CreateCustomTLSCertificate(i *fastly.CreateCustomTLSCertificateInput) (*fastly.CustomTLSCertificate, error)
	DeleteCustomTLSCertificate(i *fastly.DeleteCustomTLSCertificateInput) error
	GetCustomTLSCertificate(i *fastly.GetCustomTLSCertificateInput) (*fastly.CustomTLSCertificate, error)
	ListCustomTLSCertificates(i *fastly.ListCustomTLSCertificatesInput) ([]*fastly.CustomTLSCertificate, error)
	UpdateCustomTLSCertificate(i *fastly.UpdateCustomTLSCertificateInput) (*fastly.CustomTLSCertificate, error)

	ListTLSDomains(i *fastly.ListTLSDomainsInput) ([]*fastly.TLSDomain, error)

	CreatePrivateKey(i *fastly.CreatePrivateKeyInput) (*fastly.PrivateKey, error)
	DeletePrivateKey(i *fastly.DeletePrivateKeyInput) error
	GetPrivateKey(i *fastly.GetPrivateKeyInput) (*fastly.PrivateKey, error)
	ListPrivateKeys(i *fastly.ListPrivateKeysInput) ([]*fastly.PrivateKey, error)

	CreateBulkCertificate(i *fastly.CreateBulkCertificateInput) (*fastly.BulkCertificate, error)
	DeleteBulkCertificate(i *fastly.DeleteBulkCertificateInput) error
	GetBulkCertificate(i *fastly.GetBulkCertificateInput) (*fastly.BulkCertificate, error)
	ListBulkCertificates(i *fastly.ListBulkCertificatesInput) ([]*fastly.BulkCertificate, error)
	UpdateBulkCertificate(i *fastly.UpdateBulkCertificateInput) (*fastly.BulkCertificate, error)

	CreateTLSSubscription(i *fastly.CreateTLSSubscriptionInput) (*fastly.TLSSubscription, error)
	DeleteTLSSubscription(i *fastly.DeleteTLSSubscriptionInput) error
	GetTLSSubscription(i *fastly.GetTLSSubscriptionInput) (*fastly.TLSSubscription, error)
	ListTLSSubscriptions(i *fastly.ListTLSSubscriptionsInput) ([]*fastly.TLSSubscription, error)
	UpdateTLSSubscription(i *fastly.UpdateTLSSubscriptionInput) (*fastly.TLSSubscription, error)

	ListServiceAuthorizations(i *fastly.ListServiceAuthorizationsInput) (*fastly.ServiceAuthorizations, error)
	GetServiceAuthorization(i *fastly.GetServiceAuthorizationInput) (*fastly.ServiceAuthorization, error)
	CreateServiceAuthorization(i *fastly.CreateServiceAuthorizationInput) (*fastly.ServiceAuthorization, error)
	UpdateServiceAuthorization(i *fastly.UpdateServiceAuthorizationInput) (*fastly.ServiceAuthorization, error)
	DeleteServiceAuthorization(i *fastly.DeleteServiceAuthorizationInput) error

	CreateConfigStore(i *fastly.CreateConfigStoreInput) (*fastly.ConfigStore, error)
	DeleteConfigStore(i *fastly.DeleteConfigStoreInput) error
	GetConfigStore(i *fastly.GetConfigStoreInput) (*fastly.ConfigStore, error)
	GetConfigStoreMetadata(i *fastly.GetConfigStoreMetadataInput) (*fastly.ConfigStoreMetadata, error)
	ListConfigStores(i *fastly.ListConfigStoresInput) ([]*fastly.ConfigStore, error)
	ListConfigStoreServices(i *fastly.ListConfigStoreServicesInput) ([]*fastly.Service, error)
	UpdateConfigStore(i *fastly.UpdateConfigStoreInput) (*fastly.ConfigStore, error)

	CreateConfigStoreItem(i *fastly.CreateConfigStoreItemInput) (*fastly.ConfigStoreItem, error)
	DeleteConfigStoreItem(i *fastly.DeleteConfigStoreItemInput) error
	GetConfigStoreItem(i *fastly.GetConfigStoreItemInput) (*fastly.ConfigStoreItem, error)
	ListConfigStoreItems(i *fastly.ListConfigStoreItemsInput) ([]*fastly.ConfigStoreItem, error)
	UpdateConfigStoreItem(i *fastly.UpdateConfigStoreItemInput) (*fastly.ConfigStoreItem, error)

	CreateKVStore(i *fastly.CreateKVStoreInput) (*fastly.KVStore, error)
	ListKVStores(i *fastly.ListKVStoresInput) (*fastly.ListKVStoresResponse, error)
	DeleteKVStore(i *fastly.DeleteKVStoreInput) error
	GetKVStore(i *fastly.GetKVStoreInput) (*fastly.KVStore, error)
	ListKVStoreKeys(i *fastly.ListKVStoreKeysInput) (*fastly.ListKVStoreKeysResponse, error)
	GetKVStoreKey(i *fastly.GetKVStoreKeyInput) (string, error)
	DeleteKVStoreKey(i *fastly.DeleteKVStoreKeyInput) error
	InsertKVStoreKey(i *fastly.InsertKVStoreKeyInput) error
	BatchModifyKVStoreKey(i *fastly.BatchModifyKVStoreKeyInput) error

	CreateSecretStore(i *fastly.CreateSecretStoreInput) (*fastly.SecretStore, error)
	GetSecretStore(i *fastly.GetSecretStoreInput) (*fastly.SecretStore, error)
	DeleteSecretStore(i *fastly.DeleteSecretStoreInput) error
	ListSecretStores(i *fastly.ListSecretStoresInput) (*fastly.SecretStores, error)
	CreateSecret(i *fastly.CreateSecretInput) (*fastly.Secret, error)
	GetSecret(i *fastly.GetSecretInput) (*fastly.Secret, error)
	DeleteSecret(i *fastly.DeleteSecretInput) error
	ListSecrets(i *fastly.ListSecretsInput) (*fastly.Secrets, error)
	CreateClientKey() (*fastly.ClientKey, error)
	GetSigningKey() (ed25519.PublicKey, error)

	CreateResource(i *fastly.CreateResourceInput) (*fastly.Resource, error)
	DeleteResource(i *fastly.DeleteResourceInput) error
	GetResource(i *fastly.GetResourceInput) (*fastly.Resource, error)
	ListResources(i *fastly.ListResourcesInput) ([]*fastly.Resource, error)
	UpdateResource(i *fastly.UpdateResourceInput) (*fastly.Resource, error)

	CreateERL(i *fastly.CreateERLInput) (*fastly.ERL, error)
	DeleteERL(i *fastly.DeleteERLInput) error
	GetERL(i *fastly.GetERLInput) (*fastly.ERL, error)
	ListERLs(i *fastly.ListERLsInput) ([]*fastly.ERL, error)
	UpdateERL(i *fastly.UpdateERLInput) (*fastly.ERL, error)

	CreateCondition(i *fastly.CreateConditionInput) (*fastly.Condition, error)
	DeleteCondition(i *fastly.DeleteConditionInput) error
	GetCondition(i *fastly.GetConditionInput) (*fastly.Condition, error)
	ListConditions(i *fastly.ListConditionsInput) ([]*fastly.Condition, error)
	UpdateCondition(i *fastly.UpdateConditionInput) (*fastly.Condition, error)

	GetProduct(i *fastly.ProductEnablementInput) (*fastly.ProductEnablement, error)
	EnableProduct(i *fastly.ProductEnablementInput) (*fastly.ProductEnablement, error)
	DisableProduct(i *fastly.ProductEnablementInput) error
}

// RealtimeStatsInterface is the subset of go-fastly's realtime stats API used here.
type RealtimeStatsInterface interface {
	GetRealtimeStatsJSON(*fastly.GetRealtimeStatsInput, any) error
}

// Ensure that fastly.Client satisfies Interface.
var _ Interface = (*fastly.Client)(nil)

// Ensure that fastly.RTSClient satisfies RealtimeStatsInterface.
var _ RealtimeStatsInterface = (*fastly.RTSClient)(nil)
