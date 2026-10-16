package hugo

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// fakeHugoSite puts a hugo on PATH that runs script and returns a site that
// BuildSite treats as a Hugo site.
func fakeHugoSite(t *testing.T, script string) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("fake hugo is a shell script")
	}
	bin := t.TempDir()
	writeFiles(t, bin, map[string]string{"hugo": "#!/bin/sh\n" + script})
	if err := os.Chmod(filepath.Join(bin, "hugo"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+":/bin:/usr/bin")
	isolateHome(t)

	site := t.TempDir()
	writeFiles(t, site, map[string]string{
		"walgo.yaml": "hugo:\n  publishDir: public\n",
		"hugo.toml":  "title = 'test'\n",
	})
	return site
}

// withBrokenTerminal makes every write to os.Stdout and os.Stderr fail, as in
// a Windows GUI app with no console.
func withBrokenTerminal(t *testing.T) {
	t.Helper()
	closed, err := os.CreateTemp(t.TempDir(), "closed")
	if err != nil {
		t.Fatal(err)
	}
	_ = closed.Close()
	stdout, stderr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = closed, closed
	t.Cleanup(func() { os.Stdout, os.Stderr = stdout, stderr })
}

func TestBuildSiteSucceedsWithoutATerminal(t *testing.T) {
	site := fakeHugoSite(t, "echo building\necho warn >&2\nmkdir -p public && echo hi > public/index.html\n")
	withBrokenTerminal(t)

	if err := BuildSite(site); err != nil {
		t.Fatalf("build failed because the terminal is unwritable: %v", err)
	}
}

func TestBuildSiteFailureCarriesHugoOutput(t *testing.T) {
	site := fakeHugoSite(t, "echo 'Error: module \"ananke\" not found' >&2\nexit 1\n")
	withBrokenTerminal(t)

	err := BuildSite(site)
	var buildErr *BuildError
	if !errors.As(err, &buildErr) {
		t.Fatalf("err = %v, want a *BuildError", err)
	}
	if !strings.Contains(buildErr.Output, `module "ananke" not found`) {
		t.Errorf("Output = %q, want hugo's stderr", buildErr.Output)
	}
}
