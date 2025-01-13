package compute

import (
	"bytes"
	"regexp"
	"strings"
)

// StaticSecretEnvVars is a static list of env vars containing secrets.
//
// NOTE: Env Vars pulled from https://github.com/Puliczek/awesome-list-of-secrets-in-environment-variables
//
// The reason for not listing more environment variables is because we have a
// generalised pattern `SecretGeneralisedEnvVarPattern` that catches the
// majority of formats used.
var StaticSecretEnvVars = []string{
	"AZURE_CLIENT_ID",
	"CI_JOB_JWT",
	"CI_JOB_JWT_V2",
	"FACEBOOK_APP_ID",
	"MSI_ENDPOINT",
	"OKTA_AUTHN_GROUPID",
	"OKTA_OAUTH2_CLIENTID",
}

// SecretGeneralisedEnvVarPattern attempts to capture a secret assigned in an environment
// variable where the key follows a common pattern.
//
// Example:
// https://regex101.com/r/mf9Ymb/1
var SecretGeneralisedEnvVarPattern = regexp.MustCompile(`(?i)\b[^\s]+_(?:API|CLIENTSECRET|CREDENTIALS|KEY|PASSWORD|SECRET|TOKEN)(?:[^=]+)?=(?:\s+)?"?([^\s"]+)`) // #nosec G101 (CWE-798)

// AWSIDPattern is the pattern for an AWS ID.
var AWSIDPattern = regexp.MustCompile(`\b((?:AKIA|ABIA|ACCA|ASIA)[0-9A-Z]{16})\b`)

// AWSSecretPattern is the pattern for an AWS Secret.
var AWSSecretPattern = regexp.MustCompile(`[^A-Za-z0-9+\/]{0,1}([A-Za-z0-9+\/]{40})[^A-Za-z0-9+\/]{0,1}`)

// GitHubOAuthTokenPattern is the pattern for a GitHub OAuth token.
var GitHubOAuthTokenPattern = regexp.MustCompile(`\b((?:ghp|gho|ghu|ghs|ghr|github_pat)_[a-zA-Z0-9_]{36,255})\b`)

// GitHubOldOAuthTokenPattern is the pattern for an older GitHub OAuth token format.
var GitHubOldOAuthTokenPattern = regexp.MustCompile(`(?i)(?:github|gh|pat|token)[^\.].{0,40}[ =:'"]+([a-f0-9]{40})\b`)

// GitHubOAuth2ClientIDPattern is the pattern for a GitHub OAuth2 ClientID.
var GitHubOAuth2ClientIDPattern = regexp.MustCompile(`(?i)(?:github)(?:.|[\n\r]){0,40}\b([a-f0-9]{20})\b`)

// GitHubOAuth2ClientSecretPattern is the pattern for a GitHub OAuth2 ClientID.
var GitHubOAuth2ClientSecretPattern = regexp.MustCompile(`(?i)(?:github)(?:.|[\n\r]){0,40}\b([a-f0-9]{40})\b`)

// GitHubAppIDPattern is the pattern for a GitHub App ID.
var GitHubAppIDPattern = regexp.MustCompile(`(?i)(?:github)(?:.|[\n\r]){0,40}\b([0-9]{6})\b`)

// GitHubAppKeyPattern is the pattern for a GitHub App Key.
var GitHubAppKeyPattern = regexp.MustCompile(`(?i)(?:github)(?:.|[\n\r]){0,40}(-----BEGIN RSA PRIVATE KEY-----\s[A-Za-z0-9+\/\s]*\s-----END RSA PRIVATE KEY-----)`)

// SecretPatterns is a collection of secret identifying patterns.
//
// NOTE: Patterns pulled from https://github.com/trufflesecurity/trufflehog
var SecretPatterns = []*regexp.Regexp{
	AWSIDPattern,
	AWSSecretPattern,
	GitHubOAuthTokenPattern,
	GitHubOldOAuthTokenPattern,
	GitHubOAuth2ClientIDPattern,
	GitHubOAuth2ClientSecretPattern,
	GitHubAppIDPattern,
	GitHubAppKeyPattern,
}

// ExtendStaticSecretEnvVars mutates `StaticSecretEnvVars` to include user
// specified environment variables. The `filter` argument is comma-separated.
func ExtendStaticSecretEnvVars(filter string) {
	customFilters := strings.Split(filter, ",")
	for _, v := range customFilters {
		if v == "" {
			continue
		}
		var found bool
		for _, f := range StaticSecretEnvVars {
			if f == v {
				found = true
				break
			}
		}
		if !found {
			StaticSecretEnvVars = append(StaticSecretEnvVars, v)
		}
	}
}

