package version

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// fakeToolchain puts sui, walrus and site-builder scripts that report the given
// versions first on PATH, with HOME pointed away from any real install.
func fakeToolchain(t *testing.T, versions map[string]string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("fake tools are shell scripts")
	}
	bin := t.TempDir()
	for tool, v := range versions {
		script := "#!/bin/sh\necho \"" + tool + " " + v + "\"\n"
		if err := os.WriteFile(filepath.Join(bin, tool), []byte(script), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", bin)
	t.Setenv("HOME", t.TempDir())
}

func TestRequireDeployToolchain(t *testing.T) {
	t.Run("current tools pass", func(t *testing.T) {
		fakeToolchain(t, map[string]string{"sui": "1.75.0-abc", "walrus": "1.52.0", "site-builder": "2.12.0"})
		if err := RequireDeployToolchain(); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("outdated site-builder blocks with a fix", func(t *testing.T) {
		fakeToolchain(t, map[string]string{"sui": "1.75.0", "walrus": "1.52.0", "site-builder": "2.11.9"})
		err := RequireDeployToolchain()
		if err == nil {
			t.Fatal("expected outdated site-builder to block the deploy")
		}
		for _, want := range []string{"site-builder v2.11.9", "need v2.12.0", "walgo setup-deps"} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("error missing %q:\n%v", want, err)
			}
		}
	})
}

func TestOfferToolUpdatesNeverInstallsWithoutConsent(t *testing.T) {
	result := &CheckResult{
		Walrus:     &ToolVersion{Tool: "walrus", CurrentVersion: "1.52.0", LatestVersion: "1.60.0", UpdateRequired: true},
		HasUpdates: true,
	}

	tests := []struct {
		name        string
		quiet       bool
		interactive bool
		input       string
		wantInstall bool
	}{
		{name: "no TTY", interactive: false, input: "y\n", wantInstall: false},
		{name: "quiet", quiet: true, interactive: true, input: "y\n", wantInstall: false},
		{name: "stdin closed", interactive: true, input: "", wantInstall: false},
		{name: "declined", interactive: true, input: "n\n", wantInstall: false},
		{name: "enter accepts default", interactive: true, input: "\n", wantInstall: true},
		{name: "yes", interactive: true, input: "yes\n", wantInstall: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			installed := false
			original := installTools
			installTools = func(string) error { installed = true; return nil }
			t.Cleanup(func() { installTools = original })

			if err := offerToolUpdates(result, "mainnet", tt.quiet, tt.interactive, strings.NewReader(tt.input)); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if installed != tt.wantInstall {
				t.Errorf("installed = %v, want %v", installed, tt.wantInstall)
			}
		})
	}
}
