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

// Optimize minifies JavaScript when minifyJS is on. Minification strips comments,
// so removeComments has no effect on its own.
func (j *JSOptimizer) Optimize(content []byte) ([]byte, error) {
	if !j.config.Enabled || !j.config.MinifyJS {
		return content, nil
	}
	return j.m.Bytes(jsMime, content)
}
