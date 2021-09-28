package authtoken

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/commands/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v5/fastly"
)

// NewDeleteCommand returns a usable command registered under the parent.
func NewDeleteCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *DeleteCommand {
	var c DeleteCommand
	c.CmdClause = parent.Command("delete", "Revoke an API token").Alias("remove")
	c.Globals = globals
	c.manifest = data
	c.CmdClause.Flag("current", "Revoke the token used to authenticate the request").BoolVar(&c.current)
	c.CmdClause.Flag("file", "Revoke tokens in bulk from a newline delimited list of tokens").StringVar(&c.file)
	c.CmdClause.Flag("id", "Alphanumeric string identifying a token").StringVar(&c.id)
	return &c
}

// DeleteCommand calls the Fastly API to delete an appropriate resource.
type DeleteCommand struct {
	cmd.Base

	current  bool
	file     string
	id       string
	manifest manifest.Data
}

// Exec invokes the application logic for the command.
func (c *DeleteCommand) Exec(in io.Reader, out io.Writer) error {
	// Exit early if no token configured.
	_, s := c.Globals.Token()
	if s == config.SourceUndefined {
		return errors.ErrNoToken
	}

	if !c.current && c.file == "" && c.id == "" {
		return fmt.Errorf("error parsing arguments: must provide either the --current, --file or --id flag")
	}

	if c.current {
		err := c.Globals.Client.DeleteTokenSelf()
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

		err = c.Globals.Client.BatchDeleteTokens(input)
		if err != nil {
			c.Globals.ErrLog.Add(err)
			return err
		}

		text.Success(out, "Deleted tokens")
		if c.Globals.Verbose() {
			c.printTokens(out, input.Tokens)
		}
		return nil
	}

	if c.id != "" {
		input := c.constructInput()

		err := c.Globals.Client.DeleteToken(input)
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
			// gosec flagged this:
			// G304 (CWE-22): Potential file inclusion via variable
			// Disabling as we trust the source of the path variable.
			/* #nosec */
			if file, err = os.Open(path); err == nil {
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
	fmt.Fprintf(out, "\n")
	t := text.NewTable(out)
	t.AddHeader("TOKEN ID")
	for _, r := range rs {
		t.AddLine(r.ID)
	}
	t.Print()
}
