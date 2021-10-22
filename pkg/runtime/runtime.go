package runtime

import "runtime"

// Windows indicates if the CLI binary's runtime OS is Windows.
//
// NOTE: We use the same conditional check multiple times across the code base
// and I noticed I had a typo in a few instances where I had omitted the "s" at
// the end of "window" which meant the conditional failed to match when running
// on Windows. So this avoids that issue in case we need to add more uses of it.
var Windows = runtime.GOOS == "windows"
