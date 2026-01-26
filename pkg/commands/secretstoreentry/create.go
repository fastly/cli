package secretstoreentry

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/fastly/go-fastly/v12/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

const (
	// Maximum secret length, as defined at https://www.fastly.com/documentation/reference/api/services/resources/secret-store-secret
	maxSecretKiB = 64
	maxSecretLen = maxSecretKiB * 1024
)

// verificationKey is Fastly's Ed25519 public key, used to verify signatures
// on client keys returned by the API. Fastly holds the corresponding private
// key and signs client keys on their side.
//
// This key is meant to be long-lived and infrequently (if ever) rotated.
// Hardcoding it in the CLI provides a trust anchor that prevents MITM attacks
// where an attacker could substitute a different key.
//
// When Fastly rotates it, we will need to update this value and release a
// new version of the CLI. Users can also override this check with
// the FASTLY_USE_API_SIGNING_KEY environment variable.
var verificationKey = mustDecode("CrO/A92vkxEZjtTW7D/Sr+1EMf/q9BahC0sfLkWa+0k=")

func mustDecode(s string) []byte {
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return b
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent argparser.Registerer, g *global.Data) *CreateCommand {
	c := CreateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}

	c.CmdClause = parent.Command("create", "Create a new secret within specified store")

	// Required.
	c.RegisterFlag(secretNameFlag(&c.Input.Name))           // --name
	c.RegisterFlag(argparser.StoreIDFlag(&c.Input.StoreID)) // --store-id

	// Optional.
	c.RegisterFlag(secretFileFlag(&c.secretFile)) // --file
	c.RegisterFlagBool(c.JSONFlag())              // --json
	c.RegisterFlagBool(argparser.BoolFlagOpts{
		Name:        "recreate",
		Description: "Recreate secret by name (errors if secret doesn't already exist)",
		Dst:         &c.recreate,
		Required:    false,
	})
	c.RegisterFlagBool(argparser.BoolFlagOpts{
		Name:        "recreate-allow",
		Description: "Create or recreate secret by name",
		Dst:         &c.recreateAllow,
		Required:    false,
	})
	c.RegisterFlagBool(secretStdinFlag(&c.secretSTDIN)) // --stdin

	return &c
}

// CreateCommand calls the Fastly API to create an appropriate resource.
type CreateCommand struct {
	argparser.Base
	argparser.JSONOutput

	Input         fastly.CreateSecretInput
	recreate      bool
	recreateAllow bool
	secretFile    string
	secretSTDIN   bool
}

var errMultipleSecretValue = fsterr.RemediationError{
	Inner:       fmt.Errorf("invalid flag combination, --file and --stdin"),
	Remediation: "Use one of --file or --stdin flag",
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

	switch {
	case c.recreate && c.recreateAllow:
		return fsterr.RemediationError{
			Inner:       fmt.Errorf("invalid flag combination, --recreate and --recreate-allow"),
			Remediation: "Use either --recreate or --recreate-allow, not both.",
		}
	case c.recreate:
		c.Input.Method = http.MethodPatch
	case c.recreateAllow:
		c.Input.Method = http.MethodPut
	}

	// Read secret's value: either from STDIN, a file, or prompt.
	switch {
	case c.secretSTDIN:
		// Determine if 'in' has data available.
		if in == nil || text.IsTTY(in) {
			return fsterr.ErrNoSTDINData
		}
		var buf bytes.Buffer
		if _, err := buf.ReadFrom(in); err != nil {
			return err
		}
		c.Input.Secret = buf.Bytes()

	case c.secretFile != "":
		var err error
		// nosemgrep: trailofbits.go.questionable-assignment.questionable-assignment
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

	ck, err := c.Globals.APIClient.CreateClientKey(context.TODO())
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	apiPublicKey, err := c.Globals.APIClient.GetSigningKey(context.TODO())
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if !bytes.Equal(apiPublicKey, verificationKey) && os.Getenv("FASTLY_USE_API_SIGNING_KEY") == "" {
		err := fmt.Errorf("API public key does not match expected verification key")
		c.Globals.ErrLog.Add(err)
		return err
	}

	if !ck.VerifySignature(apiPublicKey) {
		err := fmt.Errorf("unable to verify signature of client key")
		c.Globals.ErrLog.Add(err)
		return err
	}

	wrapped, err := ck.Encrypt(c.Input.Secret)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	c.Input.Secret = wrapped
	c.Input.ClientKey = ck.PublicKey

	o, err := c.Globals.APIClient.CreateSecret(context.TODO(), &c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, o); ok {
		return err
	}

	action := "Created"
	if o.Recreated {
		action = "Recreated"
	}
	text.Success(out, "%s secret '%s' in Secret Store '%s' (digest: %s)", action, o.Name, c.Input.StoreID, hex.EncodeToString(o.Digest))
	return nil
}
