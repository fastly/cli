package config

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	fsterr "github.com/fastly/cli/pkg/errors"
)

type mockHTTPClient struct{}

func (c mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return &http.Response{}, context.DeadlineExceeded
}

// TestConfigLoad validates that when a context.DeadlineExceeded error is
// returned from a http.Client.Do request, that an appropriate remediation
// error is returned to the user.
func TestConfigLoad(t *testing.T) {
	var (
		c mockHTTPClient
		d time.Duration
		f *File
	)
	if err := f.Load("foo", c, d); err != nil {
		if !errors.As(err, &fsterr.RemediationError{}) {
			t.Errorf("expected RemediationError got: %T", err)
		}
		if err.(fsterr.RemediationError).Remediation != fsterr.NetworkRemediation {
			t.Errorf("expected NetworkRemediation got: %s", err.(fsterr.RemediationError).Remediation)
		}
	} else {
		t.Error("expected an error, got nil")
	}
}
