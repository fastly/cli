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

// UpdateCommand calls the Fastly API to update an account-level rule.
type UpdateCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	path   string
	ruleID string
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent argparser.Registerer, g *global.Data) *UpdateCommand {
	c := UpdateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("update", "Update a workspace")

	// Required.
	c.CmdClause.Flag("rule-id", "Rule ID").Required().StringVar(&c.ruleID)
	c.CmdClause.Flag("path", "Path to a json file that contains the rule schema.").Required().StringVar(&c.path)

	// Optional.
	c.RegisterFlagBool(c.JSONFlag())

	return &c
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}
	var err error

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

	input := &rules.UpdateInput{
		RuleID:             &c.ruleID,
		Actions:            []*rules.UpdateAction{},
		Conditions:         []*rules.UpdateCondition{},
		Description:        &rule.Description,
		GroupConditions:    []*rules.UpdateGroupCondition{},
		MultivalConditions: []*rules.UpdateMultivalCondition{},
		Enabled:            &rule.Enabled,
		Type:               &rule.Type,
		GroupOperator:      &rule.GroupOperator,
		RequestLogging:     &rule.RequestLogging,
		Scope: &scope.Scope{
			Type:      scope.ScopeTypeAccount,
			AppliesTo: []string{"*"},
		},
	}

	for _, action := range rule.Actions {
		input.Actions = append(input.Actions, &rules.UpdateAction{
			AllowInteractive: action.AllowInteractive,
			DeceptionType:    &action.DeceptionType,
			RedirectURL:      &action.RedirectURL,
			ResponseCode:     &action.ResponseCode,
			Signal:           &action.Signal,
			Type:             &action.Type,
		})
	}

	if rule.RateLimit != nil {
		input.RateLimit = &rules.UpdateRateLimit{
			ClientIdentifiers: []*rules.UpdateClientIdentifier{},
			Duration:          &rule.RateLimit.Duration,
			Interval:          &rule.RateLimit.Interval,
			Signal:            &rule.RateLimit.Signal,
			Threshold:         &rule.RateLimit.Threshold,
		}

		for _, rateLimit := range rule.RateLimit.ClientIdentifiers {
			input.RateLimit.ClientIdentifiers = append(input.RateLimit.ClientIdentifiers, &rules.UpdateClientIdentifier{
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
				input.Conditions = append(input.Conditions, &rules.UpdateCondition{
					Field:    &sc.Field,
					Operator: &sc.Operator,
					Value:    &sc.Value,
				})
			} else {
				return fmt.Errorf("expected SingleCondition, got %T", jsonCondition.Fields)
			}
		case "group":
			if gc, ok := jsonCondition.Fields.(rules.GroupCondition); ok {
				parsedGroupCondition := &rules.UpdateGroupCondition{
					GroupOperator: &gc.GroupOperator,
					Conditions:    []*rules.UpdateCondition{},
				}
				for _, groupSingleCondition := range gc.Conditions {
					parsedGroupCondition.Conditions = append(parsedGroupCondition.Conditions, &rules.UpdateCondition{
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
			if mvc, ok := jsonCondition.Fields.(rules.UpdateMultivalCondition); ok {
				parsedMultiValCondition := &rules.UpdateMultivalCondition{
					Field:         mvc.Field,
					GroupOperator: mvc.GroupOperator,
					Operator:      mvc.Operator,
					Conditions:    []*rules.UpdateConditionMult{},
				}
				for _, multiSingleCondition := range mvc.Conditions {
					parsedMultiValCondition.Conditions = append(parsedMultiValCondition.Conditions, &rules.UpdateConditionMult{
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

	data, err := rules.Update(context.TODO(), fc, input)
	if err != nil {
		return err
	}

	if ok, err := c.WriteJSON(out, data); ok {
		return err
	}

	text.Success(out, "Updated account-level rule with id: %s", data.RuleID)
	return nil
}
