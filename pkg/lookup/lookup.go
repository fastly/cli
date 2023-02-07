package lookup

// Source enumerates where the parameter is taken from.
type Source uint8

const (
	// SourceUndefined indicates the parameter isn't provided in any of the
	// available sources, similar to "not found".
	SourceUndefined Source = iota

	// SourceFile indicates the parameter came from a config file.
	SourceFile

	// SourceEnvironment indicates the parameter came from an env var.
	SourceEnvironment

	// SourceFlag indicates the parameter came from an explicit flag.
	SourceFlag

	// SourceDefault indicates the parameter came from a program default.
	SourceDefault
)
