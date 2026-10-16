package walrus

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ganbitlabs/walgo/internal/compress"
	"github.com/ganbitlabs/walgo/internal/ui"
)

// DeploySite publishes deployDir as a new Walrus Site with `site-builder publish`.
// network prices the cost estimate; site-builder itself follows the active Sui
// environment. An empty network falls back to that environment.
func DeploySite(ctx context.Context, deployDir, network string, epochs int) (*SiteBuilderOutput, error) {
	if epochs <= 0 {
		return nil, fmt.Errorf("epochs must be greater than 0, got %d", epochs)
	}

	icons := ui.GetIcons()

	printCostEstimate(deployDir, network, epochs, false)

	builderPath, err := execLookPath(siteBuilderCmd)
	if err != nil {
		return nil, fmt.Errorf("'%s' CLI not found in PATH", siteBuilderCmd)
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
		"publish",
		deployDir,
		"--epochs", fmt.Sprintf("%d", epochs),
	}

	if isVerbose() {
		fmt.Printf("%s Verbose mode enabled\n", icons.Wrench)
		fmt.Printf("%s Builder path: %s\n", icons.Wrench, builderPath)
		fmt.Printf("%s Arguments: %v\n", icons.Wrench, args)
	}

	wsResourcesPath := filepath.Join(deployDir, "ws-resources.json")
	priorObjectID := readWSObjectID(wsResourcesPath)

	fmt.Printf("%s Executing: %s %s\n", icons.Info, builderPath, strings.Join(args, " "))
	fmt.Printf("%s Uploading site files to Walrus...\n", icons.Upload)
	fmt.Printf("%s This may take several minutes depending on file count and network conditions...\n", icons.Hourglass)
	fmt.Printf("   (timeout: %v)\n", commandTimeout(ctx))
	fmt.Println()

	stdoutStr, stderrStr, err := runCommandWithTimeout(ctx, builderPath, args, true)
	if err != nil {
		// Build detailed error with full command and output for debugging
		debugInfo := fmt.Sprintf("\n\nCommand: %s %s\nBuilder: %s\nWalrus: %s\nContext: %s",
			builderPath, strings.Join(args, " "), builderPath, walrusPath, siteBuilderContext)
		if stdoutStr != "" {
			debugInfo += fmt.Sprintf("\nStdout: %s", strings.TrimSpace(stdoutStr))
		}
		if stderrStr != "" {
			debugInfo += fmt.Sprintf("\nStderr: %s", strings.TrimSpace(stderrStr))
		}
		return nil, fmt.Errorf("deployment failed: %w%s", err, debugInfo)
	}

	fmt.Printf("\n%s Site deployment command executed successfully.\n", icons.Success)

	combinedOutput := stdoutStr + "\n" + stderrStr
	output := parseSiteBuilderOutput(combinedOutput)
	output.Success = true

	// site-builder records the new site's object_id in ws-resources.json, so a
	// change in its output format must not turn a paid, successful publish into
	// a reported failure. An ID that was already there before publishing is
	// stale and says nothing about this run.
	if output.ObjectID == "" {
		if id := readWSObjectID(wsResourcesPath); id != "" && id != priorObjectID {
			output.ObjectID = id
		}
	}

	if output.ObjectID != "" {
		fmt.Printf("\n%s Deployment successful!\n", icons.Celebrate)
		fmt.Printf("%s Site Object ID: %s\n", icons.Clipboard, output.ObjectID)
		fmt.Printf("\n%s Next steps:\n", icons.Pencil)
		fmt.Printf("1. Save this Object ID in your walgo.yaml:\n")
		fmt.Printf("   walrus:\n")
		fmt.Printf("     projectID: \"%s\"\n", output.ObjectID)
		fmt.Printf("2. Configure a SuiNS domain: walgo domain <your-domain>\n")
		fmt.Printf("3. Update your site: walgo update\n")
		fmt.Printf("4. Check status: walgo status\n")

		if len(output.BrowseURLs) > 0 {
			fmt.Printf("\n%s Browse your site:\n", icons.Globe)
			for _, url := range output.BrowseURLs {
				fmt.Printf("   %s\n", url)
			}
		}
	}

	return output, nil
}

func readWSObjectID(path string) string {
	cfg, err := compress.ReadWSResourcesConfig(path)
	if err != nil {
		return ""
	}
	return cfg.ObjectID
}

// printCostEstimate prints walgo's own estimate of what storing deployDir costs.
// For an update it is an upper bound, since site-builder reuses unchanged files.
func printCostEstimate(deployDir, network string, epochs int, isUpdate bool) {
	icons := ui.GetIcons()

	var fileCount int
	var totalSize int64
	err := filepath.Walk(deployDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			fileCount++
			totalSize += info.Size()
		}
		return nil
	})
	if err != nil {
		fmt.Printf("   %s Cost estimate unavailable: could not analyze %s: %v\n", icons.Warning, deployDir, err)
		return
	}

	if network == "" {
		network = GetWalrusContext()
	}
	fmt.Printf("%s Estimating %s costs...\n", icons.Chart, network)

	breakdown, err := CalculateCost(CostOptions{
		SiteSize:  totalSize,
		Epochs:    epochs,
		Network:   network,
		FileCount: fileCount,
	})
	if err != nil {
		fmt.Printf("   %s Cost estimate unavailable: %v\n", icons.Warning, err)
		return
	}

	fmt.Printf("%s\n", FormatCostBreakdown(*breakdown))
	fmt.Printf("\n%s Summary: %s\n", icons.Info,
		FormatCostSummary(breakdown.TotalWAL, breakdown.GasCostSUI, breakdown.FileCount, epochs))
	if isUpdate {
		fmt.Printf("   Upper bound: assumes every file is uploaded again; unchanged files are reused.\n")
	}
	fmt.Println()
}
