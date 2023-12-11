package purge

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
)

// NewRootCommand returns a new command registered in the parent.
func NewRootCommand(parent argparser.Registerer, g *global.Data) *RootCommand {
	var c RootCommand
	c.CmdClause = parent.Command("purge", "Invalidate objects in the Fastly cache")
	c.Globals = g

	// Optional.
	c.CmdClause.Flag("all", "Purge everything from a service").BoolVar(&c.all)
	c.CmdClause.Flag("file", "Purge a service of a newline delimited list of Surrogate Keys").StringVar(&c.file)
	c.CmdClause.Flag("key", "Purge a service of objects tagged with a Surrogate Key").StringVar(&c.key)
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagServiceIDName,
		Description: argparser.FlagServiceIDDesc,
		Dst:         &g.Manifest.Flag.ServiceID,
		Short:       's',
	})
	c.RegisterFlag(argparser.StringFlagOpts{
		Action:      c.serviceName.Set,
		Name:        argparser.FlagServiceName,
		Description: argparser.FlagServiceDesc,
		Dst:         &c.serviceName.Value,
	})
	c.CmdClause.Flag("soft", "A 'soft' purge marks affected objects as stale rather than making them inaccessible").BoolVar(&c.soft)
	c.CmdClause.Flag("url", "Purge an individual URL").StringVar(&c.url)

	return &c
}

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base

	all         bool
	file        string
	key         string
	serviceName argparser.OptionalServiceNameID
	soft        bool
	url         string
}

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, out io.Writer) error {
	serviceID, source, flag, err := argparser.ServiceID(c.serviceName, *c.Globals.Manifest, c.Globals.APIClient, c.Globals.ErrLog)
	if err != nil {
		return err
	}
	if c.Globals.Verbose() {
		argparser.DisplayServiceID(serviceID, flag, source, out)
	}

	// The URL purge API call doesn't require a Service ID.
	if c.url == "" {
		if source == manifest.SourceUndefined {
			return fsterr.ErrNoServiceID
		}
	}

	if c.all {
		if c.soft {
			return fsterr.RemediationError{
				Inner:       fmt.Errorf("purge-all requests cannot be done in soft mode (--soft) and will always immediately invalidate all cached content associated with the service"),
				Remediation: "The --soft flag should not be used with --all so retry command without it.",
			}
		}
		err := c.purgeAll(serviceID, out)
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
				"Service ID": serviceID,
				"All":        c.all,
			})
			return err
		}
		return nil
	}

	if c.file != "" {
		err := c.purgeKeys(serviceID, out)
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
				"Service ID": serviceID,
				"File":       c.file,
			})
			return err
		}
		return nil
	}

	if c.key != "" {
		err := c.purgeKey(serviceID, out)
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
				"Service ID": serviceID,
				"Key":        c.key,
			})
			return err
		}
		return nil
	}

	if c.url != "" {
		err := c.purgeURL(out)
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
				"URL": c.url,
			})
			return err
		}
		return nil
	}

	return nil
}

func (c *RootCommand) purgeAll(serviceID string, out io.Writer) error {
	p, err := c.Globals.APIClient.PurgeAll(&fastly.PurgeAllInput{
		ServiceID: serviceID,
	})
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID": serviceID,
		})
		return err
	}
	text.Success(out, "Purge all status: %s", p.Status)
	return nil
}

func (c *RootCommand) purgeKeys(serviceID string, out io.Writer) error {
	keys, err := populateKeys(c.file, c.Globals.ErrLog)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID": serviceID,
		})
		return err
	}

	m, err := c.Globals.APIClient.PurgeKeys(&fastly.PurgeKeysInput{
		ServiceID: serviceID,
		Keys:      keys,
		Soft:      c.soft,
	})
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID": serviceID,
			"Keys":       keys,
			"Soft":       c.soft,
		})
		return err
	}

	sortedKeys := make([]string, 0, len(m))
	for k := range m {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	t := text.NewTable(out)
	t.AddHeader("KEY", "ID")
	for _, k := range sortedKeys {
		t.AddLine(k, m[k])
	}
	t.Print()

	return nil
}

func (c *RootCommand) purgeKey(serviceID string, out io.Writer) error {
	p, err := c.Globals.APIClient.PurgeKey(&fastly.PurgeKeyInput{
		ServiceID: serviceID,
		Key:       c.key,
		Soft:      c.soft,
	})
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID": serviceID,
			"Key":        c.key,
			"Soft":       c.soft,
		})
		return err
	}
	text.Success(out, "Purged key: %s (soft: %t). Status: %s, ID: %s", c.key, c.soft, p.Status, p.ID)
	return nil
}

func (c *RootCommand) purgeURL(out io.Writer) error {
	p, err := c.Globals.APIClient.Purge(&fastly.PurgeInput{
		URL:  c.url,
		Soft: c.soft,
	})
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"URL":  c.url,
			"Soft": c.soft,
		})
		return err
	}
	text.Success(out, "Purged URL: %s (soft: %t). Status: %s, ID: %s", c.url, c.soft, p.Status, p.ID)
	return nil
}

// populateKeys opens the given file path, initializes a scanner, and appends
// each line of the file (expected to be a surrogate key) to a slice.
func populateKeys(fpath string, errLog fsterr.LogInterface) (keys []string, err error) {
	var (
		file io.Reader
		path string
	)

	if path, err = filepath.Abs(fpath); err == nil {
		if _, err = os.Stat(path); err == nil {
			if file, err = os.Open(path); err == nil /* #nosec */ {
				scanner := bufio.NewScanner(file)
				for scanner.Scan() {
					keys = append(keys, scanner.Text())
				}
				err = scanner.Err()
			}
		}
	}

	if err != nil {
		errLog.Add(err)
		return nil, err
	}
	return keys, nil
}
