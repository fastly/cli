package rule

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/fastly/go-fastly/v13/fastly"
	"github.com/fastly/go-fastly/v13/fastly/ngwaf/v1/rules"
	"github.com/fastly/go-fastly/v13/fastly/ngwaf/v1/scope"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// CreateCommand calls the Fastly API to create account-level rules.
type CreateCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	path string
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent argparser.Registerer, g *global.Data) *CreateCommand {
	c := CreateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("create", "Create an account-level rule").Alias("add")

	// Required.
	c.CmdClause.Flag("path", "Path to a json file that contains the rule schema.").Required().StringVar(&c.path)

	// Optional.
	c.RegisterFlagBool(c.JSONFlag())

	return &c
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
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
			Type:      scope.ScopeTypeAccount,
			AppliesTo: []string{"*"},
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
				for _, groupCondition := range gc.Conditions {
					switch groupCondition.Type {
					case "single":
						if gsc, ok := groupCondition.Fields.(rules.Condition); ok {
							parsedGroupCondition.Conditions = append(parsedGroupCondition.Conditions, &rules.CreateCondition{
								Field:    &gsc.Field,
								Operator: &gsc.Operator,
								Value:    &gsc.Value,
							})
						} else {
							return fmt.Errorf("expected Condition, got %T", groupCondition.Fields)
						}
					case "multival":
						if gmvc, ok := groupCondition.Fields.(rules.MultivalCondition); ok {
							createMultivalCondition := &rules.CreateMultivalCondition{
								Field:         &gmvc.Field,
								Operator:      &gmvc.Operator,
								GroupOperator: &gmvc.GroupOperator,
								Conditions:    []*rules.CreateConditionMult{},
							}
							for _, groupMultivalSingleCondition := range gmvc.Conditions {
								createMultivalCondition.Conditions = append(createMultivalCondition.Conditions, &rules.CreateConditionMult{
									Field:    &groupMultivalSingleCondition.Field,
									Operator: &groupMultivalSingleCondition.Operator,
									Value:    &groupMultivalSingleCondition.Value,
								})
							}
							parsedGroupCondition.MultivalConditions = append(parsedGroupCondition.MultivalConditions, createMultivalCondition)
						} else {
							return fmt.Errorf("expected MultivalCondition, got %T", groupCondition.Fields)
						}

					default:
						return fmt.Errorf("unknown condition type: %s", groupCondition.Type)
					}
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

	text.Success(out, "Created account-level rule with ID %s", data.RuleID)
	return nil
}
