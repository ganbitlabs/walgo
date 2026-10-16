package hugo

import (
	"archive/zip"
	"embed"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/ganbitlabs/walgo/internal/ai"
	"github.com/ganbitlabs/walgo/internal/compress"
	"github.com/ganbitlabs/walgo/internal/config"
	"github.com/ganbitlabs/walgo/internal/deps"
	"github.com/ganbitlabs/walgo/internal/executil"
	"github.com/ganbitlabs/walgo/internal/optimizer"
	"github.com/ganbitlabs/walgo/internal/projects"
)

//go:embed tomls/*.toml
var tomlsFS embed.FS

//go:embed assets/favicon.svg
var faviconSVG []byte

// ContentType represents a detected content type in a Hugo site
type ContentType struct {
	Name      string // Name of the content type (e.g., "posts", "pages")
	Path      string // Path to the content type directory
	FileCount int    // Number of files in this content type
}

// DetectContentTypes scans the content directory and returns found content types
func DetectContentTypes(sitePath string) ([]ContentType, error) {
	contentDir := filepath.Join(sitePath, "content")

	if _, err := os.Stat(contentDir); os.IsNotExist(err) {
		return nil, nil
	}

	entries, err := os.ReadDir(contentDir)
	if err != nil {
		return nil, fmt.Errorf("reading content directory: %w", err)
	}

	var types []ContentType
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Skip hidden directories
		if name := entry.Name(); len(name) > 0 && name[0] == '.' {
			continue
		}

		// Count files in this directory
		typePath := filepath.Join(contentDir, entry.Name())
		fileCount := countMarkdownFiles(typePath)

		types = append(types, ContentType{
			Name:      entry.Name(),
			Path:      typePath,
			FileCount: fileCount,
		})
	}

	// Sort by file count (descending)
	sort.Slice(types, func(i, j int) bool {
		return types[i].FileCount > types[j].FileCount
	})

	return types, nil
}

// GetDefaultContentType returns the default content type for a Hugo site
func GetDefaultContentType(sitePath string) string {
	types, err := DetectContentTypes(sitePath)
	if err != nil || len(types) == 0 {
		return "posts"
	}

	// Return the most used content type
	return types[0].Name
}

// countMarkdownFiles counts .md files recursively in a directory.
// Returns -1 if the directory cannot be walked.
func countMarkdownFiles(dir string) int {
	count := 0
	if err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && filepath.Ext(path) == ".md" {
			count++
		}
		return nil
	}); err != nil {
		return -1
	}
	return count
}

// InitializeSite creates a new Hugo site at the given path.
func InitializeSite(sitePath string) error {
	hugoPath, err := deps.LookPath("hugo")
	if err != nil {
		return fmt.Errorf("hugo is not installed or not found in PATH")
	}

	cmd := executil.Command(hugoPath, "new", "site", ".", "--format", "toml")
	cmd.Dir = sitePath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to initialize Hugo site at %s: %v\nOutput: %s", sitePath, err, string(output))
	}

	fmt.Printf("Hugo `new site` command output:\n%s\n", string(output))

	return nil
}

// cleanPublicDir removes OS metadata files from the publish directory.
// Everything else there is either build output or the user's own files, so it stays.
func cleanPublicDir(publicDir string) error {
	var deletedCount int

	err := filepath.Walk(publicDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || info.Name() != ".DS_Store" {
			return nil
		}
		if err := os.Remove(path); err != nil {
			fmt.Printf("Warning: failed to delete %s: %v\n", path, err)
		} else {
			deletedCount++
		}
		return nil
	})

	if deletedCount > 0 {
		fmt.Printf("Cleaned up %d unnecessary files from public directory\n", deletedCount)
	}

	return err
}

// resolvePublishDir returns the absolute publish directory for sitePath.
// It refuses a directory that contains the site: hugo's --cleanDestinationDir would delete the sources,
// and a static deploy would upload walgo.yaml and everything else in the project.
func resolvePublishDir(sitePath string, cfg *config.WalgoConfig) (string, error) {
	dir := cfg.HugoConfig.PublishDir
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(sitePath, dir)
	}
	publishDir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	absSite, err := filepath.Abs(sitePath)
	if err != nil {
		return "", err
	}
	if rel, err := filepath.Rel(publishDir, absSite); err == nil && !strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("publish directory %s contains the site itself; set hugo.publishDir in walgo.yaml to a subdirectory such as public", publishDir)
	}
	return publishDir, nil
}

// BuildSite builds the site in sitePath into its publish directory, then optimizes it and writes ws-resources.json.
// The publish directory is walgo.yaml's hugo.publishDir, else the Hugo config's publishDir, else public.
// A site with no Hugo config is treated as prebuilt static files and its publish directory is used as-is.
func BuildSite(sitePath string) error {
	cfg, err := config.LoadConfigFrom(sitePath)
	if err != nil {
		return err
	}
	publishDir, err := resolvePublishDir(sitePath, cfg)
	if err != nil {
		return fmt.Errorf("resolving publish directory: %w", err)
	}

	if config.IsHugoSite(sitePath) {
		if err := runHugoBuild(sitePath, publishDir); err != nil {
			return err
		}
		if cfg.OptimizerConfig.Enabled {
			fmt.Printf("Optimizing assets...\n")
			optimizerEngine := optimizer.NewEngine(cfg.OptimizerConfig)
			stats, err := optimizerEngine.OptimizeDirectory(publishDir)
			if err != nil {
				return fmt.Errorf("failed to optimize directory: %w", err)
			}
			optimizerEngine.PrintStats(stats)
		}
	} else {
		var hints []string
		for _, file := range config.IgnoredConfigFiles(sitePath) {
			hints = append(hints, fmt.Sprintf("%s was not treated as a Hugo config; if this is a Hugo site, rename it to hugo%s", file, filepath.Ext(file)))
		}
		if info, err := os.Stat(publishDir); err != nil || !info.IsDir() {
			msg := fmt.Sprintf("no Hugo config (hugo.toml, config.toml, ...) found in %s and publish directory %s does not exist", sitePath, publishDir)
			for _, hint := range hints {
				msg += "\n  " + hint
			}
			return errors.New(msg)
		}
		fmt.Printf("No Hugo config found; deploying static files from %s as-is.\n", publishDir)
		for _, hint := range hints {
			fmt.Fprintf(os.Stderr, "Warning: %s.\n", hint)
		}
	}

	if err := cleanPublicDir(publishDir); err != nil {
		fmt.Printf("Warning: failed to clean public directory: %v\n", err)
	}

	if cfg.CompressConfig.GenerateWSResources {
		if err := writeWSResources(sitePath, publishDir, cfg); err != nil {
			return err
		}
	}

	return nil
}

