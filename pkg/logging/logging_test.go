package logging

import (
	"testing"
)

func Test_getMinorSemver(t *testing.T) {
	tests := []struct {
		name string
		arg  string
		want string
	}{
		{
			name: "default semver",
			arg:  "v0.0.0-dev",
			want: "v0.0",
		},
		{
			name: "example semver",
			arg:  "v2.0.1",
			want: "v2.0",
		},
		{
			name: "example semver with alpha notation",
			arg:  "v3.0.0-alpha.0",
			want: "v3.0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getMinorSemver(tt.arg); got != tt.want {
				t.Errorf("getMinorSemver() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDocsURL(t *testing.T) {
	tests := []struct {
		name string
		arg  string
		want string
	}{
		{
			name: "url with no leading slash",
			arg:  "features",
			want: "https://vektra.github.io/mockery/v0.0/features",
		},
		{
			name: "url with leading slash",
			arg:  "/features",
			want: "https://vektra.github.io/mockery/v0.0/features",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DocsURL(tt.arg); got != tt.want {
				t.Errorf("DocsURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
