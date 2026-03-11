package argparser_test

import (
	"testing"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/env"
)

func TestIsGlobalFlagsOnly(t *testing.T) {
	t.Setenv(env.DisableAuthCommand, "")
	tests := []struct {
		name string
		args []string
		want bool
	}{
		{
			name: "verbose only",
			args: []string{"--verbose"},
			want: true,
		},
		{
			name: "token with value",
			args: []string{"--token", "abc"},
			want: true,
		},
		{
			name: "short token with value",
			args: []string{"-t", "abc"},
			want: true,
		},
		{
			name: "subcommand present",
			args: []string{"--verbose", "version"},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := argparser.IsGlobalFlagsOnly(tt.args); got != tt.want {
				t.Errorf("IsGlobalFlagsOnly(%v) = %v, want %v", tt.args, got, tt.want)
			}
		})
	}
}

func TestIsGlobalFlagsOnlyDisabledAuth(t *testing.T) {
	t.Setenv(env.DisableAuthCommand, "1")

	tests := []struct {
		name string
		args []string
		want bool
	}{
		{
			name: "token is not global when auth disabled",
			args: []string{"--token", "abc"},
			want: false,
		},
		{
			name: "short token is not global when auth disabled",
			args: []string{"-t", "abc"},
			want: false,
		},
		{
			name: "other globals still work",
			args: []string{"--verbose"},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := argparser.IsGlobalFlagsOnly(tt.args); got != tt.want {
				t.Errorf("IsGlobalFlagsOnly(%v) = %v, want %v", tt.args, got, tt.want)
			}
		})
	}
}
