package serviceversion

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v3/fastly"
)

// LockCommand calls the Fastly API to lock a service version.
type LockCommand struct {
	cmd.Base
	manifest       manifest.Data
	Input          fastly.LockVersionInput
	serviceVersion cmd.OptionalServiceVersion
}

// NewLockCommand returns a usable command registered under the parent.
func NewLockCommand(parent cmd.Registerer, globals *config.Data) *LockCommand {
	var c LockCommand
	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("lock", "Lock a Fastly service version")
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.SetServiceVersionFlag(cmd.ServiceVersionFlagOpts{
		Dst: &c.serviceVersion.Value,
	})
	return &c
}

// Exec invokes the application logic for the command.
func (c *LockCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return errors.ErrNoServiceID
	}
	c.Input.ServiceID = serviceID

	// TODO(integralist): replace this surrounding code with cmd.ServiceDetails
	// once we have conditional boolean for the autoclone logic
	v, err := c.serviceVersion.Parse(c.Input.ServiceID, c.Globals.Client)
	if err != nil {
		return err
	}
	c.Input.ServiceVersion = v.Number

	ver, err := c.Globals.Client.LockVersion(&c.Input)
	if err != nil {
		return err
	}

	text.Success(out, "Locked service %s version %d", ver.ServiceID, c.Input.ServiceVersion)
	return nil
}
