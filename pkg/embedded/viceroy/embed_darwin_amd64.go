//go:build viceroy_embed && darwin && amd64

package viceroy

import _ "embed"

//go:embed assets/viceroy_darwin_amd64.zst
var binaryZstd []byte

const platformSupported = true
