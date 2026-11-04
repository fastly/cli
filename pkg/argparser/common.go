package argparser

var (
	// FlagCustomerIDName is the flag name.
	FlagCustomerIDName = "customer-id"
	// FlagCustomerIDDesc is the flag description.
	FlagCustomerIDDesc = "Alphanumeric string identifying the customer (falls back to FASTLY_CUSTOMER_ID)"
	// FlagJSONName is the flag name.
	FlagJSONName = "json"
	// FlagJSONDesc is the flag description.
	FlagJSONDesc = "Render output as JSON"
	// FlagNGWAFAlertID is the alert ID.
	FlagNGWAFAlertID = "alert-id"
	// FlagNGWAFAlertIDDesc is the alert ID flag description.
	FlagNGWAFAlertIDDesc = "Alphanumeric string identifying the alert"
	// FlagNGWAFWorkspaceID is the workspace ID.
	FlagNGWAFWorkspaceID = "workspace-id"
	// FlagNGWAFWorkspaceIDDesc is the workspace ID flag description.
	FlagNGWAFWorkspaceIDDesc = "Alphanumeric string identifying the NGWAF Workspace (falls back to FASTLY_WORKSPACE_ID)"
	// FlagServiceIDName is the flag name.
	FlagServiceIDName = "service-id"
	// FlagServiceIDDesc is the flag description.
	FlagServiceIDDesc = "Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"
	// FlagServiceName is the flag name.
	FlagServiceName = "service-name"
	// FlagServiceNameDesc is the flag description.
	FlagServiceNameDesc = "The name of the service"
	// FlagVersionName is the flag name.
	FlagVersionName = "version"
	// FlagVersionDesc is the flag description.
	FlagVersionDesc = "'latest', 'active', 'staged', or the number of a specific Fastly service version"
	// FlagVersionOptionalDesc is the flag description for commands where the
	// version flag is optional. It describes what OptionalServiceVersion.Parse
	// does when the flag is omitted.
	FlagVersionOptionalDesc = FlagVersionDesc + " (default: the active version, or the latest version if none is active)"
)

// PaginationDirection is a list of directions the page results can be displayed.
var PaginationDirection = []string{"ascend", "descend"}
