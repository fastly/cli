//go:build viceroy_embed && darwin && arm64

package viceroy

import _ "embed"

//go:embed assets/viceroy_darwin_arm64.zst
var binaryZstd []byte

const platformSupported = true