// FilterSecretsFromSlice returns the input slice modified such that any value
// assigned to an environment variable (identified as containing a secret) is
// redacted. Additionally, any 'value' identified as being a secret will also be
// redacted.
//
// NOTE: `data` is expected to contain "KEY=VALUE" formatted strings.
func FilterSecretsFromSlice(data []string) []string {
	copyOfData := make([]string, len(data))
	copy(copyOfData, data)

	for i, keypair := range copyOfData {
		k, v, found := strings.Cut(keypair, "=")
		if !found {
			return copyOfData
		}
		for _, f := range StaticSecretEnvVars {
			if k == f {
				copyOfData[i] = k + "=REDACTED"
				break
			}
		}
		if strings.Contains(copyOfData[i], "REDACTED") {
			continue
		}
		for _, matches := range SecretGeneralisedEnvVarPattern.FindAllStringSubmatch(keypair, -1) {
			if len(matches) == 2 {
				o := matches[0]
				n := strings.ReplaceAll(matches[0], matches[1], "REDACTED")
				copyOfData[i] = strings.ReplaceAll(keypair, o, n)
			}
		}
		if strings.Contains(copyOfData[i], "REDACTED") {
			continue
		}
		for _, pattern := range SecretPatterns {
			n := pattern.ReplaceAllString(v, "REDACTED")
			copyOfData[i] = k + "=" + n
			if n == "REDACTED" {
				break
			}
		}
	}

	return copyOfData
}

// FilterSecretsFromString returns the input string modified such that any value
// assigned to an environment variable (identified as containing a secret) is
// redacted. Additionally, any 'value' identified as being a secret will also be
// redacted.
//
// Example:
// https://go.dev/play/p/jhCcC4SlsHA
//
// NOTE: The input data should be simple (i.e. not a complex json object).
// Otherwise the `SecretGeneralisedEnvVarPattern` will unlikely match all cases.
func FilterSecretsFromString(data string) string {
	staticSecretEnvVarsPattern := regexp.MustCompile(`(?i)\b(?:` + strings.Join(StaticSecretEnvVars, "|") + `)(?:\s+)?=(?:\s+)?"?([^\s"]+)`)
	for _, matches := range staticSecretEnvVarsPattern.FindAllStringSubmatch(data, -1) {
		if len(matches) == 2 {
			o := matches[0]
			n := strings.ReplaceAll(matches[0], matches[1], "REDACTED")
			data = strings.ReplaceAll(data, o, n)
		}
	}
	for _, matches := range SecretGeneralisedEnvVarPattern.FindAllStringSubmatch(data, -1) {
		if len(matches) == 2 {
			o := matches[0]
			n := strings.ReplaceAll(matches[0], matches[1], "REDACTED")
			data = strings.ReplaceAll(data, o, n)
		}
	}
	for _, pattern := range SecretPatterns {
		data = pattern.ReplaceAllString(data, "REDACTED")
	}
	return data
}

// FilterSecretsFromBytes returns the input string modified such that any value
// assigned to an environment variable (identified as containing a secret) is
// redacted. Additionally, any 'value' identified as being a secret will also be
// redacted.
//
// Example:
// https://go.dev/play/p/jhCcC4SlsHA
//
// NOTE: The input data should be simple (i.e. not a complex json object).
// Otherwise the `SecretGeneralisedEnvVarPattern` will unlikely match all cases.
func FilterSecretsFromBytes(data []byte) []byte {
	copyOfData := make([]byte, len(data))
	copy(copyOfData, data)

	staticSecretEnvVarsPattern := regexp.MustCompile(`(?i)\b(?:` + strings.Join(StaticSecretEnvVars, "|") + `)(?:\s+)?=(?:\s+)?"?([^\s"]+)`)
	for _, matches := range staticSecretEnvVarsPattern.FindAllSubmatch(copyOfData, -1) {
		if len(matches) == 2 {
			o := matches[0]
			n := bytes.ReplaceAll(matches[0], matches[1], []byte("REDACTED"))
			copyOfData = bytes.ReplaceAll(copyOfData, o, n)
		}
	}
	for _, matches := range SecretGeneralisedEnvVarPattern.FindAllSubmatch(copyOfData, -1) {
		if len(matches) == 2 {
			o := matches[0]
			n := bytes.ReplaceAll(matches[0], matches[1], []byte("REDACTED"))
			copyOfData = bytes.ReplaceAll(copyOfData, o, n)
		}
	}
	for _, pattern := range SecretPatterns {
		copyOfData = pattern.ReplaceAll(copyOfData, []byte("REDACTED"))
	}

	return copyOfData
}
