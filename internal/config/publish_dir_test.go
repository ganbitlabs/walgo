package config

import (
	"path/filepath"
	"testing"
)

func TestResolvePublishDir(t *testing.T) {
	site := t.TempDir()
	elsewhere := filepath.Join(t.TempDir(), "out")
	tests := []struct {
		name       string
		publishDir string
		want       string
		wantErr    bool
	}{
		{"relative joins the site", "public", filepath.Join(site, "public"), false},
		{"nested relative", "build/www", filepath.Join(site, "build", "www"), false},
		{"absolute stays absolute", elsewhere, elsewhere, false},
		{"site root is refused", ".", "", true},
		{"parent of the site is refused", "..", "", true},
		{"absolute site root is refused", site, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewDefaultWalgoConfig()
			cfg.HugoConfig.PublishDir = tt.publishDir
			got, err := ResolvePublishDir(site, &cfg)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ResolvePublishDir(%q) = %q, want error", tt.publishDir, got)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("ResolvePublishDir(%q) = %q, want %q", tt.publishDir, got, tt.want)
			}
		})
	}
}
