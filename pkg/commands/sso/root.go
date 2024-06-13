package sso

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/fastly/cli/pkg/api/undocumented"
	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/auth"
	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/profile"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/cli/pkg/useragent"
)

// ForceReAuth indicates we want to force a re-auth of the user's session.
// This variable is overridden by ../../app/run.go to force a re-auth.
var ForceReAuth = false

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	profile string

	// The following fields are populated once authentication is complete.
	customerID   string
	customerName string

	// IMPORTANT: The following fields are public to the `profile` subcommands.

	// InvokedFromProfileCreate indicates if we should create a new profile.
	InvokedFromProfileCreate bool
	// ProfileCreateName indicates the new profile name.
	ProfileCreateName string
	// ProfileDefault indicates if the affected profile should become the default.
	ProfileDefault bool
	// InvokedFromProfileUpdate indicates if we should update a profile.
	InvokedFromProfileUpdate bool
	// ProfileUpdateName indicates the profile name to update.
	ProfileUpdateName string
	// InvokedFromProfileSwitch indicates if we should switch a profile.
	InvokedFromProfileSwitch bool
	// ProfileSwitchName indicates the profile name to switch to.
	ProfileSwitchName string
	// ProfileSwitchEmail indicates the profile email to reference in auth URL.
	ProfileSwitchEmail string
	// ProfileSwitchCustomerID indicates the customer ID to reference in auth URL.
	ProfileSwitchCustomerID string
}

// NewRootCommand returns a new command registered in the parent.
func NewRootCommand(parent argparser.Registerer, g *global.Data) *RootCommand {
	var c RootCommand
	c.Globals = g
	// FIXME: Unhide this command once SSO is GA.
	c.CmdClause = parent.Command("sso", "Single Sign-On authentication (defaults to current profile)")
	c.CmdClause.Arg("profile", "Profile to authenticate (i.e. create/update a token for)").Short('p').StringVar(&c.profile)
	return &c
}

// Exec implements the command interface.
func (c *RootCommand) Exec(in io.Reader, out io.Writer) error {
	profileName, _ := c.identifyProfileAndFlow()

	// We need to prompt the user, so they know we're about to open their web
	// browser, but we also need to handle the scenario where the `sso` command is
	// invoked indirectly via ../../app/run.go as that package will have its own
	// (similar) prompt before invoking this command. So to avoid a double prompt,
	// the app package will set `SkipAuthPrompt: true`.
	if !c.Globals.SkipAuthPrompt && !c.Globals.Flags.AutoYes && !c.Globals.Flags.NonInteractive {
		msg := fmt.Sprintf("We're going to authenticate the '%s' profile", profileName)
		text.Important(out, "%s. We need to open your browser to authenticate you.", msg)
		text.Break(out)
		cont, err := text.AskYesNo(out, text.BoldYellow("Do you want to continue? [y/N]: "), in)
		text.Break(out)
		if err != nil {
			return err
		}
		if !cont {
			return fsterr.SkipExitError{
				Skip: true,
				Err:  fsterr.ErrDontContinue,
			}
		}
	}

	var serverErr error
	go func() {
		err := c.Globals.AuthServer.Start()
		if err != nil {
			serverErr = err
		}
	}()
	if serverErr != nil {
		return serverErr
	}

	text.Info(out, "Starting a local server to handle the authentication flow.")

	// For creating/updating a profile we set `prompt` because we want to ensure
	// that another session (from a different profile) doesn't cause unexpected
	// errors for the user flow. This forces a re-auth.
	if c.InvokedFromProfileCreate || ForceReAuth {
		c.Globals.AuthServer.SetParam("prompt", "login select_account")
	}
	if c.InvokedFromProfileUpdate || c.InvokedFromProfileSwitch {
		c.Globals.AuthServer.SetParam("prompt", "login")
		if c.ProfileSwitchEmail != "" {
			c.Globals.AuthServer.SetParam("login_hint", c.ProfileSwitchEmail)
		}
		if c.ProfileSwitchCustomerID != "" {
			c.Globals.AuthServer.SetParam("account_hint", c.ProfileSwitchCustomerID)
		}
	}

	authorizationURL, err := c.Globals.AuthServer.AuthURL()
	if err != nil {
		return fsterr.RemediationError{
			Inner:       fmt.Errorf("failed to generate an authorization URL: %w", err),
			Remediation: auth.Remediation,
		}
	}

	text.Break(out)
	text.Description(out, "We're opening the following URL in your default web browser so you may authenticate with Fastly", authorizationURL)

	err = c.Globals.Opener(authorizationURL)
	if err != nil {
		return fmt.Errorf("failed to open your default browser: %w", err)
	}

	ar := <-c.Globals.AuthServer.GetResult()
	if ar.Err != nil || ar.SessionToken == "" {
		err := ar.Err
		if ar.Err == nil {
			err = errors.New("no session token")
		}
		return fsterr.RemediationError{
			Inner:       fmt.Errorf("failed to authorize: %w", err),
			Remediation: auth.Remediation,
		}
	}

	err = c.processCustomer(ar)
	if err != nil {
		return fmt.Errorf("failed to use session token to get customer data: %w", err)
	}

	err = c.processProfiles(ar)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return fmt.Errorf("failed to process profile data: %w", err)
	}

	textFn := text.Success
	if c.InvokedFromProfileCreate || c.InvokedFromProfileUpdate || c.InvokedFromProfileSwitch {
		textFn = text.Info
	}
	textFn(out, "Session token (persisted to your local configuration): %s", ar.SessionToken)
	return nil
}

