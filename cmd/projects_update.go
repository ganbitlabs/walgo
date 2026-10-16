package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ganbitlabs/walgo/internal/config"
	"github.com/ganbitlabs/walgo/internal/deployment"
	"github.com/ganbitlabs/walgo/internal/hugo"
	"github.com/ganbitlabs/walgo/internal/projects"
	"github.com/ganbitlabs/walgo/internal/sui"
	"github.com/ganbitlabs/walgo/internal/ui"
	"github.com/ganbitlabs/walgo/internal/version"
)

// updateProjectByRef pushes local site changes to the project's existing site on Walrus,
// through the same deployment workflow as `walgo deploy`.
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

	if proj.ObjectID == "" {
		return fmt.Errorf("project %q has no deployed site to update; deploy it first with 'walgo launch'", proj.Name)
	}

	if epochs == 0 {
		epochs = proj.Epochs
	} else if epochs != proj.Epochs {
		fmt.Printf("  %s Epochs: %d → %d\n", icons.Info, proj.Epochs, epochs)
	}

	walgoCfg, err := config.LoadConfigFrom(proj.SitePath)
	if err != nil {
		if !filepath.IsAbs(proj.SitePath) {
			return fmt.Errorf("project %q stores the relative site path %q, which does not resolve to a walgo site from this directory: %w\n\n"+
				"  Run this command from the directory the project was created in, or re-import the site", proj.Name, proj.SitePath, err)
		}
		return fmt.Errorf("failed to load config from %s: %w", proj.SitePath, err)
	}

	// The deployment saves its result by absolute site path. A relative path from an older
	// walgo is recorded as absolute only now that it has resolved to the site's walgo.yaml.
	if !filepath.IsAbs(proj.SitePath) {
		abs, err := filepath.Abs(proj.SitePath)
		if err != nil {
			return fmt.Errorf("resolving site path %q: %w", proj.SitePath, err)
		}
		proj.SitePath = abs
		if err := pm.UpdateProject(proj); err != nil {
			return fmt.Errorf("failed to record the absolute site path: %w", err)
		}
	}

	// Same guards as `walgo update`: the project's recorded network must be the
	// one site-builder will actually use, and the toolchain must be able to talk
	// to Sui at all.
	network, err := sui.ResolveDeployNetwork(proj.Network)
	if err != nil {
		if proj.Network == "" {
			return err
		}
		return fmt.Errorf("project %q is recorded on %s: %w", proj.Name, proj.Network, err)
	}
	if err := version.RequireDeployToolchain(); err != nil {
		return err
	}

	fmt.Printf("%s Pushing changes to Walrus...\n", icons.Spinner)

	if err := hugo.BuildSite(proj.SitePath); err != nil {
		return fmt.Errorf("failed to build site: %w", err)
	}

	publishDir, err := config.ResolvePublishDir(proj.SitePath, walgoCfg)
	if err != nil {
		return err
	}
	if _, err := os.Stat(publishDir); err != nil {
		return fmt.Errorf("site not built - run 'walgo build' first")
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

	// The project's object is the one to update, whatever walgo.yaml or ws-resources.json say.
	walgoCfg.WalrusConfig.ProjectID = proj.ObjectID

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	res, err := deployment.PerformDeployment(ctx, deployment.DeploymentOptions{
		SitePath:    proj.SitePath,
		PublishDir:  publishDir,
		Epochs:      epochs,
		WalgoCfg:    walgoCfg,
		Verbose:     true,
		SaveProject: true,
		ProjectName: proj.Name,
		Category:    proj.Category,
		Network:     network,
		WalletAddr:  proj.WalletAddr,
		Description: proj.Description,
		ImageURL:    proj.ImageURL,
	})
	if err != nil {
		_ = pm.RecordDeployment(&projects.DeploymentRecord{
			ProjectID: proj.ID,
			ObjectID:  proj.ObjectID,
			Network:   network,
			Epochs:    epochs,
			Success:   false,
			Error:     err.Error(),
		})
		return fmt.Errorf("update failed: %w", err)
	}

	gasFee := res.GasFee(network, epochs)

	fmt.Println()
	fmt.Printf("%s Update successful!\n", icons.Check)
	fmt.Printf("%s Object ID: %s\n", icons.Info, res.ObjectID)
	if gasFee != "" {
		fmt.Printf("%s Gas Fee: %s\n", icons.Info, gasFee)
	}
	fmt.Println()

	return nil
}
