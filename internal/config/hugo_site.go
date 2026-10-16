package config

import (
	"encoding/json"
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

// hugoConfigs returns the site's Hugo configs with lowercased top-level keys, highest precedence first.
// A config that exists but does not parse is returned empty, so hugo still runs and reports the syntax error.
func hugoConfigs(sitePath string) []map[string]any {
	var configs []map[string]any
	for _, dir := range hugoConfigDirs {
		for _, base := range []string{"hugo", "config"} {
			for _, ext := range hugoConfigExtensions {
				path := filepath.Join(sitePath, dir, base+ext)
				if fileExists(path) && (!isAmbiguousConfigName(dir, base, ext) || hasHugoSiteStructure(sitePath)) {
					configs = append(configs, parseConfigFile(path))
				}
			}
		}
	}
	return configs
}

// IsHugoSite reports whether sitePath has a Hugo config file.
func IsHugoSite(sitePath string) bool {
	return len(hugoConfigs(sitePath)) > 0
}

// IgnoredConfigFiles returns root config.{yaml,yml,json} files that were not treated as Hugo config.
func IgnoredConfigFiles(sitePath string) []string {
	if hasHugoSiteStructure(sitePath) {
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
// An absolute Hugo publishDir is ignored: callers join PublishDir onto the site path, and BuildSite
// passes --destination, so the build and every caller agree on the same directory.
func defaultPublishDir(sitePath string) string {
	for _, values := range hugoConfigs(sitePath) {
		if dir, ok := values["publishdir"].(string); ok && dir != "" && !filepath.IsAbs(dir) {
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
