package kongcompat

// The program was converted from kingpin v3 or a fork based on it.

// version is the major version of kingpin whose behavior the package
// reproduces.
const version = 3

// legacy is set for kingpin v2 releases before v2.3.0, whose help was
// closer to kingpin v3's.
const legacy = false

// v2From23 reports whether the program came from kingpin v2.3.0 or later,
// whose help shows --[no-] forms, environment variables and argument
// placeholders, and which names every missing required flag.
func v2From23() bool {
	return version == 2 && !legacy
}

// Action is the type of kingpin's callbacks, for code that kept values of
// it.
type Action func(element *ParseElement, context *ParseContext) error
