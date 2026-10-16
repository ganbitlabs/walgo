package cmd

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/ganbitlabs/walgo/internal/projects"
)

func TestSiteNetworkPrefersRecordedNetworkOverActiveEnv(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake sui is a shell script")
	}
	bin := t.TempDir()
	if err := os.WriteFile(filepath.Join(bin, "sui"), []byte("#!/bin/sh\necho mainnet\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin)
	t.Setenv("HOME", t.TempDir())
	sitePath := t.TempDir()

	got, err := siteNetwork(sitePath, "")
	if err != nil || got != "mainnet" {
		t.Fatalf("unrecorded site: got %q, %v; want the active env mainnet", got, err)
	}

	got, err = siteNetwork(sitePath, "testnet")
	if err != nil || got != "testnet" {
		t.Fatalf("configured site: got %q, %v; want walgo.yaml's testnet", got, err)
	}

	pm, err := projects.NewManager()
	if err != nil {
		t.Fatal(err)
	}
	err = pm.CreateProject(&projects.Project{Name: "site", Network: "testnet", SitePath: sitePath, ObjectID: "0x1"})
	pm.Close()
	if err != nil {
		t.Fatal(err)
	}

	got, err = siteNetwork(sitePath, "")
	if err != nil || got != "testnet" {
		t.Fatalf("recorded site: got %q, %v; want its recorded network testnet", got, err)
	}
}
