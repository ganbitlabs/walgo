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

func TestValidateGitHubURL(t *testing.T) {
	ok := []string{"https://github.com/a/b", "https://www.github.com/a/b"}
	bad := []string{
		"http://github.com/a/b", "https://evil.com/github.com/a", "https://github.com.evil.com/a/b",
		"https://evil.com/?x=github.com", "https://user@github.com/a/b", "git@github.com:a/b", "",
		"https://github.com/foo/..", "https://github.com/foo/.", "https://github.com/foo", "https://github.com/", "https://github.com/foo/../..", "https://github.com/a/b/tree/main",
	}
	for _, u := range ok {
		if err := validateGitHubURL(u); err != nil {
			t.Errorf("%q: %v", u, err)
		}
	}
	for _, u := range bad {
		if err := validateGitHubURL(u); err == nil {
			t.Errorf("%q should be rejected", u)
		}
	}
}
