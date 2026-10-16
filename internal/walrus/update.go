package walrus

import (
	"context"
	"fmt"
	"strings"

	"github.com/ganbitlabs/walgo/internal/ui"
)

// UpdateSite updates an existing site on Walrus with `site-builder update`.
// network prices the cost estimate, as in DeploySite.
func UpdateSite(ctx context.Context, deployDir, objectID, network string, epochs int) (*SiteBuilderOutput, error) {
	if err := validateObjectID(objectID); err != nil {
		return nil, fmt.Errorf("invalid object ID: %w", err)
	}

	if epochs <= 0 {
		return nil, fmt.Errorf("epochs must be greater than 0, got %d", epochs)
	}

	if err := CheckSiteBuilderSetup(); err != nil {
		return nil, fmt.Errorf("site-builder setup issue: %w\n\nRun 'walgo setup' to configure site-builder", err)
	}

	builderPath, err := execLookPath(siteBuilderCmd)
	if err != nil {
		return nil, fmt.Errorf("'%s' CLI not found. Please install it and ensure it's in your PATH", siteBuilderCmd)
	}

	// Find walrus binary path to pass to site-builder
	walrusPath, err := execLookPath("walrus")
	if err != nil {
		return nil, fmt.Errorf("'walrus' CLI not found in PATH. Please install it using:\n  suiup install walrus@mainnet\n  Or run: walgo setup-deps")
	}

	siteBuilderContext := GetWalrusContext()
	args := []string{
		"--context", siteBuilderContext,
		"--walrus-binary", walrusPath,
		"update",
		"--epochs", fmt.Sprintf("%d", epochs),
		deployDir,
		objectID,
	}

	icons := ui.GetIcons()
	fmt.Printf("%s Executing: %s %s\n", icons.Info, builderPath, strings.Join(args, " "))
	fmt.Printf("%s Updating site files on Walrus...\n", icons.Upload)

	printCostEstimate(deployDir, network, epochs, true)

	fmt.Printf("%s This may take several minutes depending on file count and network conditions...\n", icons.Hourglass)
	fmt.Printf("   (timeout: %v)\n", commandTimeout(ctx))
	fmt.Println()

	stdoutStr, stderrStr, err := runCommandWithTimeout(ctx, builderPath, args, true)
	if err != nil {
		debugInfo := fmt.Sprintf("\n\nCommand: %s %s\nBuilder: %s\nWalrus: %s\nContext: %s",
			builderPath, strings.Join(args, " "), builderPath, walrusPath, siteBuilderContext)
		if stdoutStr != "" {
			debugInfo += fmt.Sprintf("\nStdout: %s", strings.TrimSpace(stdoutStr))
		}
		if stderrStr != "" {
			debugInfo += fmt.Sprintf("\nStderr: %s", strings.TrimSpace(stderrStr))
		}
		return nil, fmt.Errorf("update failed: %w%s", err, debugInfo)
	}

	fmt.Printf("\n%s Site update command executed successfully.\n", icons.Success)

	combinedOutput := stdoutStr + "\n" + stderrStr
	output := parseSiteBuilderOutput(combinedOutput)
	output.Success = true
	output.ObjectID = objectID

	fmt.Printf("\n%s Site updated successfully!\n", icons.Success)
	fmt.Printf("%s Object ID: %s\n", icons.Clipboard, objectID)
	fmt.Printf("%s Your site should be updated at the same URLs as before\n", icons.Globe)

	return output, nil
}
