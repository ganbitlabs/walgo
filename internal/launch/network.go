package launch

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ganbitlabs/walgo/internal/config"
	"github.com/ganbitlabs/walgo/internal/projects"
	"github.com/ganbitlabs/walgo/internal/sui"
	"github.com/ganbitlabs/walgo/internal/ui"
	"gopkg.in/yaml.v3"
)

const placeholderProjectID = "YOUR_WALRUS_PROJECT_ID"

// normalizeNetwork reduces a Sui environment name to "mainnet" or "testnet",
// or "" when it is neither.
func normalizeNetwork(env string) string {
	env = strings.ToLower(strings.TrimSpace(env))
	switch {
	case strings.Contains(env, "mainnet"):
		return "mainnet"
	case strings.Contains(env, "testnet"):
		return "testnet"
	default:
		return ""
	}
}

// parseNetworkChoice maps the network prompt's answer to a network. Empty input
// takes defaultNetwork. ok is false for anything unrecognized.
func parseNetworkChoice(input, defaultNetwork string) (network string, ok bool) {
	switch strings.ToLower(strings.TrimSpace(input)) {
	case "":
		return defaultNetwork, true
	case "1", "testnet":
		return "testnet", true
	case "2", "mainnet":
		return "mainnet", true
	default:
		return "", false
	}
}

// SiteAction is what a launch does with the site walgo.yaml already records.
type SiteAction int

const (
	// SiteNew publishes a new site because none is recorded.
	SiteNew SiteAction = iota
	// SiteUpdate updates the recorded site, which lives on the chosen network.
	SiteUpdate
	// SiteOnOtherNetwork means the recorded site lives on a different network,
	// where its object ID does not exist.
	SiteOnOtherNetwork
)

// classifySite decides whether the recorded projectID can be updated on the
// chosen network. An unknown recorded network keeps the update, as before.
func classifySite(projectID, recordedNetwork, chosen string) SiteAction {
	if projectID == "" || projectID == placeholderProjectID {
		return SiteNew
	}
	if recordedNetwork != "" && recordedNetwork != chosen {
		return SiteOnOtherNetwork
	}
	return SiteUpdate
}

// recordedSiteNetwork returns the network the projectID was deployed to. The
// projects database row that holds this exact object ID wins, because it was
// written at deploy time. walgo.yaml's walrus.network comes next.
func recordedSiteNetwork(projectID, cfgNetwork string, dbProject *projects.Project) string {
	if dbProject != nil && dbProject.ObjectID == projectID {
		if n := normalizeNetwork(dbProject.Network); n != "" {
			return n
		}
	}
	return normalizeNetwork(cfgNetwork)
}

// SelectNetwork prompts for a network, defaulting to the active Sui environment
// and re-prompting on invalid input.
func SelectNetwork() (string, error) {
	icons := ui.GetIcons()

	activeEnv, err := sui.GetActiveEnv()
	defaultNetwork := normalizeNetwork(activeEnv)
	if err != nil || defaultNetwork == "" {
		defaultNetwork = "testnet"
	}
	if err == nil {
		fmt.Printf("Current network: %s\n\n", strings.TrimSpace(activeEnv))
	}

	fmt.Println("Available networks:")
	fmt.Println("  1) testnet  - For testing (1 epoch = 1 day)")
	fmt.Println("  2) mainnet  - For production (1 epoch = 2 weeks, requires SuiNS)")

	for {
		input := readlineInputWithDefault("\nSelect network", defaultNetwork)
		if network, ok := parseNetworkChoice(input, defaultNetwork); ok {
			return network, nil
		}
		fmt.Printf("%s Invalid choice %q, enter 1 (testnet) or 2 (mainnet)\n", icons.Warning, input)
	}
}

