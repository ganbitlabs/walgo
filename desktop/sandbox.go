package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// siteSandbox confines paths sent by the frontend to registered site folders.
// The webview renders user content, so every bound file method treats its path
// argument as untrusted.
type siteSandbox struct {
	roots []string
}

func newSiteSandbox(sitePaths []string) siteSandbox {
	var roots []string
	for _, p := range sitePaths {
		if p == "" || !filepath.IsAbs(p) {
			continue
		}
		real, err := filepath.EvalSymlinks(filepath.Clean(p))
		if err != nil {
			continue
		}
		roots = append(roots, real)
	}
	return siteSandbox{roots: roots}
}

func (s siteSandbox) inside(p string) bool {
	for _, r := range s.roots {
		if p == r || strings.HasPrefix(p, r+string(filepath.Separator)) {
			return true
		}
	}
	return false
}

func (s siteSandbox) isRoot(p string) bool {
	for _, r := range s.roots {
		if p == r {
			return true
		}
	}
	return false
}

// resolve returns the cleaned absolute form of p when it lies inside a site
// folder after symlinks are resolved. With allowRoot false, p must name an entry
// strictly inside a site, so the caller may create, delete, or move it without
// ever touching the site folder itself.
func (s siteSandbox) resolve(p string, allowRoot bool) (string, error) {
	if p == "" {
		return "", errors.New("access denied: empty path")
	}
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}
	parent, err := resolveExisting(filepath.Dir(abs))
	if err != nil {
		return "", fmt.Errorf("access denied: %w", err)
	}
	entry := filepath.Join(parent, filepath.Base(abs))
	real := entry
	if info, err := os.Lstat(entry); err == nil && info.Mode()&fs.ModeSymlink != 0 {
		if real, err = filepath.EvalSymlinks(entry); err != nil {
			return "", fmt.Errorf("access denied: broken symlink %s", p)
		}
	}

	if !s.inside(real) {
		return "", errors.New("access denied: path is outside the registered site folders")
	}
	if !allowRoot && (!s.inside(entry) || s.isRoot(entry) || s.isRoot(real)) {
		return "", errors.New("access denied: a site folder itself cannot be created, moved, or deleted here")
	}
	return abs, nil
}

// resolveExisting resolves symlinks in the longest existing prefix of p and
// appends the missing tail, so paths that do not exist yet can be checked.
func resolveExisting(p string) (string, error) {
	var tail []string
	cur := p
	for {
		real, err := filepath.EvalSymlinks(cur)
		if err == nil {
			for i := len(tail) - 1; i >= 0; i-- {
				real = filepath.Join(real, tail[i])
			}
			return real, nil
		}
		if !errors.Is(err, fs.ErrNotExist) {
			return "", err
		}
		if _, lerr := os.Lstat(cur); lerr == nil {
			return "", fmt.Errorf("broken symlink %s", cur)
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return "", err
		}
		tail = append(tail, filepath.Base(cur))
		cur = parent
	}
}

func (a *App) sandbox() (siteSandbox, error) {
	paths, err := a.siteRoots()
	if err != nil {
		return siteSandbox{}, fmt.Errorf("cannot load registered sites: %w", err)
	}
	return newSiteSandbox(paths), nil
}

// sitePath allows any path inside a site folder, the folder itself included.
func (a *App) sitePath(p string) (string, error) {
	s, err := a.sandbox()
	if err != nil {
		return "", err
	}
	return s.resolve(p, true)
}

// siteEntry allows only paths strictly inside a site folder.
func (a *App) siteEntry(p string) (string, error) {
	s, err := a.sandbox()
	if err != nil {
		return "", err
	}
	return s.resolve(p, false)
}
