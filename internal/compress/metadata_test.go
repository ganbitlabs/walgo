package compress

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUpdateMetadataCreatesMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ws-resources.json")

	if err := UpdateMetadata(path, MetadataOptions{SiteName: "blog", ObjectID: "0xabc"}); err != nil {
		t.Fatalf("UpdateMetadata: %v", err)
	}

	cfg, err := ReadWSResourcesConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.SiteName != "blog" || cfg.ObjectID != "0xabc" {
		t.Errorf("got site_name %q object_id %q", cfg.SiteName, cfg.ObjectID)
	}
	if cfg.Metadata == nil || cfg.Metadata.Creator != DefaultCreator {
		t.Errorf("default metadata not filled: %+v", cfg.Metadata)
	}
}

func TestUpdateMetadataRejectsCorruptFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ws-resources.json")
	if err := os.WriteFile(path, []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := UpdateMetadata(path, MetadataOptions{}); err == nil {
		t.Error("expected an error for a corrupt ws-resources.json")
	}
}
