package providerconnection

// Supported provider connection authentication types.
const (
	authTypeAPIKey = "api-key"
	authTypeAWSIAM = "aws-iam"
	authTypeGCPIAM = "gcp-iam"
)

// Provider names required by the aws-iam and gcp-iam authentication types.
const (
	providerNameBedrock = "bedrock"
	providerNameVertex  = "vertex"
)

// authTypes is the list of supported provider connection authentication types.
var authTypes = []string{authTypeAPIKey, authTypeAWSIAM, authTypeGCPIAM}
