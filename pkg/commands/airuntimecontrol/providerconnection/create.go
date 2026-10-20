package providerconnection

import (
	"context"
	"errors"
	"fmt"
	"io"
	"slices"

	"github.com/fastly/kingpin"

	"github.com/fastly/go-fastly/v17/fastly"
	"github.com/fastly/go-fastly/v17/fastly/airuntimecontrol/v1/providerconnection"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// CreateCommand calls the Fastly API to create a provider connection.
type CreateCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	name   string
	models []string

	// Optional.
	authType            argparser.OptionalString
	baseURL             argparser.OptionalString
	apiKey              argparser.OptionalString
	customerRoleARN     argparser.OptionalString
	region              argparser.OptionalString
	poolID              argparser.OptionalString
	workloadProviderID  argparser.OptionalString
	serviceAccountEmail argparser.OptionalString
	projectID           argparser.OptionalString
	projectNumber       argparser.OptionalString
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent argparser.Registerer, g *global.Data) *CreateCommand {
	c := CreateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("create", "Create a provider connection").Alias("add")

	// Required.
	c.CmdClause.Flag("name", "Human-readable name of the provider connection. For Bedrock or Vertex AI connections, use 'bedrock' or 'vertex' respectively").HintOptions(providerNameBedrock, providerNameVertex).Required().StringVar(&c.name)
	c.CmdClause.Flag("models", "Comma-separated list of allowed AI model identifiers").Required().StringsVar(&c.models, kingpin.Separator(","))

	// Optional.
	c.CmdClause.Flag("auth-type", "Authentication type for the connection. One of: api-key (default), aws-iam (Bedrock), gcp-iam (Vertex AI)").Default(authTypeAPIKey).HintOptions(authTypes...).Action(c.authType.Set).EnumVar(&c.authType.Value, authTypes...)
	c.CmdClause.Flag("base-url", "Base URL for the provider's API. Must be an absolute http or https URL (required for api-key)").Action(c.baseURL.Set).StringVar(&c.baseURL.Value)
	c.CmdClause.Flag("api-key", "Provider's secret key for authentication (required for api-key)").Action(c.apiKey.Set).StringVar(&c.apiKey.Value)
	c.CmdClause.Flag("customer-role-arn", "ARN of the IAM role in your AWS account that ARC assumes to access Bedrock (required for Bedrock)").Action(c.customerRoleARN.Set).StringVar(&c.customerRoleARN.Value)
	c.CmdClause.Flag("region", "AWS region for Bedrock or GCP region for Vertex AI (required for Bedrock and Vertex AI)").Action(c.region.Set).StringVar(&c.region.Value)
	c.CmdClause.Flag("pool-id", "GCP workload identity pool ID (required for Vertex AI)").Action(c.poolID.Set).StringVar(&c.poolID.Value)
	c.CmdClause.Flag("workload-provider-id", "GCP workload identity provider ID (required for Vertex AI)").Action(c.workloadProviderID.Set).StringVar(&c.workloadProviderID.Value)
	c.CmdClause.Flag("service-account-email", "Email of the GCP service account to impersonate (required for Vertex AI)").Action(c.serviceAccountEmail.Set).StringVar(&c.serviceAccountEmail.Value)
	c.CmdClause.Flag("project-id", "GCP project ID (required for Vertex AI)").Action(c.projectID.Set).StringVar(&c.projectID.Value)
	c.CmdClause.Flag("project-number", "GCP project number (required for Vertex AI)").Action(c.projectNumber.Set).StringVar(&c.projectNumber.Value)
	c.RegisterFlagBool(c.JSONFlag())

	return &c
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	if err := c.validate(); err != nil {
		return err
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	input := &providerconnection.CreateInput{
		Name:   &c.name,
		Models: c.models,
	}
	if c.authType.WasSet {
		input.AuthType = &c.authType.Value
	}
	if c.baseURL.WasSet {
		input.BaseURL = &c.baseURL.Value
	}
	if c.apiKey.WasSet {
		input.APIKey = &c.apiKey.Value
	}
	if c.customerRoleARN.WasSet {
		input.CustomerRoleARN = &c.customerRoleARN.Value
	}
	if c.region.WasSet {
		input.Region = &c.region.Value
	}
	if c.poolID.WasSet {
		input.PoolID = &c.poolID.Value
	}
	if c.workloadProviderID.WasSet {
		input.WorkloadProviderID = &c.workloadProviderID.Value
	}
	if c.serviceAccountEmail.WasSet {
		input.ServiceAccountEmail = &c.serviceAccountEmail.Value
	}
	if c.projectID.WasSet {
		input.ProjectID = &c.projectID.Value
	}
	if c.projectNumber.WasSet {
		input.ProjectNumber = &c.projectNumber.Value
	}

	data, err := providerconnection.Create(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, data); ok {
		return err
	}

	text.PrintProviderConnection(out, data)
	return nil
}

// validate checks that the flags provided match the selected auth type.
func (c *CreateCommand) validate() error {
	authType := c.authType.Value

	switch authType {
	case authTypeAWSIAM:
		if c.name != providerNameBedrock {
			return fmt.Errorf("error parsing arguments: --name must be %s for --auth-type %s", providerNameBedrock, authType)
		}
	case authTypeGCPIAM:
		if c.name != providerNameVertex {
			return fmt.Errorf("error parsing arguments: --name must be %s for --auth-type %s", providerNameVertex, authType)
		}
	}

	// Each auth-specific flag is required for the auth types it applies to,
	// and not supported for the others.
	authFlags := []struct {
		name      string
		set       bool
		authTypes []string
	}{
		{"base-url", c.baseURL.WasSet, []string{authTypeAPIKey}},
		{"api-key", c.apiKey.WasSet, []string{authTypeAPIKey}},
		{"customer-role-arn", c.customerRoleARN.WasSet, []string{authTypeAWSIAM}},
		{"region", c.region.WasSet, []string{authTypeAWSIAM, authTypeGCPIAM}},
		{"pool-id", c.poolID.WasSet, []string{authTypeGCPIAM}},
		{"workload-provider-id", c.workloadProviderID.WasSet, []string{authTypeGCPIAM}},
		{"service-account-email", c.serviceAccountEmail.WasSet, []string{authTypeGCPIAM}},
		{"project-id", c.projectID.WasSet, []string{authTypeGCPIAM}},
		{"project-number", c.projectNumber.WasSet, []string{authTypeGCPIAM}},
	}
	for _, f := range authFlags {
		applies := slices.Contains(f.authTypes, authType)
		if applies && !f.set {
			return fmt.Errorf("error parsing arguments: required flag --%s not provided for --auth-type %s", f.name, authType)
		}
		if !applies && f.set {
			return fmt.Errorf("error parsing arguments: --%s is not supported with --auth-type %s", f.name, authType)
		}
	}
	return nil
}
