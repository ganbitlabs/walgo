package ai

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestContentRelPath(t *testing.T) {
	good := map[string]string{
		"content/_index.md":    "_index.md",
		"content/blog/post.md": "blog/post.md",
		"content/a/../b.md":    "b.md",
	}
	for in, want := range good {
		got, err := contentRelPath(in)
		if err != nil || filepath.ToSlash(got) != want {
			t.Errorf("%q: got %q, %v; want %q", in, got, err, want)
		}
	}
	bad := []string{
		"content/../../.zshrc", "content/../x.md", "content/..", "/etc/passwd.md",
		"content/a/../../b.md", "other/x.md", "content/post.txt", "content/", "content",
	}
	for _, in := range bad {
		if got, err := contentRelPath(in); err == nil {
			t.Errorf("%q should be rejected, got %q", in, got)
		}
	}
}

func TestGeneratePageRejectsTraversal(t *testing.T) {
	root := t.TempDir()
	cfg := DefaultPipelineConfig()
	cfg.ContentDir = filepath.Join(root, "site", "content")
	g := &Generator{config: cfg}
	out := g.GeneratePage(context.Background(), &SitePlan{}, &PageSpec{ID: "x", Path: "content/../../.zshrc"})
	if out.Success || out.Error == nil {
		t.Fatalf("expected failure, got %+v", out)
	}
	if _, err := os.Stat(filepath.Join(root, ".zshrc")); err == nil {
		t.Fatal("file written outside site")
	}
}

func TestValidatePlanRejectsTraversal(t *testing.T) {
	plan := &SitePlan{ID: "p", Pages: []PageSpec{
		{ID: "home", Path: "content/_index.md", PageType: PageTypeHome},
		{ID: "evil", Path: "content/../../.zshrc"},
	}}
	if err := (&Planner{}).validatePlan(plan); err == nil || !strings.Contains(err.Error(), "content/") {
		t.Fatalf("expected traversal path to be rejected, got %v", err)
	}
}
