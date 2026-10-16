package hugo

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func writeZip(t *testing.T, entries map[string]string, links map[string]string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "t.zip")
	f, err := os.Create(p)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	for name, body := range entries {
		w, _ := zw.Create(name)
		_, _ = w.Write([]byte(body))
	}
	for name, target := range links {
		h := &zip.FileHeader{Name: name}
		h.SetMode(os.ModeSymlink | 0777)
		w, _ := zw.CreateHeader(h)
		_, _ = w.Write([]byte(target))
	}
	_ = zw.Close()
	_ = f.Close()
	return p
}

func TestExtractThemeZipRejectsZipSlip(t *testing.T) {
	root := t.TempDir()
	dest := filepath.Join(root, "themes", "t")
	z := writeZip(t, map[string]string{
		"repo-main/":              "",
		"repo-main/theme.toml":    "x",
		"repo-main/../../evil.sh": "boom",
	}, nil)
	if _, err := extractThemeZip(z, dest); err == nil {
		t.Fatal("expected zip-slip entry to be rejected")
	}
	if _, err := os.Stat(filepath.Join(root, "themes", "evil.sh")); err == nil {
		t.Fatal("file escaped theme dir")
	}
}

func TestExtractThemeZipSkipsSymlinks(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "t")
	z := writeZip(t,
		map[string]string{"repo-main/theme.toml": "x"},
		map[string]string{"repo-main/link": "/etc/passwd"})
	n, err := extractThemeZip(z, dest)
	if err != nil || n != 1 {
		t.Fatalf("n=%d err=%v", n, err)
	}
	if _, err := os.Lstat(filepath.Join(dest, "link")); err == nil {
		t.Fatal("symlink was extracted")
	}
}

func TestParseGitHubRepo(t *testing.T) {
	ok := map[string][2]string{
		"https://github.com/a/b":      {"a", "b"},
		"https://github.com/a/b/":     {"a", "b"},
		"https://github.com/a/b.git":  {"a", "b"},
		"https://github.com/a/b.git/": {"a", "b"},
		"https://www.github.com/a/b":  {"a", "b"},
	}
	for u, want := range ok {
		o, r, err := parseGitHubRepo(u)
		if err != nil || o != want[0] || r != want[1] {
			t.Errorf("%q: got %q %q %v", u, o, r, err)
		}
	}
	bad := []string{
		"", "http://github.com/a/b", "https://evil.com/github.com/a", "https://github.com.evil.com/a/b",
		"https://evil.com/?x=github.com", "https://user@github.com/a/b", "git@github.com:a/b",
		"https://github.com/foo/..", "https://github.com/foo/.", "https://github.com/foo",
		"https://github.com/", "https://github.com/foo/../..", "https://github.com/a/b/tree/main",
		"https://github.com/owner/repo#/..", "https://github.com/owner/repo?x=/..",
		"https://github.com/owner/repo/?x=/..", "https://github.com/owner/repo.git#/..",
		"https://github.com/owner/%2e%2e", "https://github.com/owner/%2E%2E/", "https://github.com/o/r%2f..",
		"https://github.com/owner/.git", "https://github.com/owner/..git", "https://github.com:444/a/b",
		"https://github.com/a\\b", "https://github.com/a/b c",
	}
	for _, u := range bad {
		if o, r, err := parseGitHubRepo(u); err == nil {
			t.Errorf("%q should be rejected, got %q %q", u, o, r)
		}
	}
}

func TestThemeChildPath(t *testing.T) {
	dir := t.TempDir()
	if p, err := themeChildPath(dir, "ananke"); err != nil || p != filepath.Join(dir, "ananke") {
		t.Fatalf("got %q %v", p, err)
	}
	for _, n := range []string{"", ".", "..", "a/b", "../x", "a\\b", "/abs"} {
		if _, err := themeChildPath(dir, n); err == nil {
			t.Errorf("%q should be rejected", n)
		}
	}
}

func TestInstallThemeFromURLRejectsTraversalURLs(t *testing.T) {
	site := t.TempDir()
	marker := filepath.Join(site, "keep.txt")
	if err := os.WriteFile(marker, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	for _, u := range []string{
		"https://github.com/owner/repo#/..", "https://github.com/owner/repo?x=/..",
		"https://github.com/owner/%2e%2e", "https://github.com/foo/..",
	} {
		if _, err := InstallThemeFromURL(site, u); err == nil {
			t.Errorf("%q should fail", u)
		}
	}
	if _, err := os.Stat(marker); err != nil {
		t.Fatal("site content was removed")
	}
}