func runHugoBuild(sitePath, publishDir string) error {
	hugoPath, err := deps.LookPath("hugo")
	if err != nil {
		return fmt.Errorf("hugo is not installed or not found in PATH")
	}

	fmt.Printf("Building Hugo site in %s...\n", sitePath)

	themesDir := filepath.Join(sitePath, "themes")
	if _, err := os.Stat(themesDir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Warning: themes/ directory not found. Hugo might fail if a theme is required.\n")
	} else {
		entries, err := os.ReadDir(themesDir)
		if err == nil && len(entries) == 0 {
			fmt.Fprintf(os.Stderr, "Warning: themes/ directory is empty. Hugo might fail if a theme is required.\n")
		}
	}

	cmd := executil.Command(hugoPath, "build", "--environment", "production", "--minify", "--gc", "--cleanDestinationDir", "--destination", publishDir)
	cmd.Dir = sitePath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "\nHugo build failed. Common issues:\n")
		fmt.Fprintf(os.Stderr, "  1. Missing theme - check if themes/ directory contains your theme\n")
		fmt.Fprintf(os.Stderr, "  2. Configuration error - verify your Hugo config (hugo.toml, config.toml, ...) is valid\n")
		fmt.Fprintf(os.Stderr, "  3. Content error - check your markdown files for syntax issues\n")
		fmt.Fprintf(os.Stderr, "\nFor more details, run: hugo build --verbose\n")
		return fmt.Errorf("failed to build Hugo site: %v (check Hugo output above for details)", err)
	}

	fmt.Println("Hugo site built successfully.")
	fmt.Printf("Static files generated in: %s\n", publishDir)
	return nil
}

func writeWSResources(sitePath, publicDir string, cfg *config.WalgoConfig) error {
	fmt.Printf("Generating ws-resources.json...\n")

	cacheConfig := compress.DefaultCacheControlConfig()
	cacheConfig.Enabled = cfg.CacheConfig.Enabled
	if cfg.CacheConfig.ImmutableMaxAge != 0 {
		cacheConfig.ImmutableMaxAge = cfg.CacheConfig.ImmutableMaxAge
	}
	if cfg.CacheConfig.MutableMaxAge != 0 {
		cacheConfig.MutableMaxAge = cfg.CacheConfig.MutableMaxAge
	}

	wsOptions := compress.WSResourcesOptions{
		CacheConfig:  cacheConfig,
		CustomRoutes: cfg.CompressConfig.CustomRoutes,
		CustomIgnore: cfg.CompressConfig.IgnorePatterns,
		SiteName:     filepath.Base(sitePath),
		ImageURL:     compress.DefaultImageURL,
		Link:         compress.DefaultLink,
		ProjectURL:   compress.DefaultProjectURL,
		Creator:      compress.DefaultCreator,
		Category:     compress.DefaultCategory,
	}

	var existingObjectID string
	pm, err := projects.NewManager()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not access project database: %v (using defaults)\n", err)
	} else {
		defer pm.Close()
		if project, err := pm.GetProjectBySitePath(sitePath); err == nil && project != nil {
			wsOptions.SiteName = project.Name
			wsOptions.Description = project.Description
			wsOptions.ImageURL = project.ImageURL
			wsOptions.Category = project.Category
			existingObjectID = project.ObjectID
		}
	}

	outputPath := filepath.Join(publicDir, "ws-resources.json")
	if existingObjectID == "" {
		if oldConfig, err := compress.ReadWSResourcesConfig(outputPath); err == nil && oldConfig.ObjectID != "" {
			existingObjectID = oldConfig.ObjectID
		}
	}

	wsConfig, err := compress.GenerateWSResourcesConfig(publicDir, wsOptions)
	if err != nil {
		return fmt.Errorf("failed to generate ws-resources.json: %w", err)
	}
	if existingObjectID != "" {
		wsConfig.ObjectID = existingObjectID
	}

	if err := compress.WriteWSResourcesConfig(wsConfig, outputPath); err != nil {
		return fmt.Errorf("failed to write ws-resources.json: %w", err)
	}
	fmt.Printf("Generated ws-resources.json (%d resources)\n", len(wsConfig.Headers))

	routes, err := compress.GenerateRoutesFromPublic(publicDir)
	if err != nil {
		return fmt.Errorf("failed to generate routes: %w", err)
	}
	if err := compress.MergeRoutesIntoWSResources(outputPath, routes); err != nil {
		return fmt.Errorf("failed to merge routes into ws-resources.json: %w", err)
	}
	return nil
}

