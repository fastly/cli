package service

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/commands/service/acl"
	"github.com/fastly/cli/pkg/commands/service/aclentry"
	"github.com/fastly/cli/pkg/commands/service/alert"
	"github.com/fastly/cli/pkg/commands/service/auth"
	"github.com/fastly/cli/pkg/commands/service/backend"
	"github.com/fastly/cli/pkg/commands/service/dictionary"
	"github.com/fastly/cli/pkg/commands/service/dictionaryentry"
	"github.com/fastly/cli/pkg/commands/service/domain"
	"github.com/fastly/cli/pkg/commands/service/healthcheck"
	"github.com/fastly/cli/pkg/commands/service/imageoptimizerdefaults"
	"github.com/fastly/cli/pkg/commands/service/logging"
	"github.com/fastly/cli/pkg/commands/service/purge"
	"github.com/fastly/cli/pkg/commands/service/ratelimit"
	"github.com/fastly/cli/pkg/commands/service/resourcelink"
	"github.com/fastly/cli/pkg/commands/service/vcl"
	"github.com/fastly/cli/pkg/commands/service/version"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	Create          CreateCommand                      `cmd:"" aliases:"add" help:"Create a Fastly service"`
	Delete          DeleteCommand                      `cmd:"" aliases:"remove" help:"Delete a Fastly service"`
	Describe        DescribeCommand                    `cmd:"" aliases:"get" help:"Show detailed information about a Fastly service"`
	List            ListCommand                        `cmd:"" help:"List Fastly services"`
	Search          SearchCommand                      `cmd:"" help:"Search for a Fastly service by name"`
	Update          UpdateCommand                      `cmd:"" help:"Update a Fastly service"`
	Purge           purge.PurgeCommand                 `cmd:"" help:"Invalidate objects in the Fastly cache"`
	Alert           alerts.RootCommand                 `cmd:"" help:"Manipulate Fastly Service Alerts"`
	ACL             acl.RootCommand                    `cmd:"" help:"Manipulate Fastly ACLs (Access Control Lists)"`
	ACLEntry        aclentry.RootCommand               `cmd:"" help:"Manipulate Fastly ACL (Access Control List) entries"`
	Auth            auth.RootCommand                   `cmd:"" help:"Allow users to access only specified services"`
	Dictionary      dictionary.RootCommand             `cmd:"" help:"Manipulate Fastly edge dictionaries"`
	VCL             vcl.RootCommand                    `cmd:"" help:"Manipulate Fastly service version VCL"`
	Logging         logging.RootCommand                `cmd:"" help:"Inspect and manage Fastly service logging"`
	Version         version.RootCommand                `cmd:"" help:"Manipulate Fastly service versions"`
	Domain          domain.RootCommand                 `cmd:"" help:"Manipulate Fastly service version domains"`
	DictionaryEntry dictionaryentry.RootCommand        `cmd:"" help:"Manipulate Fastly edge dictionary items"`
	Backend         backend.RootCommand                `cmd:"" help:"Manipulate Fastly service version backends"`
	Healthcheck     healthcheck.RootCommand            `cmd:"" help:"Manipulate Fastly service version healthchecks"`
	Imageoptimizer  imageoptimizerdefaults.RootCommand `cmd:"" help:"Manipulate Fastly service Image Optimizer default settings"`
	RateLimit       ratelimit.RootCommand              `cmd:"" help:"Manipulate rate-limiters of the Fastly API and web interface"`
	ResourceLink    resourcelink.RootCommand           `cmd:"" help:"Manipulate Fastly service resource links"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "service"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
