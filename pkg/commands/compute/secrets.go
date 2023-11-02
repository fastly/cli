package compute

import (
	"bytes"
	"regexp"
	"strings"
)

// filterSecretsPattern attempts to capture a secret assigned in an environment
// variable where the key follows a common pattern.
// https://regex101.com/r/4GnH3r/1
const filterSecretsPattern = `(?i)\b[^\s_]+_(?:API|CLIENTSECRET|CREDENTIALS|KEY|PASSWORD|SECRET|TOKEN)(?:[^=]+)?=(?:\s+)?"?([^\s"]+)` // #nosec G101 (CWE-798)

// filterEnvVarSecrets identify environment variables containing secrets.
var filterEnvVarSecrets = []string{
	"AZURE_CLIENT_ID",
	"CI_JOB_JWT",
	"CI_JOB_JWT_V2",
	"FACEBOOK_APP_ID",
	"MSI_ENDPOINT",
	"OKTA_AUTHN_GROUPID",
	"OKTA_OAUTH2_CLIENTID",
}

// ExtendEnvVarSecretsFilter mutates filterEnvVarSecrets to include user
// specified environment variables. The `filter` argument is comma-separated.
func ExtendEnvVarSecretsFilter(filter string) {
	customFilters := strings.Split(filter, ",")
	for _, v := range customFilters {
		if v == "" {
			continue
		}
		var found bool
		for _, f := range filterEnvVarSecrets {
			if f == v {
				found = true
				break
			}
		}
		if !found {
			filterEnvVarSecrets = append(filterEnvVarSecrets, v)
		}
	}
}

// FilterEnvVarSecretsFromSlice mutates the input data such that any value
// assigned to an environment variable (identified as containing a secret) is
// redacted.
func FilterEnvVarSecretsFromSlice(data []string) {
	for i, v := range data {
		for _, f := range filterEnvVarSecrets {
			k := strings.Split(v, "=")[0]
			if strings.HasPrefix(k, f) {
				data[i] = k + "=REDACTED"
			}
		}
	}
}

// FilterEnvVarSecretsFromBytes mutates the input data such that any value
// assigned to an environment variable (identified as containing a secret) is
// redacted.
//
// Example:
// https://go.dev/play/p/GYXMNc7Froz
func FilterEnvVarSecretsFromBytes(data []byte) []byte {
	re := regexp.MustCompile(filterSecretsPattern)
	for _, matches := range re.FindAllSubmatch(data, -1) {
		if len(matches) == 2 {
			o := matches[0]
			n := bytes.ReplaceAll(matches[0], matches[1], []byte("REDACTED"))
			data = bytes.ReplaceAll(data, o, n)
		}
	}
	return data
}

// FilterEnvVarSecretsFromString mutates the input data such that any value
// assigned to an environment variable (identified as containing a secret) is
// redacted.
//
// Example:
// https://go.dev/play/p/GYXMNc7Froz
func FilterEnvVarSecretsFromString(data string) string {
	re := regexp.MustCompile(filterSecretsPattern)
	for _, matches := range re.FindAllStringSubmatch(data, -1) {
		if len(matches) == 2 {
			o := matches[0]
			n := strings.ReplaceAll(matches[0], matches[1], "REDACTED")
			data = strings.ReplaceAll(data, o, n)
		}
	}
	return data
}
