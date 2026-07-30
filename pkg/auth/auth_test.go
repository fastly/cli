package auth

import "testing"

func TestValidateAudienceClaim(t *testing.T) {
	t.Parallel()

	const apiEndpoint = "https://api.fastly.com"

	tests := []struct {
		name    string
		aud     any
		wantErr bool
	}{
		{
			name: "exact match",
			aud:  apiEndpoint,
		},
		{
			name: "trailing slash match",
			aud:  apiEndpoint + "/",
		},
		{
			name:    "different audience",
			aud:     "https://api.example.com/",
			wantErr: true,
		},
		{
			name:    "non string audience",
			aud:     []string{apiEndpoint},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validateAudienceClaim(tt.aud, apiEndpoint)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateAudienceClaim() error = %v, wantErr %t", err, tt.wantErr)
			}
		})
	}
}

func TestAudienceMatchesAPIEndpoint(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		aud         string
		apiEndpoint string
		want        bool
	}{
		{
			name:        "exact match without trailing slash",
			aud:         "https://api.fastly.com",
			apiEndpoint: "https://api.fastly.com",
			want:        true,
		},
		{
			name:        "audience trailing slash accepted",
			aud:         "https://api.fastly.com/",
			apiEndpoint: "https://api.fastly.com",
			want:        true,
		},
		{
			name:        "configured endpoint trailing slash accepted",
			aud:         "https://api.fastly.com",
			apiEndpoint: "https://api.fastly.com/",
			want:        true,
		},
		{
			name:        "double trailing slash rejected",
			aud:         "https://api.fastly.com//",
			apiEndpoint: "https://api.fastly.com",
			want:        false,
		},
		{
			name:        "different audience rejected",
			aud:         "https://example.com/",
			apiEndpoint: "https://api.fastly.com",
			want:        false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := audienceMatchesAPIEndpoint(tt.aud, tt.apiEndpoint)
			if got != tt.want {
				t.Fatalf("audienceMatchesAPIEndpoint(%q, %q) = %t, want %t", tt.aud, tt.apiEndpoint, got, tt.want)
			}
		})
	}
}
