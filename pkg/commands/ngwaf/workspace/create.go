package workspace

import (
	"context"
	"errors"
	"io"
	"strconv"
	"strings"

	"github.com/fastly/go-fastly/v19/fastly"
	"github.com/fastly/go-fastly/v19/fastly/ngwaf/v1/workspaces"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// CreateCommand calls the Fastly API to create workspaces.
type CreateCommand struct {
	argparser.Base

	// Required.
	Description  string `required:"" help:"User submitted description of a workspace."`
	BlockingMode string `name:"blockingMode" required:"" help:"User configured mode blocking mode."`
	NameFlag     string `name:"name" required:"" help:"User submitted display name of a workspace."`

	// Optional.
	AttackThresholds         argparser.OptionalString `name:"attackThresholds" help:"Attack threshold parameters for system site alerts. Each threshold value is the number of attack signals per IP address that must be detected during the interval before the related IP address is flagged. Input accepted as colon separated string: Immediate:OneMinute:TenMinutes:OneHour"`
	ClientIPHeaders          argparser.OptionalString `name:"clientIPHeaders" help:"Specify the request header containing the client IP address. Input accepted as colon separated string."`
	DefaultBlockingCode      argparser.OptionalInt    `name:"defaultBlockingCode" help:"Default status code that is returned when a request to your web application is blocked."`
	DefaultRedirectURL       argparser.OptionalString `name:"defaultRedirectURL" help:"Redirect url to be used if code 301 or 302 is used."`
	IPAnonymization          argparser.OptionalString `name:"ipAnonymization" help:"Agents will anonymize IP addresses according to the option selected."`
	AutoEnableVirtualPatches argparser.OptionalBool   `name:"autoEnableNewVirtualPatches" help:"Automatically enable, in block mode, virtual patches released after this setting is enabled. Virtual patches released before it was enabled are not changed. Use --no-autoEnableNewVirtualPatches to disable."`

	argparser.JSONOutput
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}
	var err error
	input := &workspaces.CreateInput{
		Description: &c.Description,
		Mode:        &c.BlockingMode,
		Name:        &c.NameFlag,
	}
	if c.AttackThresholds.WasSet {
		input.AttackSignalThresholds, err = parseCreateAttackSignalThresholds(c.AttackThresholds.Value)
		if err != nil {
			return err
		}
	}
	if c.ClientIPHeaders.WasSet {
		input.ClientIPHeaders = strings.Split(c.ClientIPHeaders.Value, ":")
	}
	if c.DefaultBlockingCode.WasSet {
		input.DefaultBlockingResponseCode = &c.DefaultBlockingCode.Value
	}
	if c.DefaultRedirectURL.WasSet {
		input.DefaultRedirectURL = &c.DefaultRedirectURL.Value
	}
	if c.IPAnonymization.WasSet {
		input.IPAnonymization = &c.IPAnonymization.Value
	}
	if c.AutoEnableVirtualPatches.WasSet {
		input.AutoEnableNewVirtualPatches = &c.AutoEnableVirtualPatches.Value
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	data, err := workspaces.Create(context.TODO(), fc, input)
	if err != nil {
		return err
	}

	if ok, err := c.WriteJSON(out, data); ok {
		return err
	}

	text.Success(out, "Created workspace '%s' (workspace-id: %s)", data.Name, data.WorkspaceID)
	return nil
}

func parseCreateAttackSignalThresholds(thresholds string) (*workspaces.AttackSignalThresholdsCreateInput, error) {
	thresholdsArray := strings.Split(thresholds, ":")
	if len(thresholdsArray) != 4 {
		return nil, errors.New("wrong number of inputs for Attack Signal Thresholds")
	}
	immediate, err := strconv.ParseBool(thresholdsArray[0])
	if err != nil {
		return nil, err
	}
	oneMinute, err := strconv.Atoi(thresholdsArray[1])
	if err != nil {
		return nil, err
	}
	tenMinutes, err := strconv.Atoi(thresholdsArray[2])
	if err != nil {
		return nil, err
	}
	oneHour, err := strconv.Atoi(thresholdsArray[3])
	if err != nil {
		return nil, err
	}

	return &workspaces.AttackSignalThresholdsCreateInput{
		OneMinute:  &oneMinute,
		TenMinutes: &tenMinutes,
		OneHour:    &oneHour,
		Immediate:  &immediate,
	}, nil
}
