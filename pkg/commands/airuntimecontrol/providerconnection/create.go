package providerconnection

import (
	"context"
	"errors"
	"io"

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
	name    string
	models  []string
	baseURL string
	apiKey  string
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
	c.CmdClause.Flag("name", "Human-readable name of the provider").Required().StringVar(&c.name)
	c.CmdClause.Flag("models", "Comma-separated list of allowed AI model identifiers").Required().StringsVar(&c.models, kingpin.Separator(","))
	c.CmdClause.Flag("base-url", "Base URL for the provider's API").Required().StringVar(&c.baseURL)
	c.CmdClause.Flag("api-key", "Provider's secret key for authentication").Required().StringVar(&c.apiKey)

	// Optional.
	c.RegisterFlagBool(c.JSONFlag())

	return &c
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	data, err := providerconnection.Create(context.TODO(), fc, &providerconnection.CreateInput{
		Name:    &c.name,
		Models:  c.models,
		BaseURL: &c.baseURL,
		APIKey:  &c.apiKey,
	})
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