// ProfileFlow enumerates which profile flow to take.
type ProfileFlow uint8

const (
	// ProfileNone indicates we need to create a new 'default' profile as no
	// profiles currently exist.
	ProfileNone ProfileFlow = iota

	// ProfileCreate indicates we need to create a new profile using details
	// passed in either from the `sso` or `profile create` command.
	ProfileCreate

	// ProfileUpdate indicates we need to update a profile using details passed in
	// either from the `sso` or `profile update` command.
	ProfileUpdate

	// ProfileSwitch indicates we need to re-authenticate and switch profiles.
	// Triggered by user invoking `fastly profile switch` with an SSO-based profile.
	ProfileSwitch
)

// identifyProfileAndFlow identifies the profile and the specific workflow.
func (c *RootCommand) identifyProfileAndFlow() (profileName string, flow ProfileFlow) {
	var profileOverride string
	switch {
	case c.Globals.Flags.Profile != "":
		profileOverride = c.Globals.Flags.Profile
	case c.Globals.Manifest.File.Profile != "":
		profileOverride = c.Globals.Manifest.File.Profile
	}

	currentDefaultProfile, _ := profile.Default(c.Globals.Config.Profiles)
	var newDefaultProfile string
	if currentDefaultProfile == "" && len(c.Globals.Config.Profiles) > 0 {
		newDefaultProfile, c.Globals.Config.Profiles = profile.SetADefault(c.Globals.Config.Profiles)
	}

	switch {
	case profileOverride != "":
		return profileOverride, ProfileUpdate
	case c.profile != "" && profile.Get(c.profile, c.Globals.Config.Profiles) != nil:
		return c.profile, ProfileUpdate
	case c.InvokedFromProfileCreate && c.ProfileCreateName != "":
		return c.ProfileCreateName, ProfileCreate
	case c.InvokedFromProfileUpdate && c.ProfileUpdateName != "":
		return c.ProfileUpdateName, ProfileUpdate
	case c.InvokedFromProfileSwitch && c.ProfileSwitchName != "":
		return c.ProfileSwitchName, ProfileSwitch
	case currentDefaultProfile != "":
		return currentDefaultProfile, ProfileUpdate
	case newDefaultProfile != "":
		return newDefaultProfile, ProfileUpdate
	default:
		return profile.DefaultName, ProfileCreate
	}
}

func (c *RootCommand) processCustomer(ar auth.AuthorizationResult) error {
	debugMode, _ := strconv.ParseBool(c.Globals.Env.DebugMode)
	apiEndpoint, _ := c.Globals.APIEndpoint()
	// NOTE: The endpoint is documented but not implemented in go-fastly.
	data, err := undocumented.Call(undocumented.CallOptions{
		APIEndpoint: apiEndpoint,
		HTTPClient:  c.Globals.HTTPClient,
		HTTPHeaders: []undocumented.HTTPHeader{
			{
				Key:   "Accept",
				Value: "application/json",
			},
			{
				Key:   "User-Agent",
				Value: useragent.Name,
			},
		},
		Method: http.MethodGet,
		Path:   "/current_customer",
		Token:  ar.SessionToken,
		Debug:  debugMode,
	})
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return fmt.Errorf("error executing current_customer API request: %w", err)
	}

	var response CurrentCustomerResponse
	if err := json.Unmarshal(data, &response); err != nil {
		c.Globals.ErrLog.Add(err)
		return fmt.Errorf("error decoding current_customer API response: %w", err)
	}

	c.customerID = response.ID
	c.customerName = response.Name

	return nil
}

