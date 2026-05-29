package api_fetcher

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/okarahan/repo-compliance-checker/internal/api_fetcher/client"
	"github.com/okarahan/repo-compliance-checker/internal/api_fetcher/endpoints"
	"github.com/okarahan/repo-compliance-checker/internal/model"
)

// ManifestDownload is the result of downloading manifest files for a repo.
type ManifestDownload struct {
	// Dir is the local directory where files were written (mirrors repo-relative paths).
	Dir string
	// Downloaded lists repo-relative paths that were successfully fetched and written.
	Downloaded []string
	// Missing lists repo-relative paths that did not exist at the given ref (404).
	Missing []string
}

// ResolveManifestPaths returns the de-duplicated list of repo-relative paths to fetch,
// based on the detected languages (Linguist) and the manifest map config.
//
// Language names from Linguist (e.g. "Go", "JavaScript") are matched case-insensitively
// against the manifest map keys (e.g. "go", "javascript"). Global files are always included.
func ResolveManifestPaths(langs endpoints.LanguagesResponse, mm model.ManifestMapConfig) []string {
	seen := map[string]struct{}{}
	var paths []string

	add := func(p string) {
		p = strings.TrimSpace(p)
		if p == "" {
			return
		}
		if _, ok := seen[p]; ok {
			return
		}
		seen[p] = struct{}{}
		paths = append(paths, p)
	}

	for lang := range langs {
		key := strings.ToLower(strings.TrimSpace(lang))
		if targets, ok := mm.ByLanguage[key]; ok {
			for _, f := range targets.Files {
				add(f)
			}
		}
	}

	for _, f := range mm.Global.Files {
		add(f)
	}

	return paths
}

// DownloadManifestFiles resolves the manifest paths for the repo's languages and downloads
// each file via the GitHub Contents API into a temp directory.
//
// Files that do not exist (404) are recorded in Missing and are not treated as errors.
// If destDir is empty, a new temp directory is created.
func DownloadManifestFiles(
	ctx context.Context,
	c *client.Client,
	owner, repo, ref string,
	langs endpoints.LanguagesResponse,
	mm model.ManifestMapConfig,
	destDir string,
) (ManifestDownload, error) {
	if c == nil {
		return ManifestDownload{}, fmt.Errorf("client is nil")
	}

	if strings.TrimSpace(destDir) == "" {
		d, err := os.MkdirTemp("", "rcc-manifest-*")
		if err != nil {
			return ManifestDownload{}, fmt.Errorf("create temp dir: %w", err)
		}
		destDir = d
	} else if err := os.MkdirAll(destDir, 0o755); err != nil {
		return ManifestDownload{}, fmt.Errorf("create dest dir: %w", err)
	}

	paths := ResolveManifestPaths(langs, mm)

	result := ManifestDownload{Dir: destDir}
	for _, p := range paths {
		content, err := endpoints.GetFileContent(ctx, c, owner, repo, ref, p)
		if err != nil {
			if errors.Is(err, endpoints.ErrFileNotFound) {
				result.Missing = append(result.Missing, p)
				continue
			}
			return result, fmt.Errorf("download %q: %w", p, err)
		}

		outPath := filepath.Join(destDir, filepath.FromSlash(p))
		if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
			return result, fmt.Errorf("create dir for %q: %w", p, err)
		}
		if err := os.WriteFile(outPath, content, 0o644); err != nil {
			return result, fmt.Errorf("write %q: %w", p, err)
		}
		result.Downloaded = append(result.Downloaded, p)
	}

	return result, nil
}
