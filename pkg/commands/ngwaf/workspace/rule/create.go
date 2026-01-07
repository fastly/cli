package rule

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/fastly/go-fastly/v12/fastly"
	"github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/rules"
	"github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/scope"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// CreateCommand calls the Fastly API to create workspace-level rules.
type CreateCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	path        string
	workspaceID argparser.OptionalWorkspaceID
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent argparser.Registerer, g *global.Data) *CreateCommand {
	c := CreateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("create", "Create a workspace-level rule").Alias("add")

	// Required.
	c.CmdClause.Flag("path", "Path to a json file that contains the rule schema.").Required().StringVar(&c.path)
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagNGWAFWorkspaceID,
		Description: argparser.FlagNGWAFWorkspaceIDDesc,
		Dst:         &c.workspaceID.Value,
		Action:      c.workspaceID.Set,
	})

	// Optional.
	c.RegisterFlagBool(c.JSONFlag())

	return &c
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}
	if err := c.workspaceID.Parse(); err != nil {
		return err
	}
	rule := &rules.Rule{}
	if c.path != "" {
		path, err := filepath.Abs(c.path)
		if err != nil {
			return fmt.Errorf("error parsing path '%s': %q", c.path, err)
		}

		jsonFile, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("error reading path '%s': %q", c.path, err)
		}
		defer jsonFile.Close()

		byteValue, err := io.ReadAll(jsonFile)
		if err != nil {
			return fmt.Errorf("failed to read json file: %v", err)
		}

		if err := json.Unmarshal(byteValue, rule); err != nil {
			return fmt.Errorf("failed to unmarshal json data: %v", err)
		}
	}

	input := &rules.CreateInput{
		Actions:            []*rules.CreateAction{},
		Conditions:         []*rules.CreateCondition{},
		Description:        &rule.Description,
		GroupConditions:    []*rules.CreateGroupCondition{},
		MultivalConditions: []*rules.CreateMultivalCondition{},
		Enabled:            &rule.Enabled,
		Type:               &rule.Type,
		GroupOperator:      &rule.GroupOperator,
		RequestLogging:     &rule.RequestLogging,
		Scope: &scope.Scope{
			Type:      scope.ScopeTypeWorkspace,
			AppliesTo: []string{c.workspaceID.Value},
		},
	}

	for _, action := range rule.Actions {
		input.Actions = append(input.Actions, &rules.CreateAction{
			AllowInteractive: action.AllowInteractive,
			DeceptionType:    &action.DeceptionType,
			RedirectURL:      &action.RedirectURL,
			ResponseCode:     &action.ResponseCode,
			Signal:           &action.Signal,
			Type:             &action.Type,
		})
	}

	if rule.RateLimit != nil {
		input.RateLimit = &rules.CreateRateLimit{
			ClientIdentifiers: []*rules.CreateClientIdentifier{},
			Duration:          &rule.RateLimit.Duration,
			Interval:          &rule.RateLimit.Interval,
			Signal:            &rule.RateLimit.Signal,
			Threshold:         &rule.RateLimit.Threshold,
		}

		for _, rateLimit := range rule.RateLimit.ClientIdentifiers {
			input.RateLimit.ClientIdentifiers = append(input.RateLimit.ClientIdentifiers, &rules.CreateClientIdentifier{
				Key:  &rateLimit.Key,
				Name: &rateLimit.Name,
				Type: &rateLimit.Type,
			})
		}
	}

	for _, jsonCondition := range rule.Conditions {
		switch jsonCondition.Type {
		case "single":
			if sc, ok := jsonCondition.Fields.(rules.SingleCondition); ok {
				input.Conditions = append(input.Conditions, &rules.CreateCondition{
					Field:    &sc.Field,
					Operator: &sc.Operator,
					Value:    &sc.Value,
				})
			} else {
				return fmt.Errorf("expected SingleCondition, got %T", jsonCondition.Fields)
			}
		case "group":
			if gc, ok := jsonCondition.Fields.(rules.GroupCondition); ok {
				parsedGroupCondition := &rules.CreateGroupCondition{
					GroupOperator: &gc.GroupOperator,
					Conditions:    []*rules.CreateCondition{},
				}
				for _, groupSingleCondition := range gc.Conditions {
					parsedGroupCondition.Conditions = append(parsedGroupCondition.Conditions, &rules.CreateCondition{
						Field:    &groupSingleCondition.Field,
						Operator: &groupSingleCondition.Operator,
						Value:    &groupSingleCondition.Value,
					})
				}
				input.GroupConditions = append(input.GroupConditions, parsedGroupCondition)
			} else {
				return fmt.Errorf("expected GroupCondition, got %T", jsonCondition.Fields)
			}
		case "multival":
			if mvc, ok := jsonCondition.Fields.(rules.CreateMultivalCondition); ok {
				parsedMultiValCondition := &rules.CreateMultivalCondition{
					Field:         mvc.Field,
					GroupOperator: mvc.GroupOperator,
					Operator:      mvc.Operator,
					Conditions:    []*rules.CreateConditionMult{},
				}
				for _, multiSingleCondition := range mvc.Conditions {
					parsedMultiValCondition.Conditions = append(parsedMultiValCondition.Conditions, &rules.CreateConditionMult{
						Field:    multiSingleCondition.Field,
						Operator: multiSingleCondition.Operator,
						Value:    multiSingleCondition.Value,
					})
				}
				input.MultivalConditions = append(input.MultivalConditions, parsedMultiValCondition)
			} else {
				return fmt.Errorf("expected MultivalCondition, got %T", jsonCondition.Fields)
			}
		default:
			return fmt.Errorf("unknown condition type: %s", jsonCondition.Type)
		}
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	data, err := rules.Create(context.TODO(), fc, input)
	if err != nil {
		return err
	}

	if ok, err := c.WriteJSON(out, data); ok {
		return err
	}

	text.Success(out, "Created workspace-level rule with ID %s", data.RuleID)
	return nil
}
