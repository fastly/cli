package compute

import (
	"io"

	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
)

// NewOther constructs a new unsupported language instance.
func NewOther(
	fastlyManifest *manifest.File,
	globals *global.Data,
	flags Flags,
	in io.Reader,
	out io.Writer,
	spinner text.Spinner,
) *Other {
	return &Other{
		Shell: Shell{},

		autoYes:        globals.Flags.AutoYes,
		build:          fastlyManifest.Scripts.Build,
		errlog:         globals.ErrLog,
		input:          in,
		nonInteractive: globals.Flags.NonInteractive,
		output:         out,
		postBuild:      fastlyManifest.Scripts.PostBuild,
		spinner:        spinner,
		timeout:        flags.Timeout,
		verbose:        globals.Verbose(),
	}
}

// Other implements a Toolchain for languages without official support.
type Other struct {
	Shell

	// autoYes is the --auto-yes flag.
	autoYes bool
	// build is a shell command defined in fastly.toml using [scripts.build].
	build string
	// errlog is an abstraction for recording errors to disk.
	errlog fsterr.LogInterface
	// input is the user's terminal stdin stream
	input io.Reader
	// nonInteractive is the --non-interactive flag.
	nonInteractive bool
	// output is the users terminal stdout stream
	output io.Writer
	// postBuild is a custom script executed after the build but before the Wasm
	// binary is added to the .tar.gz archive.
	postBuild string
	// spinner is a terminal progress status indicator.
	spinner text.Spinner
	// timeout is the build execution threshold.
	timeout int
	// verbose indicates if the user set --verbose
	verbose bool
}

// Build implements the Toolchain interface and attempts to compile the package
// source to a Wasm binary.
func (o Other) Build() error {
	bt := BuildToolchain{
		autoYes:        o.autoYes,
		buildFn:        o.Shell.Build,
		buildScript:    o.build,
		errlog:         o.errlog,
		in:             o.input,
		nonInteractive: o.nonInteractive,
		out:            o.output,
		postBuild:      o.postBuild,
		spinner:        o.spinner,
		timeout:        o.timeout,
		verbose:        o.verbose,
	}
	return bt.Build()
}