// CurrentCustomerResponse models the Fastly API response for the
// /current_customer endpoint.
type CurrentCustomerResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// processProfiles updates the relevant profile with the returned token data.
//
// First it checks the --profile flag and the `profile` fastly.toml field.
// Second it checks to see which profile is currently the default.
// Third it identifies which profile to be modified.
// Fourth it writes the updated in-memory data back to disk.
func (c *RootCommand) processProfiles(ar auth.AuthorizationResult) error {
	profileName, flow := c.identifyProfileAndFlow()

	//nolint:exhaustive
	switch flow {
	case ProfileCreate:
		c.processCreateProfile(ar, profileName)
	case ProfileUpdate:
		err := c.processUpdateProfile(ar, profileName)
		if err != nil {
			return fmt.Errorf("failed to update profile: %w", err)
		}
	case ProfileSwitch:
		err := c.processSwitchProfile(ar, profileName)
		if err != nil {
			return fmt.Errorf("failed to switch profile: %w", err)
		}
	}

	if err := c.Globals.Config.Write(c.Globals.ConfigPath); err != nil {
		return fmt.Errorf("failed to update config file: %w", err)
	}
	return nil
}

// processCreateProfile handles creating a new profile.
func (c *RootCommand) processCreateProfile(ar auth.AuthorizationResult, profileName string) {
	isDefault := true
	if c.InvokedFromProfileCreate {
		isDefault = c.ProfileDefault
	}

	c.Globals.Config.Profiles = createNewProfile(
		profileName,
		c.customerID,
		c.customerName,
		isDefault,
		c.Globals.Config.Profiles,
		ar,
	)

	// If the user wants the newly created profile to be their new default, then
	// we'll call Set for its side effect of resetting all other profiles to have
	// their Default field set to false.
	if c.ProfileDefault { // this is set by the `profile create` command.
		if ps, ok := profile.SetDefault(c.ProfileCreateName, c.Globals.Config.Profiles); ok {
			c.Globals.Config.Profiles = ps
		}
	}
}

// processUpdateProfile handles updating a profile.
func (c *RootCommand) processUpdateProfile(ar auth.AuthorizationResult, profileName string) error {
	var isDefault bool
	if p := profile.Get(profileName, c.Globals.Config.Profiles); p != nil {
		isDefault = p.Default
	}
	if c.InvokedFromProfileUpdate {
		isDefault = c.ProfileDefault
	}
	ps, err := editProfile(
		profileName,
		c.customerID,
		c.customerName,
		isDefault,
		c.Globals.Config.Profiles,
		ar,
	)
	if err != nil {
		return err
	}
	c.Globals.Config.Profiles = ps
	return nil
}

// processSwitchProfile handles updating a profile.
func (c *RootCommand) processSwitchProfile(ar auth.AuthorizationResult, profileName string) error {
	ps, err := editProfile(
		profileName,
		c.customerID,
		c.customerName,
		c.ProfileDefault,
		c.Globals.Config.Profiles,
		ar,
	)
	if err != nil {
		return err
	}
	ps, ok := profile.SetDefault(profileName, ps)
	if !ok {
		return fmt.Errorf("failed to set '%s' to be the default profile", profileName)
	}
	c.Globals.Config.Profiles = ps
	return nil
}

// IMPORTANT: Mutates the config.Profiles map type.
// We need to return the modified type so it can be safely reassigned.
func createNewProfile(profileName, customerID, customerName string, makeDefault bool, p config.Profiles, ar auth.AuthorizationResult) config.Profiles {
	now := time.Now().Unix()
	if p == nil {
		p = make(config.Profiles)
	}
	p[profileName] = &config.Profile{
		AccessToken:         ar.Jwt.AccessToken,
		AccessTokenCreated:  now,
		AccessTokenTTL:      ar.Jwt.ExpiresIn,
		CustomerID:          customerID,
		CustomerName:        customerName,
		Default:             makeDefault,
		Email:               ar.Email,
		RefreshToken:        ar.Jwt.RefreshToken,
		RefreshTokenCreated: now,
		RefreshTokenTTL:     ar.Jwt.RefreshExpiresIn,
		Token:               ar.SessionToken,
	}
	return p
}

// editProfile mutates the given profile with JWT details returned from the SSO
// authentication process.
//
// IMPORTANT: Mutates the config.Profiles map type.
// We need to return the modified type so it can be safely reassigned.
func editProfile(profileName, customerID, customerName string, makeDefault bool, p config.Profiles, ar auth.AuthorizationResult) (config.Profiles, error) {
	ps, ok := profile.Edit(profileName, p, func(p *config.Profile) {
		now := time.Now().Unix()
		p.Default = makeDefault
		p.AccessToken = ar.Jwt.AccessToken
		p.AccessTokenCreated = now
		p.AccessTokenTTL = ar.Jwt.ExpiresIn
		p.CustomerID = customerID
		p.CustomerName = customerName
		p.Email = ar.Email
		p.RefreshToken = ar.Jwt.RefreshToken
		p.RefreshTokenCreated = now
		p.RefreshTokenTTL = ar.Jwt.RefreshExpiresIn
		p.Token = ar.SessionToken
	})
	if !ok {
		return ps, fsterr.RemediationError{
			Inner:       fmt.Errorf("failed to update '%s' profile with new token data", profileName),
			Remediation: "Run `fastly sso` to retry.",
		}
	}
	return ps, nil
}
