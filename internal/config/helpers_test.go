package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUpdateWalgoYAMLProjectIDPreservesCommentsAndOrder(t *testing.T) {
	dir := t.TempDir()
	in := `# top comment
zeta: 1
walrus:
  # network note
  network: testnet # inline
  projectID: old
alpha: 2
`
	if err := os.WriteFile(filepath.Join(dir, "walgo.yaml"), []byte(in), 0644); err != nil {
		t.Fatal(err)
	}
	if err := UpdateWalgoYAMLProjectID(dir, "0xabc"); err != nil {
		t.Fatal(err)
	}
	out, _ := os.ReadFile(filepath.Join(dir, "walgo.yaml"))
	got := string(out)
	for _, want := range []string{"# top comment", "# network note", "# inline", "projectID: \"0xabc\""} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
	if !(strings.Index(got, "zeta") < strings.Index(got, "walrus") && strings.Index(got, "walrus") < strings.Index(got, "alpha")) {
		t.Errorf("key order changed:\n%s", got)
	}
}

func TestUpdateWalgoYAMLProjectIDCreatesKeys(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "walgo.yaml"), []byte("# c\nname: x\n"), 0644)
	if err := UpdateWalgoYAMLProjectID(dir, "0x1"); err != nil {
		t.Fatal(err)
	}
	out, _ := os.ReadFile(filepath.Join(dir, "walgo.yaml"))
	got := string(out)
	if !strings.Contains(got, "# c") || !strings.Contains(got, "walrus:") || !strings.Contains(got, "projectID: \"0x1\"") {
		t.Errorf("unexpected output:\n%s", got)
	}
}

func TestUpdateWalgoYAMLProjectIDEmptyFile(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "walgo.yaml"), nil, 0644)
	if err := UpdateWalgoYAMLProjectID(dir, "0x1"); err != nil {
		t.Fatal(err)
	}
	out, _ := os.ReadFile(filepath.Join(dir, "walgo.yaml"))
	if !strings.Contains(string(out), "projectID: \"0x1\"") {
		t.Errorf("unexpected output:\n%s", out)
	}
}
