package hugo

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/ganbitlabs/walgo/internal/ai"
	"github.com/ganbitlabs/walgo/internal/config"
)

// DetectSiteType detects the site type from the Hugo theme, falling back to the content structure.
// It defaults to Blog.
func DetectSiteType(sitePath string) ai.SiteType {
	return siteTypeForTheme(GetThemeName(sitePath), sitePath)
}

// siteTypeForTheme maps a theme name to a site type, using the content structure for general-purpose themes.
func siteTypeForTheme(theme string, sitePath string) ai.SiteType {
	theme = strings.ToLower(theme)

	if strings.Contains(theme, "walgo-biolink") {
		return ai.SiteTypeBiolink
	}
	if strings.Contains(theme, "walgo-whitepaper") {
		return ai.SiteTypeWhitepaper
	}

	docThemePatterns := []string{"hugo-book", "docsy", "learn", "doks", "geekdoc", "techdoc"}
	for _, pattern := range docThemePatterns {
		if strings.Contains(theme, pattern) {
			return ai.SiteTypeDocs
		}
	}

	return detectSiteTypeFromContent(sitePath)
}

// detectSiteTypeFromContent analyzes content directory to determine site type.
func detectSiteTypeFromContent(sitePath string) ai.SiteType {
	contentDir := filepath.Join(sitePath, "content")

	// Check for whitepaper directory (whitepaper site indicator)
	if dirExists(filepath.Join(contentDir, "whitepaper")) {
		return ai.SiteTypeWhitepaper
	}

	// Check for docs directory (documentation site indicator)
	if dirExists(filepath.Join(contentDir, "docs")) {
		return ai.SiteTypeDocs
	}

	// Check for biolink indicators (links/ directory or biolink-style config)
	if dirExists(filepath.Join(contentDir, "links")) || dirExists(filepath.Join(contentDir, "biolink")) {
		return ai.SiteTypeBiolink
	}

	// Default to blog
	return ai.SiteTypeBlog
}

// dirExists checks if a directory exists.
func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// GetThemeName returns the theme set in the site's Hugo config, the first one when it lists several.
// It returns "" when no Hugo config sets a theme.
func GetThemeName(sitePath string) string {
	switch theme := config.HugoConfigValue(sitePath, "theme").(type) {
	case string:
		return theme
	case []any:
		if len(theme) > 0 {
			if first, ok := theme[0].(string); ok {
				return first
			}
		}
	}
	return ""
}
