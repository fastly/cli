package setup

import (
	"io"

	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
)

// Loggers represents the service state related to log entries defined within
// the fastly.toml [setup] configuration.
//
// NOTE: It implements the setup.Interface interface.
type Loggers struct {
	Setup  map[string]*manifest.SetupLogger
	Stdout io.Writer
}

// Logger represents the configuration parameters for creating a dictionary
// via the API client.
type Logger struct {
	Provider string
}

// Configure prompts the user for specific values related to the service resource.
func (l *Loggers) Configure() error {
	text.Break(l.Stdout)
	text.Info(l.Stdout, "The package code requires the following log endpoints to be created.")
	text.Break(l.Stdout)

	for name, settings := range l.Setup {
		text.Output(l.Stdout, "%s %s", text.Bold("Name:"), name)
		if settings.Provider != "" {
			text.Output(l.Stdout, "%s %s", text.Bold("Provider:"), settings.Provider)
		}
		text.Break(l.Stdout)
	}

	text.Description(
		l.Stdout,
		"Refer to the help documentation for each provider (if no provider shown, then select your own)",
		"fastly logging <provider> create --help",
	)

	return nil
}

// Predefined indicates if the service resource has been specified within the
// fastly.toml file using a [setup] configuration block.
func (l *Loggers) Predefined() bool {
	return len(l.Setup) > 0
}
