//go:build viceroy_embed && linux && amd64

package viceroy

import _ "embed"

//go:embed assets/viceroy_linux_amd64.zst
var binaryZstd []byte

const platformSupported = true
