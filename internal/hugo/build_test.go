package hugo

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ganbitlabs/walgo/internal/compress"
	"github.com/ganbitlabs/walgo/internal/deps"
)

func writeFiles(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for name, content := range files {
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func requireHugo(t *testing.T) {
	t.Helper()
	if _, err := deps.LookPath("hugo"); err != nil {
		t.Skip("hugo not installed")
	}
}

// isolateHome keeps BuildSite's project database lookup away from the real ~/.walgo.
func isolateHome(t *testing.T) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
}

func assertFile(t *testing.T, path, contains string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("expected %s: %v", path, err)
	}
	if !strings.Contains(string(data), contains) {
		t.Errorf("%s does not contain %q:\n%s", path, contains, data)
	}
}

func assertMissing(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("expected %s not to exist (err=%v)", path, err)
	}
}

const sampleLayout = `<!DOCTYPE html><html><head><title>{{ .Site.Title }}</title></head>
<body><p><b>bold</b> <i>it</i></p></body></html>
`

func TestBuildSiteHonorsPublishDir(t *testing.T) {
	requireHugo(t)

	tests := []struct {
		name  string
		files map[string]string
		out   string
	}{
		{
			name: "walgo.yaml publishDir wins over the Hugo config",
			files: map[string]string{
				"walgo.yaml": "hugo:\n  publishDir: dist\ncompress:\n  generateWSResources: true\n",
				"hugo.toml":  "title = \"Sample\"\npublishDir = \"ignored\"\n",
			},
			out: "dist",
		},
		{
			name: "Hugo config publishDir when walgo.yaml leaves it unset",
			files: map[string]string{
				"walgo.yaml": "compress:\n  generateWSResources: true\n",
				"hugo.yaml":  "title: Sample\npublishDir: out\n",
			},
			out: "out",
		},
		{
			name: "config/_default with JSON config",
			files: map[string]string{
				"walgo.yaml":                "compress:\n  generateWSResources: true\n",
				"config/_default/hugo.json": `{"title": "Sample", "publishdir": "site"}`,
			},
			out: "site",
		},
		{
			name: "default public with config.yml beside archetypes",
			files: map[string]string{
				"walgo.yaml":            "compress:\n  generateWSResources: true\n",
				"config.yml":            "title: Sample\n",
				"archetypes/default.md": "",
			},
			out: "public",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isolateHome(t)
			site := t.TempDir()
			tt.files["layouts/index.html"] = sampleLayout
			writeFiles(t, site, tt.files)

			if err := BuildSite(site); err != nil {
				t.Fatalf("BuildSite: %v", err)
			}

			assertFile(t, filepath.Join(site, tt.out, "index.html"), "<title>Sample</title>")
			assertFile(t, filepath.Join(site, tt.out, "ws-resources.json"), `"/index.html"`)
			if tt.out != "public" {
				assertMissing(t, filepath.Join(site, "public"))
			}
		})
	}
}

func TestBuildSiteKeepsUserCompressedAndSourceMapFiles(t *testing.T) {
	requireHugo(t)
	isolateHome(t)
	site := t.TempDir()
	writeFiles(t, site, map[string]string{
		"walgo.yaml":           "compress:\n  generateWSResources: true\n",
		"hugo.toml":            "title = \"Sample\"\n",
		"layouts/index.html":   sampleLayout,
		"static/data.json.gz":  "gz",
		"static/fonts/a.br":    "br",
		"static/js/app.js.map": "{}",
		"static/.DS_Store":     "junk",
	})

	if err := BuildSite(site); err != nil {
		t.Fatalf("BuildSite: %v", err)
	}

	public := filepath.Join(site, "public")
	for _, kept := range []string{"data.json.gz", "fonts/a.br", "js/app.js.map"} {
		assertFile(t, filepath.Join(public, kept), "")
	}
	assertMissing(t, filepath.Join(public, ".DS_Store"))

	ws, err := compress.ReadWSResourcesConfig(filepath.Join(public, "ws-resources.json"))
	if err != nil {
		t.Fatal(err)
	}
	for path := range ws.Headers {
		if strings.HasSuffix(path, ".DS_Store") {
			t.Errorf("ws-resources.json lists %s, which cleanup deleted", path)
		}
	}
}

func TestBuildSiteKeepsInlineWhitespaceWithDefaultOptimizer(t *testing.T) {
	requireHugo(t)
	isolateHome(t)
	site := t.TempDir()
	writeFiles(t, site, map[string]string{
		"walgo.yaml":         "optimizer:\n  enabled: true\n  html:\n    enabled: true\n    minifyHTML: true\n    removeWhitespace: true\n  js:\n    enabled: true\n    minifyJS: true\n",
		"hugo.toml":          "title = \"Sample\"\n",
		"layouts/index.html": sampleLayout,
		"static/app.js":      "const a = \"/*\"; const keep = 1; const b = \"*/\";\n",
	})

	if err := BuildSite(site); err != nil {
		t.Fatalf("BuildSite: %v", err)
	}

	assertFile(t, filepath.Join(site, "public", "index.html"), "<b>bold</b> <i>it</i>")
	assertFile(t, filepath.Join(site, "public", "app.js"), `b="*/"`)
}

