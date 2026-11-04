package certificate

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/text"
)

// UpdateCommand calls the Fastly API to update an appropriate resource.
type UpdateCommand struct {
	argparser.Base

	// Required
	// cert-blob and cert-path are mutually exclusive. One is required.
	CertBlob string `help:"The PEM-formatted certificate blob, mutually exclusive with --cert-path"`
	CertPath string `help:"Filepath to a PEM-formatted certificate, mutually exclusive with --cert-blob"`
	ID       string `required:"" help:"Alphanumeric string identifying a TLS certificate"`

	// Optional.
	NameFlag string `name:"name" help:"A customizable name for your certificate. Defaults to the certificate's Common Name or first Subject Alternative Names (SAN) entry"`
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	input, err := c.constructInput()
	if err != nil {
		return err
	}

	r, err := c.Globals.APIClient.UpdateCustomTLSCertificate(context.TODO(), input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"TLS Certificate ID":   c.ID,
			"TLS Certificate Name": c.NameFlag,
		})
		return err
	}

	if c.NameFlag != "" {
		text.Success(out, "Updated TLS Certificate '%s' (previously: '%s')", r.Name, input.Name)
	} else {
		text.Success(out, "Updated TLS Certificate '%s'", r.ID)
	}
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) constructInput() (*fastly.UpdateCustomTLSCertificateInput, error) {
	var input fastly.UpdateCustomTLSCertificateInput

	if c.CertPath == "" && c.CertBlob == "" {
		return nil, fmt.Errorf("neither --cert-path or --cert-blob provided, one must be provided")
	}

	if c.CertPath != "" && c.CertBlob != "" {
		return nil, fmt.Errorf("cert-path and cert-blob provided, only one can be specified")
	}

	input.ID = c.ID

	if c.CertBlob != "" {
		input.CertBlob = c.CertBlob
	}

	if c.CertPath != "" {
		path, err := filepath.Abs(c.CertPath)
		if err != nil {
			return nil, fmt.Errorf("error parsing cert-path '%s': %q", c.CertPath, err)
		}

		data, err := os.ReadFile(path) // #nosec
		if err != nil {
			return nil, fmt.Errorf("error reading cert-path '%s': %q", c.CertPath, err)
		}

		input.CertBlob = string(data)
	}

	if c.NameFlag != "" {
		input.Name = c.NameFlag
	}

	return &input, nil
}
