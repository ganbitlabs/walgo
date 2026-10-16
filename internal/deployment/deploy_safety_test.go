package deployment

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/ganbitlabs/walgo/internal/cache"
	"github.com/ganbitlabs/walgo/internal/config"
)

// fakeSiteBuilder puts sui, walrus and site-builder on PATH and returns a file
// that exists only if any of them ran.
func fakeSiteBuilder(t *testing.T) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("fake tools are shell scripts")
	}
	bin := t.TempDir()
	marker := filepath.Join(t.TempDir(), "invoked")
	for _, tool := range []string{"sui", "walrus", "site-builder"} {
		script := "#!/bin/sh\necho \"" + tool + " $*\" >> \"" + marker + "\"\necho testnet\n"
		if err := os.WriteFile(filepath.Join(bin, tool), []byte(script), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", bin)
	t.Setenv("HOME", t.TempDir())
	return marker
}

func TestDeployReportsSuccessWhenBookkeepingFails(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake tools are shell scripts")
	}
	const objectID = "0x2222222222222222222222222222222222222222222222222222222222222222"
	bin := t.TempDir()
	tools := map[string]string{
		// No active address, so no gas lookup goes to the network.
		"sui":    "if [ \"$2\" = active-env ]; then echo testnet; else exit 1; fi\n",
		"walrus": "exit 0\n",
		// Prints nothing parseable; records the new site the way site-builder does.
		"site-builder": "printf '{\"object_id\":\"" + objectID + "\"}' > \"$6/ws-resources.json\"\n",
	}
	for name, body := range tools {
		if err := os.WriteFile(filepath.Join(bin, name), []byte("#!/bin/sh\n"+body), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", bin)
	t.Setenv("HOME", t.TempDir())

	// No walgo.yaml, so recording the project ID fails after the publish.
	sitePath := t.TempDir()
	publishDir := filepath.Join(sitePath, "public")
	// An unreadable subdirectory keeps the cost estimator off the network.
	locked := filepath.Join(publishDir, "locked")
	if err := os.MkdirAll(locked, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(publishDir, "ws-resources.json"), []byte(`{}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(locked, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(locked, 0o755) })

	result, err := PerformDeployment(context.Background(), DeploymentOptions{
		SitePath:   sitePath,
		PublishDir: publishDir,
		Epochs:     1,
		WalgoCfg:   &config.WalgoConfig{},
		Network:    "testnet",
		Quiet:      true,
	})
	if err != nil {
		t.Fatalf("on-chain success reported as failure: %v", err)
	}
	if !result.Success || result.ObjectID != objectID {
		t.Errorf("result = %+v, want success with object %s", result, objectID)
	}
}

func TestDryRunNeverInvokesTools(t *testing.T) {
	tests := []struct {
		name  string
		setup func(t *testing.T, sitePath, publishDir string)
	}{
		{
			name: "cache unavailable",
			setup: func(t *testing.T, sitePath, publishDir string) {
				// A file where the cache directory belongs makes the cache fail to open.
				if err := os.WriteFile(filepath.Join(sitePath, cache.CacheDir), nil, 0o644); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "change analysis fails",
			setup: func(t *testing.T, sitePath, publishDir string) {
				if err := os.Chmod(publishDir, 0o000); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = os.Chmod(publishDir, 0o755) })
			},
		},
		{
			name:  "cache healthy",
			setup: func(t *testing.T, sitePath, publishDir string) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			marker := fakeSiteBuilder(t)

			sitePath := t.TempDir()
			publishDir := filepath.Join(sitePath, "public")
			if err := os.MkdirAll(publishDir, 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(publishDir, "index.html"), []byte("<h1>hi</h1>"), 0o644); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(publishDir, "ws-resources.json"), []byte(`{}`), 0o644); err != nil {
				t.Fatal(err)
			}
			tt.setup(t, sitePath, publishDir)

			result, err := PerformDeployment(context.Background(), DeploymentOptions{
				SitePath:   sitePath,
				PublishDir: publishDir,
				Epochs:     1,
				WalgoCfg:   &config.WalgoConfig{},
				Network:    "testnet",
				DryRun:     true,
			})
			if err != nil {
				t.Fatalf("dry run returned error: %v", err)
			}
			if !result.Success || result.ObjectID != "" {
				t.Errorf("dry run result = %+v, want success with no object", result)
			}
			if calls, err := os.ReadFile(marker); err == nil {
				t.Fatalf("dry run invoked external tools:\n%s", calls)
			}
		})
	}
}
