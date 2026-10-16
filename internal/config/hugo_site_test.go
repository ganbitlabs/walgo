package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigFromPublishDir(t *testing.T) {
	tests := []struct {
		name  string
		files map[string]string
		want  string
		hugo  bool
	}{
		{"no Hugo config", map[string]string{}, "public", false},
		{"walgo.yaml wins", map[string]string{"walgo.yaml": "hugo:\n  publishDir: dist\n", "hugo.toml": "publishDir = \"x\""}, "dist", true},
		{"hugo.toml", map[string]string{"hugo.toml": "publishDir = \"out\""}, "out", true},
		{"config.yaml beside archetypes", map[string]string{"config.yaml": "publishDir: out", "archetypes/default.md": ""}, "out", true},
		{"config.json beside content markdown", map[string]string{"config.json": `{"publishDir": "out"}`, "content/posts/a.md": "# a"}, "out", true},
		{"config.json with Hugo-looking keys alone is not Hugo", map[string]string{"config.json": `{"theme": "dark", "baseUrl": "https://api.example.com", "module": "app", "publishDir": "x"}`, "layouts/main.html": ""}, "public", false},
		{"hugo.json lowercase key", map[string]string{"hugo.json": `{"publishdir": "out"}`}, "out", true},
		{"config/_default", map[string]string{"config/_default/config.toml": "baseURL = \"https://e.com/\"\npublishDir = \"out\""}, "out", true},
		{"production overrides _default", map[string]string{
			"config/_default/hugo.toml":   "publishDir = \"default\"",
			"config/production/hugo.toml": "publishDir = \"prod\"",
		}, "prod", true},
		{"Hugo config without publishDir", map[string]string{"hugo.yml": "title: x"}, "public", true},
		{"generic config.json is not Hugo", map[string]string{"config.json": `{"publishDir": "x", "apiUrl": "y"}`}, "public", false},
		{"absolute Hugo publishDir is ignored", map[string]string{"hugo.toml": "publishDir = \"/srv/site\""}, "public", true},
		{"config.toml is Hugo", map[string]string{"config.toml": "title = \"x\""}, "public", true},
		{"malformed hugo.toml is still Hugo", map[string]string{"hugo.toml": "title = "}, "public", true},
		{"malformed config.toml is still Hugo", map[string]string{"config.toml": "title = "}, "public", true},
		{"malformed config.json alone is not Hugo", map[string]string{"config.json": "{"}, "public", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			if _, ok := tt.files["walgo.yaml"]; !ok {
				tt.files["walgo.yaml"] = "walrus:\n  network: testnet\n"
			}
			for name, content := range tt.files {
				path := filepath.Join(dir, name)
				if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			cfg, err := LoadConfigFrom(dir)
			if err != nil {
				t.Fatal(err)
			}
			if cfg.HugoConfig.PublishDir != tt.want {
				t.Errorf("PublishDir = %q, want %q", cfg.HugoConfig.PublishDir, tt.want)
			}
			if IsHugoSite(dir) != tt.hugo {
				t.Errorf("IsHugoSite = %v, want %v", IsHugoSite(dir), tt.hugo)
			}
		})
	}
}
