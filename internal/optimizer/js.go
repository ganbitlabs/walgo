package optimizer

import (
	"regexp"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/js"
)

const jsMime = "application/javascript"

var jsMimeRegexp = regexp.MustCompile(`^(application|text)/(x-)?(java|ecma)script$|^module$`)

// JSOptimizer handles JavaScript optimization
type JSOptimizer struct {
	config JSConfig
	m      *minify.M
}

// NewJSOptimizer creates a new JavaScript optimizer
func NewJSOptimizer(config JSConfig) *JSOptimizer {
	m := minify.New()
	m.Add(jsMime, &js.Minifier{KeepVarNames: !config.Obfuscate})
	return &JSOptimizer{config: config, m: m}
}

// Optimize optimizes JavaScript content
func (j *JSOptimizer) Optimize(content []byte) ([]byte, error) {
	cfg := j.config
	if !cfg.Enabled || !(cfg.MinifyJS || cfg.RemoveComments) {
		return content, nil
	}
	return j.m.Bytes(jsMime, content)
}
