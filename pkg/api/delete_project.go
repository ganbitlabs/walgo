package api

import (
	"context"
	"fmt"
	"strings"
	"time"

	sb "github.com/ganbitlabs/walgo/internal/deployer/sitebuilder"
	"github.com/ganbitlabs/walgo/internal/projects"
	"github.com/ganbitlabs/walgo/internal/sui"
)

var (
	activeSuiEnv     = sui.GetActiveEnv
	activeSuiAddress = sui.GetActiveAddress
	destroySite      = func(ctx context.Context, objectID string) error {
		return sb.New().Destroy(ctx, objectID)
	}
)

// DeleteProjectParams holds delete project parameters
type DeleteProjectParams struct {
	ProjectID int64 `json:"projectId"`
	// DeleteSiteFolder also removes the project's site folder from disk.
	DeleteSiteFolder bool `json:"deleteSiteFolder"`
	// SkipOnChain removes the project locally without destroying its on-chain
	// site. Callers set it only after the user confirms a failed destroy.
	SkipOnChain bool `json:"skipOnChain"`
}

// DeleteProjectResult holds delete result
type DeleteProjectResult struct {
	Success          bool   `json:"success"`
	Message          string `json:"message"`
	Error            string `json:"error"`
	OnChainDestroyed bool   `json:"onChainDestroyed"`
	// OnChainFailed means the on-chain destroy was refused or failed, so
	// nothing was deleted locally either.
	OnChainFailed     bool   `json:"onChainFailed"`
	SiteFolderDeleted bool   `json:"siteFolderDeleted"`
	EstimatedGasCost  string `json:"estimatedGasCost,omitempty"`
}

// DeleteProject destroys the project's on-chain site (when it has one), then
// removes the project record, and the site folder only when asked. A failed
// destroy stops before any local change.
func DeleteProject(params DeleteProjectParams) DeleteProjectResult {
	var result DeleteProjectResult

	pm, err := projects.NewManager()
	if err != nil {
		result.Error = fmt.Sprintf("failed to create project manager: %v", err)
		return result
	}
	defer pm.Close()

	proj, err := pm.GetProject(params.ProjectID)
	if err != nil {
		result.Error = fmt.Sprintf("failed to get project: %v", err)
		return result
	}

	if proj.ObjectID != "" && !params.SkipOnChain {
		result.EstimatedGasCost = projects.EstimateDestroyCost(proj.Network)
		if err := checkDestroyWallet(proj); err != nil {
			result.OnChainFailed = true
			result.Error = fmt.Sprintf("%v. Nothing was deleted.", err)
			return result
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()
		if err := destroySite(ctx, proj.ObjectID); err != nil {
			result.OnChainFailed = true
			result.Error = fmt.Sprintf("on-chain destroy failed: %v. Nothing was deleted.", err)
			return result
		}
		result.OnChainDestroyed = true
	}

	if err := pm.DeleteProjectWithOptions(params.ProjectID, params.DeleteSiteFolder); err != nil {
		if result.OnChainDestroyed {
			result.Error = fmt.Sprintf("the site was destroyed on-chain, but local cleanup failed: %v", err)
		} else {
			result.Error = fmt.Sprintf("failed to delete project: %v", err)
		}
		return result
	}

	result.Success = true
	result.SiteFolderDeleted = params.DeleteSiteFolder && proj.SitePath != ""
	result.Message = describeDeletion(proj, result)
	return result
}

// checkDestroyWallet refuses a destroy that would run against a different
// network or wallet than the one the site was deployed with.
func checkDestroyWallet(proj *projects.Project) error {
	if want := strings.ToLower(strings.TrimSpace(proj.Network)); want != "" {
		env, err := activeSuiEnv()
		if err != nil {
			return fmt.Errorf("cannot read the active Sui environment: %w", err)
		}
		if got := strings.ToLower(strings.TrimSpace(env)); got != want {
			return fmt.Errorf("the active Sui environment is %q but this site was deployed on %q; switch networks and try again", got, want)
		}
	}
	if want := strings.TrimSpace(proj.WalletAddr); want != "" {
		addr, err := activeSuiAddress()
		if err != nil {
			return fmt.Errorf("cannot read the active Sui address: %w", err)
		}
		if got := strings.TrimSpace(addr); !strings.EqualFold(got, want) {
			return fmt.Errorf("the active Sui address is %s but this site was deployed from %s; switch addresses and try again", got, want)
		}
	}
	return nil
}

func describeDeletion(proj *projects.Project, result DeleteProjectResult) string {
	var parts []string
	switch {
	case result.OnChainDestroyed:
		parts = append(parts, "Destroyed the site on-chain.")
	case proj.ObjectID != "":
		parts = append(parts, fmt.Sprintf("Skipped the on-chain destroy; site object %s still exists on %s.", proj.ObjectID, proj.Network))
	}
	parts = append(parts, "Removed the project record and its deployment history.")
	switch {
	case result.SiteFolderDeleted:
		parts = append(parts, fmt.Sprintf("Deleted the site folder %s.", proj.SitePath))
	case proj.SitePath != "":
		parts = append(parts, fmt.Sprintf("Kept the site folder %s.", proj.SitePath))
	}
	return strings.Join(parts, " ")
}

// ProjectSitePaths returns the site folder of every registered project.
func ProjectSitePaths() ([]string, error) {
	pm, err := projects.NewManager()
	if err != nil {
		return nil, fmt.Errorf("failed to create project manager: %w", err)
	}
	defer pm.Close()

	projs, err := pm.ListProjects("", "")
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}
	paths := make([]string, 0, len(projs))
	for _, p := range projs {
		if p.SitePath != "" {
			paths = append(paths, p.SitePath)
		}
	}
	return paths, nil
}
