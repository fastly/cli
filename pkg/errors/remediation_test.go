package errors_test

import (
	"strings"
	"testing"

	"github.com/fastly/cli/pkg/errors"
)

func TestRemediationDisableAuthCommand(t *testing.T) {
	t.Setenv("FASTLY_DISABLE_AUTH_COMMAND", "")
	t.Run("AuthRemediation includes auth login by default", func(t *testing.T) {
		msg := errors.AuthRemediation()
		if !strings.Contains(msg, "fastly auth login") {
			t.Errorf("expected AuthRemediation to mention 'fastly auth login', got: %s", msg)
		}
	})

	t.Run("AuthRemediation omits auth login and --token when env var set", func(t *testing.T) {
		t.Setenv("FASTLY_DISABLE_AUTH_COMMAND", "1")
		msg := errors.AuthRemediation()
		if strings.Contains(msg, "fastly auth login") {
			t.Errorf("expected AuthRemediation to omit 'fastly auth login', got: %s", msg)
		}
		if strings.Contains(msg, "--token") {
			t.Errorf("expected AuthRemediation to omit --token, got: %s", msg)
		}
		if !strings.Contains(msg, "FASTLY_API_TOKEN") {
			t.Errorf("expected AuthRemediation to mention FASTLY_API_TOKEN, got: %s", msg)
		}
	})

	t.Run("ForbiddenRemediation includes auth login by default", func(t *testing.T) {
		msg := errors.ForbiddenRemediation()
		if !strings.Contains(msg, "fastly auth login") {
			t.Errorf("expected ForbiddenRemediation to mention 'fastly auth login', got: %s", msg)
		}
	})

	t.Run("ForbiddenRemediation omits auth login and whoami when env var set", func(t *testing.T) {
		t.Setenv("FASTLY_DISABLE_AUTH_COMMAND", "1")
		msg := errors.ForbiddenRemediation()
		if strings.Contains(msg, "fastly auth login") {
			t.Errorf("expected ForbiddenRemediation to omit 'fastly auth login', got: %s", msg)
		}
		if strings.Contains(msg, "fastly whoami") {
			t.Errorf("expected ForbiddenRemediation to omit 'fastly whoami', got: %s", msg)
		}
		if !strings.Contains(msg, "FASTLY_API_TOKEN") {
			t.Errorf("expected ForbiddenRemediation to mention FASTLY_API_TOKEN, got: %s", msg)
		}
	})

	t.Run("ErrNoToken remediation responds to env var", func(t *testing.T) {
		t.Setenv("FASTLY_DISABLE_AUTH_COMMAND", "1")
		re := errors.ErrNoToken()
		if strings.Contains(re.Remediation, "fastly auth login") {
			t.Errorf("expected ErrNoToken remediation to omit 'fastly auth login', got: %s", re.Remediation)
		}
	})

	t.Run("NonInteractiveAuthRemediation includes --token by default", func(t *testing.T) {
		t.Setenv("FASTLY_DISABLE_AUTH_COMMAND", "")
		msg := errors.NonInteractiveAuthRemediation()
		if !strings.Contains(msg, "--token") {
			t.Errorf("expected NonInteractiveAuthRemediation to mention --token, got: %s", msg)
		}
	})

	t.Run("NonInteractiveAuthRemediation omits --token when env var set", func(t *testing.T) {
		t.Setenv("FASTLY_DISABLE_AUTH_COMMAND", "1")
		msg := errors.NonInteractiveAuthRemediation()
		if strings.Contains(msg, "--token") {
			t.Errorf("expected NonInteractiveAuthRemediation to omit --token, got: %s", msg)
		}
		if !strings.Contains(msg, "FASTLY_API_TOKEN") {
			t.Errorf("expected NonInteractiveAuthRemediation to mention FASTLY_API_TOKEN, got: %s", msg)
		}
	})

	t.Run("ErrNonInteractiveNoToken omits --token when env var set", func(t *testing.T) {
		t.Setenv("FASTLY_DISABLE_AUTH_COMMAND", "1")
		re := errors.ErrNonInteractiveNoToken()
		if strings.Contains(re.Remediation, "--token") {
			t.Errorf("expected ErrNonInteractiveNoToken remediation to omit --token, got: %s", re.Remediation)
		}
		if !strings.Contains(re.Remediation, "FASTLY_API_TOKEN") {
			t.Errorf("expected ErrNonInteractiveNoToken remediation to mention FASTLY_API_TOKEN, got: %s", re.Remediation)
		}
	})

	t.Run("ProfileRemediation omits --token when env var set", func(t *testing.T) {
		t.Setenv("FASTLY_DISABLE_AUTH_COMMAND", "1")
		msg := errors.ProfileRemediation()
		if strings.Contains(msg, "--token") {
			t.Errorf("expected ProfileRemediation to omit --token, got: %s", msg)
		}
		if !strings.Contains(msg, "FASTLY_API_TOKEN") {
			t.Errorf("expected ProfileRemediation to mention FASTLY_API_TOKEN, got: %s", msg)
		}
	})
}
