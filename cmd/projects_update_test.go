package cmd

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/ganbitlabs/walgo/internal/projects"
)

const projectsUpdateObjectID = "0x4444444444444444444444444444444444444444444444444444444444444444"

// fakeProjectsUpdateEnv puts fake sui (on testnet, with no active address), walrus, site-builder
// and hugo on PATH under a fresh HOME, and returns a Hugo site plus the file site-builder logs its calls to.
func fakeProjectsUpdateEnv(t *testing.T, siteBuilder string) (sitePath, builderLog string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("fake tools are shell scripts")
	}
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeTestFile(t, filepath.Join(home, ".config", "walrus", "sites-config.yaml"), "contexts: {}\n")

	builderLog = filepath.Join(t.TempDir(), "site-builder.log")
	if siteBuilder == "" {
		siteBuilder = "[ \"$1\" = --version ] && exit 0\necho \"$*\" >> \"" + builderLog + "\"\n"
	}
	bin := t.TempDir()
	tools := map[string]string{
		"sui":          "case \"$2\" in\n  active-env) echo testnet ;;\n  *) exit 1 ;;\nesac\n",
		"walrus":       "exit 0\n",
		"site-builder": siteBuilder,
		// An unreadable subdirectory keeps the cost estimator off the network.
		"hugo": "mkdir -p public/locked && echo '<h1>hi</h1>' > public/index.html && echo '{}' > public/ws-resources.json && chmod 000 public/locked\n",
	}
	for name, body := range tools {
		writeTestFile(t, filepath.Join(bin, name), "#!/bin/sh\n"+body)
		if err := os.Chmod(filepath.Join(bin, name), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+"/bin:/usr/bin")

	sitePath = t.TempDir()
	stale := "0x" + strings.Repeat("9", 64)
	writeTestFile(t, filepath.Join(sitePath, "walgo.yaml"), "walrus:\n  projectID: \""+stale+"\"\nhugo:\n  publishDir: public\noptimizer:\n  enabled: false\n")
	writeTestFile(t, filepath.Join(sitePath, "hugo.toml"), "title = 'test'\n")
	t.Cleanup(func() { _ = os.Chmod(filepath.Join(sitePath, "public", "locked"), 0o755) })
	return sitePath, builderLog
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func createTestProject(t *testing.T, p *projects.Project) (*projects.Project, *projects.Manager) {
	t.Helper()
	pm, err := projects.NewManager()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = pm.Close() })
	if err := pm.CreateProject(p); err != nil {
		t.Fatal(err)
	}
	return p, pm
}

