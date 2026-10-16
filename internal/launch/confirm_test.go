package launch

import (
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
