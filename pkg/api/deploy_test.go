package api

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/ganbitlabs/walgo/internal/projects"
)

const testObjectID = "0x1111111111111111111111111111111111111111111111111111111111111111"

type fakeEnv struct {
	sitePath   string
	builderLog string // every site-builder call, one per line
	bin        string
}

// fakeDeployEnv puts fake sui, walrus, site-builder and hugo on PATH under a
// fresh HOME. sui answers active-env with activeEnv and active-address with
// activeAddr, or fails when activeAddr is empty, which keeps the gas lookup
// off the network.
func fakeDeployEnv(t *testing.T, activeEnv, activeAddr string) fakeEnv {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("fake tools are shell scripts")
	}

	home := t.TempDir()
	t.Setenv("HOME", home)
	sitesConfig := filepath.Join(home, ".config", "walrus", "sites-config.yaml")
	writeFile(t, sitesConfig, "contexts: {}\n")

	builderLog := filepath.Join(t.TempDir(), "site-builder.log")
	addr := "exit 1"
	if activeAddr != "" {
		addr = "echo " + activeAddr
	}
	bin := t.TempDir()
	tools := map[string]string{
		"sui": "case \"$2\" in\n" +
			"  active-env) echo " + activeEnv + " ;;\n" +
			"  active-address) " + addr + " ;;\n" +
			"  *) exit 1 ;;\n" +
			"esac\n",
		"walrus":       "exit 0\n",
		"site-builder": "[ \"$1\" = --version ] && exit 0\necho \"$*\" >> \"" + builderLog + "\"\n",
		// An unreadable subdirectory keeps the cost estimator off the network.
		"hugo": "mkdir -p public/locked && echo '<h1>hi</h1>' > public/index.html && echo '{}' > public/ws-resources.json && chmod 000 public/locked\n",
	}
	for name, body := range tools {
		writeFile(t, filepath.Join(bin, name), "#!/bin/sh\n"+body)
		if err := os.Chmod(filepath.Join(bin, name), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+"/bin:/usr/bin")

	sitePath := t.TempDir()
	writeFile(t, filepath.Join(sitePath, "walgo.yaml"), "hugo:\n  publishDir: public\noptimizer:\n  enabled: false\n")
	writeFile(t, filepath.Join(sitePath, "hugo.toml"), "title = 'test'\n")
	t.Cleanup(func() { _ = os.Chmod(filepath.Join(sitePath, "public", "locked"), 0o755) })
	return fakeEnv{sitePath: sitePath, builderLog: builderLog, bin: bin}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func createProject(t *testing.T, p *projects.Project) *projects.Project {
	t.Helper()
	pm, err := projects.NewManager()
	if err != nil {
		t.Fatal(err)
	}
	defer pm.Close()
	if err := pm.CreateProject(p); err != nil {
		t.Fatal(err)
	}
	return p
}

func deployments(t *testing.T, projectID int64) []*projects.DeploymentRecord {
	t.Helper()
	pm, err := projects.NewManager()
	if err != nil {
		t.Fatal(err)
	}
	defer pm.Close()
	records, err := pm.GetProjectDeployments(projectID)
	if err != nil {
		t.Fatal(err)
	}
	return records
}

func TestUpdateSiteRunsSharedDeployment(t *testing.T) {
	env := fakeDeployEnv(t, "testnet", "")
	proj := createProject(t, &projects.Project{
		Name: "blog", Network: "testnet", ObjectID: testObjectID, SitePath: env.sitePath, Epochs: 2,
	})

	result := UpdateSite(context.Background(), UpdateSiteParams{ProjectID: proj.ID, Epochs: 3})
	if !result.Success {
		t.Fatalf("UpdateSite failed: %s", result.Error)
	}
	if result.ObjectID != testObjectID {
		t.Errorf("ObjectID = %q, want %q", result.ObjectID, testObjectID)
	}

	calls, err := os.ReadFile(env.builderLog)
	if err != nil {
		t.Fatalf("site-builder never ran: %v", err)
	}
	if !strings.Contains(string(calls), "update --epochs 3") || !strings.Contains(string(calls), testObjectID) {
		t.Errorf("site-builder was not asked to update the project's site:\n%s", calls)
	}

	yaml, err := os.ReadFile(filepath.Join(env.sitePath, "walgo.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(yaml), testObjectID) {
		t.Errorf("walgo.yaml does not record the object ID:\n%s", yaml)
	}
	ws, err := os.ReadFile(filepath.Join(env.sitePath, "public", "ws-resources.json"))
	if err != nil || !strings.Contains(string(ws), testObjectID) {
		t.Errorf("ws-resources.json does not record the object ID: %s (%v)", ws, err)
	}

	records := deployments(t, proj.ID)
	if len(records) != 1 || !records[0].Success || records[0].Epochs != 3 {
		t.Errorf("deployment history = %+v, want one successful 3-epoch record", records)
	}
}

func TestUpdateSiteRefusesWrongTarget(t *testing.T) {
	tests := []struct {
		name       string
		activeEnv  string
		activeAddr string
		wantErr    string
	}{
		{"other network", "mainnet", "", "network mismatch"},
		{"other wallet", "testnet", "0xbbb", "wallet mismatch"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := fakeDeployEnv(t, tt.activeEnv, tt.activeAddr)
			proj := createProject(t, &projects.Project{
				Name: "blog", Network: "testnet", ObjectID: testObjectID, SitePath: env.sitePath, WalletAddr: "0xaaa", Epochs: 1,
			})

			result := UpdateSite(context.Background(), UpdateSiteParams{ProjectID: proj.ID})
			if result.Success {
				t.Fatal("UpdateSite succeeded against the wrong target")
			}
			if !strings.Contains(result.Error, tt.wantErr) {
				t.Errorf("error = %q, want it to mention %q", result.Error, tt.wantErr)
			}
			if calls, err := os.ReadFile(env.builderLog); err == nil {
				t.Errorf("site-builder ran despite the refusal:\n%s", calls)
			}
		})
	}
}

func TestUpdateSiteRecordsFailure(t *testing.T) {
	env := fakeDeployEnv(t, "testnet", "")
	writeFile(t, filepath.Join(env.bin, "site-builder"), "#!/bin/sh\necho boom >&2\nexit 1\n")
	proj := createProject(t, &projects.Project{
		Name: "blog", Network: "testnet", ObjectID: testObjectID, SitePath: env.sitePath, Epochs: 1,
	})

	result := UpdateSite(context.Background(), UpdateSiteParams{ProjectID: proj.ID})
	if result.Success {
		t.Fatal("UpdateSite succeeded although site-builder failed")
	}
	records := deployments(t, proj.ID)
	if len(records) != 1 || records[0].Success || records[0].Error == "" {
		t.Errorf("deployment history = %+v, want one failed record with its error", records)
	}
}

func TestUpdateSiteCancelledBeforeUpload(t *testing.T) {
	env := fakeDeployEnv(t, "testnet", "")
	proj := createProject(t, &projects.Project{
		Name: "blog", Network: "testnet", ObjectID: testObjectID, SitePath: env.sitePath, Epochs: 1,
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	result := UpdateSite(ctx, UpdateSiteParams{ProjectID: proj.ID})
	if result.Success || !strings.Contains(result.Error, "cancelled before anything was uploaded") {
		t.Errorf("result = %+v, want a cancellation error", result)
	}
	if calls, err := os.ReadFile(env.builderLog); err == nil {
		t.Errorf("site-builder ran after cancellation:\n%s", calls)
	}
}

func TestLaunchWizardRefusesNetworkMismatch(t *testing.T) {
	env := fakeDeployEnv(t, "testnet", "")

	result := LaunchWizard(context.Background(), LaunchWizardParams{SitePath: env.sitePath, Network: "mainnet", Epochs: 1})
	if result.Success || !strings.Contains(result.Error, "network mismatch") {
		t.Errorf("result = %+v, want a network mismatch error", result)
	}
	if calls, err := os.ReadFile(env.builderLog); err == nil {
		t.Errorf("site-builder ran despite the refusal:\n%s", calls)
	}
}

func TestUpdateSiteShowsHugoOutputOnBuildFailure(t *testing.T) {
	env := fakeDeployEnv(t, "testnet", "")
	writeFile(t, filepath.Join(env.bin, "hugo"), "#!/bin/sh\necho 'Error: module \"ananke\" not found' >&2\nexit 1\n")
	proj := createProject(t, &projects.Project{
		Name: "blog", Network: "testnet", ObjectID: testObjectID, SitePath: env.sitePath, Epochs: 1,
	})

	result := UpdateSite(context.Background(), UpdateSiteParams{ProjectID: proj.ID})
	if result.Success {
		t.Fatal("UpdateSite succeeded although hugo failed")
	}
	if !strings.Contains(result.Error, `module "ananke" not found`) {
		t.Errorf("error = %q, want hugo's own output", result.Error)
	}
	if calls, err := os.ReadFile(env.builderLog); err == nil {
		t.Errorf("site-builder ran after a failed build:\n%s", calls)
	}
}

func TestUpdateSiteDeploysAbsolutePublishDir(t *testing.T) {
	env := fakeDeployEnv(t, "testnet", "")
	out := filepath.Join(t.TempDir(), "built")
	writeFile(t, filepath.Join(out, "index.html"), "<h1>static</h1>")
	if err := os.Remove(filepath.Join(env.sitePath, "hugo.toml")); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(env.sitePath, "walgo.yaml"), "hugo:\n  publishDir: "+out+"\n")
	proj := createProject(t, &projects.Project{
		Name: "static", Network: "testnet", ObjectID: testObjectID, SitePath: env.sitePath, Epochs: 1,
	})

	result := UpdateSite(context.Background(), UpdateSiteParams{ProjectID: proj.ID})
	if !result.Success {
		t.Fatalf("UpdateSite failed: %s", result.Error)
	}
	calls, err := os.ReadFile(env.builderLog)
	if err != nil {
		t.Fatalf("site-builder never ran: %v", err)
	}
	if !strings.Contains(string(calls), out) || strings.Contains(string(calls), filepath.Join(env.sitePath, out)) {
		t.Errorf("site-builder was not pointed at the absolute publish dir %s:\n%s", out, calls)
	}
}
