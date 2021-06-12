package env

const (
	// Token is the env var we look in for the Fastly API token.
	// gosec flagged this:
	// G101 (CWE-798): Potential hardcoded credentials
	// Disabling as we use the value in the command help output.
	/* #nosec */
	Token = "FASTLY_API_TOKEN"

	// Endpoint is the env var we look in for the API endpoint.
	Endpoint = "FASTLY_API_ENDPOINT"

	// ServiceID is a relatively ubiquitous parameter which can be tedious to
	// specify, so we also look in this env var for a value.
	ServiceID = "FASTLY_SERVICE_ID"
)
