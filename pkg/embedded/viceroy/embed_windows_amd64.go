//go:build viceroy_embed && windows && amd64

package viceroy

import _ "embed"

//go:embed assets/viceroy_windows_amd64.zst
var binaryZstd []byte

const platformSupported = true
