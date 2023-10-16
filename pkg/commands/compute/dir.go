package compute

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fastly/cli/pkg/manifest"
)

// EnvManifestMsg informs the user that an environment manifest is being used.
const EnvManifestMsg = "Using the '%s' environment manifest (it will be packaged up as %s)\n\n"

// ProjectDirMsg informs the user that we've changed the project directory.
const ProjectDirMsg = "Changed project directory to '%s'\n\n"

// EnvironmentManifest returns the relevant manifest filename, taking into
// account the user passing an --env flag.
func EnvironmentManifest(env string) (manifestFilename string) {
	manifestFilename = manifest.Filename
	if env != "" {
		manifestFilename = fmt.Sprintf("fastly.%s.toml", env)
	}
	return manifestFilename
}

// ChangeProjectDirectory moves into `dir` and returns its absolute path.
func ChangeProjectDirectory(dir string) (projectDirectory string, err error) {
	if dir != "" {
		projectDirectory, err = filepath.Abs(dir)
		if err != nil {
			return "", fmt.Errorf("failed to construct absolute path to directory '%s': %w", dir, err)
		}
		if err := os.Chdir(projectDirectory); err != nil {
			return "", fmt.Errorf("failed to change working directory to '%s': %w", projectDirectory, err)
		}
	}
	return projectDirectory, nil
}
