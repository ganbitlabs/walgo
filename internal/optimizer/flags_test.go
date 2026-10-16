package optimizer

import "testing"

func TestMinifyFlagOffLeavesContentUnchanged(t *testing.T) {
	cases := []struct {
		name string
		opt  interface{ Optimize([]byte) ([]byte, error) }
		in   string
	}{
		{"js", NewJSOptimizer(JSConfig{Enabled: true, RemoveComments: true}), "const a = 1; // c\n"},
		{"css", NewCSSOptimizer(CSSConfig{Enabled: true, RemoveComments: true, CompressColors: true}), "a { color: #ffffff; } /* c */\n"},
		{"html", NewHTMLOptimizer(HTMLConfig{Enabled: true, RemoveComments: true, CompressInlineCSS: true, CompressInlineJS: true}), "<p class=\"x\">  a  </p>\n"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.opt.Optimize([]byte(tc.in))
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != tc.in {
				t.Errorf("minify flag is off but output changed: %q", got)
			}
		})
	}
}