// ServeSite starts the Hugo development server
func ServeSite(sitePath string) error {
	hugoPath, err := deps.LookPath("hugo")
	if err != nil {
		return fmt.Errorf("hugo is not installed or not found in PATH")
	}

	fmt.Printf("Starting Hugo development server...\n")
	cmd := executil.Command(hugoPath, "server", "--environment", "development", "--bind", "0.0.0.0",
		"--port", "1313", "--buildDrafts", "--buildFuture", "--disableFastRender", "--noHTTPCache")
	cmd.Dir = sitePath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

// CreateContent creates new content using Hugo's new command
func CreateContent(sitePath, contentPath string) error {
	// Validate contentPath to prevent path traversal
	if strings.Contains(contentPath, "..") || filepath.IsAbs(contentPath) {
		return fmt.Errorf("invalid content path: must be relative without '..'")
	}

	hugoPath, err := deps.LookPath("hugo")
	if err != nil {
		return fmt.Errorf("hugo is not installed or not found in PATH")
	}

	cmd := executil.Command(hugoPath, "new", contentPath)
	cmd.Dir = sitePath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("creating content: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// SiteType represents the type of Hugo site
type SiteType string

const (
	SiteTypeBlog       SiteType = "blog"
	SiteTypeDocs       SiteType = "docs"
	SiteTypeBiolink    SiteType = "biolink"
	SiteTypeWhitepaper SiteType = "whitepaper"
)

// GetSiteConfig returns the embedded TOML config for a site type
func getSiteConfig(siteType SiteType) ([]byte, error) {
	filename := fmt.Sprintf("tomls/%s.toml", siteType)
	return tomlsFS.ReadFile(filename)
}

// SetupSiteConfigWithName copies the appropriate TOML config to the site directory with site name
func SetupSiteConfigWithName(sitePath string, siteType SiteType, siteName string) error {
	configData, err := getSiteConfig(siteType)
	if err != nil {
		return fmt.Errorf("getting config for %s: %w", siteType, err)
	}

	// Replace site name placeholder (all occurrences)
	configStr := string(configData)
	if siteName == "" {
		// Use directory name as default site name
		siteName = filepath.Base(sitePath)
	}
	// Sanitize siteName to prevent TOML injection (escape quotes and newlines)
	siteName = strings.ReplaceAll(siteName, `"`, `\"`)
	siteName = strings.ReplaceAll(siteName, "\n", " ")
	siteName = strings.ReplaceAll(siteName, "\r", "")
	configStr = strings.ReplaceAll(configStr, "{{SITE_NAME}}", siteName)

	configPath := filepath.Join(sitePath, "hugo.toml")
	if err := os.WriteFile(configPath, []byte(configStr), 0644); err != nil {
		return fmt.Errorf("writing hugo.toml: %w", err)
	}

	return nil
}

// SetupArchetypes creates archetypes based on theme analysis
// This replaces static embedded archetypes with theme-aware dynamic ones
func SetupArchetypes(sitePath, themeName string) error {
	return ai.SetupDynamicArchetypes(sitePath, themeName)
}

// ThemeInfo contains information about a Hugo theme
type ThemeInfo struct {
	Name    string // Theme name
	DirName string // Directory name in themes/
}

// GetThemeInfo returns theme information for a site type
// IMPORTANT: DirName must match the theme name in Hugo config files (tomls/*.toml)
// - blog.toml uses: theme = "ananke"
// - docs.toml uses: theme = "hugo-book"
func GetThemeInfo(siteType SiteType) ThemeInfo {
	switch siteType {
	case SiteTypeBlog:
		return ThemeInfo{
			Name:    "Ananke",
			DirName: "ananke", // Must match: theme = "ananke" in blog.toml
		}
	case SiteTypeDocs:
		return ThemeInfo{
			Name:    "Book",
			DirName: "hugo-book", // Must match: theme = "hugo-book" in docs.toml
		}
	case SiteTypeBiolink:
		return ThemeInfo{
			Name:    "Walgo Biolink",
			DirName: "walgo-biolink", // Must match: theme = "walgo-biolink" in biolink.toml
		}
	case SiteTypeWhitepaper:
		return ThemeInfo{
			Name:    "Walgo Whitepaper",
			DirName: "walgo-whitepaper", // Must match: theme = "walgo-whitepaper" in whitepaper.toml
		}
	default:
		// Default to blog theme
		return ThemeInfo{
			Name:    "Ananke",
			DirName: "ananke", // Must match: theme = "ananke" in blog.toml
		}
	}
}

// InstallTheme installs the default theme for a site type.
// It first tries to download the real theme from GitHub using the RepoURL
// in ai.DefaultThemes. If the download fails (no network, timeout, etc.),
// it falls back to scaffolding a blank theme with `hugo new theme`.
func InstallTheme(sitePath string, siteType SiteType) error {
	theme := GetThemeInfo(siteType)
	themePath := filepath.Join(sitePath, "themes", theme.DirName)

	// Check if theme already exists
	if _, err := os.Stat(themePath); err == nil {
		return nil // Theme already installed
	}

	// Try to download the real theme from GitHub
	aiTheme := ai.GetDefaultTheme(ai.SiteType(siteType))
	if aiTheme.RepoURL != "" {
		fmt.Printf("Downloading theme %s from %s...\n", theme.Name, aiTheme.RepoURL)
		if err := downloadThemeZip(theme.DirName, aiTheme.RepoURL, themePath); err != nil {
			fmt.Printf("Warning: Failed to download theme: %v\n", err)
			fmt.Printf("Falling back to blank theme scaffold...\n")
		} else {
			// Download succeeded — update config for theme-specific params
			fmt.Printf("Updating site configuration for theme '%s'...\n", theme.Name)
			if err := ai.UpdateSiteConfigForTheme(sitePath, theme.DirName); err != nil {
				fmt.Printf("Warning: Could not update config for theme: %v\n", err)
			}
			return nil
		}
	}

	// Fallback: scaffold a blank theme with hugo new theme
	hugoPath, err := deps.LookPath("hugo")
	if err != nil {
		return fmt.Errorf("hugo is not installed or not found in PATH")
	}

	fmt.Printf("Creating theme %s...\n", theme.Name)
	cmd := executil.Command(hugoPath, "new", "theme", theme.DirName)
	cmd.Dir = sitePath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create theme %s: %v\nOutput: %s", theme.DirName, err, string(output))
	}

	// Update site config with theme-specific params
	fmt.Printf("Updating site configuration for theme '%s'...\n", theme.Name)
	if err := ai.UpdateSiteConfigForTheme(sitePath, theme.DirName); err != nil {
		fmt.Printf("Warning: Could not update config for theme: %v\n", err)
	}

	return nil
}

// CopyExampleSite copies the exampleSite directory from a downloaded theme into the
// site root. This provides ready-to-use sample content and configuration from the
// theme author. Files in the site root are NOT overwritten if they already exist
// (e.g., hugo.toml created by SetupSiteConfigWithName is preserved).
func CopyExampleSite(sitePath string, siteType SiteType) error {
	theme := GetThemeInfo(siteType)
	exampleDir := filepath.Join(sitePath, "themes", theme.DirName, "exampleSite")

	if _, err := os.Stat(exampleDir); os.IsNotExist(err) {
		return fmt.Errorf("no exampleSite found in theme %s", theme.DirName)
	}

	copied := 0
	err := filepath.Walk(exampleDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip unreadable files
		}

		relPath, err := filepath.Rel(exampleDir, path)
		if err != nil {
			return nil
		}
		if relPath == "." {
			return nil
		}

		targetPath := filepath.Join(sitePath, relPath)

		if info.IsDir() {
			return os.MkdirAll(targetPath, 0755)
		}

		// Skip config files — our TOML template takes precedence
		base := filepath.Base(relPath)
		if base == "hugo.toml" || base == "config.toml" || base == "hugo.yaml" || base == "config.yaml" {
			return nil
		}

		// Don't overwrite existing files
		if _, err := os.Stat(targetPath); err == nil {
			return nil
		}

		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return err
		}

		// Copy file
		if err := copyFileContents(path, targetPath, info.Mode()); err != nil {
			return err
		}
		copied++
		return nil
	})

	if err != nil {
		return fmt.Errorf("copying exampleSite: %w", err)
	}

	fmt.Printf("Copied %d files from theme exampleSite\n", copied)
	return nil
}

