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
	hugoConfigBaseNames  = []string{"hugo", "config"}
	hugoConfigExtensions = []string{".toml", ".yaml", ".yml", ".json"}
	// Directories searched for a Hugo config, highest precedence first.
	// walgo builds with --environment production, so that environment's directory wins.
	hugoConfigDirs = []string{".", filepath.Join("config", "production"), filepath.Join("config", "_default")}
)

// hugoConfigFiles returns the Hugo config files present in sitePath, highest precedence first.
func hugoConfigFiles(sitePath string) []string {
	var files []string
	for _, dir := range hugoConfigDirs {
		for _, base := range hugoConfigBaseNames {
			for _, ext := range hugoConfigExtensions {
				path := filepath.Join(sitePath, dir, base+ext)
				if info, err := os.Stat(path); err == nil && !info.IsDir() {
					files = append(files, path)
				}
			}
		}
	}
	return files
}

// IsHugoSite reports whether sitePath has a Hugo config file.
func IsHugoSite(sitePath string) bool {
	return len(hugoConfigFiles(sitePath)) > 0
}

// defaultPublishDir returns the publishDir set in the site's Hugo config, else Hugo's default "public".
func defaultPublishDir(sitePath string) string {
	for _, file := range hugoConfigFiles(sitePath) {
		if dir := readHugoPublishDir(file); dir != "" {
			return dir
		}
	}
	return "public"
}

func readHugoPublishDir(file string) string {
	data, err := os.ReadFile(file) // #nosec G304 - path is built from the site directory and fixed names
	if err != nil {
		return ""
	}
	values := map[string]any{}
	switch filepath.Ext(file) {
	case ".toml":
		err = toml.Unmarshal(data, &values)
	case ".json":
		err = json.Unmarshal(data, &values)
	default:
		err = yaml.Unmarshal(data, &values)
	}
	if err != nil {
		return ""
	}
	// Hugo config keys are case-insensitive.
	for key, value := range values {
		if strings.EqualFold(key, "publishDir") {
			if dir, ok := value.(string); ok {
				return dir
			}
		}
	}
	return ""
}
