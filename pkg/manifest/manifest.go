package manifest

// Source enumerates where a manifest parameter is taken from.
type Source uint8

const (
	// Filename is the name of the package manifest file.
	// It is expected to be a project specific configuration file.
	Filename = "fastly.toml"

	// ManifestLatestVersion represents the latest known manifest schema version
	// supported by the CLI.
	//
	// NOTE: The CLI is the primary consumer of the fastly.toml manifest so its
	// code is typically coupled to the specification.
	ManifestLatestVersion = 3

	// FilePermissions represents a read/write file mode.
	FilePermissions = 0o666

	// SourceUndefined indicates the parameter isn't provided in any of the
	// available sources, similar to "not found".
	SourceUndefined Source = iota

	// SourceFile indicates the parameter came from a manifest file.
	SourceFile

	// SourceEnv indicates the parameter came from the user's shell environment.
	SourceEnv

	// SourceFlag indicates the parameter came from an explicit flag.
	SourceFlag

	// SpecIntro informs the user of what the manifest file is for.
	SpecIntro = "This file describes a Fastly Compute@Edge package. To learn more visit:"

	// SpecURL points to the fastly.toml manifest specification reference.
	SpecURL = "https://developer.fastly.com/reference/fastly-toml/"
)
