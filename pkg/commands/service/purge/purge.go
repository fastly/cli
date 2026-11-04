package purge

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
)

// CommandName is the string to be used to invoke this command.
const CommandName = "purge"

// PurgeCommand calls the Fastly API to purge items from the cache.
type PurgeCommand struct { //revive:disable:exported
	argparser.Base

	// Optional.
	All         bool                            `help:"Purge everything from a service"`
	File        string                          `help:"Purge a service of a newline delimited list of Surrogate Keys"`
	Key         string                          `help:"Purge a service of objects tagged with a Surrogate Key"`
	ServiceID   *string                         `short:"s" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`
	ServiceName argparser.OptionalServiceNameID `help:"The name of the service"`
	Soft        bool                            `help:"A 'soft' purge marks affected objects as stale rather than making them inaccessible"`
	URL         string                          `help:"Purge an individual URL"`
}

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *PurgeCommand) AfterApply() error {
	if c.ServiceID != nil {
		c.Globals.Manifest.Flag.ServiceID = *c.ServiceID
	}
	return nil
}

// Exec implements the command interface.
func (c *PurgeCommand) Exec(_ io.Reader, out io.Writer) error {
	serviceID, source, flag, err := argparser.ServiceID(c.ServiceName, *c.Globals.Manifest, c.Globals.APIClient, c.Globals.ErrLog)
	if err != nil {
		return err
	}
	if c.Globals.Verbose() {
		argparser.DisplayServiceID(serviceID, flag, source, out)
	}

	// The URL purge API call doesn't require a Service ID.
	if c.URL == "" {
		if source == manifest.SourceUndefined {
			return fsterr.ErrNoServiceID
		}
	}

	if c.All {
		if c.Soft {
			return fsterr.RemediationError{
				Inner:       fmt.Errorf("purge-all requests cannot be done in soft mode (--soft) and will always immediately invalidate all cached content associated with the service"),
				Remediation: "The --soft flag should not be used with --all so retry command without it.",
			}
		}
		err := c.purgeAll(serviceID, out)
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
				"Service ID": serviceID,
				"All":        c.All,
			})
			return err
		}
		return nil
	}

	if c.File != "" {
		err := c.purgeKeys(serviceID, out)
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
				"Service ID": serviceID,
				"File":       c.File,
			})
			return err
		}
		return nil
	}

	if c.Key != "" {
		err := c.purgeKey(serviceID, out)
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
				"Service ID": serviceID,
				"Key":        c.Key,
			})
			return err
		}
		return nil
	}

	if c.URL != "" {
		err := c.purgeURL(out)
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
				"URL": c.URL,
			})
			return err
		}
		return nil
	}

	return nil
}

func (c *PurgeCommand) purgeAll(serviceID string, out io.Writer) error {
	p, err := c.Globals.APIClient.PurgeAll(context.TODO(), &fastly.PurgeAllInput{
		ServiceID: serviceID,
	})
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID": serviceID,
		})
		return err
	}
	text.Success(out, "Purge all status: %s", fastly.ToValue(p.Status))
	return nil
}

// purgeKey now uses the bulk purge endpoint to avoid serialization of the 'key' field values.
// This serialization occurs due to the nature of the POST /service/{service_id}/purge/{surrogate_key}
// endpoint storing the 'key' as part of the URL.
func (c *PurgeCommand) purgeKey(serviceID string, out io.Writer) error {
	m, err := c.Globals.APIClient.PurgeKeys(context.TODO(), &fastly.PurgeKeysInput{
		ServiceID: serviceID,
		Keys:      []string{c.Key},
		Soft:      c.Soft,
	})
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID": serviceID,
			"Key":        c.Key,
			"Soft":       c.Soft,
		})
		return err
	}
	purgeID, ok := m[c.Key]
	if !ok {
		return fmt.Errorf("no purge ID returned for key: %s", c.Key)
	}
	// The bulk purge endpoint doesn't return a 'Status' field like the single-key
	// endpoint did. To avoid a breaking change in the CLI output, we hardcode
	// 'Status: ok' in the success message to maintain consistent behavior.
	text.Success(out, "Purged key: %s (soft: %t). Status: ok, ID: %s", c.Key, c.Soft, purgeID)
	return nil
}

func (c *PurgeCommand) purgeKeys(serviceID string, out io.Writer) error {
	keys, err := populateKeys(c.File, c.Globals.ErrLog)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID": serviceID,
		})
		return err
	}

	return c.purgeBulkKeys(serviceID, keys, out)
}

func (c *PurgeCommand) purgeBulkKeys(serviceID string, keys []string, out io.Writer) error {
	m, err := c.Globals.APIClient.PurgeKeys(context.TODO(), &fastly.PurgeKeysInput{
		ServiceID: serviceID,
		Keys:      keys,
		Soft:      c.Soft,
	})
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID": serviceID,
			"Keys":       keys,
			"Soft":       c.Soft,
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

func (c *PurgeCommand) purgeURL(out io.Writer) error {
	p, err := c.Globals.APIClient.Purge(context.TODO(), &fastly.PurgeInput{
		URL:  c.URL,
		Soft: c.Soft,
	})
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"URL":  c.URL,
			"Soft": c.Soft,
		})
		return err
	}
	text.Success(out, "Purged URL: %s (soft: %t). Status: %s, ID: %s", c.URL, c.Soft, fastly.ToValue(p.Status), fastly.ToValue(p.PurgeID))
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