// CopyExampleSiteWithConfig copies the entire exampleSite from a downloaded theme,
// including config files (hugo.toml, config.toml, etc.), and replaces the title
// with the provided site name. This is used for non-blog site types (docs, biolink,
// whitepaper) where the exampleSite config should be used directly instead of our
// embedded TOML templates.
func CopyExampleSiteWithConfig(sitePath string, siteType SiteType, siteName string) error {
	theme := GetThemeInfo(siteType)
	exampleDir := filepath.Join(sitePath, "themes", theme.DirName, "exampleSite")

	if _, err := os.Stat(exampleDir); os.IsNotExist(err) {
		return fmt.Errorf("no exampleSite found in theme %s", theme.DirName)
	}

	copied := 0
	err := filepath.Walk(exampleDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip unreadable files
		}

		relPath, err := filepath.Rel(exampleDir, path)
		if err != nil {
			return nil
		}
		if relPath == "." {
			return nil
		}

		targetPath := filepath.Join(sitePath, relPath)

		if info.IsDir() {
			return os.MkdirAll(targetPath, 0755)
		}

		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return err
		}

		// Copy file (overwrite existing — exampleSite takes precedence)
		if err := copyFileContents(path, targetPath, info.Mode()); err != nil {
			return err
		}
		copied++
		return nil
	})

	if err != nil {
		return fmt.Errorf("copying exampleSite: %w", err)
	}

	// Fix config for standalone use: title, baseURL, enableGitInfo
	if err := fixExampleSiteConfig(sitePath, siteName); err != nil {
		return fmt.Errorf("fixing example site config: %w", err)
	}

	fmt.Printf("Copied %d files from theme exampleSite (with config)\n", copied)
	return nil
}

// fixExampleSiteConfig adjusts the exampleSite's config for standalone use:
// - Sets the title to the site name
// - Sets baseURL to "/" (Walrus uses relative URLs)
// - Removes enableGitInfo (site may not be a git repo)
func fixExampleSiteConfig(sitePath, siteName string) error {
	// Sanitize siteName to prevent TOML injection
	safeName := strings.ReplaceAll(siteName, `"`, `\"`)
	safeName = strings.ReplaceAll(safeName, "\n", " ")
	safeName = strings.ReplaceAll(safeName, "\r", "")

	for _, configName := range []string{"hugo.toml", "config.toml"} {
		configPath := filepath.Join(sitePath, configName)
		data, err := os.ReadFile(configPath)
		if err != nil {
			continue
		}

		lines := strings.Split(string(data), "\n")
		// Track whether we're at top-level scope (before any [section] header).
		// Only replace title/baseURL/enableGitInfo at top level to avoid
		// clobbering identically-named keys inside [params] or other sections.
		topLevel := true
		for i, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "[") {
				topLevel = false
			}
			if !topLevel {
				continue
			}
			if safeName != "" && strings.HasPrefix(trimmed, "title") && !strings.HasPrefix(trimmed, "titleSeparator") {
				lines[i] = fmt.Sprintf(`title = "%s"`, safeName)
			} else if strings.HasPrefix(trimmed, "baseURL") {
				lines[i] = `baseURL = "/"`
			} else if strings.HasPrefix(trimmed, "enableGitInfo") {
				lines[i] = "enableGitInfo = false"
			}
		}

		if err := os.WriteFile(configPath, []byte(strings.Join(lines, "\n")), 0644); err != nil {
			return fmt.Errorf("writing config %s: %w", configName, err)
		}
		return nil
	}
	return nil
}

// copyFileContents copies a file from src to dst, closing handles promptly on error.
func copyFileContents(src, dst string, mode os.FileMode) error {
	srcFile, err := os.Open(src) // #nosec G304 - path validated by caller
	if err != nil {
		return err
	}

	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode) // #nosec G304 - path validated by caller
	if err != nil {
		srcFile.Close()
		return err
	}

	_, copyErr := io.Copy(dstFile, srcFile)
	srcFile.Close()
	closeErr := dstFile.Close()

	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

