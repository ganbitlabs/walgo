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
		KeepWhitespace:      !config.RemoveWhitespace,
	})
	if config.CompressInlineCSS {
		m.Add(cssMime, &css.Minifier{})
	}
	if config.CompressInlineJS {
		m.AddRegexp(jsMimeRegexp, &js.Minifier{KeepVarNames: true})
	}
	return &HTMLOptimizer{config: config, m: m}
}

// Optimize minifies HTML when minifyHTML is on. The other HTML flags refine that pass
// and have no effect on their own.
func (h *HTMLOptimizer) Optimize(content []byte) ([]byte, error) {
	if !h.config.Enabled || !h.config.MinifyHTML {
		return content, nil
	}
	return h.m.Bytes(htmlMime, content)
}
