package app_test

import (
	"bytes"
	stderrors "errors"
	"io"
	"strings"
	"testing"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/env"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/testutil"
)

func TestTokenFlagDisabledWhenAuthCommandDisabled(t *testing.T) {
	t.Setenv(env.DisableAuthCommand, "1")

	t.Run("--token flag rejected", func(t *testing.T) {
		var stdout bytes.Buffer
		args := testutil.SplitArgs("--token abc version")

		app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
			return testutil.MockGlobalData(args, &stdout), nil
		}

		err := app.Run(args, nil)
		if err == nil {
			t.Fatal("expected error when using --token with FASTLY_DISABLE_AUTH_COMMAND set")
		}
		errStr := err.Error()
		if !strings.Contains(errStr, "unknown long flag") {
			t.Errorf("expected unknown flag error, got: %s", errStr)
		}
	})

	t.Run("-t flag rejected", func(t *testing.T) {
		var stdout bytes.Buffer
		args := testutil.SplitArgs("-t abc version")

		app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
			return testutil.MockGlobalData(args, &stdout), nil
		}

		err := app.Run(args, nil)
		if err == nil {
			t.Fatal("expected error when using -t with FASTLY_DISABLE_AUTH_COMMAND set")
		}
		errStr := err.Error()
		if !strings.Contains(errStr, "unknown short flag") {
			t.Errorf("expected unknown flag error, got: %s", errStr)
		}
	})

	t.Run("help output omits --token", func(t *testing.T) {
		var stdout bytes.Buffer
		args := testutil.SplitArgs("--help")

		app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
			return testutil.MockGlobalData(args, &stdout), nil
		}

		err := app.Run(args, nil)

		var output string
		if err != nil {
			var re errors.RemediationError
			if stderrors.As(err, &re) {
				output = re.Prefix
			}
		}
		output += stdout.String()

		if strings.Contains(output, "--token") {
			t.Errorf("expected --token to be absent from help output when FASTLY_DISABLE_AUTH_COMMAND is set, got:\n%s", output)
		}
	})
}

func TestTokenFlagAvailableByDefault(t *testing.T) {
	t.Setenv(env.DisableAuthCommand, "")
	t.Run("help output includes --token", func(t *testing.T) {
		var stdout bytes.Buffer
		args := testutil.SplitArgs("--help")

		app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
			return testutil.MockGlobalData(args, &stdout), nil
		}

		err := app.Run(args, nil)

		var output string
		if err != nil {
			var re errors.RemediationError
			if stderrors.As(err, &re) {
				output = re.Prefix
			}
		}
		output += stdout.String()

		if !strings.Contains(output, "--token") {
			t.Errorf("expected --token in help output, got:\n%s", output)
		}
	})
}
