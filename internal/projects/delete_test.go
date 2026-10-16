package projects

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDeleteProjectRefusesToRemoveHomeOrRoot(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	manager, err := NewManager()
	if err != nil {
		t.Fatal(err)
	}
	defer manager.Close()

	for i, sitePath := range []string{home, filepath.Dir(home), "/", "relative/site"} {
		project := &Project{Name: filepath.Base(sitePath) + string(rune('a'+i)), Network: "testnet", SitePath: sitePath, Status: "active"}
		if err := manager.CreateProject(project); err != nil {
			t.Fatal(err)
		}
		if err := manager.DeleteProjectWithOptions(project.ID, true); err == nil {
			t.Errorf("deleting site folder %q should be refused", sitePath)
		}
		if _, err := manager.GetProject(project.ID); err != nil {
			t.Errorf("refused delete of %q must keep the record: %v", sitePath, err)
		}
	}
	if _, err := os.Stat(home); err != nil {
		t.Fatalf("home must survive: %v", err)
	}
}

func TestCheckRemovableSiteFolder_SymlinkAndCaseVariantsOfHome(t *testing.T) {
	real := filepath.Join(t.TempDir(), "realhome")
	if err := os.MkdirAll(real, 0o755); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(t.TempDir(), "home")
	if err := os.Symlink(real, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	t.Setenv("HOME", link)

	if err := CheckRemovableSiteFolder(real); err == nil {
		t.Error("the real target of a symlinked home must be refused")
	}
	if err := CheckRemovableSiteFolder(filepath.Dir(real)); err == nil {
		t.Error("an ancestor of the real home must be refused")
	}

	upper := filepath.Join(filepath.Dir(real), strings.ToUpper(filepath.Base(real)))
	if _, err := os.Stat(upper); err != nil {
		t.Skip("filesystem is case-sensitive")
	}
	if err := CheckRemovableSiteFolder(upper); err == nil {
		t.Errorf("case variant %s of home must be refused", upper)
	}
}
