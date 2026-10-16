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
