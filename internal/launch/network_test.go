package launch

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ganbitlabs/walgo/internal/config"
	"github.com/ganbitlabs/walgo/internal/projects"
	"gopkg.in/yaml.v3"
)

func TestParseNetworkChoice(t *testing.T) {
	tests := []struct {
		input, def, want string
		ok               bool
	}{
		{"", "mainnet", "mainnet", true},
		{"", "testnet", "testnet", true},
		{"1", "mainnet", "testnet", true},
		{"2", "testnet", "mainnet", true},
		{" Mainnet ", "testnet", "mainnet", true},
		{"testnet", "mainnet", "testnet", true},
		{"3", "testnet", "", false},
		{"0", "testnet", "", false},
		{"devnet", "testnet", "", false},
		{"yes", "mainnet", "", false},
	}
	for _, tt := range tests {
		got, ok := parseNetworkChoice(tt.input, tt.def)
		if got != tt.want || ok != tt.ok {
			t.Errorf("parseNetworkChoice(%q, %q) = (%q, %v), want (%q, %v)", tt.input, tt.def, got, ok, tt.want, tt.ok)
		}
	}
}

func TestSelectNetworkRepromptsOnInvalidInput(t *testing.T) {
	defer resetReader()
	setTestReader("3\nfoo\nmainnet\n")
	got, err := SelectNetwork()
	if err != nil {
		t.Fatal(err)
	}
	if got != "mainnet" {
		t.Errorf("SelectNetwork() = %q after invalid answers, want mainnet", got)
	}
}

func TestNormalizeNetwork(t *testing.T) {
	for in, want := range map[string]string{
		"mainnet":       "mainnet",
		" Testnet\n":    "testnet",
		"sui-mainnet-1": "mainnet",
		"devnet":        "",
		"":              "",
	} {
		if got := normalizeNetwork(in); got != want {
			t.Errorf("normalizeNetwork(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestClassifySite(t *testing.T) {
	tests := []struct {
		name, projectID, recorded, chosen string
		want                              SiteAction
	}{
		{"no project", "", "mainnet", "testnet", SiteNew},
		{"same network", "0xabc", "mainnet", "mainnet", SiteUpdate},
		{"unknown network keeps update", "0xabc", "", "testnet", SiteUpdate},
		{"other network", "0xabc", "mainnet", "testnet", SiteOnOtherNetwork},
		{"other network reversed", "0xabc", "testnet", "mainnet", SiteOnOtherNetwork},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := classifySite(tt.projectID, tt.recorded, tt.chosen); got != tt.want {
				t.Errorf("classifySite = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExistingObjectID(t *testing.T) {
	db := &projects.Project{ObjectID: "0xdb"}
	tests := []struct {
		name, projectID, ws string
		db                  *projects.Project
		want                string
	}{
		{"walgo.yaml wins", "0xyaml", "0xws", db, "0xyaml"},
		{"placeholder falls through to ws-resources", placeholderProjectID, "0xws", db, "0xws"},
		{"ws-resources before db", "", "0xws", db, "0xws"},
		{"db last", "", "", db, "0xdb"},
		{"nothing", " ", "", nil, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := existingObjectID(tt.projectID, tt.ws, tt.db); got != tt.want {
				t.Errorf("existingObjectID = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRecordedSiteNetwork(t *testing.T) {
	tests := []struct {
		name       string
		cfgNetwork string
		db         *projects.Project
		want       string
	}{
		{"walgo.yaml only", "mainnet", nil, "mainnet"},
		{"nothing recorded", "", nil, ""},
		{"db row for this object wins", "testnet", &projects.Project{ObjectID: "0xabc", Network: "mainnet"}, "mainnet"},
		{"db row for another object ignored", "testnet", &projects.Project{ObjectID: "0xdef", Network: "mainnet"}, "testnet"},
		{"db row without network falls back", "testnet", &projects.Project{ObjectID: "0xabc"}, "testnet"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := recordedSiteNetwork("0xabc", tt.cfgNetwork, tt.db); got != tt.want {
				t.Errorf("recordedSiteNetwork = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPersistNetwork(t *testing.T) {
	tests := []struct {
		name     string
		before   string
		keepText []string
	}{
		{
			name:     "updates existing key and keeps comments",
			before:   "# my site\nhugo:\n  publishDir: public\nwalrus:\n  projectID: 0xabc # deployed\n  network: testnet\n",
			keepText: []string{"# my site", "# deployed", "projectID: 0xabc", "publishDir: public"},
		},
		{
			name:     "adds key to walrus section",
			before:   "walrus:\n  projectID: 0xabc\n",
			keepText: []string{"projectID: 0xabc"},
		},
		{
			name:     "adds walrus section",
			before:   "hugo:\n  publishDir: public\n",
			keepText: []string{"publishDir: public"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, config.DefaultConfigFileName)
			if err := os.WriteFile(path, []byte(tt.before), 0644); err != nil {
				t.Fatal(err)
			}
			if err := PersistNetwork(dir, "mainnet"); err != nil {
				t.Fatal(err)
			}
			after, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			for _, s := range tt.keepText {
				if !strings.Contains(string(after), s) {
					t.Errorf("walgo.yaml lost %q:\n%s", s, after)
				}
			}
			cfg, err := config.LoadConfigFrom(dir)
			if err != nil {
				t.Fatal(err)
			}
			if cfg.WalrusConfig.Network != "mainnet" {
				t.Errorf("walrus.network = %q, want mainnet:\n%s", cfg.WalrusConfig.Network, after)
			}
		})
	}
}

func TestPersistNetworkLeavesMatchingFileUntouched(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, config.DefaultConfigFileName)
	before := "walrus:\n    network:   mainnet   # chosen\n"
	if err := os.WriteFile(path, []byte(before), 0644); err != nil {
		t.Fatal(err)
	}
	if err := PersistNetwork(dir, "mainnet"); err != nil {
		t.Fatal(err)
	}
	after, _ := os.ReadFile(path)
	if string(after) != before {
		t.Errorf("walgo.yaml rewritten although network already matched:\n%s", after)
	}
}

func TestSetWalrusNetworkRejectsNonMapping(t *testing.T) {
	for _, in := range []string{"- a\n- b\n", "walrus: 3\n"} {
		if _, err := setWalrusNetwork([]byte(in), "mainnet"); err == nil {
			t.Errorf("setWalrusNetwork(%q) succeeded, want error", in)
		}
	}
	var out map[string]any
	got, err := setWalrusNetwork(nil, "testnet")
	if err != nil {
		t.Fatal(err)
	}
	if err := yaml.Unmarshal(got, &out); err != nil || out["walrus"].(map[string]any)["network"] != "testnet" {
		t.Errorf("empty walgo.yaml = %q, want walrus.network testnet", got)
	}
}
