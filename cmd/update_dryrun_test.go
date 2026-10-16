package cmd

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestUpdateDryRunNeverRunsSiteBuilderWhenAnalysisFails(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake tools are shell scripts")
	}
	bin := t.TempDir()
	marker := filepath.Join(t.TempDir(), "site-builder-calls")
	tools := map[string]string{
		"sui":          `if [ "$1" = --version ]; then echo "sui 1.75.0"; else echo testnet; fi`,
		"walrus":       `echo "walrus 1.52.0"`,
		"site-builder": `if [ "$1" = --version ]; then echo "site-builder 2.12.0"; else echo "$*" >> "` + marker + `"; fi`,
	}
	for name, body := range tools {
		if err := os.WriteFile(filepath.Join(bin, name), []byte("#!/bin/sh\n"+body+"\n"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", bin)
	t.Setenv("HOME", t.TempDir())

	sitePath := t.TempDir()
	originalWd, _ := os.Getwd()
	if err := os.Chdir(sitePath); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(originalWd) }()

	if err := os.WriteFile("walgo.yaml", []byte("walrus:\n  network: testnet\nhugo:\n  publishDir: public\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	publicDir := filepath.Join(sitePath, "public")
	if err := os.MkdirAll(publicDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// The cache opens, but an unreadable publish dir makes the change analysis fail.
	if err := os.Chmod(publicDir, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(publicDir, 0o755) })

	objectID := "0x" + "3333333333333333333333333333333333333333333333333333333333333333"
	if _, err := executeCommand(rootCmd, "update", objectID, "--dry-run"); err != nil {
		t.Fatalf("dry run returned error: %v", err)
	}
	if calls, err := os.ReadFile(marker); err == nil {
		t.Fatalf("dry run invoked site-builder:\n%s", calls)
	}
}