func TestProjectsUpdateRunsSharedDeployment(t *testing.T) {
	sitePath, builderLog := fakeProjectsUpdateEnv(t, "")
	proj, pm := createTestProject(t, &projects.Project{
		Name: "blog", Network: "testnet", ObjectID: projectsUpdateObjectID, SitePath: sitePath, Epochs: 2, Description: "my blog",
	})

	if err := updateProjectByRef(proj, 3); err != nil {
		t.Fatalf("projects update failed: %v", err)
	}

	calls, err := os.ReadFile(builderLog)
	if err != nil {
		t.Fatalf("site-builder never ran: %v", err)
	}
	if !strings.Contains(string(calls), "update --epochs 3") || !strings.Contains(string(calls), projectsUpdateObjectID) {
		t.Errorf("site-builder was not asked to update the project's object:\n%s", calls)
	}

	ws, err := os.ReadFile(filepath.Join(sitePath, "public", "ws-resources.json"))
	if err != nil || !strings.Contains(string(ws), projectsUpdateObjectID) || !strings.Contains(string(ws), "my blog") {
		t.Errorf("ws-resources.json lacks the project's object ID and metadata: %s (%v)", ws, err)
	}
	yaml, err := os.ReadFile(filepath.Join(sitePath, "walgo.yaml"))
	if err != nil || !strings.Contains(string(yaml), projectsUpdateObjectID) {
		t.Errorf("walgo.yaml does not record the project's object ID: %s (%v)", yaml, err)
	}

	got, err := pm.GetProject(proj.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Epochs != 3 || got.ObjectID != projectsUpdateObjectID {
		t.Errorf("project after update = epochs %d, object %s; want 3, %s", got.Epochs, got.ObjectID, projectsUpdateObjectID)
	}
	records, err := pm.GetProjectDeployments(proj.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 1 || !records[0].Success || records[0].Epochs != 3 {
		t.Errorf("deployment history = %+v, want one successful 3-epoch record", records)
	}
	all, err := pm.ListProjects("", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 1 {
		t.Errorf("projects after update = %d, want the one updated project", len(all))
	}
}

func TestProjectsUpdateRecordsFailure(t *testing.T) {
	sitePath, _ := fakeProjectsUpdateEnv(t, "[ \"$1\" = --version ] && exit 0\necho boom >&2\nexit 1\n")
	proj, pm := createTestProject(t, &projects.Project{
		Name: "blog", Network: "testnet", ObjectID: projectsUpdateObjectID, SitePath: sitePath, Epochs: 1,
	})

	if err := updateProjectByRef(proj, 0); err == nil {
		t.Fatal("projects update succeeded although site-builder failed")
	}
	records, err := pm.GetProjectDeployments(proj.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 1 || records[0].Success || records[0].Error == "" {
		t.Errorf("deployment history = %+v, want one failed record with its error", records)
	}
}

func TestProjectsUpdateRefusesUndeployedProject(t *testing.T) {
	sitePath, builderLog := fakeProjectsUpdateEnv(t, "")
	proj, _ := createTestProject(t, &projects.Project{Name: "draft", Network: "testnet", SitePath: sitePath, Epochs: 1})

	err := updateProjectByRef(proj, 0)
	if err == nil || !strings.Contains(err.Error(), "no deployed site") {
		t.Fatalf("err = %v, want a refusal naming the missing site", err)
	}
	if calls, err := os.ReadFile(builderLog); err == nil {
		t.Errorf("site-builder ran for an undeployed project:\n%s", calls)
	}
}

func setStoredSitePath(t *testing.T, pm *projects.Manager, proj *projects.Project, sitePath string) {
	t.Helper()
	proj.SitePath = sitePath
	if err := pm.UpdateProject(proj); err != nil {
		t.Fatal(err)
	}
}

func TestProjectsUpdateLeavesUnresolvableLegacyPathAlone(t *testing.T) {
	sitePath, builderLog := fakeProjectsUpdateEnv(t, "")
	proj, pm := createTestProject(t, &projects.Project{
		Name: "legacy", Network: "testnet", ObjectID: projectsUpdateObjectID, SitePath: sitePath, Epochs: 1,
	})
	legacy := filepath.Join("sites", "legacy")
	setStoredSitePath(t, pm, proj, legacy)
	t.Chdir(sitePath)

	err := updateProjectByRef(proj, 0)
	if err == nil || !strings.Contains(err.Error(), "relative site path") {
		t.Fatalf("err = %v, want a refusal naming the relative path", err)
	}
	if got, _ := pm.GetProject(proj.ID); got == nil || got.SitePath != legacy {
		t.Errorf("stored site path = %+v, want it left as %q", got, legacy)
	}
	if calls, err := os.ReadFile(builderLog); err == nil {
		t.Errorf("site-builder ran for an unresolvable site:\n%s", calls)
	}
}

func TestProjectsUpdateRecordsResolvedLegacyPath(t *testing.T) {
	sitePath, _ := fakeProjectsUpdateEnv(t, "")
	proj, pm := createTestProject(t, &projects.Project{
		Name: "legacy", Network: "testnet", ObjectID: projectsUpdateObjectID, SitePath: sitePath, Epochs: 1,
	})
	t.Chdir(filepath.Dir(sitePath))
	setStoredSitePath(t, pm, proj, filepath.Base(sitePath))

	if err := updateProjectByRef(proj, 0); err != nil {
		t.Fatalf("projects update failed: %v", err)
	}
	wd, _ := os.Getwd()
	want := filepath.Join(wd, filepath.Base(sitePath))
	if got, _ := pm.GetProject(proj.ID); got == nil || got.SitePath != want {
		t.Errorf("stored site path = %+v, want %q", got, want)
	}
	all, _ := pm.ListProjects("", "")
	if len(all) != 1 {
		t.Errorf("projects after update = %d, want the one legacy project", len(all))
	}
}
