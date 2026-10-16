package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// newSiteApp returns an App whose only registered site is a fresh folder, plus
// an unregistered sibling folder standing in for the rest of the disk.
func newSiteApp(t *testing.T) (app *App, site, outside string) {
	t.Helper()
	base := t.TempDir()
	site = filepath.Join(base, "site")
	outside = filepath.Join(base, "outside")
	for _, d := range []string{filepath.Join(site, "content"), outside} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(site, "content", "post.md"), []byte("post"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(outside, "secret"), []byte("secret"), 0o644); err != nil {
		t.Fatal(err)
	}
	return &App{siteRoots: func() ([]string, error) { return []string{site}, nil }}, site, outside
}

func assertDenied(t *testing.T, what string, success bool, errMsg string) {
	t.Helper()
	if success || !strings.Contains(errMsg, "access denied") {
		t.Errorf("%s: expected access denied, got success=%v err=%q", what, success, errMsg)
	}
}

func TestSandbox_AllowsFilesInsideSite(t *testing.T) {
	app, site, _ := newSiteApp(t)
	post := filepath.Join(site, "content", "post.md")

	if r := app.ReadFile(post); !r.Success || r.Content != "post" {
		t.Fatalf("ReadFile inside site: %+v", r)
	}
	if r := app.WriteFile(post, "edited"); !r.Success {
		t.Fatalf("WriteFile inside site: %+v", r)
	}
	if r := app.ListFiles(site); !r.Success || len(r.Files) != 1 {
		t.Fatalf("ListFiles on site root: %+v", r)
	}
	if r := app.CreateFile(filepath.Join(site, "content", "new.md"), "x"); !r.Success {
		t.Fatalf("CreateFile inside site: %+v", r)
	}
	if r := app.DeleteFile(post); !r.Success {
		t.Fatalf("DeleteFile inside site: %+v", r)
	}
}

func TestSandbox_RejectsPathsOutsideSites(t *testing.T) {
	app, site, outside := newSiteApp(t)
	secret := filepath.Join(outside, "secret")
	home, _ := os.UserHomeDir()

	r := app.ReadFile(secret)
	assertDenied(t, "ReadFile outside", r.Success, r.Error)
	traversal := filepath.Join(site, "content", "..", "..", "outside", "secret")
	r = app.ReadFile(traversal)
	assertDenied(t, "ReadFile traversal", r.Success, r.Error)
	w := app.WriteFile(secret, "pwned")
	assertDenied(t, "WriteFile outside", w.Success, w.Error)
	d := app.DeleteFile(home)
	assertDenied(t, "DeleteFile home", d.Success, d.Error)
	l := app.ListFiles(home)
	assertDenied(t, "ListFiles home", l.Success, l.Error)
	if err := app.OpenInFinder(outside); err == nil || !strings.Contains(err.Error(), "access denied") {
		t.Errorf("OpenInFinder outside: %v", err)
	}

	if data, _ := os.ReadFile(secret); string(data) != "secret" {
		t.Errorf("outside file was modified: %q", data)
	}
}

func TestSandbox_NeverRemovesOrReplacesSiteRoot(t *testing.T) {
	app, site, outside := newSiteApp(t)

	d := app.DeleteFile(site)
	assertDenied(t, "DeleteFile root", d.Success, d.Error)
	m := app.MoveFile(site, filepath.Join(outside, "moved"))
	assertDenied(t, "MoveFile root", m.Success, m.Error)
	rn := app.RenameFile(site, "renamed")
	assertDenied(t, "RenameFile root", rn.Success, rn.Error)
	c := app.CreateDirectory(site)
	assertDenied(t, "CreateDirectory root", c.Success, c.Error)

	if _, err := os.Stat(filepath.Join(site, "content", "post.md")); err != nil {
		t.Fatalf("site content must survive: %v", err)
	}
}

