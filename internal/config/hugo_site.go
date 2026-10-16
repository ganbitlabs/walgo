package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

var (
	hugoConfigExtensions = []string{".toml", ".yaml", ".yml", ".json"}
	// Directories searched for a Hugo config, highest precedence first.
	// walgo builds with --environment production, so that environment's directory wins.
	hugoConfigDirs = []string{".", filepath.Join("config", "production"), filepath.Join("config", "_default")}
)

// isAmbiguousConfigName reports whether a config file name is common outside Hugo.
// A static site misread as Hugo would be built by hugo with --cleanDestinationDir, so these
// names also need a Hugo-only site structure. Their contents are not trusted as a signal.
func isAmbiguousConfigName(dir, base, ext string) bool {
	return dir == "." && base == "config" && ext != ".toml"
}

// hasHugoSiteStructure reports whether sitePath has archetypes/ or Markdown under content/.
func hasHugoSiteStructure(sitePath string) bool {
	if info, err := os.Stat(filepath.Join(sitePath, "archetypes")); err == nil && info.IsDir() {
		return true
	}
	matches, _ := filepath.Glob(filepath.Join(sitePath, "content", "*.md"))
	if len(matches) > 0 {
		return true
	}
	matches, _ = filepath.Glob(filepath.Join(sitePath, "content", "*", "*.md"))
	return len(matches) > 0
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// hugoConfigFiles returns the site's Hugo config files, highest precedence first.
// An ambiguous root config.{yaml,yml,json} counts only beside a Hugo site structure or with site.kind: hugo.
func hugoConfigFiles(sitePath string) []string {
	trustAmbiguous := hasHugoSiteStructure(sitePath) || siteKindOverride(sitePath) == SiteKindHugo
	var files []string
	for _, dir := range hugoConfigDirs {
		for _, base := range []string{"hugo", "config"} {
			for _, ext := range hugoConfigExtensions {
				path := filepath.Join(sitePath, dir, base+ext)
				if fileExists(path) && (!isAmbiguousConfigName(dir, base, ext) || trustAmbiguous) {
					files = append(files, path)
				}
			}
		}
	}
	return files
}

// hugoConfigs returns the site's Hugo configs with lowercased top-level keys, highest precedence first.
// A config that exists but does not parse is returned empty, so hugo still runs and reports the syntax error.
func hugoConfigs(sitePath string) []map[string]any {
	var configs []map[string]any
	for _, file := range hugoConfigFiles(sitePath) {
		configs = append(configs, parseConfigFile(file))
	}
	return configs
}

// HugoConfigFile returns the Hugo config file to read or edit site-wide settings in: a root config
// (hugo.* before config.*), else config/_default, else config/production. It returns "" for a site
// with no Hugo config.
func HugoConfigFile(sitePath string) string {
	files := hugoConfigFiles(sitePath)
	for _, dir := range []string{sitePath, filepath.Join(sitePath, "config", "_default")} {
		for _, file := range files {
			if filepath.Dir(file) == filepath.Clean(dir) {
				return file
			}
		}
	}
	if len(files) > 0 {
		return files[0]
	}
	return ""
}

// HugoTOMLConfigFile returns HugoConfigFile when it is TOML, the only format walgo edits in place.
func HugoTOMLConfigFile(sitePath string) (string, error) {
	file := HugoConfigFile(sitePath)
	switch {
	case file == "":
		return "", fmt.Errorf("no Hugo config (hugo.toml, hugo.yaml, config.toml, config/_default/...) found in %s", sitePath)
	case filepath.Ext(file) != ".toml":
		return "", fmt.Errorf("walgo edits only TOML Hugo configs; update %s by hand", file)
	}
	return file, nil
}

// HugoConfigValue returns the top-level key from the site's Hugo config, matched case-insensitively,
// taking the highest-precedence file that sets it. It returns nil when no config sets the key.
func HugoConfigValue(sitePath, key string) any {
	key = strings.ToLower(key)
	for _, values := range hugoConfigs(sitePath) {
		if v, ok := values[key]; ok {
			return v
		}
	}
	return nil
}

// IsHugoConfigFileName reports whether name is a Hugo config file name (hugo.* or config.*).
func IsHugoConfigFileName(name string) bool {
	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)
	if base != "hugo" && base != "config" {
		return false
	}
	for _, e := range hugoConfigExtensions {
		if ext == e {
			return true
		}
	}
	return false
}

// siteKindOverride returns site.kind from sitePath's walgo.yaml, or "" when it is unset or unreadable.
func siteKindOverride(sitePath string) SiteKind {
	data, err := os.ReadFile(filepath.Join(sitePath, DefaultConfigFileName)) // #nosec G304 - fixed name in the site directory
	if err != nil {
		return ""
	}
	var cfg struct {
		Site SiteConfig `yaml:"site"`
	}
	if yaml.Unmarshal(data, &cfg) != nil {
		return ""
	}
	return cfg.Site.Kind
}

// IsHugoSite reports whether walgo builds sitePath with hugo: walgo.yaml's site.kind when set,
// else whether the site has a Hugo config file.
func IsHugoSite(sitePath string) bool {
	switch siteKindOverride(sitePath) {
	case SiteKindHugo:
		return true
	case SiteKindStatic:
		return false
	}
	return len(hugoConfigFiles(sitePath)) > 0
}

// IgnoredConfigFiles returns root config.{yaml,yml,json} files that were not treated as Hugo config.
// It returns nil when walgo.yaml sets site.kind, since detection did not decide.
func IgnoredConfigFiles(sitePath string) []string {
	if hasHugoSiteStructure(sitePath) || siteKindOverride(sitePath) != "" {
		return nil
	}
	var files []string
	for _, ext := range hugoConfigExtensions {
		path := filepath.Join(sitePath, "config"+ext)
		if isAmbiguousConfigName(".", "config", ext) && fileExists(path) {
			files = append(files, path)
		}
	}
	return files
}

// defaultPublishDir returns the publishDir set in the site's Hugo config, else Hugo's default "public".
func defaultPublishDir(sitePath string) string {
	for _, values := range hugoConfigs(sitePath) {
		if dir, ok := values["publishdir"].(string); ok && dir != "" {
			return dir
		}
	}
	return "public"
}

// parseConfigFile returns the file's top-level keys lowercased, as Hugo matches them case-insensitively.
// An unreadable or malformed file yields an empty map.
func parseConfigFile(file string) map[string]any {
	values := map[string]any{}
	data, err := os.ReadFile(file) // #nosec G304 - path is built from the site directory and fixed names
	if err != nil {
		return values
	}
	raw := map[string]any{}
	switch filepath.Ext(file) {
	case ".toml":
		err = toml.Unmarshal(data, &raw)
	case ".json":
		err = json.Unmarshal(data, &raw)
	default:
		err = yaml.Unmarshal(data, &raw)
	}
	if err != nil {
		return values
	}
	for key, value := range raw {
		values[strings.ToLower(key)] = value
	}
	return values
}
