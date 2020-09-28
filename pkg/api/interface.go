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
	SearchService(*fastly.SearchServiceInput) (*fastly.Service, error)

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

	GetPackage(*fastly.GetPackageInput) (*fastly.Package, error)
	UpdatePackage(*fastly.UpdatePackageInput) (*fastly.Package, error)

	CreateDictionary(*fastly.CreateDictionaryInput) (*fastly.Dictionary, error)
	GetDictionary(*fastly.GetDictionaryInput) (*fastly.Dictionary, error)

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

	GetUser(*fastly.GetUserInput) (*fastly.User, error)

	GetRegions() (*fastly.RegionsResponse, error)
	GetStatsJSON(*fastly.GetStatsInput, interface{}) error
}

// RealtimeStatsInterface is the subset of go-fastly's realtime stats API used here.
type RealtimeStatsInterface interface {
	GetRealtimeStatsJSON(*fastly.GetRealtimeStatsInput, interface{}) error
}

// Ensure that fastly.Client satisfies Interface.
var _ Interface = (*fastly.Client)(nil)

// Ensure that fastly.RTSClient satisfies RealtimeStatsInterface.
var _ RealtimeStatsInterface = (*fastly.RTSClient)(nil)