func TestSandbox_NestedSiteRootIsProtectedInAnyOrder(t *testing.T) {
	parent := t.TempDir()
	inner := filepath.Join(parent, "blog")
	if err := os.MkdirAll(filepath.Join(inner, "content"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, roots := range [][]string{{parent, inner}, {inner, parent}} {
		app := &App{siteRoots: func() ([]string, error) { return roots, nil }}
		d := app.DeleteFile(inner)
		assertDenied(t, "DeleteFile nested root", d.Success, d.Error)
		rn := app.RenameFile(inner, "renamed")
		assertDenied(t, "RenameFile nested root", rn.Success, rn.Error)
	}
	if _, err := os.Stat(filepath.Join(inner, "content")); err != nil {
		t.Fatalf("nested site must survive: %v", err)
	}
}

func TestSandbox_AncestorOfNestedRootIsProtected(t *testing.T) {
	parent := t.TempDir()
	inner := filepath.Join(parent, "group", "blog")
	if err := os.MkdirAll(filepath.Join(inner, "content"), 0o755); err != nil {
		t.Fatal(err)
	}
	app := &App{siteRoots: func() ([]string, error) { return []string{parent, inner}, nil }}
	group := filepath.Join(parent, "group")

	d := app.DeleteFile(group)
	assertDenied(t, "DeleteFile ancestor of nested root", d.Success, d.Error)
	m := app.MoveFile(group, filepath.Join(parent, "moved"))
	assertDenied(t, "MoveFile ancestor of nested root", m.Success, m.Error)
	if _, err := os.Stat(filepath.Join(inner, "content")); err != nil {
		t.Fatalf("nested site must survive: %v", err)
	}
}

func TestUploadFile_OnlyAcceptsTheDialogPick(t *testing.T) {
	app, site, outside := newSiteApp(t)
	secret := filepath.Join(outside, "secret")
	target := filepath.Join(site, "content")

	r := app.UploadFile(secret, target)
	assertDenied(t, "UploadFile without a dialog pick", r.Success, r.Error)

	app.uploadSource = secret
	if r := app.UploadFile(secret, target); !r.Success {
		t.Fatalf("UploadFile of the picked file: %+v", r)
	}
	r = app.UploadFile(secret, target)
	assertDenied(t, "UploadFile reusing a consumed pick", r.Success, r.Error)
}

func TestSandbox_ResolvesSymlinks(t *testing.T) {
	app, site, outside := newSiteApp(t)
	link := filepath.Join(site, "content", "escape")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	fileLink := filepath.Join(site, "content", "secret-link")
	if err := os.Symlink(filepath.Join(outside, "secret"), fileLink); err != nil {
		t.Fatal(err)
	}
	dangling := filepath.Join(site, "content", "dangling")
	if err := os.Symlink(filepath.Join(outside, "new-file"), dangling); err != nil {
		t.Fatal(err)
	}

	r := app.ReadFile(filepath.Join(link, "secret"))
	assertDenied(t, "ReadFile through dir symlink", r.Success, r.Error)
	r = app.ReadFile(fileLink)
	assertDenied(t, "ReadFile file symlink", r.Success, r.Error)
	w := app.WriteFile(dangling, "pwned")
	assertDenied(t, "WriteFile dangling symlink", w.Success, w.Error)
	w = app.WriteFile(filepath.Join(link, "planted"), "pwned")
	assertDenied(t, "WriteFile through dir symlink", w.Success, w.Error)
	if _, err := os.Stat(filepath.Join(outside, "new-file")); err == nil {
		t.Error("write followed a dangling symlink out of the site")
	}

	inner := filepath.Join(site, "content-link")
	if err := os.Symlink(filepath.Join(site, "content"), inner); err != nil {
		t.Fatal(err)
	}
	if r := app.ReadFile(filepath.Join(inner, "post.md")); !r.Success {
		t.Errorf("a symlink that stays inside the site should work: %+v", r)
	}
}

func TestSandbox_RegisteredSiteOutsideHomeIsAllowed(t *testing.T) {
	app, site, _ := newSiteApp(t)
	if r := app.ReadFile(filepath.Join(site, "content", "post.md")); !r.Success {
		t.Fatalf("site under the temp dir should be reachable: %+v", r)
	}
}

func TestSandbox_NoRegisteredSitesDeniesEverything(t *testing.T) {
	app := &App{siteRoots: func() ([]string, error) { return nil, nil }}
	home, _ := os.UserHomeDir()
	r := app.ListFiles(home)
	assertDenied(t, "ListFiles with no sites", r.Success, r.Error)
}

func TestOpenInEditor_RejectsUndetectedCommand(t *testing.T) {
	app, site, _ := newSiteApp(t)
	for _, cmd := range []string{"sh", "/bin/sh", "osascript", "code; rm -rf ~"} {
		err := app.OpenInEditor(cmd, site)
		if err == nil || !strings.Contains(err.Error(), "not a detected code editor") {
			t.Errorf("OpenInEditor(%q) should be refused, got %v", cmd, err)
		}
	}
}

func TestOpenInBrowser_RejectsNonHTTP(t *testing.T) {
	app := &App{}
	for _, u := range []string{"file:///etc/passwd", "javascript:alert(1)", "vscode://x", "/relative", ""} {
		if err := app.OpenInBrowser(u); err == nil {
			t.Errorf("OpenInBrowser(%q) should be refused", u)
		}
	}
}
