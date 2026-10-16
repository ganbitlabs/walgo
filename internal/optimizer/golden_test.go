package optimizer

import (
	"flag"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "rewrite testdata/golden/*.golden from current output")

const goldenDir = "testdata/golden"

func goldenInputs(t *testing.T) []string {
	t.Helper()
	entries, err := os.ReadDir(goldenDir)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	present := map[string]bool{}
	for _, e := range entries {
		present[e.Name()] = true
		if !strings.HasSuffix(e.Name(), ".golden") {
			names = append(names, e.Name())
		}
	}
	for name := range present {
		if input, ok := strings.CutSuffix(name, ".golden"); ok && !present[input] {
			t.Errorf("%s has no input file %s (check .gitignore)", name, input)
		}
	}
	return names
}

// optimizeWithDefaults runs the default engine over a copy of the input and returns the rewritten file.
func optimizeWithDefaults(t *testing.T, name string) (input, output []byte) {
	t.Helper()
	input, err := os.ReadFile(filepath.Join(goldenDir, name))
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, input, 0o644); err != nil {
		t.Fatal(err)
	}
	stats, err := NewEngine(NewDefaultOptimizerConfig()).OptimizeDirectory(dir)
	if err != nil {
		t.Fatal(err)
	}
	if stats.FilesError != 0 {
		t.Fatalf("%s: optimizer reported %d errors", name, stats.FilesError)
	}
	output, err = os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return input, output
}

func TestGolden(t *testing.T) {
	for _, name := range goldenInputs(t) {
		t.Run(name, func(t *testing.T) {
			_, got := optimizeWithDefaults(t, name)
			goldenPath := filepath.Join(goldenDir, name+".golden")
			if *update {
				if err := os.WriteFile(goldenPath, got, 0o644); err != nil {
					t.Fatal(err)
				}
				return
			}
			want, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("missing golden file, run go test ./internal/optimizer -run TestGolden -update: %v", err)
			}
			if string(got) != string(want) {
				t.Errorf("output differs from %s\ngot:\n%s\nwant:\n%s", goldenPath, got, want)
			}
		})
	}
}

func TestMinifiedJSBehavesLikeOriginal(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node not installed")
	}
	run := func(t *testing.T, src []byte) string {
		t.Helper()
		path := filepath.Join(t.TempDir(), "run.js")
		prog := string(src) + "\n;console.log(JSON.stringify(globalThis.result));\n"
		if err := os.WriteFile(path, []byte(prog), 0o644); err != nil {
			t.Fatal(err)
		}
		out, err := exec.Command(node, path).CombinedOutput()
		if err != nil {
			t.Fatalf("node failed: %v\n%s\n--- program ---\n%s", err, out, prog)
		}
		return string(out)
	}
	for _, name := range goldenInputs(t) {
		if filepath.Ext(name) != ".js" {
			continue
		}
		t.Run(name, func(t *testing.T) {
			input, output := optimizeWithDefaults(t, name)
			if want, got := run(t, input), run(t, output); got != want {
				t.Errorf("minified result %s, original %s", got, want)
			}
		})
	}
}

func TestHTMLKeepsWhitespaceBetweenInlineElements(t *testing.T) {
	_, got := optimizeWithDefaults(t, "inline-elements.html")
	if !strings.Contains(string(got), "<b>bold</b> <i>it</i>") {
		t.Errorf("space between inline elements was dropped: %q", got)
	}
}
