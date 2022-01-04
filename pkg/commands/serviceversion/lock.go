package serviceversion

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v5/fastly"
)

// LockCommand calls the Fastly API to lock a service version.
type LockCommand struct {
	cmd.Base
	manifest       manifest.Data
	Input          fastly.LockVersionInput
	serviceName    cmd.OptionalServiceNameID
	serviceVersion cmd.OptionalServiceVersion
}

// NewLockCommand returns a usable command registered under the parent.
func NewLockCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *LockCommand {
	var c LockCommand
	c.Globals = globals
	c.manifest = data
	c.CmdClause = parent.Command("lock", "Lock a Fastly service version")
	c.RegisterServiceIDFlag(&c.manifest.Flag.ServiceID)
	c.RegisterServiceNameFlag(c.serviceName.Set, &c.serviceName.Value)
	c.RegisterServiceVersionFlag(cmd.ServiceVersionFlagOpts{
		Dst: &c.serviceVersion.Value,
	})
	return &c
}

// Exec invokes the application logic for the command.
func (c *LockCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, serviceVersion, err := cmd.ServiceDetails(cmd.ServiceDetailsOpts{
		AllowActiveLocked:  true,
		Client:             c.Globals.Client,
		Manifest:           c.manifest,
		Out:                out,
		ServiceNameFlag:    c.serviceName,
		ServiceVersionFlag: c.serviceVersion,
		VerboseMode:        c.Globals.Flag.Verbose,
	})
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
			"Service ID":      serviceID,
			"Service Version": errors.ServiceVersion(serviceVersion),
		})
		return err
	}

	c.Input.ServiceID = serviceID
	c.Input.ServiceVersion = serviceVersion.Number

	ver, err := c.Globals.Client.LockVersion(&c.Input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
			"Service ID":      serviceID,
			"Service Version": serviceVersion.Number,
		})
		return err
	}

	text.Success(out, "Locked service %s version %d", ver.ServiceID, c.Input.ServiceVersion)
	return nil
}
