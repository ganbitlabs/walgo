package launch

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ganbitlabs/walgo/internal/config"
)

func TestConfirm(t *testing.T) {
	defer resetReader()
	tests := []struct {
		name       string
		input      string
		defaultYes bool
		want       bool
	}{
		{"enter takes default yes", "\n", true, true},
		{"enter takes default no", "\n", false, false},
		{"y", "y\n", false, true},
		{"YES", " YES \n", false, true},
		{"n", "n\n", true, false},
		{"other answer is no", "sure\n", true, false},
		{"EOF is no even when default is yes", "", true, false},
		{"partial line at EOF is no", "y", true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setTestReader(tt.input)
			if got := Confirm("? ", tt.defaultYes); got != tt.want {
				t.Errorf("Confirm(%q, %v) = %v, want %v", tt.input, tt.defaultYes, got, tt.want)
			}
		})
	}
}

func TestReconcileExistingSiteEOFDeclines(t *testing.T) {
	defer resetReader()
	setTestReader("")
	cfg := &config.WalgoConfig{}
	cfg.WalrusConfig.ProjectID = "0xabc"
	cfg.WalrusConfig.Network = "mainnet"
	if _, err := ReconcileExistingSite(t.TempDir(), cfg, "testnet"); err == nil {
		t.Error("ReconcileExistingSite approved a new site on EOF")
	}
}

func TestReconcileExistingSiteSeesWSResourcesObjectID(t *testing.T) {
	defer resetReader()
	setTestReader("")
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "public"), 0755); err != nil {
		t.Fatal(err)
	}
	ws := []byte(`{"object_id": "0xfeed"}`)
	if err := os.WriteFile(filepath.Join(dir, "public", "ws-resources.json"), ws, 0644); err != nil {
		t.Fatal(err)
	}
	cfg := &config.WalgoConfig{}
	cfg.HugoConfig.PublishDir = "public"
	cfg.WalrusConfig.Network = "mainnet"
	if _, err := ReconcileExistingSite(dir, cfg, "testnet"); err == nil {
		t.Error("ReconcileExistingSite let a mainnet site from ws-resources.json be updated on testnet")
	}
}
