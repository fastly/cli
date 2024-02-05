package authtoken

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// NewDeleteCommand returns a usable command registered under the parent.
func NewDeleteCommand(parent argparser.Registerer, g *global.Data) *DeleteCommand {
	c := DeleteCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("delete", "Revoke an API token").Alias("remove")

	c.CmdClause.Flag("current", "Revoke the token used to authenticate the request").BoolVar(&c.current)
	c.CmdClause.Flag("file", "Revoke tokens in bulk from a newline delimited list of tokens").StringVar(&c.file)
	c.CmdClause.Flag("id", "Alphanumeric string identifying a token").StringVar(&c.id)
	return &c
}

// DeleteCommand calls the Fastly API to delete an appropriate resource.
type DeleteCommand struct {
	argparser.Base

	current bool
	file    string
	id      string
}

// Exec invokes the application logic for the command.
func (c *DeleteCommand) Exec(_ io.Reader, out io.Writer) error {
	if !c.current && c.file == "" && c.id == "" {
		return fmt.Errorf("error parsing arguments: must provide either the --current, --file or --id flag")
	}

	if c.current {
		err := c.Globals.APIClient.DeleteTokenSelf()
		if err != nil {
			c.Globals.ErrLog.Add(err)
			return err
		}

		text.Success(out, "Deleted current token")
		return nil
	}

	if c.file != "" {
		input, err := c.constructInputBatch()
		if err != nil {
			c.Globals.ErrLog.Add(err)
			return err
		}

		err = c.Globals.APIClient.BatchDeleteTokens(input)
		if err != nil {
			c.Globals.ErrLog.Add(err)
			return err
		}

		text.Success(out, "Deleted tokens")
		if c.Globals.Verbose() {
			text.Break(out)
			c.printTokens(out, input.Tokens)
		}
		return nil
	}

	if c.id != "" {
		input := c.constructInput()

		err := c.Globals.APIClient.DeleteToken(input)
		if err != nil {
			c.Globals.ErrLog.Add(err)
			return err
		}

		text.Success(out, "Deleted token '%s'", c.id)
		return nil
	}

	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *DeleteCommand) constructInput() *fastly.DeleteTokenInput {
	var input fastly.DeleteTokenInput

	input.TokenID = c.id

	return &input
}

// constructInputBatch transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *DeleteCommand) constructInputBatch() (*fastly.BatchDeleteTokensInput, error) {
	var (
		err    error
		file   io.Reader
		input  fastly.BatchDeleteTokensInput
		path   string
		tokens []*fastly.BatchToken
	)

	if path, err = filepath.Abs(c.file); err == nil {
		if _, err = os.Stat(path); err == nil {
			if file, err = os.Open(path); err == nil /* #nosec */ {
				scanner := bufio.NewScanner(file)
				for scanner.Scan() {
					tokens = append(tokens, &fastly.BatchToken{ID: scanner.Text()})
				}
				err = scanner.Err()
			}
		}
	}

	input.Tokens = tokens

	if err != nil {
		return nil, err
	}

	return &input, nil
}

// printTokens displays the tokens provided by a user.
func (c *DeleteCommand) printTokens(out io.Writer, rs []*fastly.BatchToken) {
	t := text.NewTable(out)
	t.AddHeader("TOKEN ID")
	for _, r := range rs {
		t.AddLine(r.ID)
	}
	t.Print()
}
