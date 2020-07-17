package compute

import (
	"crypto/sha512"
	"fmt"
	"io"
	"os"
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
	version  common.OptionalInt
}

// NewDeployCommand returns a usable command registered under the parent.
func NewDeployCommand(parent common.Registerer, client api.HTTPClient, globals *config.Data) *DeployCommand {
	var c DeployCommand
	c.Globals = globals
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("deploy", "Deploy a package to a Fastly Compute@Edge service")
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("version", "Number of version to activate").Action(c.version.Set).IntVar(&c.version.Value)
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

	var (
		version *fastly.Version
	)

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

	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return fmt.Errorf("error reading service: no service ID found. Please provide one via the --service-id flag or within your package manifest")
	}

	// Set the version we want to operate on.
	// If version not provided infer the latest ideal version from the service.
	if !c.version.Valid {
		progress.Step("Fetching latest version...")
		versions, err := c.Globals.Client.ListVersions(&fastly.ListVersionsInput{
			Service: serviceID,
		})
		if err != nil {
			return fmt.Errorf("error listing service versions: %w", err)
		}

		version, err = getLatestIdealVersion(versions)
		if err != nil {
			return fmt.Errorf("error finding latest service version")
		}
	} else {
		version = &fastly.Version{Number: c.version.Value}
	}

	progress.Step("Validating package...")

	if err := validate(c.path); err != nil {
		return err
	}

	// Compare local package hashsum against existing service package version
	// and exit early with message if identical.
	if p, err := c.Globals.Client.GetPackage(&fastly.GetPackageInput{
		Service: serviceID,
		Version: version.Number,
	}); err == nil {
		hashSum, err := getHashSum(c.path)
		if err != nil {
			return fmt.Errorf("error getting package hashsum: %w", err)
		}

		if hashSum == p.Metadata.HashSum {
			progress.Done()
			text.Info(out, "Skipping package deployment, local and service version are identical. (service %v, version %v) ", serviceID, version.Number)
			return nil
		}
	}

	// If a version wasn't supplied and the ideal version is currently active
	// or locked, clone it.
	if !c.version.Valid && version.Active || version.Locked {
		progress.Step("Cloning latest version...")
		version, err = c.Globals.Client.CloneVersion(&fastly.CloneVersionInput{
			Service: serviceID,
			Version: version.Number,
		})
		if err != nil {
			return fmt.Errorf("error cloning latest service version: %w", err)
		}
	}

	progress.Step("Uploading package...")
	_, err = c.Globals.Client.UpdatePackage(&fastly.UpdatePackageInput{
		Service:     serviceID,
		Version:     version.Number,
		PackagePath: c.path,
	})
	if err != nil {
		return fmt.Errorf("error uploading package: %w", err)
	}

	progress.Step("Activating version...")

	_, err = c.Globals.Client.ActivateVersion(&fastly.ActivateVersionInput{
		Service: serviceID,
		Version: version.Number,
	})
	if err != nil {
		return fmt.Errorf("error activating version: %w", err)
	}

	progress.Step("Updating package manifest...")

	fmt.Fprintf(progress, "Setting version in manifest to %d...\n", version.Number)
	c.manifest.File.Version = version.Number

	if err := c.manifest.File.Write(ManifestFilename); err != nil {
		return fmt.Errorf("error saving package manifest: %w", err)
	}

	progress.Done()

	text.Break(out)

	text.Description(out, "Manage this service at", fmt.Sprintf("%s%s", manageServiceBaseURL, serviceID))

	if domains, err := c.Globals.Client.ListDomains(&fastly.ListDomainsInput{
		Service: serviceID,
		Version: version.Number,
	}); err == nil {
		text.Description(out, "View this service at", fmt.Sprintf("https://%s", domains[0].Name))
	}

	text.Success(out, "Deployed package (service %s, version %v)", serviceID, version.Number)
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

func getHashSum(path string) (hash string, err error) {
	// gosec flagged this:
	// G304 (CWE-22): Potential file inclusion via variable
	// Disabling as we trust the source of the filepath variable.
	/* #nosec */
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() {
		cerr := f.Close()
		if err == nil {
			err = cerr
		}
	}()

	h := sha512.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
