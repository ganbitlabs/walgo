package optimizer

import (
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
)

const cssMime = "text/css"

// CSSOptimizer handles CSS optimization
type CSSOptimizer struct {
	config CSSConfig
	m      *minify.M
}

// NewCSSOptimizer creates a new CSS optimizer
func NewCSSOptimizer(config CSSConfig) *CSSOptimizer {
	m := minify.New()
	m.Add(cssMime, &css.Minifier{})
	return &CSSOptimizer{config: config, m: m}
}

// Optimize optimizes CSS content
func (c *CSSOptimizer) Optimize(content []byte) ([]byte, error) {
	cfg := c.config
	if !cfg.Enabled || !(cfg.MinifyCSS || cfg.RemoveComments || cfg.CompressColors) {
		return content, nil
	}
	return c.m.Bytes(cssMime, content)
}
