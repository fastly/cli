package zone

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