// CreateTheme scaffolds a new Hugo theme using `hugo new theme` and updates the site config.
func CreateTheme(sitePath, themeName string) error {
	themePath := filepath.Join(sitePath, "themes", themeName)

	// Check if theme already exists
	if _, err := os.Stat(themePath); err == nil {
		return fmt.Errorf("theme %q already exists at %s", themeName, themePath)
	}

	hugoPath, err := deps.LookPath("hugo")
	if err != nil {
		return fmt.Errorf("hugo is not installed or not found in PATH")
	}

	cmd := executil.Command(hugoPath, "new", "theme", themeName)
	cmd.Dir = sitePath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create theme %s: %v\nOutput: %s", themeName, err, string(output))
	}

	// Update hugo.toml with the new theme name
	if err := updateHugoConfigTheme(sitePath, themeName); err != nil {
		return fmt.Errorf("failed to update config with theme name: %w", err)
	}

	return nil
}

// parseGitHubRepo validates a GitHub repository URL and returns its owner and
// repo name. Only https://github.com/owner/repo (optionally with a trailing
// slash and/or .git suffix) is accepted; queries, fragments, userinfo, percent
// escapes and extra path segments are rejected. Callers must derive every
// filesystem path and download URL from the returned parts, never from raw.
func parseGitHubRepo(raw string) (owner, repo string, err error) {
	if strings.ContainsAny(raw, "?#%\\") {
		return "", "", fmt.Errorf("GitHub URL must not contain query, fragment, escapes or backslashes")
	}
	trimmed := strings.TrimSuffix(raw, "/")
	trimmed = strings.TrimSuffix(trimmed, ".git")
	trimmed = strings.TrimSuffix(trimmed, "/")
	u, err := url.Parse(trimmed)
	if err != nil {
		return "", "", fmt.Errorf("invalid GitHub URL: %w", err)
	}
	host := strings.ToLower(u.Hostname())
	if u.Scheme != "https" || (host != "github.com" && host != "www.github.com") ||
		u.User != nil || u.Port() != "" || u.RawQuery != "" || u.Fragment != "" || u.Opaque != "" {
		return "", "", fmt.Errorf("only https://github.com URLs are supported")
	}
	segs := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")
	if len(segs) != 2 {
		return "", "", fmt.Errorf("GitHub URL must be of the form https://github.com/owner/repo")
	}
	for _, seg := range segs {
		if !isSafeDirName(seg) {
			return "", "", fmt.Errorf("invalid GitHub URL path segment %q", seg)
		}
	}
	return segs[0], segs[1], nil
}

// isSafeDirName reports whether name is a plain single path element made of
// letters, digits, '.', '_' and '-' (and not "." or "..").
func isSafeDirName(name string) bool {
	if name == "" || name == "." || name == ".." {
		return false
	}
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '.', r == '_', r == '-':
		default:
			return false
		}
	}
	return true
}

// themeChildPath returns filepath.Join(themesDir, name) after checking that it
// is a direct child of themesDir. Run it before any extract or RemoveAll.
func themeChildPath(themesDir, name string) (string, error) {
	if !isSafeDirName(name) {
		return "", fmt.Errorf("invalid theme directory name %q", name)
	}
	target := filepath.Join(themesDir, name)
	rel, err := filepath.Rel(themesDir, target)
	if err != nil || rel != name || strings.ContainsRune(rel, filepath.Separator) {
		return "", fmt.Errorf("theme path %q is not a direct child of %s", name, themesDir)
	}
	return target, nil
}

// safeExtractPath joins an archive entry name onto dest and rejects entries
// whose cleaned target escapes dest.
func safeExtractPath(dest, name string) (string, error) {
	target := filepath.Join(dest, filepath.FromSlash(name))
	rel, err := filepath.Rel(dest, target)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return "", fmt.Errorf("entry %q escapes theme directory", name)
	}
	return target, nil
}

// downloadThemeZip downloads a theme as a ZIP archive from GitHub and extracts it
func downloadThemeZip(themeName, gitURL, themePath string) error {
	// Convert git URL to ZIP download URL
	// Example: https://github.com/theNewDynamic/gohugo-theme-ananke.git
	//       -> https://github.com/theNewDynamic/gohugo-theme-ananke/archive/refs/heads/master.zip
	owner, repo, err := parseGitHubRepo(gitURL)
	if err != nil {
		return err
	}
	repoURL := "https://github.com/" + owner + "/" + repo
	if !isSafeDirName(themeName) || filepath.Base(themePath) != themeName {
		return fmt.Errorf("invalid theme target %q", themePath)
	}

	// Try main branch first, then master (GitHub default branch varies)
	branches := []string{"main", "master"}
	var zipURL string
	var resp *http.Response

	downloadTimeout := 2 * time.Minute
	if envTimeout := os.Getenv("WALGO_THEME_DOWNLOAD_TIMEOUT"); envTimeout != "" {
		if d, err := time.ParseDuration(envTimeout); err == nil && d > 0 {
			downloadTimeout = d
		}
	}
	client := &http.Client{Timeout: downloadTimeout}

	for _, branch := range branches {
		zipURL = repoURL + "/archive/refs/heads/" + branch + ".zip"
		fmt.Printf("Trying to download theme from %s branch...\n", branch)

		req, err := http.NewRequest(http.MethodGet, zipURL, nil)
		if err != nil {
			continue
		}
		req.Header.Set("User-Agent", "walgo-theme-installer")

		resp, err = client.Do(req)
		if err != nil {
			// resp may be non-nil even on error; close body if present
			if resp != nil {
				resp.Body.Close()
				resp = nil
			}
			continue
		}

		if resp.StatusCode == http.StatusOK {
			// Found the correct branch
			fmt.Printf("Found theme on %s branch\n", branch)
			break
		}
		resp.Body.Close()
		resp = nil
	}

	if resp == nil || resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download theme from any branch (tried: %v)", branches)
	}
	defer resp.Body.Close()

	// Create temporary file for download
	tmpFile, err := os.CreateTemp("", "hugo-theme-*.zip")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpName := tmpFile.Name()
	defer os.Remove(tmpName)

	// Write response body directly to temp file (no close-reopen)
	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		tmpFile.Close()
		return fmt.Errorf("saving theme download: %w", err)
	}
	tmpFile.Close()

	// Extract ZIP archive
	fmt.Printf("Extracting theme to %s...\n", themePath)
	filesExtracted, err := extractThemeZip(tmpName, themePath)
	if err != nil {
		return err
	}

	// Verify theme was installed correctly by checking for theme.toml or theme.yaml
	themeMetaFiles := []string{"theme.toml", "theme.yaml", "theme.yml"}
	themeValid := false
	for _, metaFile := range themeMetaFiles {
		if _, err := os.Stat(filepath.Join(themePath, metaFile)); err == nil {
			themeValid = true
			break
		}
	}

	if !themeValid {
		fmt.Printf("Warning: Theme metadata file not found in %s\n", themePath)
		fmt.Printf("Theme may not work correctly. Extracted files: %d\n", filesExtracted)
	}

	// Verify the directory exists and is accessible
	entries, err := os.ReadDir(themePath)
	if err != nil {
		return fmt.Errorf("theme extracted but directory is not readable: %w", err)
	}

	fmt.Printf("Theme %s installed successfully (%d files, %d top-level entries)\n",
		themeName, filesExtracted, len(entries))

	return nil
}

