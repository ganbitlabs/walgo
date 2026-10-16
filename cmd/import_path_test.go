package cmd

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/ganbitlabs/walgo/internal/projects"
)

func TestImportRecordsAbsoluteSitePathForRelativeParentDir(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake hugo is a shell script")
	}
	t.Setenv("HOME", t.TempDir())
	bin := t.TempDir()
	hugo := "#!/bin/sh\n" +
		"if [ \"$1\" = new ]; then mkdir -p content archetypes && echo \"title = 'x'\" > hugo.toml; exit 0; fi\n" +
		"mkdir -p public && echo '<h1>hi</h1>' > public/index.html\n"
	writeTestFile(t, filepath.Join(bin, "hugo"), hugo)
	if err := os.Chmod(filepath.Join(bin, "hugo"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+"/bin:/usr/bin")

	vault := t.TempDir()
	writeTestFile(t, filepath.Join(vault, "note.md"), "# Note\n")

	work := t.TempDir()
	t.Chdir(work)
	if err := os.Mkdir("sites", 0o755); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_ = importCmd.Flags().Set("parent-dir", "")
		_ = importCmd.Flags().Set("site-name", "")
	})
	if _, err := executeCommand(rootCmd, "import", vault, "--parent-dir", "sites", "--site-name", "notes"); err != nil {
		t.Fatalf("import failed: %v", err)
	}

	pm, err := projects.NewManager()
	if err != nil {
		t.Fatal(err)
	}
	defer pm.Close()
	proj, err := pm.GetProjectByName("notes")
	if err != nil {
		t.Fatal(err)
	}
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(wd, "sites", "notes")
	if proj.SitePath != want {
		t.Errorf("recorded SitePath = %q, want the absolute %q", proj.SitePath, want)
	}
}
