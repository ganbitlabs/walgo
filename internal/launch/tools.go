package launch

import (
	"fmt"

	"github.com/ganbitlabs/walgo/internal/ui"
	"github.com/ganbitlabs/walgo/internal/version"
)

// CheckToolVersions offers to update sui, walrus and site-builder before a
// mainnet deploy. It asks through Confirm, so a closed stdin never starts an
// update. Testnet deploys skip the check.
func CheckToolVersions(network string) error {
	if network != "mainnet" {
		return nil
	}
	icons := ui.GetIcons()
	fmt.Printf("  %s Checking tool versions for mainnet deployment...\n", icons.Info)

	result, err := version.CheckAllVersions()
	if err != nil {
		return fmt.Errorf("failed to check versions: %w", err)
	}
	outdated := outdatedTools(result)
	if len(outdated) == 0 {
		fmt.Printf("  %s All tools are up to date\n", icons.Check)
		return nil
	}

	fmt.Printf("\n%s Updates available for mainnet deployment:\n\n", icons.Warning)
	for _, t := range outdated {
		fmt.Printf("  • %s: %s → %s\n", t.Tool, t.CurrentVersion, t.LatestVersion)
	}
	fmt.Printf("\n%s For mainnet deployments, it's recommended to use the latest versions.\n", icons.Lightbulb)

	if !Confirm("\nWould you like to update now? [Y/n]: ", true) {
		fmt.Printf("\n%s Continuing with current versions...\n\n", icons.Warning)
		return nil
	}
	fmt.Println("\nUpdating tools...")
	if err := version.UpdateAllTools(network); err != nil {
		return fmt.Errorf("failed to update tools: %w", err)
	}
	fmt.Printf("\n%s All tools updated successfully!\n\n", icons.Check)
	return nil
}

func outdatedTools(result *version.CheckResult) []*version.ToolVersion {
	var out []*version.ToolVersion
	for _, t := range []*version.ToolVersion{result.Sui, result.Walrus, result.SiteBuilder} {
		if t != nil && t.UpdateRequired {
			out = append(out, t)
		}
	}
	return out
}
