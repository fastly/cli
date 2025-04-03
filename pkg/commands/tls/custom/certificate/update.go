package certificate

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/fastly/go-fastly/v10/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent argparser.Registerer, g *global.Data) *UpdateCommand {
	var c UpdateCommand
	c.CmdClause = parent.Command("update", "Replace a TLS certificate with a newly reissued TLS certificate, or update a TLS certificate's name")
	c.Globals = g

	// Required
	// cert-blob and cert-path are mutually exclusive. One is required.
	c.CmdClause.Flag("cert-blob", "The PEM-formatted certificate blob, mutually exclusive with --cert-path").StringVar(&c.certBlob)
	c.CmdClause.Flag("cert-path", "Filepath to a PEM-formatted certificate, mutually exclusive with --cert-blob").StringVar(&c.certPath)
	c.CmdClause.Flag("id", "Alphanumeric string identifying a TLS certificate").Required().StringVar(&c.id)

	// Optional.
	c.CmdClause.Flag("name", "A customizable name for your certificate. Defaults to the certificate's Common Name or first Subject Alternative Names (SAN) entry").StringVar(&c.name)
	return &c
}

// UpdateCommand calls the Fastly API to update an appropriate resource.
type UpdateCommand struct {
	argparser.Base

	certBlob string
	certPath string
	id       string
	name     string
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	input, err := c.constructInput()
	if err != nil {
		return err
	}

	r, err := c.Globals.APIClient.UpdateCustomTLSCertificate(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"TLS Certificate ID":   c.id,
			"TLS Certificate Name": c.name,
		})
		return err
	}

	if c.name != "" {
		text.Success(out, "Updated TLS Certificate '%s' (previously: '%s')", r.Name, input.Name)
	} else {
		text.Success(out, "Updated TLS Certificate '%s'", r.ID)
	}
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) constructInput() (*fastly.UpdateCustomTLSCertificateInput, error) {
	var input fastly.UpdateCustomTLSCertificateInput

	if c.certPath == "" && c.certBlob == "" {
		return nil, fmt.Errorf("neither --cert-path or --cert-blob provided, one must be provided")
	}

	if c.certPath != "" && c.certBlob != "" {
		return nil, fmt.Errorf("cert-path and cert-blob provided, only one can be specified")
	}

	input.ID = c.id

	if c.certBlob != "" {
		input.CertBlob = c.certBlob
	}

	if c.certPath != "" {
		path, err := filepath.Abs(c.certPath)
		if err != nil {
			return nil, fmt.Errorf("error parsing cert-path '%s': %q", c.certPath, err)
		}

		data, err := os.ReadFile(path) // #nosec
		if err != nil {
			return nil, fmt.Errorf("error reading cert-path '%s': %q", c.certPath, err)
		}

		input.CertBlob = string(data)
	}

	if c.name != "" {
		input.Name = c.name
	}

	return &input, nil
}
