package cmd

var (
	FlagCustomerIDName = "customer-id"
	FlagCustomerIDDesc = "Alphanumeric string identifying the customer (falls back to FASTLY_CUSTOMER_ID)"
	FlagJSONName       = "json"
	FlagJSONDesc       = "Render output as JSON"
	FlagServiceIDName  = "service-id"
	FlagServiceIDDesc  = "Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"
	FlagServiceName    = "service-name"
	FlagServiceDesc    = "The name of the service"
	FlagVersionName    = "version"
	FlagVersionDesc    = "'latest', 'active', or the number of a specific version"
)

var PaginationDirection = []string{"ascend", "descend"}
