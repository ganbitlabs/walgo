package projects

import (
	"os"
	"path/filepath"
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