// extractThemeZip extracts a GitHub zip archive into themePath, stripping the
// archive's root directory. Entries that escape themePath are rejected and
// symlinks are skipped. It returns the number of files written.
func extractThemeZip(zipPath, themePath string) (int, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return 0, fmt.Errorf("opening theme archive: %w", err)
	}
	defer r.Close()

	// GitHub ZIP archives have a root directory named {repo}-{branch}
	// We need to extract contents and strip this root directory
	var rootDir string
	filesExtracted := 0
	for _, f := range r.File {
		if rootDir == "" {
			// First entry should be the root directory
			parts := strings.Split(f.Name, "/")
			if len(parts) > 0 {
				rootDir = parts[0] + "/"
				fmt.Printf("Detected archive root directory: %s\n", strings.TrimSuffix(rootDir, "/"))
			}
		}

		// Never extract symlinks: they can point outside the theme directory
		if f.Mode()&os.ModeSymlink != 0 {
			continue
		}

		// Skip the root directory itself
		if f.Name == strings.TrimSuffix(rootDir, "/") || f.Name == rootDir {
			continue
		}

		// Strip root directory from path
		relativePath := strings.TrimPrefix(f.Name, rootDir)
		if relativePath == "" {
			continue
		}

		targetPath, err := safeExtractPath(themePath, relativePath)
		if err != nil {
			return 0, fmt.Errorf("unsafe path in theme archive: %w", err)
		}

		if f.FileInfo().IsDir() {
			// Create directory
			if err := os.MkdirAll(targetPath, f.Mode()); err != nil {
				return 0, fmt.Errorf("creating directory %s: %w", targetPath, err)
			}
		} else {
			// Create parent directory if needed
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return 0, fmt.Errorf("creating parent directory for %s: %w", targetPath, err)
			}

			// Extract file
			rc, err := f.Open()
			if err != nil {
				return 0, fmt.Errorf("opening file in archive: %w", err)
			}

			outFile, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				rc.Close()
				return 0, fmt.Errorf("creating file %s: %w", targetPath, err)
			}

			if _, err := io.Copy(outFile, rc); err != nil {
				outFile.Close()
				rc.Close()
				return 0, fmt.Errorf("extracting file %s: %w", targetPath, err)
			}

			outFile.Close()
			rc.Close()
			filesExtracted++
		}
	}

	if filesExtracted == 0 {
		return 0, fmt.Errorf("no files were extracted from theme archive - archive may be empty or corrupted")
	}
	return filesExtracted, nil
}

