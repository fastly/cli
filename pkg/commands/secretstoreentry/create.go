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

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
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

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *CreateCommand) AfterApply() error {
	c.Input.Name = c.NameFlag
	c.Input.StoreID = c.StoreID
	return nil
}

// CreateCommand calls the Fastly API to create an appropriate resource.
type CreateCommand struct {
	argparser.Base

	// Required.
	NameFlag string `name:"name" short:"n" required:"" help:"Secret name"`
	StoreID  string `short:"s" required:"" help:"Store ID"`

	// Optional.
	SecretFile string `name:"file" short:"f" help:"Read secret value from file instead of prompt"`
	argparser.JSONOutput

	Input         fastly.CreateSecretInput `kong:"-"`
	Recreate      bool                     `help:"Recreate secret by name (errors if secret doesn't already exist)"`
	RecreateAllow bool                     `help:"Create or recreate secret by name"`
	SecretSTDIN   bool                     `name:"stdin" help:"Read secret value from STDIN instead of prompt"`
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
	if c.SecretFile != "" && c.SecretSTDIN {
		return errMultipleSecretValue
	}

	switch {
	case c.Recreate && c.RecreateAllow:
		return fsterr.RemediationError{
			Inner:       fmt.Errorf("invalid flag combination, --recreate and --recreate-allow"),
			Remediation: "Use either --recreate or --recreate-allow, not both.",
		}
	case c.Recreate:
		c.Input.Method = http.MethodPatch
	case c.RecreateAllow:
		c.Input.Method = http.MethodPut
	}

	// Read secret's value: either from STDIN, a file, or prompt.
	switch {
	case c.SecretSTDIN:
		// Determine if 'in' has data available.
		if in == nil || text.IsTTY(in) {
			return fsterr.ErrNoSTDINData
		}
		var buf bytes.Buffer
		if _, err := buf.ReadFrom(in); err != nil {
			return err
		}
		c.Input.Secret = buf.Bytes()

	case c.SecretFile != "":
		var err error
		// nosemgrep: trailofbits.go.questionable-assignment.questionable-assignment
		if c.Input.Secret, err = os.ReadFile(c.SecretFile); err != nil {
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
