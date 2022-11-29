package secretstore

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"os"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v7/fastly"
)

const (
	// Maximum secret length, as defined at https://developer.fastly.com/reference/api/secret-store.
	maxSecretKiB = 64
	maxSecretLen = maxSecretKiB * 1024
)

// NewCreateSecretCommand returns a usable command registered under the parent.
func NewCreateSecretCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *CreateSecretCommand {
	c := CreateSecretCommand{
		Base: cmd.Base{
			Globals: globals,
		},
		manifest: data,
	}
	
	c.CmdClause = parent.Command("create", "Create secret")
	c.RegisterFlag(storeIDFlag(&c.Input.ID))
	c.RegisterFlag(secretNameFlag(&c.Input.Name))
	c.RegisterFlag(secretFileFlag(&c.secretFile))
	c.RegisterFlagBool(secretStdinFlag(&c.secretSTDIN))
	c.RegisterFlagBool(c.jsonFlag())
	return &c
}

// CreateSecretCommand calls the Fastly API to create a secret.
type CreateSecretCommand struct {
	cmd.Base
	jsonOutput
	
	Input       fastly.CreateSecretInput
	manifest    manifest.Data
	secretFile  string
	secretSTDIN bool
}

var errMultipleSecretValue = fsterr.RemediationError{
	Inner:       fmt.Errorf("invalid flag combination, --secret-file and --secret-stdin"),
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
func (cmd *CreateSecretCommand) Exec(in io.Reader, out io.Writer) error {
	if cmd.Globals.Verbose() && cmd.jsonOutput.enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}
	if cmd.secretFile != "" && cmd.secretSTDIN {
		return errMultipleSecretValue
	}

	// Read secret's value: either from STDIN, a file, or prompt.
	switch {
	case cmd.secretSTDIN:
		// Determine if 'in' has data available.
		if in == nil || text.IsTTY(in) {
			return errNoSTDINData
		}
		var buf bytes.Buffer
		if _, err := buf.ReadFrom(in); err != nil {
			return err
		}		
		cmd.Input.Secret = buf.Bytes()

	case cmd.secretFile != "":
		var err error
		if cmd.Input.Secret, err = os.ReadFile(cmd.secretFile); err != nil {
			return err
		}

	default:
		secret, err := text.InputSecure(out, "Secret: ", in)
		if err != nil {
			return err
		}
		cmd.Input.Secret = []byte(secret)
	}

	if len(cmd.Input.Secret) > maxSecretLen {
		return errMaxSecretLength
	}

	o, err := cmd.Globals.APIClient.CreateSecret(&cmd.Input)
	if err != nil {
		cmd.Globals.ErrLog.Add(err)
		return err
	}

	// TODO: Use this approach across the code base.
	if ok, err := cmd.WriteJSON(out, o); ok {
		return err
	}

	text.Success(out, "Created secret %s in store %s (digest %s)", o.Name, cmd.Input.ID, hex.EncodeToString(o.Digest))

	return nil
}
