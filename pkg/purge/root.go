package purge

import (
	"bufio"
	"io"
	"os"
	"path/filepath"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v3/fastly"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	cmd.Base

	all      bool
	file     string
	key      string
	manifest manifest.Data
	soft     bool
	url      string
}

// NewRootCommand returns a new command registered in the parent.
func NewRootCommand(parent cmd.Registerer, globals *config.Data) *RootCommand {
	var c RootCommand
	c.CmdClause = parent.Command("purge", "Remove an object from the Fastly cache")
	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)

	// Optional flags
	c.CmdClause.Flag("all", "Purge everything from a service").BoolVar(&c.all)
	c.CmdClause.Flag("file", "Purge a service with a line separated list of Surrogate Keys").StringVar(&c.file)
	c.CmdClause.Flag("key", "Purge a service of items tagged with a Surrogate Key").StringVar(&c.key)
	c.RegisterServiceIDFlag(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("soft", "A 'soft' purge marks the affected object as stale rather than making it inaccessible").BoolVar(&c.soft)
	c.CmdClause.Flag("url", "Purge an individual URL").StringVar(&c.url)

	return &c
}

// Exec implements the command interface.
func (c *RootCommand) Exec(in io.Reader, out io.Writer) error {
	// Exit early if no token configured.
	_, s := c.Globals.Token()
	if s == config.SourceUndefined {
		return errors.ErrNoToken
	}

	// The URL purge API call doesn't require a Service ID.
	var serviceID string
	var source manifest.Source
	if c.url == "" {
		serviceID, source = c.manifest.ServiceID()
		if source == manifest.SourceUndefined {
			return errors.ErrNoServiceID
		}
	}

	if c.all {
		err := c.purgeAll(serviceID, out)
		if err != nil {
			return err
		}
		return nil
	}

	if c.file != "" {
		err := c.purgeKeys(serviceID, out)
		if err != nil {
			return err
		}
		return nil
	}

	if c.key != "" {
		err := c.purgeKey(serviceID, out)
		if err != nil {
			return err
		}
		return nil
	}

	if c.url != "" {
		err := c.purgeURL(serviceID, out)
		if err != nil {
			return err
		}
		return nil
	}

	return nil
}

func (c *RootCommand) purgeAll(serviceID string, out io.Writer) error {
	p, err := c.Globals.Client.PurgeAll(&fastly.PurgeAllInput{
		ServiceID: serviceID,
	})
	if err != nil {
		return err
	}
	text.Success(out, "Purge all status: %s", p.Status)
	return nil
}

func (c *RootCommand) purgeKeys(serviceID string, out io.Writer) error {
	keys, err := populateKeys(c.file)
	if err != nil {
		return err
	}

	m, err := c.Globals.Client.PurgeKeys(&fastly.PurgeKeysInput{
		ServiceID: serviceID,
		Keys:      keys,
		Soft:      c.soft,
	})
	if err != nil {
		return err
	}

	t := text.NewTable(out)
	t.AddHeader("KEY", "ID")
	for k, v := range m {
		t.AddLine(k, v)
	}
	t.Print()

	return nil
}

func (c *RootCommand) purgeKey(serviceID string, out io.Writer) error {
	p, err := c.Globals.Client.PurgeKey(&fastly.PurgeKeyInput{
		ServiceID: serviceID,
		Key:       c.key,
		Soft:      c.soft,
	})
	if err != nil {
		return err
	}
	text.Success(out, "Purged key: %s (soft: %t). Status: %s, ID: %s", c.key, c.soft, p.Status, p.ID)
	return nil
}

func (c *RootCommand) purgeURL(serviceID string, out io.Writer) error {
	p, err := c.Globals.Client.Purge(&fastly.PurgeInput{
		URL:  c.url,
		Soft: c.soft,
	})
	if err != nil {
		return err
	}
	text.Success(out, "Purged url: %s (soft: %t). Status: %s, ID: %s", c.url, c.soft, p.Status, p.ID)
	return nil
}

// populateKeys opens the given file path, initializes a scanner, and appends
// each line of the file (expected to be a surrogate key) to a slice.
func populateKeys(fpath string) (keys []string, err error) {
	var (
		file io.Reader
		path string
	)

	if path, err = filepath.Abs(fpath); err == nil {
		if _, err = os.Stat(path); err == nil {
			// gosec flagged this:
			// G304 (CWE-22): Potential file inclusion via variable
			// Disabling as we trust the source of the fpath variable.
			/* #nosec */
			if file, err = os.Open(path); err == nil {
				scanner := bufio.NewScanner(file)
				for scanner.Scan() {
					keys = append(keys, scanner.Text())
				}
				err = scanner.Err()
			}
		}
	}

	if err != nil {
		return nil, err
	}
	return keys, nil
}
