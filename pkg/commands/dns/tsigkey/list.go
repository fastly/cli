package tsigkey

import (
	"context"
	"errors"
	"fmt"
	"io"
	"slices"

	"github.com/alecthomas/kong"

	"github.com/fastly/go-fastly/v19/fastly"
	"github.com/fastly/go-fastly/v19/fastly/dns/v1/tsigkeys"

	"github.com/fastly/cli/internal/kongcompat"
	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// ListCommand calls the Fastly API to list TSIG keys.
type ListCommand struct {
	argparser.Base

	// Optional.
	NameFlag argparser.OptionalString `name:"name" help:"Filter TSIG keys to only those containing the provided name."`
	Sort     argparser.OptionalString `hints:"name_asc,name_desc,created_at_asc,created_at_desc" help:"Order in which to list results. Valid values are: name_asc, name_desc, created_at_asc, created_at_desc."`
	argparser.JSONOutput
}

// AfterApply checks the values kong cannot check by itself.
func (c *ListCommand) AfterApply(ctx *kong.Context) error {
	if (c.Sort.Value != "" || kongcompat.Given(ctx, "sort")) && !slices.Contains([]string{"name_asc", "name_desc", "created_at_asc", "created_at_desc"}, c.Sort.Value) {
		return fmt.Errorf("enum value must be one of name_asc,name_desc,created_at_asc,created_at_desc, got '%s'", c.Sort.Value)
	}
	return nil
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	input := &tsigkeys.ListInput{}
	if c.NameFlag.WasSet {
		input.Name = &c.NameFlag.Value
	}
	if c.Sort.WasSet {
		v := sortAPIValue[c.Sort.Value]
		input.Sort = &v
	}

	keys, err := tsigkeys.List(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, keys); ok {
		return err
	}

	text.PrintTSIGKeyTbl(out, keys)
	return nil
}
