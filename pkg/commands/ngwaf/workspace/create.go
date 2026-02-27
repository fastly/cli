package workspace

import (
	"context"
	"errors"
	"io"
	"strconv"
	"strings"

	"github.com/fastly/go-fastly/v13/fastly"
	"github.com/fastly/go-fastly/v13/fastly/ngwaf/v1/workspaces"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// CreateCommand calls the Fastly API to create workspaces.
type CreateCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	description  string
	blockingMode string
	name         string

	// Optional.
	attackThresholds    argparser.OptionalString
	defaultBlockingCode argparser.OptionalInt
	defaultRedirectURL  argparser.OptionalString
	clientIPHeaders     argparser.OptionalString
	ipAnonimization     argparser.OptionalString
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent argparser.Registerer, g *global.Data) *CreateCommand {
	c := CreateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("create", "Create a workspace").Alias("add")

	// Required.
	c.CmdClause.Flag("description", "User submitted description of a workspace.").Required().StringVar(&c.description)
	c.CmdClause.Flag("blockingMode", "User configured mode blocking mode.").Required().StringVar(&c.blockingMode)
	c.CmdClause.Flag("name", "User submitted display name of a workspace.").Required().StringVar(&c.name)

	// Optional.
	c.CmdClause.Flag("attackThresholds", "Attack threshold parameters for system site alerts. Each threshold value is the number of attack signals per IP address that must be detected during the interval before the related IP address is flagged. Input accepted as colon separated string: Immediate:OneMinute:TenMinutes:OneHour").Action(c.attackThresholds.Set).StringVar(&c.attackThresholds.Value)
	c.CmdClause.Flag("clientIPHeaders", "Specify the request header containing the client IP address. Input accepted as colon separated string.").Action(c.clientIPHeaders.Set).StringVar(&c.clientIPHeaders.Value)
	c.CmdClause.Flag("defaultBlockingCode", "Default status code that is returned when a request to your web application is blocked.").Action(c.defaultBlockingCode.Set).IntVar(&c.defaultBlockingCode.Value)
	c.CmdClause.Flag("defaultRedirectURL", "Redirect url to be used if code 301 or 302 is used.").Action(c.defaultRedirectURL.Set).StringVar(&c.defaultRedirectURL.Value)
	c.CmdClause.Flag("ipAnonimization", "Agents will anonymize IP addresses according to the option selected.").Action(c.ipAnonimization.Set).StringVar(&c.ipAnonimization.Value)
	c.RegisterFlagBool(c.JSONFlag())

	return &c
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}
	var err error
	input := &workspaces.CreateInput{
		Description: &c.description,
		Mode:        &c.blockingMode,
		Name:        &c.name,
	}
	if c.attackThresholds.WasSet {
		input.AttackSignalThresholds, err = parseCreateAttackSignalThresholds(c.attackThresholds.Value)
		if err != nil {
			return err
		}
	}
	if c.clientIPHeaders.WasSet {
		input.ClientIPHeaders = strings.Split(c.clientIPHeaders.Value, ":")
	}
	if c.defaultBlockingCode.WasSet {
		input.DefaultBlockingResponseCode = &c.defaultBlockingCode.Value
	}
	if c.defaultRedirectURL.WasSet {
		input.DefaultRedirectURL = &c.defaultRedirectURL.Value
	}
	if c.ipAnonimization.WasSet {
		input.IPAnonymization = &c.ipAnonimization.Value
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
