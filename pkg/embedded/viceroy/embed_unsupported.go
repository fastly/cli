//go:build viceroy_embed && !(darwin && amd64) && !(darwin && arm64) && !(linux && amd64) && !(linux && arm64) && !(windows && amd64)

package viceroy

const platformSupported = false

var binaryZstd []byte
