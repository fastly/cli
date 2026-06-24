//go:build viceroy_embed && linux && arm64

package viceroy

import _ "embed"

//go:embed assets/viceroy_linux_arm64.zst
var binaryZstd []byte

const platformSupported = true
