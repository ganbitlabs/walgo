package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// Each flow that hands a site path to the editor must register the site before
// the editor's first file call, or the sandbox would lock the user out.

func isolatedHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	return home
}

func requireHugo(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("hugo"); err != nil {
		t.Skip("hugo not installed")
	}
}

func assertEditable(t *testing.T, app *App, sitePath string) {
	t.Helper()
	if r := app.ListFiles(sitePath); !r.Success {
		t.Fatalf("ListFiles(%s) after creation: %s", sitePath, r.Error)
	}
	probe := filepath.Join(sitePath, "content", "probe.md")
	if r := app.CreateFile(probe, "x"); !r.Success {
		t.Fatalf("CreateFile in new site: %s", r.Error)
	}
	if r := app.ReadFile(probe); !r.Success {
		t.Fatalf("ReadFile in new site: %s", r.Error)
	}
}

func TestSiteFlow_InitSiteRegistersBeforeEditing(t *testing.T) {
	requireHugo(t)
	home := isolatedHome(t)
	app := NewApp()

	res := app.InitSite(filepath.Join(home, "sites"), "blog")
	if !res.Success {
		t.Fatalf("InitSite: %s", res.Error)
	}
	assertEditable(t, app, res.SitePath)
}

func TestSiteFlow_ImportObsidianRegistersBeforeEditing(t *testing.T) {
	requireHugo(t)
	home := isolatedHome(t)
	vault := filepath.Join(home, "vault")
	if err := os.MkdirAll(vault, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(vault, "note.md"), []byte("# Note\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	parent := filepath.Join(home, "sites")
	if err := os.MkdirAll(parent, 0o755); err != nil {
		t.Fatal(err)
	}
	app := NewApp()

	res := app.ImportObsidian(ImportObsidianParams{VaultPath: vault, SiteName: "notes", ParentDir: parent})
	if !res.Success {
		t.Fatalf("ImportObsidian: %s", res.Error)
	}
	assertEditable(t, app, res.SitePath)
}

func TestSiteFlow_UnregisteredFolderIsDenied(t *testing.T) {
	home := isolatedHome(t)
	stray := filepath.Join(home, "stray")
	if err := os.MkdirAll(stray, 0o755); err != nil {
		t.Fatal(err)
	}
	r := NewApp().ListFiles(stray)
	assertDenied(t, "ListFiles unregistered", r.Success, r.Error)
}
