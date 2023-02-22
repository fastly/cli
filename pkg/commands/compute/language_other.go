package compute

import (
	"io"

	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/theckman/yacspin"
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

// Build implements the Toolchain interface and attempts to compile the package
// source to a Wasm binary.
func (o Other) Build(out io.Writer, spinner *yacspin.Spinner, verbose bool, callback func() error) error {
	bt := BuildToolchain{
		buildFn:           o.Shell.Build,
		buildScript:       o.build,
		errlog:            o.errlog,
		postBuild:         o.postBuild,
		timeout:           o.timeout,
		out:               out,
		postBuildCallback: callback,
		spinner:           spinner,
		verbose:           verbose,
	}
	return bt.Build()
}
