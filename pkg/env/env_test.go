package env

import (
	"runtime"
	"testing"

	"golang.org/x/exp/slices"
)

func TestVars(t *testing.T) {
	tcs := []struct {
		os       string
		vars     map[string]string
		expected []string
	}{
		{
			os:       "windows",
			expected: []string{"%HOME%", "%PATH%"},
		},
		{
			os:       "darwin",
			expected: []string{"\\$HOME", "\\$PATH"},
		},
		{
			os:       "linux",
			expected: []string{"\\$HOME", "\\$PATH"},
		},
	}
	for _, tc := range tcs {
		t.Run(tc.os, func(t *testing.T) {
			vars := Vars()
			if runtime.GOOS == tc.os {
				for _, v := range tc.expected {
					if !slices.Contains(vars, v) {
						t.Errorf("expected %s in %v", v, vars)
					}
				}
			} else {
				t.Skip()
			}
		})
	}
}
