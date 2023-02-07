package secretstoreentry

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"os"

	"github.com/fastly/cli/pkg/cmd"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v7/fastly"
)

const (
	// Maximum secret length, as defined at https://developer.fastly.com/reference/api/secret-store.
	maxSecretKiB = 64
	maxSecretLen = maxSecretKiB * 1024
)

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent cmd.Registerer, g *global.Data, m manifest.Data) *CreateCommand {
	c := CreateCommand{
		Base: cmd.Base{
			Globals: g,
		},
		manifest: m,
	}

	c.CmdClause = parent.Command("create", "Create a new secret within specified store")

	// Required.
	c.RegisterFlag(secretNameFlag(&c.Input.Name)) // --name
	c.RegisterFlag(cmd.StoreIDFlag(&c.Input.ID))  // --store-id

	// Optional.
	c.RegisterFlag(secretFileFlag(&c.secretFile))       // --file
	c.RegisterFlagBool(c.JSONFlag())                    // --json
	c.RegisterFlagBool(secretStdinFlag(&c.secretSTDIN)) // --stdin

	return &c
}

// CreateCommand calls the Fastly API to create an appropriate resource.
type CreateCommand struct {
	cmd.Base
	cmd.JSONOutput

	Input       fastly.CreateSecretInput
	manifest    manifest.Data
	secretFile  string
	secretSTDIN bool
}

var errMultipleSecretValue = fsterr.RemediationError{
	Inner:       fmt.Errorf("invalid flag combination, --file and --stdin"),
	Remediation: "Use one of --file or --stdin flag",
}

var errNoSTDINData = fsterr.RemediationError{
	Inner:       fmt.Errorf("unable to read from STDIN"),
	Remediation: "Provide data to STDIN, or use --file to read from a file",
}

var errMaxSecretLength = fsterr.RemediationError{
	Inner:       fmt.Errorf("max secret size exceeded"),
	Remediation: fmt.Sprintf("Maximum secret size is %dKiB", maxSecretKiB),
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(in io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}
	if c.secretFile != "" && c.secretSTDIN {
		return errMultipleSecretValue
	}

	// Read secret's value: either from STDIN, a file, or prompt.
	switch {
	case c.secretSTDIN:
		// Determine if 'in' has data available.
		if in == nil || text.IsTTY(in) {
			return errNoSTDINData
		}
		var buf bytes.Buffer
		if _, err := buf.ReadFrom(in); err != nil {
			return err
		}
		c.Input.Secret = buf.Bytes()

	case c.secretFile != "":
		var err error
		if c.Input.Secret, err = os.ReadFile(c.secretFile); err != nil {
			return err
		}

	default:
		secret, err := text.InputSecure(out, "Secret: ", in)
		if err != nil {
			return err
		}
		c.Input.Secret = []byte(secret)
	}

	if len(c.Input.Secret) > maxSecretLen {
		return errMaxSecretLength
	}

	o, err := c.Globals.APIClient.CreateSecret(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	// TODO: Use this approach across the code base.
	if ok, err := c.WriteJSON(out, o); ok {
		return err
	}

	text.Success(out, "Created secret %s in store %s (digest %s)", o.Name, c.Input.ID, hex.EncodeToString(o.Digest))

	return nil
}
