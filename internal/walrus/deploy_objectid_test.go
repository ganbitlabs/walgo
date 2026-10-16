package walrus

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

const freshObjectID = "0x1111111111111111111111111111111111111111111111111111111111111111"

// fakeDeployTools puts sui, walrus and a site-builder whose publish prints
// nothing parseable on PATH. When newID is set, publish writes it into the
// deploy dir's ws-resources.json the way site-builder records a new site.
func fakeDeployTools(t *testing.T, newID string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("fake tools are shell scripts")
	}
	bin := t.TempDir()
	write := func(name, body string) {
		if err := os.WriteFile(filepath.Join(bin, name), []byte("#!/bin/sh\n"+body), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	write("sui", "echo testnet\n")
	write("walrus", "exit 0\n")
	sb := "echo 'Site published'\n"
	if newID != "" {
		// args: --context C --walrus-binary W publish DIR --epochs N
		sb += "printf '{\"object_id\":\"" + newID + "\"}' > \"$6/ws-resources.json\"\n"
	}
	write("site-builder", sb)
	t.Setenv("PATH", bin)
	t.Setenv("HOME", t.TempDir())
}

// siteWithoutEstimate builds a deploy dir the cost estimator cannot walk, so the
// test never reaches the network for gas prices.
func siteWithoutEstimate(t *testing.T, wsResources string) string {
	t.Helper()
	dir := t.TempDir()
	if wsResources != "" {
		if err := os.WriteFile(filepath.Join(dir, "ws-resources.json"), []byte(wsResources), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	locked := filepath.Join(dir, "locked")
	if err := os.Mkdir(locked, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(locked, 0o755) })
	return dir
}

func TestDeploySiteFallsBackToWSResourcesObjectID(t *testing.T) {
	fakeDeployTools(t, freshObjectID)
	dir := siteWithoutEstimate(t, "")

	out, err := DeploySite(context.Background(), dir, "testnet", 1)
	if err != nil {
		t.Fatalf("DeploySite: %v", err)
	}
	if out.ObjectID != freshObjectID {
		t.Errorf("ObjectID = %q, want %q from ws-resources.json", out.ObjectID, freshObjectID)
	}
}

func TestDeploySiteIgnoresStaleWSResourcesObjectID(t *testing.T) {
	fakeDeployTools(t, "")
	dir := siteWithoutEstimate(t, `{"object_id":"`+freshObjectID+`"}`)

	out, err := DeploySite(context.Background(), dir, "testnet", 1)
	if err != nil {
		t.Fatalf("DeploySite: %v", err)
	}
	if out.ObjectID != "" {
		t.Errorf("ObjectID = %q, want empty: the ID predates this publish", out.ObjectID)
	}
}

func TestRunCommandTimeoutReportsCallerDeadline(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses sleep")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, _, err := runCommandWithTimeout(ctx, "sleep", []string{"10"}, false)
	if err == nil || !strings.Contains(err.Error(), "timed out after 2s") {
		t.Fatalf("want the caller's 2s deadline in the error, got %v", err)
	}
}
