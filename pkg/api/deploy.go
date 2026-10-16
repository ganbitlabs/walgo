package api

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ganbitlabs/walgo/internal/config"
	"github.com/ganbitlabs/walgo/internal/deployment"
	"github.com/ganbitlabs/walgo/internal/hugo"
	"github.com/ganbitlabs/walgo/internal/projects"
	"github.com/ganbitlabs/walgo/internal/sui"
	"github.com/ganbitlabs/walgo/internal/version"
)

// ProjectSitePath returns the site directory of a saved project.
func ProjectSitePath(projectID int64) (string, error) {
	pm, err := projects.NewManager()
	if err != nil {
		return "", fmt.Errorf("failed to create project manager: %w", err)
	}
	defer pm.Close()
	proj, err := pm.GetProject(projectID)
	if err != nil || proj == nil {
		return "", fmt.Errorf("project not found")
	}
	return proj.SitePath, nil
}

// checkDeployTarget refuses a deploy that site-builder would send somewhere
// other than intended. site-builder signs with the active Sui environment and
// address, so the network must match the recorded one and, when the site was
// deployed before, the active address must be the wallet that owns it. It
// returns the network the deploy will run on.
func checkDeployTarget(network, wallet string) (string, error) {
	resolved, err := sui.ResolveDeployNetwork(network)
	if err != nil {
		if network != "" {
			err = fmt.Errorf("this site deploys to %s: %w", network, err)
		}
		return "", err
	}

	if wallet = strings.TrimSpace(wallet); wallet != "" {
		active, err := sui.GetActiveAddress()
		if err != nil {
			return "", fmt.Errorf("could not determine the active Sui address: %w", err)
		}
		if active = strings.TrimSpace(active); !strings.EqualFold(active, wallet) {
			return "", fmt.Errorf("wallet mismatch: this site belongs to %s but the active Sui address is %s\n\n"+
				"  Only the owning wallet can update the site.\n\n"+
				"  To fix:  sui client switch --address %s", wallet, active, wallet)
		}
	}

	if err := version.RequireDeployToolchain(); err != nil {
		return "", err
	}
	return resolved, nil
}

// siteDeploy is one desktop deploy of the site at SitePath, after
// checkDeployTarget has settled Network.
type siteDeploy struct {
	SitePath    string
	Network     string
	WalletAddr  string
	Epochs      int
	ProjectName string
	Category    string
	Description string
	ImageURL    string
}

type siteDeployResult struct {
	ObjectID string
	IsUpdate bool
	GasFee   string
}

// deploySite builds the site and runs the shared deployment workflow, the same
// one `walgo deploy` uses, saving the result to the projects database. A
// failure is recorded against the project when one exists.
func deploySite(ctx context.Context, d siteDeploy) (*siteDeployResult, error) {
	walgoCfg, err := config.LoadConfigFrom(d.SitePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	if err := hugo.BuildSite(d.SitePath); err != nil {
		return nil, buildFailure(err)
	}

	publishDir := filepath.Join(d.SitePath, walgoCfg.HugoConfig.PublishDir)
	if _, err := os.Stat(publishDir); err != nil {
		return nil, fmt.Errorf("publish directory not found: %s", publishDir)
	}

	if err := ctx.Err(); err != nil {
		return nil, deployContextError(err, false)
	}

	res, err := deployment.PerformDeployment(ctx, deployment.DeploymentOptions{
		SitePath:    d.SitePath,
		PublishDir:  publishDir,
		Epochs:      d.Epochs,
		WalgoCfg:    walgoCfg,
		Verbose:     true,
		SaveProject: true,
		ProjectName: d.ProjectName,
		Category:    d.Category,
		Network:     d.Network,
		WalletAddr:  d.WalletAddr,
		Description: d.Description,
		ImageURL:    d.ImageURL,
	})
	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			err = deployContextError(ctxErr, true)
		}
		recordFailedDeployment(d, err)
		return nil, err
	}

	return &siteDeployResult{
		ObjectID: res.ObjectID,
		IsUpdate: res.IsUpdate,
		GasFee:   gasFeeText(res, d.Network, d.Epochs),
	}, nil
}

// deployContextError explains a cancelled or timed-out deploy. Once
// site-builder has started, its transaction may already be on-chain.
func deployContextError(err error, started bool) error {
	msg := "deployment cancelled"
	if errors.Is(err, context.DeadlineExceeded) {
		msg = "deployment timed out"
	}
	if !started {
		return errors.New(msg + " before anything was uploaded")
	}
	return fmt.Errorf("%s while site-builder was running\n\n"+
		"  If its transaction was already submitted, the site may exist on-chain.\n"+
		"  Check your wallet before deploying again:  sui client objects", msg)
}

func recordFailedDeployment(d siteDeploy, deployErr error) {
	pm, err := projects.NewManager()
	if err != nil {
		return
	}
	defer pm.Close()
	proj, err := pm.GetProjectBySitePath(d.SitePath)
	if err != nil || proj == nil {
		return
	}
	_ = pm.RecordDeployment(&projects.DeploymentRecord{
		ProjectID: proj.ID,
		ObjectID:  proj.ObjectID,
		Network:   d.Network,
		Epochs:    d.Epochs,
		Success:   false,
		Error:     deployErr.Error(),
	})
}

// gasFeeText reports the cost read from the chain, or an estimate when the
// chain query found nothing.
func gasFeeText(res *deployment.DeploymentResult, network string, epochs int) string {
	switch {
	case res.ActualWAL > 0 && res.ActualGasSUI > 0:
		return fmt.Sprintf("%.6f WAL + %.6f SUI", res.ActualWAL, res.ActualGasSUI)
	case res.ActualWAL > 0:
		return fmt.Sprintf("%.6f WAL", res.ActualWAL)
	case res.ActualGasSUI > 0:
		return fmt.Sprintf("%.6f SUI", res.ActualGasSUI)
	default:
		return projects.EstimateGasFeeWithEpochs(network, res.SiteSize, epochs)
	}
}

// buildFailure explains a failed site build. The desktop app has no terminal,
// so hugo's own output goes into the message instead of "see output above".
func buildFailure(err error) error {
	var buildErr *hugo.BuildError
	if errors.As(err, &buildErr) && strings.TrimSpace(buildErr.Output) != "" {
		return fmt.Errorf("failed to build site: hugo build failed: %v\n\n%s", buildErr.Err, strings.TrimSpace(buildErr.Output))
	}
	return fmt.Errorf("failed to build site: %w", err)
}
