package main

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
)

func TestDeployRunsRejectsConcurrentDeployOfSameSite(t *testing.T) {
	var runs deployRuns
	site := t.TempDir()

	_, done, err := runs.start(site)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := runs.start(filepath.Join(site, ".")); err == nil || !strings.Contains(err.Error(), "already running") {
		t.Fatalf("second deploy of the same site: err = %v, want already running", err)
	}
	if _, otherDone, err := runs.start(t.TempDir()); err != nil {
		t.Fatalf("deploy of another site was rejected: %v", err)
	} else {
		otherDone()
	}

	done()
	_, done, err = runs.start(site)
	if err != nil {
		t.Fatalf("deploy after the first finished was rejected: %v", err)
	}
	done()
}

func TestCancelDeployCancelsOnlyThatSite(t *testing.T) {
	app := NewApp()
	site, other := t.TempDir(), t.TempDir()

	ctx, done, err := app.deploys.start(site)
	if err != nil {
		t.Fatal(err)
	}
	defer done()
	otherCtx, otherDone, err := app.deploys.start(other)
	if err != nil {
		t.Fatal(err)
	}
	defer otherDone()

	if !app.CancelDeploy(site) {
		t.Fatal("CancelDeploy reported no running deployment")
	}
	if ctx.Err() == nil {
		t.Error("deployment context was not cancelled")
	}
	if otherCtx.Err() != nil {
		t.Error("cancelling one site cancelled another")
	}
	if app.CancelDeploy(t.TempDir()) {
		t.Error("CancelDeploy reported a deployment for an idle site")
	}
}

func TestBeforeCloseAsksWhileDeploying(t *testing.T) {
	tests := []struct {
		name       string
		deploying  bool
		answer     bool
		wantPrompt bool
		wantKeep   bool
	}{
		{name: "idle", deploying: false, wantPrompt: false, wantKeep: false},
		{name: "deploying, user keeps it", deploying: true, answer: false, wantPrompt: true, wantKeep: true},
		{name: "deploying, user quits", deploying: true, answer: true, wantPrompt: true, wantKeep: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := NewApp()
			prompted := false
			app.confirm = func(_, _ string) bool {
				prompted = true
				return tt.answer
			}

			var ctxErr func() error
			if tt.deploying {
				ctx, done, err := app.deploys.start(t.TempDir())
				if err != nil {
					t.Fatal(err)
				}
				defer done()
				ctxErr = ctx.Err
			}

			keepOpen := app.beforeClose(context.Background())
			if prompted != tt.wantPrompt {
				t.Errorf("prompted = %v, want %v", prompted, tt.wantPrompt)
			}
			if keepOpen != tt.wantKeep {
				t.Errorf("beforeClose kept the app open = %v, want %v", keepOpen, tt.wantKeep)
			}
			if ctxErr != nil && (ctxErr() != nil) == tt.wantKeep {
				t.Errorf("deployment cancelled = %v, want %v", ctxErr() != nil, !tt.wantKeep)
			}
		})
	}
}
