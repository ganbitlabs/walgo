package optimizer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEngine_OptimizeWithAllOptimizersEnabled(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	htmlContent := `<!DOCTYPE html>
<html>
<head>
	<style>
		body {
			margin: 0;
			padding: 0;
		}
	</style>
	<script>
		function init() {
			console.log("Initialized");
		}
	</script>
</head>
<body onload="init()">
	<h1 style="  color:  blue;  ">Hello World</h1>
</body>
</html>`

	cssContent := `/* Main styles */
body {
	font-family: Arial, sans-serif;
	background-color: #ffffff;
}

/* Unused class */
.unused {
	display: none;
}`

	jsContent := `// Initialize app
function initApp() {
	// Log message
	console.log("App started");

	// Setup event listeners
	document.addEventListener("click", function(e) {
		console.log("Clicked");
	});
}

// Call init
initApp();`

	// Write test files
	htmlFile := filepath.Join(tmpDir, "index.html")
	cssFile := filepath.Join(tmpDir, "style.css")
	jsFile := filepath.Join(tmpDir, "app.js")

	if err := os.WriteFile(htmlFile, []byte(htmlContent), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cssFile, []byte(cssContent), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(jsFile, []byte(jsContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create engine with all optimizations enabled
	config := OptimizerConfig{
		Enabled: true,
		HTML: HTMLConfig{
			Enabled:           true,
			MinifyHTML:        true,
			RemoveComments:    true,
			RemoveWhitespace:  true,
			CompressInlineCSS: true,
			CompressInlineJS:  true,
		},
		CSS: CSSConfig{
			Enabled:        true,
			MinifyCSS:      true,
			RemoveComments: true,
			RemoveUnused:   true,
			CompressColors: true,
		},
		JS: JSConfig{
			Enabled:        true,
			MinifyJS:       true,
			RemoveComments: true,
		},
		Verbose: true,
	}

	engine := NewEngine(config)
	stats, err := engine.OptimizeDirectory(tmpDir)

	if err != nil {
		t.Fatalf("Optimize() error = %v", err)
	}

	// Verify stats
	if stats.FilesProcessed == 0 {
		t.Error("No files were processed")
	}
	if stats.FilesOptimized == 0 {
		t.Error("No files were optimized")
	}
	if stats.OriginalSize == 0 {
		t.Error("Original size is 0")
	}
	if stats.OptimizedSize >= stats.OriginalSize {
		t.Error("Optimized size should be less than original")
	}

	// Check that files were actually modified
	htmlOptimized, err := os.ReadFile(htmlFile)
	if err != nil {
		t.Fatalf("Failed to read optimized HTML file: %v", err)
	}
	if len(htmlOptimized) >= len(htmlContent) {
		t.Error("HTML file was not optimized")
	}

	cssOptimized, err := os.ReadFile(cssFile)
	if err != nil {
		t.Fatalf("Failed to read optimized CSS file: %v", err)
	}
	if len(cssOptimized) >= len(cssContent) {
		t.Error("CSS file was not optimized")
	}

	jsOptimized, err := os.ReadFile(jsFile)
	if err != nil {
		t.Fatalf("Failed to read optimized JS file: %v", err)
	}
	if len(jsOptimized) >= len(jsContent) {
		t.Error("JS file was not optimized")
	}
}

func TestEngine_OptimizeWithSkipPatterns(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	file1 := filepath.Join(tmpDir, "optimize.css")
	file2 := filepath.Join(tmpDir, "skip.min.css")

	cssContent := `body { margin: 0; padding: 0; }`

	if err := os.WriteFile(file1, []byte(cssContent), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file2, []byte(cssContent), 0644); err != nil {
		t.Fatal(err)
	}

	config := OptimizerConfig{
		Enabled: true,
		CSS: CSSConfig{
			Enabled:   true,
			MinifyCSS: true,
		},
		SkipPatterns: []string{"*.min.css"},
	}

	engine := NewEngine(config)
	stats, err := engine.OptimizeDirectory(tmpDir)

	if err != nil {
		t.Fatalf("Optimize() error = %v", err)
	}

	// Check that skip.min.css was skipped
	if stats.FilesSkipped == 0 {
		t.Error("Expected files to be skipped")
	}

	// Check that optimize.css was processed
	if stats.FilesOptimized == 0 {
		t.Error("Expected files to be optimized")
	}
}
