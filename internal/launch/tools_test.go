package launch

import (
	"testing"

	"github.com/ganbitlabs/walgo/internal/version"
)

func TestCheckToolVersionsSkipsTestnet(t *testing.T) {
	if err := CheckToolVersions("testnet"); err != nil {
		t.Errorf("CheckToolVersions(testnet) = %v, want nil", err)
	}
}

func TestOutdatedTools(t *testing.T) {
	result := &version.CheckResult{
		Sui:         &version.ToolVersion{Tool: "sui", UpdateRequired: true},
		Walrus:      &version.ToolVersion{Tool: "walrus"},
		SiteBuilder: &version.ToolVersion{Tool: "site-builder", UpdateRequired: true},
		Hugo:        &version.ToolVersion{Tool: "hugo", UpdateRequired: true},
	}
	got := outdatedTools(result)
	if len(got) != 2 || got[0].Tool != "sui" || got[1].Tool != "site-builder" {
		t.Errorf("outdatedTools = %v, want sui and site-builder", got)
	}
	if got := outdatedTools(&version.CheckResult{}); len(got) != 0 {
		t.Errorf("outdatedTools(empty) = %v, want none", got)
	}
}
