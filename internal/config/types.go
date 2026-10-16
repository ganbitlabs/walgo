package config

import (
	"github.com/ganbitlabs/walgo/internal/compress"
	"github.com/ganbitlabs/walgo/internal/optimizer"
)

// WalgoConfig is the top-level configuration for Walgo.
// It will be stored in walgo.yaml in the site root.
type WalgoConfig struct {
	SiteConfig      SiteConfig                `mapstructure:"site" yaml:"site,omitempty"`
	HugoConfig      HugoConfig                `mapstructure:"hugo" yaml:"hugo"`
	WalrusConfig    WalrusConfig              `mapstructure:"walrus" yaml:"walrus"`
	ObsidianConfig  ObsidianConfig            `mapstructure:"obsidian" yaml:"obsidian,omitempty"`
	OptimizerConfig optimizer.OptimizerConfig `mapstructure:"optimizer" yaml:"optimizer,omitempty"`
	CompressConfig  CompressConfig            `mapstructure:"compress" yaml:"compress,omitempty"`
	CacheConfig     CacheConfig               `mapstructure:"cache" yaml:"cache,omitempty"`
	// Note: AI credentials are stored in ~/.walgo/ai-credentials.yaml, not in walgo.yaml
}

// SiteKind says how walgo builds a site.
type SiteKind string

const (
	// SiteKindHugo builds the site with hugo before deploying its publish directory.
	SiteKindHugo SiteKind = "hugo"
	// SiteKindStatic deploys the publish directory as-is.
	SiteKindStatic SiteKind = "static"
)

// SiteConfig holds settings about the site as a whole.
type SiteConfig struct {
	// Kind overrides detection, which treats a site with a Hugo config file as Hugo and anything else as static.
	Kind SiteKind `mapstructure:"kind" yaml:"kind,omitempty"`
}

// HugoConfig holds Hugo-specific settings relevant to Walgo.
// Note: Most Hugo configurations are in hugo.toml (or config.toml/yaml/json).
// This section is for settings Walgo might need to override or know about.
type HugoConfig struct {
	Version     string `mapstructure:"version" yaml:"version,omitempty"`         // Hugo version to target/ensure
	BaseURL     string `mapstructure:"baseURL" yaml:"baseURL,omitempty"`         // Overrides Hugo's baseURL for specific deployments
	PublishDir  string `mapstructure:"publishDir" yaml:"publishDir,omitempty"`   // Default: "public"
	ContentDir  string `mapstructure:"contentDir" yaml:"contentDir,omitempty"`   // Default: "content"
	ResourceDir string `mapstructure:"resourceDir" yaml:"resourceDir,omitempty"` // Default: "resources"
}

// WalrusConfig holds settings for deploying to Walrus Sites.
type WalrusConfig struct {
	ProjectID   string `mapstructure:"projectID" yaml:"projectID"`               // Walrus Project ID or name
	BucketName  string `mapstructure:"bucketName" yaml:"bucketName,omitempty"`   // Optional: specific bucket if not default
	Entrypoint  string `mapstructure:"entrypoint" yaml:"entrypoint,omitempty"`   // Default: "index.html"
	SuiNSDomain string `mapstructure:"suinsDomain" yaml:"suinsDomain,omitempty"` // SuiNS domain to associate

	// Network selection (testnet or mainnet)
	// Gas budget is managed in ~/.config/walrus/sites-config.yaml
	Network string `mapstructure:"network" yaml:"network,omitempty"` // Default: testnet
}

// ObsidianConfig holds settings for importing from Obsidian vaults.
type ObsidianConfig struct {
	VaultPath         string `mapstructure:"vaultPath" yaml:"vaultPath,omitempty"`         // Default Obsidian vault path
	AttachmentDir     string `mapstructure:"attachmentDir" yaml:"attachmentDir,omitempty"` // Where to put attachments (relative to static/)
	ConvertWikilinks  bool   `mapstructure:"convertWikilinks" yaml:"convertWikilinks"`     // Convert [[wikilinks]] to [markdown](links)
	IncludeDrafts     bool   `mapstructure:"includeDrafts" yaml:"includeDrafts"`           // Include files marked as drafts
	FrontmatterFormat string `mapstructure:"frontmatterFormat" yaml:"frontmatterFormat"`   // yaml, toml, json
	LinkStyle         string `mapstructure:"linkStyle" yaml:"linkStyle,omitempty"`         // "markdown" (default) or "relref" - markdown avoids REF_NOT_FOUND errors
}

// CompressConfig holds settings for Brotli compression
type CompressConfig struct {
	Enabled             bool              `mapstructure:"enabled" yaml:"enabled"`                         // Enable compression
	Level               int               `mapstructure:"level" yaml:"level,omitempty"`                   // Brotli level 0-11, default: 6
	GenerateWSResources bool              `mapstructure:"generateWSResources" yaml:"generateWSResources"` // Generate ws-resources.json
	CustomRoutes        map[string]string `mapstructure:"customRoutes" yaml:"customRoutes,omitempty"`     // Custom route mappings for ws-resources.json
	IgnorePatterns      []string          `mapstructure:"ignorePatterns" yaml:"ignorePatterns,omitempty"` // Additional files/folders to ignore during upload
}

// CacheConfig holds settings for caching and cache-control headers
type CacheConfig struct {
	Enabled         bool `mapstructure:"enabled" yaml:"enabled"`                           // Enable cache-control headers
	ImmutableMaxAge int  `mapstructure:"immutableMaxAge" yaml:"immutableMaxAge,omitempty"` // Max-age for immutable assets (default: 31536000)
	MutableMaxAge   int  `mapstructure:"mutableMaxAge" yaml:"mutableMaxAge,omitempty"`     // Max-age for HTML (default: 300)
}

// CacheControl returns the ws-resources.json cache settings: compress's defaults overlaid with
// these values. A zero max-age keeps the default.
func (c CacheConfig) CacheControl() compress.CacheControlConfig {
	cc := compress.DefaultCacheControlConfig()
	cc.Enabled = c.Enabled
	if c.ImmutableMaxAge != 0 {
		cc.ImmutableMaxAge = c.ImmutableMaxAge
	}
	if c.MutableMaxAge != 0 {
		cc.MutableMaxAge = c.MutableMaxAge
	}
	return cc
}

// NewDefaultWalgoConfig creates a WalgoConfig with sensible defaults.
func NewDefaultWalgoConfig() WalgoConfig {
	defaultCache := compress.DefaultCacheControlConfig()
	return WalgoConfig{
		HugoConfig: HugoConfig{
			PublishDir: "public",
			ContentDir: "content",
		},
		WalrusConfig: WalrusConfig{
			ProjectID:  "YOUR_WALRUS_PROJECT_ID", // User needs to fill this
			Entrypoint: "index.html",
		},
		ObsidianConfig: ObsidianConfig{
			AttachmentDir:     "images",
			ConvertWikilinks:  true,
			IncludeDrafts:     false,
			FrontmatterFormat: "yaml",
		},
		OptimizerConfig: optimizer.NewDefaultOptimizerConfig(),
		CompressConfig: CompressConfig{
			Enabled:             false, // Disabled by default during build, enabled during deploy
			Level:               6,     // Balanced compression
			GenerateWSResources: true,
		},
		CacheConfig: CacheConfig{
			Enabled:         defaultCache.Enabled,
			ImmutableMaxAge: defaultCache.ImmutableMaxAge,
			MutableMaxAge:   defaultCache.MutableMaxAge,
		},
	}
}
