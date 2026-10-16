package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ganbitlabs/walgo/internal/config"
	"github.com/ganbitlabs/walgo/internal/deployer"
	sb "github.com/ganbitlabs/walgo/internal/deployer/sitebuilder"
	"github.com/ganbitlabs/walgo/internal/hugo"
	"github.com/ganbitlabs/walgo/internal/projects"
	"github.com/ganbitlabs/walgo/internal/ui"
	"github.com/ganbitlabs/walgo/internal/version"
	"github.com/ganbitlabs/walgo/internal/walrus"
)

// updateProjectByRef pushes local site changes to Walrus blockchain.
func updateProjectByRef(proj *projects.Project, epochs int) error {
	icons := ui.GetIcons()
	pm, err := projects.NewManager()
	if err != nil {
		return fmt.Errorf("failed to initialize project manager: %w", err)
	}
	defer pm.Close()

	fmt.Println()
	fmt.Printf("%s Updating project on Walrus: %s\n", icons.Rocket, proj.Name)
	fmt.Println()

	if epochs == 0 {
		epochs = proj.Epochs
	} else if epochs != proj.Epochs {
		fmt.Printf("  %s Epochs: %d → %d\n", icons.Info, proj.Epochs, epochs)
	}

	walgoCfg, err := config.LoadConfigFrom(proj.SitePath)
	if err != nil {
		return fmt.Errorf("failed to load config from %s: %w", proj.SitePath, err)
	}

	publishDir := filepath.Join(proj.SitePath, walgoCfg.HugoConfig.PublishDir)

	if _, err := os.Stat(publishDir); os.IsNotExist(err) {
		return fmt.Errorf("site not built - run 'walgo build' first")
	}

	// Same guards as `walgo update`: the project's recorded network must be the
	// one site-builder will actually use, and the toolchain must be able to talk
	// to Sui at all.
	network, err := resolveDeployNetwork(proj.Network)
	if err != nil {
		if proj.Network == "" {
			return err
		}
		return fmt.Errorf("project %q is recorded on %s: %w", proj.Name, proj.Network, err)
	}
	if err := version.RequireDeployToolchain(); err != nil {
		return err
	}
	if proj.Network == "" {
		proj.Network = network
	}

	fmt.Printf("%s Pushing changes to Walrus...\n", icons.Spinner)

	err = hugo.BuildSite(proj.SitePath)
	if err != nil {
		return fmt.Errorf("failed to build site: %w", err)
	}

	var siteSize int64
	var fileCount int
	_ = filepath.Walk(publishDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			siteSize += info.Size()
			fileCount++
		}
		return nil
	})

	checkExpiryBeforeUpdate(updateExpiryCheck{
		ObjectID:  proj.ObjectID,
		Network:   network,
		Epochs:    epochs,
		SiteSize:  siteSize,
		FileCount: fileCount,
	})

	d := sb.New()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	output, err := d.Update(ctx, publishDir, proj.ObjectID, deployer.DeployOptions{
		Epochs:    epochs,
		Verbose:   true,
		WalrusCfg: walgoCfg.WalrusConfig,
		Network:   network,
	})

	if err != nil {
		deployment := &projects.DeploymentRecord{
			ProjectID: proj.ID,
			ObjectID:  proj.ObjectID,
			Network:   network,
			Epochs:    epochs,
			Success:   false,
			Error:     err.Error(),
		}
		_ = pm.RecordDeployment(deployment)

		return fmt.Errorf("update failed: %w", err)
	}

	if !output.Success {
		return fmt.Errorf("update failed: operation unsuccessful")
	}

	// Try to get actual gas from blockchain, fallback to estimate
	var gasFee string
	gasInfo, err := walrus.GetLatestTransactionGas(proj.WalletAddr, network)
	if err == nil && gasInfo != nil {
		if gasInfo.TotalWAL > 0 && gasInfo.TotalGasSUI > 0 {
			gasFee = fmt.Sprintf("%.6f WAL + %.6f SUI", gasInfo.TotalWAL, gasInfo.TotalGasSUI)
		} else if gasInfo.TotalWAL > 0 {
			gasFee = fmt.Sprintf("%.6f WAL", gasInfo.TotalWAL)
		} else if gasInfo.TotalGasSUI > 0 {
			gasFee = fmt.Sprintf("%.6f SUI", gasInfo.TotalGasSUI)
		}
	}
	// Fallback to estimate if actual gas not available
	if gasFee == "" {
		gasFee = projects.EstimateGasFeeWithEpochs(network, siteSize, epochs)
	}

	proj.Epochs = epochs
	proj.LastDeployAt = time.Now()
	proj.GasFee = gasFee

	if err := pm.UpdateProject(proj); err != nil {
		return fmt.Errorf("failed to update project record: %w", err)
	}

	deployment := &projects.DeploymentRecord{
		ProjectID: proj.ID,
		ObjectID:  proj.ObjectID,
		Network:   network,
		Epochs:    epochs,
		GasFee:    gasFee,
		Success:   true,
	}
	if err := pm.RecordDeployment(deployment); err != nil {
		fmt.Fprintf(os.Stderr, "%s Warning: Failed to record deployment: %v\n", icons.Warning, err)
	}

	fmt.Println()
	fmt.Printf("%s Update successful!\n", icons.Check)
	fmt.Printf("%s Object ID: %s\n", icons.Info, proj.ObjectID)
	if gasFee != "" {
		fmt.Printf("%s Gas Fee: %s\n", icons.Info, gasFee)
	}
	fmt.Println()

	return nil
}
