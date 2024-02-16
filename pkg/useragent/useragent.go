package useragent

import (
	"fmt"

	"github.com/fastly/cli/v10/pkg/revision"
)

// Name is the user agent which we report in all HTTP requests.
var Name = fmt.Sprintf("%s/%s", "FastlyCLI", revision.AppVersion)
