package auth

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v6/fastly"
)

// ignoreCommands represents commands that do not require a token.
var ignoreCommands = []string{"configure", "ip-list", "update", "version"}

// App is the authentication service responsible for authenticating the user
// and returning a short-lived access token.
var App = "https://developer.fastly.com"

// Required determines if the OAuth flow is required.
func Required(cmd, token string, s config.Source, stdout io.Writer) (required bool) {
	for _, name := range ignoreCommands {
		if cmd == name {
			return false
		}
	}

	// No token is defined so yes we require the OAuth flow.
	if s == config.SourceUndefined {
		text.Info(stdout, "An access token is required but has not been provided.")
		return true
	}

	if s == config.SourceFile {
		apiClient, _ := fastly.NewClient(token)
		if status, err := ValidateToken(token, apiClient); err != nil {
			text.Info(stdout, "We were not able to validate your access token.")
			required = true
		} else {
			switch status {
			case http.StatusUnauthorized:
				text.Info(stdout, "Your access token has expired.")
				required = true
			case http.StatusForbidden:
				text.Warning(stdout, "Your access token is invalid.")
				required = true
			case http.StatusOK:
				// The token was validated successfully so no OAuth flow required.
			}
		}
	}

	return required
}

// Init starts the OAuth flow and returns a new access token.
//
// NOTE: This function blocks the command execution until a token has been
// pulled from the relevant channel.
func Init(stdin io.Reader, stdout io.Writer, o Opener) (string, error) {
	text.Break(stdout)
	text.Output(stdout, "We are about to initialise a new authentication flow, which requires the CLI to open your web browser for you.")
	text.Break(stdout)

	label := "Would you like to continue? [y/N] "
	cont, err := text.AskYesNo(stdout, label, stdin)
	if err != nil {
		return "", err
	}
	if !cont {
		return "", errors.New("a token is required but the authentication process was cancelled")
	}

	token := make(chan string)
	port := make(chan int)

	go ListenAndServe(port, token)

	p := <-port
	path := "/auth/login"
	callback := fmt.Sprintf("http://127.0.0.1:%d/auth-callback", p)
	url := fmt.Sprintf("%s%s?redirect_uri=%s", App, path, callback)
	err = o(url)
	if err != nil {
		return "", err
	}

	t := <-token
	if t == "" {
		return t, errors.New("no token received from authentication service")
	}

	return t, nil
}