// UseNetwork makes network the active Sui environment for the rest of the
// launch, since walrus and site-builder follow it. restore puts back the
// environment that was active before.
func UseNetwork(network string) (restore func(), err error) {
	previous, prevErr := sui.GetActiveEnv()
	previous = strings.TrimSpace(previous)
	if prevErr == nil && normalizeNetwork(previous) == network {
		return func() {}, nil
	}
	if err := sui.SwitchEnv(network); err != nil {
		return nil, fmt.Errorf("failed to switch to %s: %w", network, err)
	}
	if prevErr != nil || previous == "" {
		return func() {}, nil
	}
	return func() {
		icons := ui.GetIcons()
		if err := sui.SwitchEnv(previous); err != nil {
			fmt.Fprintf(os.Stderr, "%s Warning: failed to switch Sui back to %s: %v\n", icons.Warning, previous, err)
			return
		}
		fmt.Printf("%s Sui environment switched back to %s\n", icons.Info, previous)
	}, nil
}

// ReconcileExistingSite checks the site walgo.yaml records against the chosen
// network. When that site lives on another network it asks whether to publish
// a new site instead, and returns an error if the user declines.
func ReconcileExistingSite(sitePath string, cfg *config.WalgoConfig, network string) (SiteAction, error) {
	projectID := strings.TrimSpace(cfg.WalrusConfig.ProjectID)

	var dbProject *projects.Project
	if pm, err := projects.NewManager(); err == nil {
		dbProject, _ = pm.GetProjectBySitePath(sitePath)
		_ = pm.Close()
	}

	recorded := recordedSiteNetwork(projectID, cfg.WalrusConfig.Network, dbProject)
	action := classifySite(projectID, recorded, network)
	if action != SiteOnOtherNetwork {
		return action, nil
	}

	icons := ui.GetIcons()
	fmt.Printf("\n%s walgo.yaml points at a site on %s (%s).\n", icons.Warning, recorded, projectID)
	fmt.Printf("   That object does not exist on %s, so it cannot be updated there.\n", network)
	fmt.Printf("   Launch can publish a new site on %s. walgo.yaml will then point at the new site,\n", network)
	fmt.Printf("   so note the %s object ID above if you still need it.\n", recorded)

	if !Confirm(fmt.Sprintf("\nPublish a new site on %s? [y/N]: ", network), false) {
		return action, fmt.Errorf("cancelled: choose %s to update the existing site", recorded)
	}
	return action, nil
}

// PersistNetwork records network as walrus.network in sitePath/walgo.yaml,
// keeping the rest of the file, comments included, as it is.
func PersistNetwork(sitePath, network string) error {
	path := filepath.Join(sitePath, config.DefaultConfigFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read walgo.yaml: %w", err)
	}
	updated, err := setWalrusNetwork(data, network)
	if err != nil {
		return err
	}
	if bytes.Equal(updated, data) {
		return nil
	}
	// #nosec G306 - config file needs to be readable
	if err := os.WriteFile(path, updated, 0644); err != nil {
		return fmt.Errorf("failed to write walgo.yaml: %w", err)
	}
	return nil
}

func setWalrusNetwork(data []byte, network string) ([]byte, error) {
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("failed to parse walgo.yaml: %w", err)
	}
	if len(doc.Content) == 0 {
		doc = yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{{Kind: yaml.MappingNode}}}
	}
	root := doc.Content[0]
	if root.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("walgo.yaml is not a mapping")
	}

	walrus := mappingValue(root, "walrus")
	if walrus == nil {
		walrus = &yaml.Node{Kind: yaml.MappingNode}
		root.Content = append(root.Content, scalar("walrus"), walrus)
	}
	if walrus.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("walgo.yaml: walrus is not a mapping")
	}

	if value := mappingValue(walrus, "network"); value != nil {
		if value.Value == network {
			return data, nil
		}
		value.Kind, value.Tag, value.Value = yaml.ScalarNode, "!!str", network
	} else {
		walrus.Content = append(walrus.Content, scalar("network"), scalar(network))
	}

	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(&doc); err != nil {
		return nil, fmt.Errorf("failed to marshal walgo.yaml: %w", err)
	}
	if err := enc.Close(); err != nil {
		return nil, fmt.Errorf("failed to marshal walgo.yaml: %w", err)
	}
	return buf.Bytes(), nil
}

func mappingValue(m *yaml.Node, key string) *yaml.Node {
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == key {
			return m.Content[i+1]
		}
	}
	return nil
}

func scalar(v string) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: v}
}
