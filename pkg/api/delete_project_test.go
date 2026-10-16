package api

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ganbitlabs/walgo/internal/projects"
)

type fakeChain struct {
	env, addr  string
	destroyErr error
	destroyed  []string
}

func installFakeChain(t *testing.T, c *fakeChain) {
	t.Helper()
	origEnv, origAddr, origDestroy := activeSuiEnv, activeSuiAddress, destroySite
	activeSuiEnv = func() (string, error) { return c.env, nil }
	activeSuiAddress = func() (string, error) { return c.addr, nil }
	destroySite = func(_ context.Context, id string) error {
		if c.destroyErr != nil {
			return c.destroyErr
		}
		c.destroyed = append(c.destroyed, id)
		return nil
	}
	t.Cleanup(func() { activeSuiEnv, activeSuiAddress, destroySite = origEnv, origAddr, origDestroy })
}

func seedProject(t *testing.T, objectID string) (int64, string) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	site := filepath.Join(home, "sites", "blog")
	if err := os.MkdirAll(filepath.Join(site, "content"), 0o755); err != nil {
		t.Fatal(err)
	}
	pm, err := projects.NewManager()
	if err != nil {
		t.Fatal(err)
	}
	defer pm.Close()
	p := &projects.Project{
		Name: "blog", Network: "testnet", ObjectID: objectID,
		WalletAddr: "0xabc", SitePath: site, Status: "active",
	}
	if err := pm.CreateProject(p); err != nil {
		t.Fatal(err)
	}
	return p.ID, site
}

func projectExists(t *testing.T, id int64) bool {
	t.Helper()
	pm, err := projects.NewManager()
	if err != nil {
		t.Fatal(err)
	}
	defer pm.Close()
	_, err = pm.GetProject(id)
	return err == nil
}

func dirExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func TestDeleteProject_KeepsFolderByDefault(t *testing.T) {
	chain := &fakeChain{env: "testnet", addr: "0xabc"}
	installFakeChain(t, chain)
	id, site := seedProject(t, "0x1234567890")

	res := DeleteProject(DeleteProjectParams{ProjectID: id})

	if !res.Success || !res.OnChainDestroyed {
		t.Fatalf("expected success with on-chain destroy, got %+v", res)
	}
	if projectExists(t, id) {
		t.Error("project record should be gone")
	}
	if !dirExists(site) {
		t.Error("site folder must be kept unless DeleteSiteFolder is set")
	}
	if res.SiteFolderDeleted || !strings.Contains(res.Message, "Kept the site folder") {
		t.Errorf("message should say the folder was kept: %q", res.Message)
	}
}

func TestDeleteProject_DeletesFolderWhenAsked(t *testing.T) {
	installFakeChain(t, &fakeChain{env: "testnet", addr: "0xabc"})
	id, site := seedProject(t, "0x1234567890")

	res := DeleteProject(DeleteProjectParams{ProjectID: id, DeleteSiteFolder: true})

	if !res.Success || !res.SiteFolderDeleted {
		t.Fatalf("expected folder deletion, got %+v", res)
	}
	if dirExists(site) {
		t.Error("site folder should be deleted")
	}
}

func TestDeleteProject_DestroyFailureKeepsEverything(t *testing.T) {
	installFakeChain(t, &fakeChain{env: "testnet", addr: "0xabc", destroyErr: errors.New("insufficient gas")})
	id, site := seedProject(t, "0x1234567890")

	res := DeleteProject(DeleteProjectParams{ProjectID: id, DeleteSiteFolder: true})

	if res.Success || !res.OnChainFailed || !strings.Contains(res.Error, "insufficient gas") {
		t.Fatalf("expected a surfaced destroy failure, got %+v", res)
	}
	if !projectExists(t, id) || !dirExists(site) {
		t.Error("a failed destroy must leave the record and folder in place")
	}
}

func TestDeleteProject_WrongNetworkOrAddressRefusesDestroy(t *testing.T) {
	for name, chain := range map[string]*fakeChain{
		"network": {env: "mainnet", addr: "0xabc"},
		"address": {env: "testnet", addr: "0xdef"},
	} {
		t.Run(name, func(t *testing.T) {
			installFakeChain(t, chain)
			id, site := seedProject(t, "0x1234567890")

			res := DeleteProject(DeleteProjectParams{ProjectID: id, DeleteSiteFolder: true})

			if res.Success || !res.OnChainFailed {
				t.Fatalf("expected refusal, got %+v", res)
			}
			if len(chain.destroyed) != 0 {
				t.Error("destroy must not run against the wrong wallet")
			}
			if !projectExists(t, id) || !dirExists(site) {
				t.Error("refusal must leave the record and folder in place")
			}
		})
	}
}

func TestDeleteProject_UnsafeFolderRefusedBeforeDestroy(t *testing.T) {
	chain := &fakeChain{env: "testnet", addr: "0xabc"}
	installFakeChain(t, chain)
	t.Setenv("HOME", t.TempDir())
	home, _ := os.UserHomeDir()
	pm, err := projects.NewManager()
	if err != nil {
		t.Fatal(err)
	}
	p := &projects.Project{Name: "home-site", Network: "testnet", ObjectID: "0x1234567890", WalletAddr: "0xabc", SitePath: home}
	if err := pm.CreateProject(p); err != nil {
		t.Fatal(err)
	}
	pm.Close()

	res := DeleteProject(DeleteProjectParams{ProjectID: p.ID, DeleteSiteFolder: true})

	if res.Success || len(chain.destroyed) != 0 {
		t.Fatalf("an unremovable folder must stop the delete before the destroy: %+v", res)
	}
	if !projectExists(t, p.ID) || !dirExists(home) {
		t.Error("record and home must remain")
	}
}

// Rows from earlier releases, drafts and CLI deploys outside home alike, are
// the sandbox roots, so existing projects stay editable after upgrading.
func TestProjectSitePaths_IncludesEveryExistingRow(t *testing.T) {
	_, deployed := seedProject(t, "0x1234567890")
	pm, err := projects.NewManager()
	if err != nil {
		t.Fatal(err)
	}
	draft := filepath.Join(t.TempDir(), "draft")
	if err := pm.CreateDraftProject("draft", draft); err != nil {
		t.Fatal(err)
	}
	pm.Close()

	paths, err := ProjectSitePaths()
	if err != nil {
		t.Fatal(err)
	}
	got := strings.Join(paths, "\n")
	for _, want := range []string{deployed, draft} {
		if !strings.Contains(got, want) {
			t.Errorf("ProjectSitePaths missing %s: %v", want, paths)
		}
	}
}

func TestDeleteProject_SkipOnChainRemovesLocally(t *testing.T) {
	chain := &fakeChain{env: "mainnet", addr: "0xdef"}
	installFakeChain(t, chain)
	id, site := seedProject(t, "0x1234567890")

	res := DeleteProject(DeleteProjectParams{ProjectID: id, SkipOnChain: true})

	if !res.Success || res.OnChainDestroyed || len(chain.destroyed) != 0 {
		t.Fatalf("expected local-only removal, got %+v", res)
	}
	if projectExists(t, id) || !dirExists(site) {
		t.Error("record should be gone and folder kept")
	}
	if !strings.Contains(res.Message, "still exists on testnet") {
		t.Errorf("message should warn the site is still on-chain: %q", res.Message)
	}
}
