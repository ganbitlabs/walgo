package cmd

import (
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestDeployToolchainGate(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake tools are shell scripts")
	}
	bin := t.TempDir()
	versions := map[string]string{"sui": "1.75.0", "walrus": "1.52.0", "site-builder": "2.11.0"}
	for tool, v := range versions {
		if err := os.WriteFile(filepath.Join(bin, tool), []byte("#!/bin/sh\necho \""+tool+" "+v+"\"\n"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", bin)
	t.Setenv("HOME", t.TempDir())

	if err := deployToolchainGate(false, false); err == nil {
		t.Fatal("outdated site-builder should block a real deploy")
	}

	stderr := captureStderr(t, func() {
		if err := deployToolchainGate(true, false); err != nil {
			t.Fatalf("--skip-version-check should not block: %v", err)
		}
	})
	for _, want := range []string{"--skip-version-check", "site-builder v2.11.0"} {
		if !strings.Contains(stderr, want) {
			t.Errorf("skip warning missing %q:\n%s", want, stderr)
		}
	}

	if err := deployToolchainGate(false, true); err != nil {
		t.Fatalf("dry run should only warn: %v", err)
	}
}

func captureStderr(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	original := os.Stderr
	os.Stderr = w
	fn()
	os.Stderr = original
	_ = w.Close()
	out, _ := io.ReadAll(r)
	return string(out)
}
