package cmd

var (
	// FlagCustomerIDName is the flag name.
	FlagCustomerIDName = "customer-id"
	// FlagCustomerIDDesc is the flag description.
	FlagCustomerIDDesc = "Alphanumeric string identifying the customer (falls back to FASTLY_CUSTOMER_ID)"
	// FlagJSONName is the flag name.
	FlagJSONName = "json"
	// FlagJSONDesc is the flag description.
	FlagJSONDesc = "Render output as JSON"
	// FlagServiceIDName is the flag name.
	FlagServiceIDName = "service-id"
	// FlagServiceIDDesc is the flag description.
	FlagServiceIDDesc = "Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"
	// FlagServiceName is the flag name.
	FlagServiceName = "service-name"
	// FlagServiceDesc is the flag description.
	FlagServiceDesc = "The name of the service"
	// FlagVersionName is the flag name.
	FlagVersionName = "version"
	// FlagVersionDesc is the flag description.
	FlagVersionDesc = "'latest', 'active', or the number of a specific version"
)

// PaginationDirection is a list of directions the page results can be displayed.
var PaginationDirection = []string{"ascend", "descend"}
