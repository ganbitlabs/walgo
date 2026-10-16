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

// Optimize minifies CSS when minifyCSS is on. Minification strips comments and shortens colors,
// so removeComments and compressColors have no effect on their own.
func (c *CSSOptimizer) Optimize(content []byte) ([]byte, error) {
	if !c.config.Enabled || !c.config.MinifyCSS {
		return content, nil
	}
	return c.m.Bytes(cssMime, content)
}
