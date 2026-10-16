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

	// config.* is a common name outside Hugo, so it only counts with one of these signals.
	// A misclassified static site would be built by hugo with --cleanDestinationDir.
	hugoOnlyKeys = map[string]bool{
		"baseurl": true, "theme": true, "languagecode": true, "defaultcontentlanguage": true,
		"module": true, "markup": true, "taxonomies": true, "permalinks": true,
	}
	hugoSiteDirs = []string{"layouts", "themes", "archetypes"}
)

// hugoConfigs returns the site's Hugo configs with lowercased top-level keys, highest precedence first.
// A config that exists but does not parse is returned empty, so hugo still runs and reports the syntax error.
func hugoConfigs(sitePath string) []map[string]any {
	var configs []map[string]any
	for _, dir := range hugoConfigDirs {
		for _, base := range []string{"hugo", "config"} {
			for _, ext := range hugoConfigExtensions {
				values, exists, parsed := parseConfigFile(filepath.Join(sitePath, dir, base+ext))
				if !exists {
					continue
				}
				if base == "hugo" || (parsed && hasHugoKey(values)) || hasHugoSiteDir(sitePath) {
					configs = append(configs, values)
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

// defaultPublishDir returns the publishDir set in the site's Hugo config, else Hugo's default "public".
func defaultPublishDir(sitePath string) string {
	for _, values := range hugoConfigs(sitePath) {
		if dir, ok := values["publishdir"].(string); ok && dir != "" {
			return dir
		}
	}
	return "public"
}

func parseConfigFile(file string) (values map[string]any, exists, parsed bool) {
	info, err := os.Stat(file)
	if err != nil || info.IsDir() {
		return nil, false, false
	}
	data, err := os.ReadFile(file) // #nosec G304 - path is built from the site directory and fixed names
	if err != nil {
		return nil, true, false
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
		return nil, true, false
	}
	// Hugo config keys are case-insensitive.
	values = make(map[string]any, len(raw))
	for key, value := range raw {
		values[strings.ToLower(key)] = value
	}
	return values, true, true
}

func hasHugoKey(values map[string]any) bool {
	for key := range values {
		if hugoOnlyKeys[key] {
			return true
		}
	}
	return false
}

func hasHugoSiteDir(sitePath string) bool {
	for _, dir := range hugoSiteDirs {
		if info, err := os.Stat(filepath.Join(sitePath, dir)); err == nil && info.IsDir() {
			return true
		}
	}
	return false
}
