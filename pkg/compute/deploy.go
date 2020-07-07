package compute

import (
	"fmt"
	"io"
	"path/filepath"
	"sort"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/fastly"
	"github.com/kennygrant/sanitize"
)

// DeployCommand deploys an artifact previously produced by build.
type DeployCommand struct {
	common.Base
	manifest manifest.Data
	path     string
	version  int
}

// NewDeployCommand returns a usable command registered under the parent.
func NewDeployCommand(parent common.Registerer, client api.HTTPClient, globals *config.Data) *DeployCommand {
	var c DeployCommand
	c.Globals = globals
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("deploy", "Deploy a package to a Fastly Compute@Edge service")
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("version", "Number of version to activate").IntVar(&c.version)
	c.CmdClause.Flag("path", "Path to package").Short('p').StringVar(&c.path)
	return &c
}

// Exec implements the command interface.
func (c *DeployCommand) Exec(in io.Reader, out io.Writer) (err error) {
	var progress text.Progress
	if c.Globals.Verbose() {
		progress = text.NewVerboseProgress(out)
	} else {
		progress = text.NewQuietProgress(out)
	}

	defer func() {
		if err != nil {
			progress.Fail() // progress.Done is handled inline
		}
	}()

	// If path flag was empty, default to package tar inside pkg directory
	// and get filename from the manifest.
	if c.path == "" {
		progress.Step("Reading package manifest...")

		name, source := c.manifest.Name()
		if source == manifest.SourceUndefined {
			return fmt.Errorf("error reading package manifest")
		}

		c.path = filepath.Join("pkg", fmt.Sprintf("%s.tar.gz", sanitize.BaseName(name)))
	}

	progress.Step("Validating package...")

	if err := validate(c.path); err != nil {
		return err
	}

	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return fmt.Errorf("error reading service: no service ID found. Please provide one via the --service-id flag or within your package manifest")
	}

	if c.version == 0 {
		progress.Step("Fetching latest version...")
		versions, err := c.Globals.Client.ListVersions(&fastly.ListVersionsInput{
			Service: serviceID,
		})
		if err != nil {
			return fmt.Errorf("error listing service versions: %w", err)
		}

		version, err := getLatestIdealVersion(versions)
		if err != nil {
			return fmt.Errorf("error finding latest service version")
		}

		if version.Active || version.Locked {
			progress.Step("Cloning latest version...")
			version, err = c.Globals.Client.CloneVersion(&fastly.CloneVersionInput{
				Service: serviceID,
				Version: version.Number,
			})
			if err != nil {
				return fmt.Errorf("error cloning latest service version: %w", err)
			}
		}

		c.version = version.Number
	}

	progress.Step("Uploading package...")
	_, err = c.Globals.Client.UpdatePackage(&fastly.UpdatePackageInput{
		Service:     serviceID,
		Version:     c.version,
		PackagePath: c.path,
	})
	if err != nil {
		return fmt.Errorf("error uploading package: %w", err)
	}

	progress.Step("Activating version...")

	_, err = c.Globals.Client.ActivateVersion(&fastly.ActivateVersionInput{
		Service: serviceID,
		Version: c.version,
	})
	if err != nil {
		return fmt.Errorf("error activating version: %w", err)
	}

	progress.Step("Updating package manifest...")

	fmt.Fprintf(progress, "Setting version in manifest to %d...\n", c.version)
	c.manifest.File.Version = c.version

	if err := c.manifest.File.Write(ManifestFilename); err != nil {
		return fmt.Errorf("error saving package manifest: %w", err)
	}

	progress.Done()

	text.Break(out)

	text.Description(out, "Manage this service at", fmt.Sprintf("%s%s", manageServiceBaseURL, serviceID))

	if domains, err := c.Globals.Client.ListDomains(&fastly.ListDomainsInput{
		Service: serviceID,
		Version: c.version,
	}); err == nil {
		text.Description(out, "View this service at", fmt.Sprintf("https://%s", domains[0].Name))
	}

	text.Success(out, "Deployed package (service %s, version %v)", serviceID, c.version)
	return nil
}

// getLatestIdealVersion gets the most ideal service version using the following logic:
// - Find the active version and return
// - If no active version, find the latest locked version and return
// - Otherwise return the latest version
func getLatestIdealVersion(versions []*fastly.Version) (*fastly.Version, error) {
	sort.Slice(versions, func(i, j int) bool {
		return versions[i].UpdatedAt.Before(*versions[j].UpdatedAt)
	})

	var active, locked, latest *fastly.Version
	for i := 0; i < len(versions); i++ {
		v := versions[i]
		if v.Active {
			active = v
		}
		if v.Locked {
			locked = v
		}
		latest = v
	}

	var version *fastly.Version
	if active != nil {
		version = active
	} else if locked != nil {
		version = locked
	} else {
		version = latest
	}

	if version == nil {
		return nil, fmt.Errorf("error finding latest service version")
	}

	return version, nil
}