// InstallThemeFromURL installs a Hugo theme from a GitHub URL
// If themes already exist, they are backed up first
// After installation, it tries to build the site
// If build fails, it restores the old theme and removes the new one
func InstallThemeFromURL(sitePath, githubURL string) (string, error) {
	// Validate URL
	if githubURL == "" {
		return "", fmt.Errorf("github URL is required")
	}

	// Validate it's a GitHub URL; everything below derives from the parsed parts
	// Example: https://github.com/theNewDynamic/gohugo-theme-ananke -> gohugo-theme-ananke
	owner, themeName, err := parseGitHubRepo(githubURL)
	if err != nil {
		return "", err
	}
	githubURL = "https://github.com/" + owner + "/" + themeName

	themesDir := filepath.Join(sitePath, "themes")
	if _, err := themeChildPath(themesDir, themeName); err != nil {
		return "", err
	}

	// Create themes directory if it doesn't exist
	if err := os.MkdirAll(themesDir, 0755); err != nil {
		return "", fmt.Errorf("creating themes directory: %w", err)
	}

	// Backup existing themes and config
	var backupDir string

	var oldConfigContent []byte
	var configPath string

	// Find config file
	configFiles := []string{
		filepath.Join(sitePath, "hugo.toml"),
		filepath.Join(sitePath, "config.toml"),
	}
	for _, cf := range configFiles {
		if _, err := os.Stat(cf); err == nil {
			configPath = cf
			oldConfigContent, err = os.ReadFile(cf)
			if err != nil {
				return "", fmt.Errorf("failed to read config %s for backup: %w", cf, err)
			}
			break
		}
	}

	// Backup existing themes
	entries, err := os.ReadDir(themesDir)
	if err == nil && len(entries) > 0 {
		// Create backup directory
		backupDir, err = os.MkdirTemp("", "walgo-theme-backup-*")
		if err != nil {
			return "", fmt.Errorf("creating backup directory: %w", err)
		}

		fmt.Printf("Backing up existing themes...\n")
		for _, entry := range entries {
			if entry.IsDir() {
				srcPath := filepath.Join(themesDir, entry.Name())
				dstPath := filepath.Join(backupDir, entry.Name())
				fmt.Printf("  Backing up: %s\n", entry.Name())

				// Copy theme to backup
				if err := copyDir(srcPath, dstPath); err != nil {
					os.RemoveAll(backupDir)
					return "", fmt.Errorf("failed to backup theme %s: %w", entry.Name(), err)
				}

				// Remove original
				if err := os.RemoveAll(srcPath); err != nil {
					os.RemoveAll(backupDir)
					return "", fmt.Errorf("failed to remove theme %s: %w", entry.Name(), err)
				}
			}
		}
	}

	// Helper function to restore on failure — returns error if restoration itself fails
	restoreBackup := func() error {
		if backupDir == "" {
			return nil
		}
		var restoreErrors []string

		fmt.Printf("Restoring previous themes...\n")
		// Remove new theme
		newThemePath, err := themeChildPath(themesDir, themeName)
		if err != nil {
			return err
		}
		os.RemoveAll(newThemePath)

		// Restore old themes
		backupEntries, err := os.ReadDir(backupDir)
		if err != nil {
			return fmt.Errorf("failed to read backup directory: %w", err)
		}
		for _, entry := range backupEntries {
			if entry.IsDir() {
				srcPath := filepath.Join(backupDir, entry.Name())
				dstPath := filepath.Join(themesDir, entry.Name())
				if err := copyDir(srcPath, dstPath); err != nil {
					restoreErrors = append(restoreErrors, fmt.Sprintf("theme %s: %v", entry.Name(), err))
				} else {
					fmt.Printf("  Restored: %s\n", entry.Name())
				}
			}
		}

		// Restore old config
		if configPath != "" && oldConfigContent != nil {
			if err := os.WriteFile(configPath, oldConfigContent, 0644); err != nil {
				restoreErrors = append(restoreErrors, fmt.Sprintf("config %s: %v", configPath, err))
			}
		}

		// Cleanup backup
		os.RemoveAll(backupDir)

		if len(restoreErrors) > 0 {
			return fmt.Errorf("partial restore failure: %s", strings.Join(restoreErrors, "; "))
		}
		return nil
	}

	// Download and install new theme
	fmt.Printf("Installing theme: %s\n", themeName)
	themePath, err := themeChildPath(themesDir, themeName)
	if err != nil {
		return "", err
	}

	if err := downloadThemeZip(themeName, githubURL+".git", themePath); err != nil {
		if restoreErr := restoreBackup(); restoreErr != nil {
			return "", fmt.Errorf("failed to download theme: %w (additionally, restore failed: %v)", err, restoreErr)
		}
		return "", fmt.Errorf("failed to download theme (previous themes restored): %w", err)
	}

	// Update hugo.toml with the new theme name
	if err := updateHugoConfigTheme(sitePath, themeName); err != nil {
		fmt.Printf("Warning: Could not update hugo.toml with theme name: %v\n", err)
		fmt.Printf("Please manually set 'theme = \"%s\"' in your hugo.toml\n", themeName)
	}

	// Update site config with theme-specific params
	fmt.Printf("Updating site configuration for theme '%s'...\n", themeName)
	if err := ai.UpdateSiteConfigForTheme(sitePath, themeName); err != nil {
		fmt.Printf("Warning: Could not update config for theme: %v\n", err)
	}

	// Setup archetypes for the new theme
	fmt.Printf("Setting up archetypes for theme '%s'...\n", themeName)
	if err := SetupArchetypes(sitePath, themeName); err != nil {
		fmt.Printf("Warning: Could not set up archetypes for theme: %v\n", err)
	}

	// Setup favicon with theme-aware placement
	fmt.Printf("Setting up favicon for theme '%s'...\n", themeName)
	if err := SetupFaviconForTheme(sitePath, themeName); err != nil {
		fmt.Printf("Warning: Could not set up favicon for theme: %v\n", err)
	}

	// Build succeeded, cleanup backup
	if backupDir != "" {
		fmt.Printf("Build successful, cleaning up backups...\n")
		os.RemoveAll(backupDir)
	}

	fmt.Printf("Theme '%s' installed and verified successfully!\n", themeName)
	return themeName, nil
}

// copyDir recursively copies a directory
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate destination path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		// Copy file
		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		dstFile, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode())
		if err != nil {
			return err
		}
		defer dstFile.Close()

		_, err = io.Copy(dstFile, srcFile)
		return err
	})
}

// updateHugoConfigTheme updates the theme setting in hugo.toml or config.toml
func updateHugoConfigTheme(sitePath, themeName string) error {
	// Try hugo.toml first, then config.toml
	configFiles := []string{
		filepath.Join(sitePath, "hugo.toml"),
		filepath.Join(sitePath, "config.toml"),
	}

	var configPath string
	for _, cf := range configFiles {
		if _, err := os.Stat(cf); err == nil {
			configPath = cf
			break
		}
	}

	if configPath == "" {
		return fmt.Errorf("no hugo.toml or config.toml found")
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("reading config: %w", err)
	}

	configStr := string(content)

	// Check if theme line exists and replace it
	// Match "theme" as a standalone key (not "themeDir", "themes", etc.)
	lines := strings.Split(configStr, "\n")
	themeFound := false
	themeLine := fmt.Sprintf("theme = \"%s\"", themeName)
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Skip comments
		if strings.HasPrefix(trimmed, "#") {
			continue
		}
		// Match "theme" followed by optional whitespace and "="
		if (strings.HasPrefix(trimmed, "theme ") || strings.HasPrefix(trimmed, "theme=")) &&
			strings.Contains(trimmed, "=") {
			lines[i] = themeLine
			themeFound = true
			break
		}
	}

	// If theme line not found, add it after baseURL or at the beginning
	if !themeFound {
		newLines := make([]string, 0, len(lines)+1)
		added := false
		for _, line := range lines {
			newLines = append(newLines, line)
			trimmed := strings.TrimSpace(line)
			if !added && strings.HasPrefix(trimmed, "baseURL") {
				newLines = append(newLines, themeLine)
				added = true
			}
		}
		if !added {
			// Add after any initial comments, or at end if file is all comments
			insertAt := len(lines)
			for i, line := range lines {
				trimmed := strings.TrimSpace(line)
				if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
					insertAt = i
					break
				}
			}
			// Build new slice without mutating original
			newLines = make([]string, 0, len(lines)+1)
			newLines = append(newLines, lines[:insertAt]...)
			newLines = append(newLines, themeLine)
			newLines = append(newLines, lines[insertAt:]...)
		}
		lines = newLines
	}

	newContent := strings.Join(lines, "\n")
	if err := os.WriteFile(configPath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	fmt.Printf("Updated %s with theme = \"%s\"\n", filepath.Base(configPath), themeName)
	return nil
}

