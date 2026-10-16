package api

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ganbitlabs/walgo/internal/projects"
)

func TestGetProjectMeasuresConfiguredPublishDir(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	site := t.TempDir()
	if err := os.WriteFile(filepath.Join(site, "walgo.yaml"), []byte("hugo:\n  publishDir: dist\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(site, "dist"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(site, "dist", "index.html"), []byte("<p>hi</p>"), 0o644); err != nil {
		t.Fatal(err)
	}

	pm, err := projects.NewManager()
	if err != nil {
		t.Fatal(err)
	}
	proj := &projects.Project{Name: "site", SitePath: site, Network: "testnet"}
	if err := pm.CreateProject(proj); err != nil {
		t.Fatal(err)
	}
	pm.Close()

	got, err := GetProject(proj.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.FileCount != 1 {
		t.Errorf("FileCount = %d, want 1 (from dist/)", got.FileCount)
	}
}
