package env

import (
	"fmt"
	"os"
	"strings"

	"github.com/fastly/cli/pkg/runtime"
)

const (
	// Token is the env var we look in for the Fastly API token.
	// gosec flagged this:
	// G101 (CWE-798): Potential hardcoded credentials
	// Disabling as we use the value in the command help output.
	/* #nosec */
	Token = "FASTLY_API_TOKEN"

	// Endpoint is the env var we look in for the API endpoint.
	Endpoint = "FASTLY_API_ENDPOINT"

	// ServiceID is the env var we look in for the required Service ID.
	ServiceID = "FASTLY_SERVICE_ID"

	// CustomerID is the env var we look in for a Customer ID.
	CustomerID = "FASTLY_CUSTOMER_ID"
)

// Vars returns a slice of environment variables appropriate to platform.
// *nix: $HOME, $USER, ...
// Windows: %HOME%, %USER%, ...
func Vars() []string {
	vars := []string{}
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		vars = append(vars, toVar(pair[0]))
	}
	return vars
}

func toVar(v string) string {
	if runtime.Windows {
		return toWin(v)
	}
	return toNix(v)
}

func toNix(v string) string {
	return fmt.Sprintf("\\$%s", v)
}

func toWin(v string) string {
	return fmt.Sprintf("%%%s%%", v)
}
