package compute

import (
	"fmt"
	"io"
	"os"
	"time"

	fsterr "github.com/fastly/cli/pkg/errors"
	fstexec "github.com/fastly/cli/pkg/exec"
	"github.com/fastly/cli/pkg/manifest"
)

// Other implements a Toolchain for languages without official support.
type Other struct {
	Shell

	build     string
	errlog    fsterr.LogInterface
	postBuild string
	timeout   int
}

// NewOther constructs a new unsupported language instance.
func NewOther(timeout int, scripts manifest.Scripts, errlog fsterr.LogInterface) *Other {
	return &Other{
		Shell:     Shell{},
		build:     scripts.Build,
		errlog:    errlog,
		postBuild: scripts.PostBuild,
		timeout:   timeout,
	}
}

// Initialize is a no-op.
func (o Other) Initialize(out io.Writer) error {
	return nil
}

// Verify is a no-op.
func (o Other) Verify(out io.Writer) error {
	return nil
}

// Build implements the Toolchain interface and attempts to compile the package
// source to a Wasm binary.
func (o Other) Build(out, progress io.Writer, verbose bool) error {
	if o.build == "" {
		err := fmt.Errorf("error reading custom build instructions from fastly.toml manifest")
		o.errlog.Add(err)
		return fsterr.RemediationError{
			Inner:       err,
			Remediation: fsterr.ComputeBuildRemediation,
		}
	}
	cmd, args := o.Shell.Build(o.build)

	err := o.execCommand(cmd, args, out, progress, verbose)
	if err != nil {
		return err
	}

	if o.postBuild != "" {
		cmd, args := o.Shell.Build(o.postBuild)
		err := o.execCommand(cmd, args, out, progress, verbose)
		if err != nil {
			return err
		}
	}

	return nil
}

func (o Other) execCommand(cmd string, args []string, out, progress io.Writer, verbose bool) error {
	s := fstexec.Streaming{
		Command:  cmd,
		Args:     args,
		Env:      os.Environ(),
		Output:   out,
		Progress: progress,
		Verbose:  verbose,
	}
	if o.timeout > 0 {
		s.Timeout = time.Duration(o.timeout) * time.Second
	}
	if err := s.Exec(); err != nil {
		o.errlog.Add(err)
		return err
	}
	return nil
}
