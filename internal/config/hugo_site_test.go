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
		{"absolute Hugo publishDir is kept", map[string]string{"hugo.toml": "publishDir = \"/srv/site\""}, "/srv/site", true},
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

func writeSite(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestHugoConfigFile(t *testing.T) {
	tests := []struct {
		name  string
		files map[string]string
		want  string
	}{
		{"none", map[string]string{}, ""},
		{"hugo.yaml", map[string]string{"hugo.yaml": "title: x"}, "hugo.yaml"},
		{"hugo before config", map[string]string{"config.toml": "", "hugo.json": "{}"}, "hugo.json"},
		{"root before config dir", map[string]string{"config/_default/hugo.toml": "", "config.toml": ""}, "config.toml"},
		{"_default before production", map[string]string{"config/production/hugo.toml": "", "config/_default/params.toml": "", "config/_default/config.yaml": "title: x"}, filepath.Join("config", "_default", "config.yaml")},
		{"production only", map[string]string{"config/production/hugo.toml": ""}, filepath.Join("config", "production", "hugo.toml")},
		{"ambiguous config.yaml alone", map[string]string{"config.yaml": "title: x"}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := writeSite(t, tt.files)
			got := HugoConfigFile(dir)
			want := ""
			if tt.want != "" {
				want = filepath.Join(dir, tt.want)
			}
			if got != want {
				t.Errorf("HugoConfigFile = %q, want %q", got, want)
			}
		})
	}
}

func TestHugoTOMLConfigFile(t *testing.T) {
	if _, err := HugoTOMLConfigFile(writeSite(t, map[string]string{"hugo.yaml": "title: x"})); err == nil {
		t.Error("YAML config: want an error naming the file")
	}
	if _, err := HugoTOMLConfigFile(t.TempDir()); err == nil {
		t.Error("no config: want an error")
	}
	dir := writeSite(t, map[string]string{"config/_default/hugo.toml": "title = \"x\""})
	got, err := HugoTOMLConfigFile(dir)
	if err != nil || got != filepath.Join(dir, "config", "_default", "hugo.toml") {
		t.Errorf("HugoTOMLConfigFile = %q, %v", got, err)
	}
}

func TestHugoConfigValue(t *testing.T) {
	dir := writeSite(t, map[string]string{
		"config/_default/hugo.toml":   "Theme = \"base\"\ntitle = \"Default\"",
		"config/production/hugo.yaml": "theme: prod",
	})
	if got := HugoConfigValue(dir, "theme"); got != "prod" {
		t.Errorf("theme = %v, want prod", got)
	}
	if got := HugoConfigValue(dir, "TITLE"); got != "Default" {
		t.Errorf("title = %v, want Default", got)
	}
	if got := HugoConfigValue(dir, "missing"); got != nil {
		t.Errorf("missing = %v, want nil", got)
	}
}

func TestIsHugoConfigFileName(t *testing.T) {
	for name, want := range map[string]bool{
		"hugo.toml": true, "hugo.yml": true, "config.json": true, "config.yaml": true,
		"params.toml": false, "hugo.txt": false, "myconfig.toml": false,
	} {
		if got := IsHugoConfigFileName(name); got != want {
			t.Errorf("IsHugoConfigFileName(%q) = %v, want %v", name, got, want)
		}
	}
}

func TestSiteKind(t *testing.T) {
	tests := []struct {
		name          string
		files         map[string]string
		wantHugo      bool
		wantPublish   string
		wantLoadError bool
	}{
		{"static overrides a Hugo config", map[string]string{"walgo.yaml": "site:\n  kind: static\n", "hugo.toml": "publishDir = \"out\""}, false, "public", false},
		{"hugo overrides missing config", map[string]string{"walgo.yaml": "site:\n  kind: hugo\n"}, true, "public", false},
		{"hugo makes an ambiguous config.yaml count", map[string]string{"walgo.yaml": "site:\n  kind: hugo\n", "config.yaml": "publishDir: out"}, true, "out", false},
		{"unset detects", map[string]string{"walgo.yaml": "walrus:\n  network: testnet\n", "hugo.toml": "publishDir = \"out\""}, true, "out", false},
		{"unknown kind is a load error", map[string]string{"walgo.yaml": "site:\n  kind: jekyll\n"}, false, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := writeSite(t, tt.files)
			cfg, err := LoadConfigFrom(dir)
			if tt.wantLoadError {
				if err == nil {
					t.Fatal("want a load error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got := IsHugoSite(dir); got != tt.wantHugo {
				t.Errorf("IsHugoSite = %v, want %v", got, tt.wantHugo)
			}
			if cfg.HugoConfig.PublishDir != tt.wantPublish {
				t.Errorf("PublishDir = %q, want %q", cfg.HugoConfig.PublishDir, tt.wantPublish)
			}
			if files := IgnoredConfigFiles(dir); len(files) != 0 {
				t.Errorf("IgnoredConfigFiles = %v, want none when site.kind decides", files)
			}
		})
	}
}
