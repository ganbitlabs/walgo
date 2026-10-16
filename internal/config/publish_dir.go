package config

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ResolvePublishDir returns the absolute publish directory for the site at sitePath.
// An absolute hugo.publishDir is used as-is; a relative one is joined onto sitePath.
// It refuses a directory that contains the site: hugo's --cleanDestinationDir would delete the sources,
// and a deploy would upload walgo.yaml and everything else in the project.
func ResolvePublishDir(sitePath string, cfg *WalgoConfig) (string, error) {
	dir := cfg.HugoConfig.PublishDir
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(sitePath, dir)
	}
	publishDir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	absSite, err := filepath.Abs(sitePath)
	if err != nil {
		return "", err
	}
	if rel, err := filepath.Rel(publishDir, absSite); err == nil && !strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("publish directory %s contains the site itself; set hugo.publishDir in walgo.yaml to a subdirectory such as public", publishDir)
	}
	return publishDir, nil
}