func TestBuildSiteStaticSiteSkipsHugo(t *testing.T) {
	isolateHome(t)
	site := t.TempDir()
	page := "<html>\n  <body>\n    <p><b>bold</b> <i>it</i></p>\n  </body>\n</html>\n"
	writeFiles(t, site, map[string]string{
		"walgo.yaml":          "hugo:\n  publishDir: dist\noptimizer:\n  enabled: true\ncompress:\n  generateWSResources: true\n",
		"dist/index.html":     page,
		"dist/bundle.js.map":  "{}",
		"dist/assets/a.js.gz": "gz",
	})

	if err := BuildSite(site); err != nil {
		t.Fatalf("BuildSite: %v", err)
	}

	dist := filepath.Join(site, "dist")
	got, err := os.ReadFile(filepath.Join(dist, "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != page {
		t.Errorf("static site file was rewritten:\n%s", got)
	}
	assertFile(t, filepath.Join(dist, "bundle.js.map"), "{}")
	assertFile(t, filepath.Join(dist, "assets/a.js.gz"), "gz")
	assertFile(t, filepath.Join(dist, "ws-resources.json"), `"/index.html"`)
}

func TestBuildSiteStaticSiteWithGenericConfigFile(t *testing.T) {
	isolateHome(t)
	site := t.TempDir()
	writeFiles(t, site, map[string]string{
		"walgo.yaml":        "hugo:\n  publishDir: dist\n",
		"config.json":       `{"theme": "dark", "baseUrl": "https://api.example.com", "module": "app"}`,
		"config.yaml":       "theme: dark\nport: 8080\n",
		"layouts/main.html": "<main></main>",
		"themes/dark.css":   "body{}",
		"dist/index.html":   "<p>app</p>",
	})

	if err := BuildSite(site); err != nil {
		t.Fatalf("BuildSite: %v", err)
	}
	assertFile(t, filepath.Join(site, "dist", "index.html"), "<p>app</p>")
	assertMissing(t, filepath.Join(site, "dist", "sitemap.xml"))
}

func TestBuildSiteMissingPublishDirNamesIgnoredConfig(t *testing.T) {
	isolateHome(t)
	site := t.TempDir()
	writeFiles(t, site, map[string]string{
		"walgo.yaml":  "compress:\n  generateWSResources: true\n",
		"config.yaml": "title: My Hugo site\n",
	})

	err := BuildSite(site)
	if err == nil || !strings.Contains(err.Error(), "hugo.yaml") {
		t.Fatalf("expected the error to suggest renaming config.yaml to hugo.yaml, got %v", err)
	}
}

func TestBuildSiteStaticSiteWithoutPublishDir(t *testing.T) {
	isolateHome(t)
	site := t.TempDir()
	writeFiles(t, site, map[string]string{"walgo.yaml": "hugo:\n  publishDir: dist\n"})

	err := BuildSite(site)
	if err == nil || !strings.Contains(err.Error(), "no Hugo config") {
		t.Fatalf("expected a no-Hugo-config error, got %v", err)
	}
}

func TestBuildSiteRefusesPublishDirContainingSite(t *testing.T) {
	for name, files := range map[string]map[string]string{
		"hugo site":   {"hugo.toml": "title = \"Sample\"\n"},
		"static site": {},
	} {
		t.Run(name, func(t *testing.T) {
			isolateHome(t)
			site := t.TempDir()
			files["walgo.yaml"] = "hugo:\n  publishDir: .\ncompress:\n  generateWSResources: true\n"
			files["content/_index.md"] = "keep me"
			writeFiles(t, site, files)

			err := BuildSite(site)
			if err == nil || !strings.Contains(err.Error(), "contains the site") {
				t.Fatalf("expected refusal, got %v", err)
			}
			assertFile(t, filepath.Join(site, "content/_index.md"), "keep me")
			assertMissing(t, filepath.Join(site, "ws-resources.json"))
		})
	}
}

func TestBuildSiteRefusesPublishDirInsideHugoSources(t *testing.T) {
	for _, dir := range []string{"content", "static/out", "layouts"} {
		t.Run(dir, func(t *testing.T) {
			isolateHome(t)
			site := t.TempDir()
			writeFiles(t, site, map[string]string{
				"walgo.yaml":         "hugo:\n  publishDir: " + dir + "\n",
				"hugo.toml":          "title = \"Sample\"\n",
				"content/_index.md":  "keep me",
				"static/logo.svg":    "<svg/>",
				"layouts/index.html": sampleLayout,
			})

			err := BuildSite(site)
			if err == nil || !strings.Contains(err.Error(), "Hugo source directory") {
				t.Fatalf("expected refusal, got %v", err)
			}
			assertFile(t, filepath.Join(site, "content/_index.md"), "keep me")
			assertFile(t, filepath.Join(site, "static/logo.svg"), "<svg/>")
			assertFile(t, filepath.Join(site, "layouts/index.html"), "<title>")
		})
	}
}

func TestBuildSiteMalformedHugoConfigFails(t *testing.T) {
	requireHugo(t)
	isolateHome(t)
	site := t.TempDir()
	writeFiles(t, site, map[string]string{
		"walgo.yaml":        "compress:\n  generateWSResources: true\n",
		"hugo.toml":         "title = ",
		"public/index.html": "stale",
	})

	if err := BuildSite(site); err == nil {
		t.Fatal("expected the hugo build to fail on a malformed hugo.toml")
	}
}
