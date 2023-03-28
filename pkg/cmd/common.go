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
	FlagVersionDesc = "'latest', 'active', or the number of a specific Fastly service version"
)

// PaginationDirection is a list of directions the page results can be displayed.
var PaginationDirection = []string{"ascend", "descend"}

// CursorFlag returns a cursor flag definition.
func CursorFlag(dst *string) StringFlagOpts {
	return StringFlagOpts{
		Name:        "cursor",
		Short:       'c',
		Description: "Pagination cursor (Use 'next_cursor' value from list output)",
		Dst:         dst,
	}
}

// LimitFlag returns a limit flag definition.
func LimitFlag(dst *int) IntFlagOpts {
	return IntFlagOpts{
		Name:        "limit",
		Short:       'l',
		Description: "Maximum number of items to list",
		Default:     50,
		Dst:         dst,
	}
}

// StoreIDFlag returns a store-id flag definition.
func StoreIDFlag(dst *string) StringFlagOpts {
	return StringFlagOpts{
		Name:        "store-id",
		Short:       's',
		Description: "Store ID",
		Dst:         dst,
		Required:    true,
	}
}