// GetInstalledThemes returns list of installed themes in the themes directory
func GetInstalledThemes(sitePath string) ([]string, error) {
	themesDir := filepath.Join(sitePath, "themes")

	if _, err := os.Stat(themesDir); os.IsNotExist(err) {
		return nil, nil
	}

	entries, err := os.ReadDir(themesDir)
	if err != nil {
		return nil, fmt.Errorf("reading themes directory: %w", err)
	}

	var themes []string
	for _, entry := range entries {
		if entry.IsDir() {
			themes = append(themes, entry.Name())
		}
	}

	return themes, nil
}

// SetupFaviconForTheme copies the embedded favicon to the location determined by theme analysis.
// The themeName parameter should be the theme directory name (e.g., "ananke", "hugo-book").
func SetupFaviconForTheme(sitePath, themeName string) error {
	// Analyze theme config to find favicon path
	configAnalysis := ai.AnalyzeThemeConfig(sitePath, themeName)

	// Determine target directory from theme config
	var targetDir string
	var faviconName string = "favicon.svg"

	// Check if theme specifies a favicon path in recommended params
	if faviconParam, ok := configAnalysis.RecommendedParams["favicon"]; ok {
		if faviconPath, ok := faviconParam.(string); ok && faviconPath != "" {
			// Extract directory and filename from favicon path
			// e.g., "/images/favicon.png" -> "static/images", "favicon.svg"
			faviconPath = strings.TrimPrefix(faviconPath, "/")
			dir := filepath.Dir(faviconPath)
			if dir != "." {
				targetDir = filepath.Join(sitePath, "static", dir)
			} else {
				targetDir = filepath.Join(sitePath, "static")
			}
		}
	}

	// Check for BookFavicon (hugo-book theme)
	if bookFavicon, ok := configAnalysis.RecommendedParams["BookFavicon"]; ok {
		if faviconPath, ok := bookFavicon.(string); ok && faviconPath != "" {
			targetDir = filepath.Join(sitePath, "static")
			faviconName = faviconPath
		}
	}

	// Default location if no specific path found
	if targetDir == "" {
		targetDir = filepath.Join(sitePath, "static")
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("creating favicon directory: %w", err)
	}

	// Write favicon (always use walgo's SVG favicon)
	faviconPath := filepath.Join(targetDir, faviconName)
	if err := os.WriteFile(faviconPath, faviconSVG, 0644); err != nil {
		return fmt.Errorf("writing favicon: %w", err)
	}

	return nil
}

// SetupDocsThemeOverrides creates necessary layout overrides for hugo-book theme
// This improves SVG favicon support with proper type specification
func SetupDocsThemeOverrides(sitePath string) error {
	layoutsDir := filepath.Join(sitePath, "layouts", "_partials", "docs")

	// Create layouts directory if it doesn't exist
	if err := os.MkdirAll(layoutsDir, 0755); err != nil {
		return fmt.Errorf("creating layouts directory: %w", err)
	}

	// Check if html-head-favicon.html already exists (don't overwrite user customizations)
	faviconPath := filepath.Join(layoutsDir, "html-head-favicon.html")
	if _, err := os.Stat(faviconPath); err == nil {
		return nil // Already exists, skip
	}

	// Create html-head-favicon.html override with proper SVG type
	faviconContent := `{{- $favicon := .Site.Params.BookFavicon | default "favicon.svg" -}}
<link rel="icon" href="{{ $favicon | relURL }}" type="image/svg+xml">
<link rel="apple-touch-icon" href="{{ $favicon | relURL }}">
`

	if err := os.WriteFile(faviconPath, []byte(faviconContent), 0644); err != nil {
		return fmt.Errorf("writing html-head-favicon.html: %w", err)
	}

	return nil
}

// UpdateDocsParams updates hugo-book theme params based on AI plan
func UpdateDocsParams(sitePath, description string) error {
	configPath := filepath.Join(sitePath, "hugo.toml")
	content, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("reading config: %w", err)
	}

	configStr := string(content)

	// Update description in homepage _index.md if it exists
	indexPath := filepath.Join(sitePath, "content", "_index.md")
	if _, err := os.Stat(indexPath); err == nil {
		indexContent, err := os.ReadFile(indexPath)
		if err == nil {
			indexStr := string(indexContent)
			// Update description in frontmatter if empty or generic
			if strings.Contains(indexStr, "description: ''") || strings.Contains(indexStr, `description: ""`) {
				indexStr = strings.Replace(indexStr, "description: ''", fmt.Sprintf("description: '%s'", description), 1)
				indexStr = strings.Replace(indexStr, `description: ""`, fmt.Sprintf(`description: "%s"`, description), 1)
				if err := os.WriteFile(indexPath, []byte(indexStr), 0644); err != nil {
					return fmt.Errorf("writing index: %w", err)
				}
			}
		}
	}

	// Write back config if changed
	if configStr != string(content) {
		if err := os.WriteFile(configPath, []byte(configStr), 0644); err != nil {
			return fmt.Errorf("writing config: %w", err)
		}
	}

	return nil
}
