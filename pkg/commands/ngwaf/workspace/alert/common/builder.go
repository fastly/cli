package common

// GetDefaultEvents returns the hardcoded events value for all alerts.
// Currently the only supported value is "flag".
func GetDefaultEvents() *[]string {
	events := []string{"flag"}
	return &events
}
