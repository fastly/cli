package authenticate

import (
	"fmt"
	"io"
	"net/http"

	"github.com/hashicorp/cap/oidc"
	"github.com/skratchdot/open-golang/open"

	"github.com/fastly/cli/pkg/auth"
	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/profile"
	"github.com/fastly/cli/pkg/text"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	cmd.Base
}

// NewRootCommand returns a new command registered in the parent.
func NewRootCommand(parent cmd.Registerer, g *global.Data) *RootCommand {
	var c RootCommand
	c.Globals = g
	c.CmdClause = parent.Command("authenticate", "Authenticate with Fastly (returns temporary, auto-rotated, API token)")
	return &c
}

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, out io.Writer) error {
	verifier, err := oidc.NewCodeVerifier()
	if err != nil {
		return fsterr.RemediationError{
			Inner:       fmt.Errorf("failed to generate a code verifier: %w", err),
			Remediation: auth.Remediation,
		}
	}

	result := make(chan auth.AuthorizationResult)
	endpoint, _ := c.Globals.Endpoint()

	s := auth.Server{
		HTTPClient:      c.Globals.HTTPClient,
		Result:          result,
		Router:          http.NewServeMux(),
		SessionEndpoint: endpoint,
		Verifier:        verifier,
	}
	s.Routes()

	var serverErr error

	go func() {
		err := s.Start()
		if err != nil {
			serverErr = err
		}
	}()

	if serverErr != nil {
		return serverErr
	}

	text.Info(out, "Starting a local server to handle the authentication flow.")

	authorizationURL, err := auth.GenURL(verifier)
	if err != nil {
		return fsterr.RemediationError{
			Inner:       fmt.Errorf("failed to generate an authorization URL: %w", err),
			Remediation: auth.Remediation,
		}
	}

	text.Break(out)
	text.Description(out, "We're opening the following URL in your default web browser so you may authenticate with Fastly", authorizationURL)

	err = open.Run(authorizationURL)
	if err != nil {
		return fmt.Errorf("failed to open your default browser: %w", err)
	}

	ar := <-result
	if ar.Err != nil || ar.SessionToken == "" {
		return fsterr.RemediationError{
			Inner:       fmt.Errorf("failed to authorize: %w", ar.Err),
			Remediation: auth.Remediation,
		}
	}

	profileName, _ := profile.Default(c.Globals.Config.Profiles)

	if profileName == "" {
		c.Globals.Config.Profiles[profile.DefaultName] = &config.Profile{
			Default: true,
			Email:   ar.Email,
			Token:   ar.SessionToken,
		}
	} else {
		ps, ok := profile.Edit(profileName, c.Globals.Config.Profiles, func(p *config.Profile) {
			p.Token = ar.SessionToken
		})
		if !ok {
			return fsterr.RemediationError{
				Inner:       fmt.Errorf("failed to update default profile with new session token"),
				Remediation: "Run `fastly profile update` and manually paste in the session token.",
			}
		}
		c.Globals.Config.Profiles = ps
	}

	if err := c.Globals.Config.Write(c.Globals.ConfigPath); err != nil {
		c.Globals.ErrLog.Add(err)
		return fmt.Errorf("error saving config file: %w", err)
	}

	// FIXME: Don't just update the default profile.
	// Allow user to configure this via a --profile flag.

	text.Success(out, "Session token (persisted to your local configuration): %s", ar.SessionToken)
	return nil
}
