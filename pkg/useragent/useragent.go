package useragent

import (
	"fmt"

	"github.com/fastly/cli/pkg/revision"
)

// Name is the user agent which we report in all HTTP requests.
var Name = fmt.Sprintf("%s/%s", "FastlyCLI", revision.AppVersion)

func SetExtension(extension string) {
	Name = fmt.Sprintf("%s, %s", Name, extension)
}
