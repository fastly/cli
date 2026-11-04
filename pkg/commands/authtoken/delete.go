package authtoken

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/text"
)

// DeleteCommand calls the Fastly API to delete an appropriate resource.
type DeleteCommand struct {
	argparser.Base

	Current bool   `help:"Revoke the token used to authenticate the request"`
	File    string `help:"Revoke tokens in bulk from a newline delimited list of tokens"`
	ID      string `help:"Alphanumeric string identifying a token"`
}

// Exec invokes the application logic for the command.
func (c *DeleteCommand) Exec(_ io.Reader, out io.Writer) error {
	if !c.Globals.Flags.Quiet {
		text.Deprecated("The 'auth-token' command tree will be removed in a future release. Use the Fastly API directly to manage API tokens.\n\n")
	}

	if !c.Current && c.File == "" && c.ID == "" {
		return fmt.Errorf("error parsing arguments: must provide either the --current, --file or --id flag")
	}

	if c.Current {
		err := c.Globals.APIClient.DeleteTokenSelf(context.TODO())
		if err != nil {
			c.Globals.ErrLog.Add(err)
			return err
		}

		text.Success(out, "Deleted current token")
		return nil
	}

	if c.File != "" {
		input, err := c.constructInputBatch()
		if err != nil {
			c.Globals.ErrLog.Add(err)
			return err
		}

		err = c.Globals.APIClient.BatchDeleteTokens(context.TODO(), input)
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

	if c.ID != "" {
		input := c.constructInput()

		err := c.Globals.APIClient.DeleteToken(context.TODO(), input)
		if err != nil {
			c.Globals.ErrLog.Add(err)
			return err
		}

		text.Success(out, "Deleted token '%s'", c.ID)
		return nil
	}

	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *DeleteCommand) constructInput() *fastly.DeleteTokenInput {
	var input fastly.DeleteTokenInput

	input.TokenID = c.ID

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

	if path, err = filepath.Abs(c.File); err == nil {
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
