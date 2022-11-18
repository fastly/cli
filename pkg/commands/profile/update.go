package profile

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/profile"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v7/fastly"
)

// APIClientFactory allows the profile command to regenerate the global Fastly
// API client when a new token is provided, in order to validate that token.
// It's a redeclaration of the app.APIClientFactory to avoid an import loop.
type APIClientFactory func(token, endpoint string) (api.Interface, error)

// UpdateCommand represents a Kingpin command.
type UpdateCommand struct {
	cmd.Base

	clientFactory APIClientFactory
	profile       string
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent cmd.Registerer, cf APIClientFactory, globals *config.Data) *UpdateCommand {
	var c UpdateCommand
	c.Globals = globals
	c.CmdClause = parent.Command("update", "Update user profile")
	c.CmdClause.Arg("profile", "Profile to update (default 'user')").Default("user").Short('p').StringVar(&c.profile)
	c.clientFactory = cf
	return &c
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(in io.Reader, out io.Writer) error {
	name, p := profile.Get(c.profile, c.Globals.File.Profiles)
	if name == "" {
		msg := fmt.Sprintf(profile.DoesNotExist, c.profile)
		return fsterr.RemediationError{
			Inner:       fmt.Errorf(msg),
			Remediation: fsterr.ProfileRemediation,
		}
	}

	opts := []profile.EditOption{}

	text.Break(out)

	token, err := text.InputSecure(out, text.BoldYellow("Profile token: (leave blank to skip): "), in)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}
	if token != "" {
		opts = append(opts, func(p *config.Profile) {
			p.Token = token
		})
	}

	text.Break(out)
	text.Break(out)

	def, err := text.AskYesNo(out, "Make profile the default? [y/N] ", in)
	if err != nil {
		return err
	}
	opts = append(opts, func(p *config.Profile) {
		p.Default = def
	})

	// User didn't want to change their token value so reassign original.
	if token == "" {
		token = p.Token
	}

	text.Break(out)
	text.Break(out)

	progress := text.NewProgress(out, c.Globals.Verbose())
	defer func() {
		if err != nil {
			c.Globals.ErrLog.Add(err)
			progress.Fail() // progress.Done is handled inline
		}
	}()

	endpoint, _ := c.Globals.Endpoint()

	u, err := c.validateToken(token, endpoint, progress)
	if err != nil {
		return err
	}
	opts = append(opts, func(p *config.Profile) {
		p.Email = u.Login
	})

	var ok bool

	ps, ok := profile.Edit(c.profile, c.Globals.File.Profiles, opts...)
	if !ok {
		msg := fmt.Sprintf(profile.DoesNotExist, c.profile)
		return fsterr.RemediationError{
			Inner:       fmt.Errorf(msg),
			Remediation: fsterr.ProfileRemediation,
		}
	}
	c.Globals.File.Profiles = ps

	if err := c.Globals.File.Write(c.Globals.Path); err != nil {
		c.Globals.ErrLog.Add(err)
		return fmt.Errorf("error saving config file: %w", err)
	}

	progress.Done()

	text.Success(out, "Profile '%s' updated", c.profile)
	return nil
}

// validateToken ensures the token can be used to acquire user data.
func (c *UpdateCommand) validateToken(token, endpoint string, progress text.Progress) (*fastly.User, error) {
	progress.Step("Validating token...")

	client, err := c.clientFactory(token, endpoint)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Endpoint": endpoint,
		})
		return nil, fmt.Errorf("error regenerating Fastly API client: %w", err)
	}

	t, err := client.GetTokenSelf()
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return nil, fmt.Errorf("error validating token: %w", err)
	}

	user, err := client.GetUser(&fastly.GetUserInput{
		ID: t.UserID,
	})
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"User ID": t.UserID,
		})
		return nil, fmt.Errorf("error fetching token user: %w", err)
	}

	return user, nil
}
