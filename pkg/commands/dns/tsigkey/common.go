package tsigkey

// SortOptions are the valid values for the --sort flag.
// Since Kingpin interprets any '-' as a flag, we'll need to remap the API enums
// to more friendly values and then send back to the API accordingly.
var SortOptions = []string{"name_asc", "name_desc", "created_at_asc", "created_at_desc"}

// sortAPIValue maps CLI sort values to the API sort parameter.
var sortAPIValue = map[string]string{
	"name_asc":        "name",
	"name_desc":       "-name",
	"created_at_asc":  "created_at",
	"created_at_desc": "-created_at",
}

// AlgorithmOptions are the valid values for the --algorithm flag.
var AlgorithmOptions = []string{"hmac-sha224", "hmac-sha256", "hmac-sha384", "hmac-sha512"}
