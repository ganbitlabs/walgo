package optimizer

import (
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/js"
)

const htmlMime = "text/html"

// HTMLOptimizer handles HTML optimization
type HTMLOptimizer struct {
	config HTMLConfig
	m      *minify.M
}

// NewHTMLOptimizer creates a new HTML optimizer
func NewHTMLOptimizer(config HTMLConfig) *HTMLOptimizer {
	m := minify.New()
	// The Keep* settings mirror Hugo's minify defaults, so output matches what `hugo --minify` produces.
	m.Add(htmlMime, &html.Minifier{
		KeepComments:        !config.RemoveComments,
		KeepSpecialComments: true,
		KeepDefaultAttrVals: true,
		KeepDocumentTags:    true,
		KeepEndTags:         true,
		KeepWhitespace:      !config.MinifyHTML || !config.RemoveWhitespace,
	})
	if config.CompressInlineCSS {
		m.Add(cssMime, &css.Minifier{})
	}
	if config.CompressInlineJS {
		m.AddRegexp(jsMimeRegexp, &js.Minifier{KeepVarNames: true})
	}
	return &HTMLOptimizer{config: config, m: m}
}

// Optimize optimizes HTML content
func (h *HTMLOptimizer) Optimize(content []byte) ([]byte, error) {
	c := h.config
	if !c.Enabled || (!c.MinifyHTML && !c.RemoveComments && !c.CompressInlineCSS && !c.CompressInlineJS) {
		return content, nil
	}
	return h.m.Bytes(htmlMime, content)
}
