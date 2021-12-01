package compute

import (
	"fmt"
	"io"
	"time"

	fsterr "github.com/fastly/cli/pkg/errors"
	fstexec "github.com/fastly/cli/pkg/exec"
)

// Other implements a Toolchain for languages without official support.
type Other struct {
	Shell

	build   string
	errlog  fsterr.LogInterface
	timeout int
}

// NewOther constructs a new unsupported language instance.
func NewOther(timeout int, build string, errlog fsterr.LogInterface) *Other {
	return &Other{
		Shell:   Shell{},
		build:   build,
		errlog:  errlog,
		timeout: timeout,
	}
}

// No-op
func (o Other) Initialize(out io.Writer) error {
	return nil
}

// No-op
func (o Other) Verify(out io.Writer) error {
	return nil
}

// Build implements the Toolchain interface and attempts to compile the package
// source to a Wasm binary.
func (o Other) Build(out io.Writer, verbose bool) error {
	if o.build == "" {
		err := fmt.Errorf("error reading custom build instructions from fastly.toml manifest")
		o.errlog.Add(err)
		return fsterr.RemediationError{
			Inner:       err,
			Remediation: fsterr.ComputeBuildRemediation,
		}
	}
	cmd, args := o.Shell.Build(o.build)

	s := fstexec.Streaming{
		Command: cmd,
		Args:    args,
		Env:     []string{},
		Output:  out,
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
