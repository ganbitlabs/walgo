package config

import (
	"testing"

	"github.com/ganbitlabs/walgo/internal/compress"
)

func TestCacheControl(t *testing.T) {
	def := compress.DefaultCacheControlConfig()

	got := CacheConfig{Enabled: true}.CacheControl()
	if got.ImmutableMaxAge != def.ImmutableMaxAge || got.MutableMaxAge != def.MutableMaxAge {
		t.Errorf("zero max-ages = %d/%d, want defaults %d/%d", got.ImmutableMaxAge, got.MutableMaxAge, def.ImmutableMaxAge, def.MutableMaxAge)
	}
	if len(got.ImmutablePatterns) != len(def.ImmutablePatterns) {
		t.Errorf("immutable patterns not inherited from defaults")
	}

	got = CacheConfig{Enabled: false, ImmutableMaxAge: 60, MutableMaxAge: 5}.CacheControl()
	if got.Enabled || got.ImmutableMaxAge != 60 || got.MutableMaxAge != 5 {
		t.Errorf("overrides = %+v", got)
	}

	cfg := NewDefaultWalgoConfig()
	if cfg.CacheConfig.ImmutableMaxAge != def.ImmutableMaxAge || cfg.CacheConfig.MutableMaxAge != def.MutableMaxAge || cfg.CacheConfig.Enabled != def.Enabled {
		t.Errorf("walgo.yaml defaults %+v drift from compress defaults", cfg.CacheConfig)
	}
}

func TestOmittedSectionsKeepDefaults(t *testing.T) {
	def := NewDefaultWalgoConfig()
	tests := []struct {
		name          string
		yaml          string
		wantOptimizer bool
		wantMinifyJS  bool
		wantCache     bool
	}{
		{"sections omitted", "walrus:\n  network: testnet\n", true, true, true},
		{"explicitly disabled", "optimizer:\n  enabled: false\ncache:\n  enabled: false\n", false, true, false},
		{"partial optimizer block", "optimizer:\n  js:\n    minifyJS: false\n", true, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := writeSite(t, map[string]string{"walgo.yaml": tt.yaml})
			cfg, err := LoadConfigFrom(dir)
			if err != nil {
				t.Fatal(err)
			}
			if cfg.OptimizerConfig.Enabled != tt.wantOptimizer || cfg.OptimizerConfig.JS.MinifyJS != tt.wantMinifyJS {
				t.Errorf("optimizer enabled=%v minifyJS=%v, want %v %v", cfg.OptimizerConfig.Enabled, cfg.OptimizerConfig.JS.MinifyJS, tt.wantOptimizer, tt.wantMinifyJS)
			}
			if cfg.OptimizerConfig.HTML.MinifyHTML != def.OptimizerConfig.HTML.MinifyHTML {
				t.Errorf("unset html.minifyHTML = %v, want default %v", cfg.OptimizerConfig.HTML.MinifyHTML, def.OptimizerConfig.HTML.MinifyHTML)
			}
			if got := cfg.CacheConfig.CacheControl().Enabled; got != tt.wantCache {
				t.Errorf("cache enabled = %v, want %v", got, tt.wantCache)
			}
		})
	}
}
