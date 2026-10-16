package main

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/ganbitlabs/walgo/pkg/api"
)

// deployTimeout bounds one deployment, site build included.
const deployTimeout = 30 * time.Minute

// deployRuns tracks the deployment running for each site, keyed by its
// cleaned absolute path. A site has at most one deployment at a time.
type deployRuns struct {
	mu      sync.Mutex
	cancels map[string]context.CancelFunc
}

// start claims sitePath for a new deployment. done releases it and must be
// called when the deployment returns.
func (r *deployRuns) start(sitePath string) (ctx context.Context, done func(), err error) {
	key, err := deployKey(sitePath)
	if err != nil {
		return nil, nil, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if _, running := r.cancels[key]; running {
		return nil, nil, fmt.Errorf("a deployment of this site is already running")
	}
	if r.cancels == nil {
		r.cancels = map[string]context.CancelFunc{}
	}
	ctx, cancel := context.WithTimeout(context.Background(), deployTimeout)
	r.cancels[key] = cancel

	done = func() {
		cancel()
		r.mu.Lock()
		delete(r.cancels, key)
		r.mu.Unlock()
	}
	return ctx, done, nil
}

// cancel stops the deployment of sitePath and reports whether one was running.
func (r *deployRuns) cancel(sitePath string) bool {
	key, err := deployKey(sitePath)
	if err != nil {
		return false
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	cancel, running := r.cancels[key]
	if running {
		cancel()
	}
	return running
}

func (r *deployRuns) cancelAll() {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, cancel := range r.cancels {
		cancel()
	}
}

func (r *deployRuns) active() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.cancels)
}

func deployKey(sitePath string) (string, error) {
	if sitePath == "" {
		return "", fmt.Errorf("site path is required")
	}
	abs, err := filepath.Abs(sitePath)
	if err != nil {
		return "", fmt.Errorf("invalid site path: %w", err)
	}
	return filepath.Clean(abs), nil
}

// UpdateSiteParams holds update site parameters
type UpdateSiteParams = api.UpdateSiteParams

// UpdateSiteResult holds update site result
type UpdateSiteResult = api.UpdateSiteResult

// UpdateSite re-deploys an existing project's site. CancelDeploy with the
// project's site path stops it.
func (a *App) UpdateSite(params UpdateSiteParams) UpdateSiteResult {
	sitePath, err := api.ProjectSitePath(params.ProjectID)
	if err != nil {
		return UpdateSiteResult{Error: err.Error()}
	}
	ctx, done, err := a.deploys.start(sitePath)
	if err != nil {
		return UpdateSiteResult{Error: err.Error()}
	}
	defer done()
	return api.UpdateSite(ctx, params)
}

type LaunchWizardParams = api.LaunchWizardParams
type LaunchWizardResult = api.LaunchWizardResult

// LaunchWizard builds and publishes a site. CancelDeploy with the same site
// path stops it.
func (a *App) LaunchWizard(params LaunchWizardParams) LaunchWizardResult {
	ctx, done, err := a.deploys.start(params.SitePath)
	if err != nil {
		return LaunchWizardResult{Error: err.Error()}
	}
	defer done()
	return api.LaunchWizard(ctx, params)
}

// CancelDeploy stops the deployment running for sitePath and reports whether
// there was one.
func (a *App) CancelDeploy(sitePath string) bool {
	return a.deploys.cancel(sitePath)
}
