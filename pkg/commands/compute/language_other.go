package compute

import (
	"io"

	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
)

// NewOther constructs a new unsupported language instance.
func NewOther(scripts manifest.Scripts, errlog fsterr.LogInterface, timeout int) *Other {
	return &Other{
		Shell: Shell{},

		build:     scripts.Build,
		errlog:    errlog,
		postBuild: scripts.PostBuild,
		timeout:   timeout,
	}
}

// Other implements a Toolchain for languages without official support.
type Other struct {
	Shell

	// build is a shell command defined in fastly.toml using [scripts.build].
	build string
	// errlog is an abstraction for recording errors to disk.
	errlog fsterr.LogInterface
	// postBuild is a custom script executed after the build but before the Wasm
	// binary is added to the .tar.gz archive.
	postBuild string
	// timeout is the build execution threshold.
	timeout int
}

// Initialize is a no-op.
func (o Other) Initialize(_ io.Writer) error {
	return nil
}

// Verify is a no-op.
func (o Other) Verify(_ io.Writer) error {
	return nil
}

// Build implements the Toolchain interface and attempts to compile the package
// source to a Wasm binary.
func (o Other) Build(out io.Writer, progress text.Progress, verbose bool, callback func() error) error {
	return build(buildOpts{
		buildScript: o.build,
		buildFn:     o.Shell.Build,
		errlog:      o.errlog,
		postBuild:   o.postBuild,
		timeout:     o.timeout,
	}, out, progress, verbose, nil, callback)
}
